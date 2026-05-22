// =============================================================================
// RTSP客户端转WebSocket/H264流媒体服务器
// =============================================================================
//
// 功能概述:
// 1. 连接到RTSP视频流 (如 rtsp://localhost:8554/live)
// 2. 接收RTP数据包并解析H.264视频帧
// 3. 通过WebSocket向Web前端传输H.264数据
//
// 数据流程:
//
//	[RTSP流] --RTP--> [gortsplib] --回调--> [processRTPPacket]
//	    |                                            |
//	    | 解析NALU类型                               提取SPS/PPS
//	    | 处理分片(Fu-A)                             |
//	    v                                            v
//	[H264Frame] --JSON序列化--> [WebSocket] --Base64--> [Web前端]
//
// 关键概念:
//
// RTP (Real-time Transport Protocol):
//   - 用于传输实时音视频数据的网络协议
//   - 每个RTP包包含一个H.264 NALU或其分片
//
// H.264 NALU (Network Abstraction Layer Unit):
//   - H.264视频流的基本传输单元
//   - 类型: SPS(7), PPS(8), IDR(5), P帧(1)等
//   - 可能通过Fu-A分片传输大数据NALU
//
// Fu-A (Fragmentation Unit Type A):
//   - 当NALU大小超过MTU时使用分片
//   - 每个分片包含: Fu-Indicator + Fu-Header + 片段数据
//   - isStart=true: 第一个分片
//   - isEnd=true: 最后一个分片
//
// WebSocket消息格式:
//
//	{"data": "Base64编码的视频数据", "timestamp": 微秒时间戳, "is_key": 是否关键帧}
//
// =============================================================================
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

// [输入]: RTSP 源、HTTP/WebSocket 请求、可选内置 UI 资源
// [输出]: WebSocket 视频流、统计接口、可选调试页面
// [定位]: RTSP 到 WebSocket 的独立服务进程入口
// [同步]: README.md、ui_disabled.go、ui_embedded.go、index.html

// 后端诊断统计。RtpLastTimeNs / RtpMaxGapNs 只在 RTP 单一回调 goroutine 内读写，
// 非 atomic；其余计数器跨 goroutine (HTTP handler vs RTP 回调) 访问用 atomic。
type ServerStats struct {
	RtpRecv     atomic.Int64
	FrameBuilt  atomic.Int64
	RtpErrors   atomic.Int64
	KeyWaiting  atomic.Int64
	RtpMaxGapNs atomic.Int64 // 由 RTP 回调写、HTTP handler Swap 读
}

var (
	stats         ServerStats
	rtpLastTimeNs int64 // RTP 回调私有状态，用于计算间隔
)

type StatsSnapshot struct {
	RtpRecv     int64 `json:"rtp_recv"`
	FrameBuilt  int64 `json:"frame_built"`
	RtpErrors   int64 `json:"rtp_errors"`
	KeyWaiting  int64 `json:"key_waiting"`
	RtpMaxGapMs int64 `json:"rtp_max_gap_ms"`
}

func getStatsSnapshot() StatsSnapshot {
	return StatsSnapshot{
		RtpRecv:     stats.RtpRecv.Swap(0),
		FrameBuilt:  stats.FrameBuilt.Swap(0),
		RtpErrors:   stats.RtpErrors.Swap(0),
		KeyWaiting:  stats.KeyWaiting.Swap(0),
		RtpMaxGapMs: stats.RtpMaxGapNs.Swap(0) / 1e6,
	}
}

// =============================================================================
// 数据结构定义
// =============================================================================

// H264Frame 定义WebSocket传输的视频帧结构
// =============================================================================
// 传输到Web前端的JSON消息格式
//
// 字段说明:
// - Data: H.264 NALU数据，包含起始码 [00 00 00 01]
// - Timestamp: 时间戳(微秒)，用于视频同步
// - IsKey: 是否为关键帧(IDR帧)
//
// 消息示例:
//
//	{
//	  "data": "AAAAAUGa7knhDyZTAl/68374odTt1V9UuOJBMcL1trrwXZdtyom//dQdis1H98i7r7ewt/l6pp3B073bcUYJ4GkjLro...",
//	  "timestamp": 1234567890,
//	  "is_key": true
//	}
//
// =============================================================================
type H264Frame struct {
	Data      []byte `json:"data"`
	Timestamp uint64 `json:"timestamp"`
	IsKey     bool   `json:"is_key"`
}

// packBinaryFrame 将H264Frame打包为二进制格式
// 格式: [1byte flags][8bytes timestamp big-endian][H264 data...]
// flags: bit0 = isKey
func packBinaryFrame(frame *H264Frame) []byte {
	buf := make([]byte, 9+len(frame.Data))
	if frame.IsKey {
		buf[0] = 1
	}
	binary.BigEndian.PutUint64(buf[1:9], frame.Timestamp)
	copy(buf[9:], frame.Data)
	return buf
}

// H264Writer 管理H.264视频流的写入和分发
// =============================================================================
// 核心数据结构，负责:
//
// 1. 维护WebSocket客户端列表
// 2. 缓存SPS/PPS参数集
// 3. 处理RTP分片重组
// 4. 广播视频帧到所有客户端
//
// 线程安全: 使用sync.Mutex保护所有成员变量
// =============================================================================
type H264Writer struct {
	mu                 sync.Mutex                  // 互斥锁，保护共享数据
	clients            map[string]*WebSocketClient // WebSocket客户端映射
	startTime          time.Time                   // 起始时间(首帧时记录，用于计算相对时间戳)
	sps                []byte                      // 序列参数集 (Sequence Parameter Set)
	pps                []byte                      // 图像参数集 (Picture Parameter Set)
	streamNeedKeyframe atomic.Bool                 // 流级别丢包标记：RTP层丢包后丢弃所有P帧，等待下一个IDR
}

// WebSocketClient 表示一个连接的WebSocket客户端
// =============================================================================
// 客户端结构简单，仅包含连接和引用
// 实际管理通过H264Writer.clients映射进行
// =============================================================================
type WebSocketClient struct {
	conn         *websocket.Conn // WebSocket连接
	writer       *H264Writer     // 所属H264Writer引用
	clientID     string          // 客户端标识符 (如 "client-0")
	frameChan    chan []byte     // 异步发送通道，避免阻塞RTP处理
	closed       atomic.Bool     // 标记是否已关闭
	needKeyframe atomic.Bool     // 丢帧后标记，等待下一个关键帧恢复
}

// writeLoop 独立的写goroutine，从channel读取帧并发送
// 设置写超时，避免慢客户端阻塞
func (c *WebSocketClient) writeLoop() {
	for data := range c.frameChan {
		c.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			log.Printf("Error sending frame to client %s: %v", c.clientID, err)
			c.close()
			return
		}
	}
}

// close 关闭客户端连接和channel
func (c *WebSocketClient) close() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.frameChan)
		c.conn.Close()
	}
}

// drainChan 清空channel中所有积压的帧
func (c *WebSocketClient) drainChan() {
	for {
		select {
		case <-c.frameChan:
		default:
			return
		}
	}
}

// send 非阻塞发送帧数据到channel
// 核心策略：一旦发生丢帧，跳过所有后续P帧，直到下一个关键帧到来时
// 清空channel重新开始，避免花屏
func (c *WebSocketClient) send(data []byte) {
	if c.closed.Load() {
		return
	}

	// 二进制协议: data[0] bit0 = isKey
	isKey := len(data) > 0 && (data[0]&1) != 0

	// 如果之前丢过帧，必须等关键帧才能恢复
	if c.needKeyframe.Load() {
		if !isKey {
			return // 丢弃P帧，等待关键帧
		}
		// 关键帧到了，清空积压的旧帧，从关键帧重新开始
		c.drainChan()
		c.needKeyframe.Store(false)
		log.Printf("Client %s: 收到关键帧，从丢帧状态恢复", c.clientID)
	}

	select {
	case c.frameChan <- data:
	default:
		// channel满，标记需要等待关键帧
		c.needKeyframe.Store(true)
		if isKey {
			// 当前就是关键帧，清空channel直接发送
			c.drainChan()
			c.needKeyframe.Store(false)
			select {
			case c.frameChan <- data:
			default:
			}
		}
		// 非关键帧直接丢弃，后续P帧也会被丢弃直到下一个关键帧
	}
}

// =============================================================================
// 工具函数
// =============================================================================

// minInt 返回两个整数中的较小值
// =============================================================================
// 用途: 用于日志输出时限制打印的字节数量
// =============================================================================
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// WebSocket配置
// =============================================================================

// upgrader HTTP升级为WebSocket的配置
// =============================================================================
// CheckOrigin: 允许所有来源的跨域请求
// 在生产环境中应该限制为特定的域名
// =============================================================================
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// =============================================================================
// H264Writer方法实现
// =============================================================================

// NewH264Writer 创建新的H264Writer实例
// =============================================================================
// 初始化客户端映射
// 其他字段使用零值，后续会从RTSP SDP中填充SPS/PPS
// =============================================================================
func NewH264Writer() *H264Writer {
	return &H264Writer{
		clients: make(map[string]*WebSocketClient),
	}
}

// broadcastFrame 广播视频帧到所有连接的客户端
// =============================================================================
// 工作流程:
// 1. 获取互斥锁，确保线程安全
// 2. 遍历所有客户端
// 3. 将H264Frame序列化为JSON
// 4. 通过WebSocket发送二进制消息
// 5. 处理发送失败，关闭连接并移除客户端
//
// 消息类型: websocket.BinaryMessage (二进制消息)
//
// 注意: 序列化的JSON包含Base64编码的数据，因为[]byte默认会Base64编码
// =============================================================================
func (w *H264Writer) broadcastFrame(frame *H264Frame) {
	// 先序列化，避免持锁时做序列化
	data := packBinaryFrame(frame)

	w.mu.Lock()
	defer w.mu.Unlock()

	for id, client := range w.clients {
		if client.closed.Load() {
			delete(w.clients, id)
			continue
		}
		if len(frame.Data) > 4 {
			// naluType := frame.Data[4] & 0x1F
			// log.Printf("Sending frame: NALU type=%d, isKey=%v, dataLen=%d", naluType, frame.IsKey, len(frame.Data))
		}
		// 非阻塞发送到客户端channel，不会阻塞RTP处理
		client.send(data)
	}
}

// buildFrameFromNALUs 从 gortsplib 解包器返回的 NALU 列表构建 H264Frame
// =============================================================================
// 输入: nalus [][]byte - 一个 access unit 中的所有 NALU（由 rtph264.Decoder 解包）
//
// 输出: *H264Frame（含 Annex B 起始码，可直接广播）或 nil（仅含 SPS/PPS 时）
//
// 处理逻辑:
// 1. 遍历所有 NALU
// 2. SPS(type=7)/PPS(type=8) → 缓存到 writer，不加入帧数据
// 3. IDR(type=5) → 标记为关键帧，在帧数据前拼接 SPS+PPS
// 4. 其他 Slice(type=1-12) → 拼接到帧数据
// 5. 每个 NALU 前添加 4 字节 Annex B 起始码 [00 00 00 01]
// =============================================================================
func (w *H264Writer) buildFrameFromNALUs(nalus [][]byte) *H264Frame {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 先扫描一遍判断是否包含关键帧（IDR）和更新 SPS/PPS
	hasIDR := false
	for _, nalu := range nalus {
		if len(nalu) == 0 {
			continue
		}
		nalType := nalu[0] & 0x1F
		switch nalType {
		case 5: // IDR
			hasIDR = true
		case 7: // SPS
			w.sps = make([]byte, len(nalu))
			copy(w.sps, nalu)
			log.Printf("RTP decoder: 更新 SPS, len=%d", len(nalu))
		case 8: // PPS
			w.pps = make([]byte, len(nalu))
			copy(w.pps, nalu)
			log.Printf("RTP decoder: 更新 PPS, len=%d", len(nalu))
		}
	}

	// 流级别丢包恢复：如果之前丢过包，只有关键帧才能恢复
	if w.streamNeedKeyframe.Load() {
		if !hasIDR {
			stats.KeyWaiting.Add(1)
			return nil // 丢弃所有P帧，等关键帧
		}
		w.streamNeedKeyframe.Store(false)
		log.Printf("收到关键帧，流从丢包状态恢复")
	}

	startCode := []byte{0x00, 0x00, 0x00, 0x01}
	var frameData []byte
	isKey := hasIDR

	// 关键帧前先拼 SPS+PPS
	if isKey && len(w.sps) > 0 && len(w.pps) > 0 {
		frameData = append(frameData, startCode...)
		frameData = append(frameData, w.sps...)
		frameData = append(frameData, startCode...)
		frameData = append(frameData, w.pps...)
	}

	// 拼接所有 slice NALU（跳过 SPS/PPS，已经在上面拼过）
	for _, nalu := range nalus {
		if len(nalu) == 0 {
			continue
		}
		nalType := nalu[0] & 0x1F
		if nalType >= 1 && nalType <= 12 && nalType != 7 && nalType != 8 {
			frameData = append(frameData, startCode...)
			frameData = append(frameData, nalu...)
		}
	}

	if frameData == nil {
		return nil // 只有 SPS/PPS，不构建帧
	}

	return &H264Frame{
		Data:      frameData,
		Timestamp: w.getTimestamp(),
		IsKey:     isKey,
	}
}

// getTimestamp 获取相对时间戳
// =============================================================================
// 返回从连接开始经过的微秒数
//
// 实现逻辑:
// 1. 第一次调用时记录起始时间
// 2. 后续调用返回 (当前时间 - 起始时间) 的微秒数
//
// 注意: 返回值会随时间递增，用于视频帧同步
// =============================================================================
func (w *H264Writer) getTimestamp() uint64 {
	if w.startTime.IsZero() {
		w.startTime = time.Now()
	}
	return uint64(time.Since(w.startTime).Microseconds())
}

// =============================================================================
// WebSocket服务器
// =============================================================================

// startWebSocketServer 启动WebSocket服务器
// =============================================================================
// 在端口8080监听WebSocket连接
//
// 功能:
// 1. HTTP处理器: /ws 端点用于WebSocket升级
// 2. 客户端管理: 记录连接/断开，维护客户端列表
// 3. 配置发送: 客户端连接时发送SPS/PPS
// 4. 消息处理: 读取客户端消息(用于保持连接活跃)
//
// 客户端连接流程:
// 1. 客户端发起HTTP请求到 /ws
// 2. 服务器升级为WebSocket
// 3. 分配clientID，加入客户端列表
// 4. 发送SPS/PPS配置数据
// 5. 等待接收视频帧 (通过broadcastFrame)
// 6. 客户端断开时清理
//
// 注意:
// - SPS/PPS在客户端连接时发送一次
// - 后续的视频帧通过broadcastFrame持续发送
// =============================================================================
func startWebSocketServer(h264Writer *H264Writer, port int, enableUI bool) {
	// 创建HTTP多路复用器
	httpMux := http.NewServeMux()

	// 诊断统计端点
	httpMux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(getStatsSnapshot())
	})

	// WebSocket端点
	httpMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 升级HTTP到WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		// 生成客户端ID
		clientID := fmt.Sprintf("client-%d", len(h264Writer.clients))
		client := &WebSocketClient{
			conn:      conn,
			writer:    h264Writer,
			clientID:  clientID,
			frameChan: make(chan []byte, 60), // 缓冲60帧(~2秒@30fps)，减少溢出触发丢帧
		}

		// 启动独立的写goroutine
		go client.writeLoop()

		// =================================================================
		// 发送SPS/PPS配置
		// =================================================================
		// 重要: 必须在任何视频帧之前发送SPS/PPS
		// 解码器需要这些参数才能正确解码视频
		// =================================================================
		h264Writer.mu.Lock()
		h264Writer.clients[clientID] = client
		currentTime := uint64(time.Now().UnixNano() / 1000)
		var initFrames []*H264Frame

		// 发送SPS (Sequence Parameter Set)
		// =================================================================
		// SPS包含视频的基本编码参数
		// 需要在PPS之前发送
		// =================================================================
		if len(h264Writer.sps) > 0 {
			// 格式: [00 00 00 01] + SPS数据
			spsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.sps...)
			log.Printf("Preparing SPS for client %s: total length=%d, first 10 bytes=%v", clientID, len(spsData), spsData[:minInt(10, len(spsData))])
			initFrames = append(initFrames, &H264Frame{
				Data:      spsData,
				Timestamp: currentTime,
				IsKey:     false,
			})
		}

		// 发送PPS (Picture Parameter Set)
		// =================================================================
		// PPS包含图像的具体编码参数
		// 需要在SPS之后发送
		// =================================================================
		if len(h264Writer.pps) > 0 {
			ppsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.pps...)
			log.Printf("Preparing PPS for client %s: total length=%d, first 10 bytes=%v", clientID, len(ppsData), ppsData[:minInt(10, len(ppsData))])
			initFrames = append(initFrames, &H264Frame{
				Data:      ppsData,
				Timestamp: currentTime,
				IsKey:     false,
			})
		}
		h264Writer.mu.Unlock()

		for _, frame := range initFrames {
			client.send(packBinaryFrame(frame))
		}

		log.Printf("Client connected: %s", clientID)

		// =================================================================
		// 消息循环: 保持连接活跃
		// =================================================================
		// 读取客户端消息(ping等)
		// 连接断开时退出循环
		// =================================================================
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// 清理断开连接的客户端
		h264Writer.mu.Lock()
		delete(h264Writer.clients, clientID)
		h264Writer.mu.Unlock()

		client.close()
		log.Printf("Client disconnected: %s", clientID)
	})

	registerUIRoutes(httpMux, enableUI)

	addr := fmt.Sprintf(":%d", port)
	go func() {
		log.Printf("WebSocket server listening on %s", addr)
		if err := http.ListenAndServe(addr, httpMux); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()
}

// =============================================================================
// 辅助函数 (未使用，保留备用)
// =============================================================================

// parseFramesFromRTP 从RTP负载解析H.264帧
// =============================================================================
// 注意: 此函数未使用
//
// 原始实现假设RTP使用字节长度前缀格式
// 但实际RTSP流使用分片(Fu-A)格式
//
// 字节长度格式:
// [4字节长度] [Nalu数据] [4字节长度] [Nalu数据] ...
// =============================================================================
func parseFramesFromRTP(pkt *rtp.Packet) [][]byte {
	payload := pkt.Payload
	var frames [][]byte
	offset := 0

	for offset < len(payload) {
		if offset+4 > len(payload) {
			break
		}

		// 读取4字节长度前缀
		naluLen := int(payload[offset])<<24 | int(payload[offset+1])<<16 |
			int(payload[offset+2])<<8 | int(payload[offset+3])

		offset += 4

		if offset+naluLen > len(payload) {
			break
		}

		// 构建带起始码的NALU
		nalu := make([]byte, naluLen+4)
		copy(nalu, []byte{0x00, 0x00, 0x00, 0x01})
		copy(nalu[4:], payload[offset:offset+naluLen])

		frames = append(frames, nalu)
		offset += naluLen
	}

	return frames
}

// findNALUStartCode 在数据中查找NALU起始码
// =============================================================================
// 注意: 此函数未使用
//
// 查找:
// - 4字节起始码: [00 00 00 01]
// - 3字节起始码: [00 00 01]
//
// 返回起始码的位置索引
// =============================================================================
func findNALUStartCode(data []byte) int {
	for i := 0; i <= len(data)-4; i++ {
		if bytes.Equal(data[i:i+4], []byte{0x00, 0x00, 0x00, 0x01}) {
			return i
		}
		if i <= len(data)-5 && bytes.Equal(data[i:i+3], []byte{0x00, 0x00, 0x01}) {
			return i
		}
	}
	return -1
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// rtspURL := flag.String("url", "rtsp://172.16.50.66:8554/live/video", "RTSP 流地址")
	// rtspURL := flag.String("url", "rtsp://172.16.50.122:554/ch2", "RTSP 流地址")
	// rtspURL := flag.String("url", "rtsp://172.16.50.130:554/ch2", "RTSP 流地址")
	// rtspURL := flag.String("url", "rtsp://172.16.50.40", "RTSP 流地址")
	// rtspURL := flag.String("url", "rtsp://172.16.50.126:554/ch2", "RTSP 流地址")
	rtspURL := flag.String("url", "rtsp://172.16.50.99/ch1", "RTSP 流地址")

	// rtspURL := flag.String("url", "rtsp://169.254.11.32:554/ch2", "RTSP 流地址")
	// rtspURL := flag.String("url", "rtsp://169.254.11.31:554/ch2", "RTSP 流地址")
	enableUI := flag.Bool("ui", false, "启用内置调试页面（仅 UI 构建可用）")

	port := flag.Int("port", 8080, "WebSocket/HTTP 服务端口")
	flag.Parse()

	h264Writer := NewH264Writer()
	startWebSocketServer(h264Writer, *port, *enableUI)

	u, err := base.ParseURL(*rtspURL)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", *rtspURL, err)
		panic(err)
	}

	log.Printf("Parsed URL - Scheme: %s, Host: %s, Path: %s", u.Scheme, u.Host, u.Path)
	fmt.Println("RTSP_SERVER_READY")

	// 创建RTSP客户端
	c := gortsplib.Client{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	// 启动RTSP客户端连接
	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// 获取媒体描述 (SDP)
	// =============================================================================
	// Describe返回会话描述协议(SDP)信息
	// 包含媒体类型、编码格式、SPS/PPS等参数
	// =============================================================================
	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	// =================================================================
	// 从SDP中查找H264 format和对应的media
	// =================================================================
	var h264Format *format.H264
	var h264Media *description.Media

	log.Printf("Loading SPS/PPS from SDP:")
	for _, media := range desc.Medias {
		for _, f := range media.Formats {
			if h264, ok := f.(*format.H264); ok {
				h264Format = h264
				h264Media = media
				log.Printf("  Found H264 format, SPS len=%d, PPS len=%d", len(h264.SPS), len(h264.PPS))
				if len(h264.SPS) > 0 {
					h264Writer.sps = h264.SPS
					log.Printf("  Loaded SPS from SDP: %v", h264Writer.sps)
				}
				if len(h264.PPS) > 0 {
					h264Writer.pps = h264.PPS
					log.Printf("  Loaded PPS from SDP: %v", h264Writer.pps)
				}
			}
		}
	}

	if h264Format == nil {
		panic("RTSP流中未找到H264格式")
	}

	// 创建 gortsplib 内置的 RTP/H264 解包器
	// =================================================================
	// 相比手写的 FU-A 重组，内置解包器提供:
	// - FU-A 分片重组 + RTP 序列号校验（丢包检测）
	// - STAP-A 聚合包解析（4K编码器常用）
	// - Marker bit 帧边界检测
	// - 自动处理 PacketizationMode
	// =================================================================
	rtpDec, err := h264Format.CreateDecoder()
	if err != nil {
		panic(err)
	}
	log.Printf("RTP H264 decoder created successfully")

	// 设置传输会话
	// ● c.SetupAll(desc.BaseURL, desc.Medias) 的作用是向 RTSP 服务器发送 SETUP 请求，为 SDP 中描述的所有媒体流建立传输会话。
	// 具体来说:
	// 1. 协议层面：对 desc.Medias 里的每一个媒体（视频、音频等）发送一个 RTSP SETUP 命令，协商传输参数（RTP/RTCP 端口、传输模式 UDP/TCP
	// 等）。
	// 2. desc.BaseURL 的作用：作为拼接每个 media control URL 的基址。SDP 里每个 media 有一个 a=control: 属性（可能是相对路径也可能是绝对
	// URL），gortsplib 用 BaseURL + control 拼出每个 media 的 SETUP 目标 URL。
	// 3. 为什么必须调用：
	//   - 没有 SETUP，服务器不会为客户端分配 RTP 通道，后续的 c.Play() 会失败。
	//   - SETUP 完成后，gortsplib 内部才知道该在哪些通道上接收 RTP 包，OnPacketRTP(h264Media, ...) 注册的回调才能真正收到包。
	// 4. 与 Setup 的区别：SetupAll 是便捷方法，一次性为所有 media 调用 Setup。如果你只想订阅视频轨，可以只对 h264Media 调
	// Setup，能省掉音频流的带宽。
	// 在 main.go 的流程里顺序是：Describe（拿 SDP） → SetupAll（建立传输） → OnPacketRTP（注册回调） → Play（开始推流）。
	err = c.SetupAll(desc.BaseURL, desc.Medias)
	if err != nil {
		panic(err)
	}

	// =================================================================
	// 注册RTP包回调（使用 OnPacketRTP 替代 OnPacketRTPAny）
	// =================================================================
	// OnPacketRTP 自动过滤出指定 format 的 RTP 包
	// rtpDec.Decode() 返回完整的 NALU 列表（一个 access unit）
	// =================================================================
	c.OnPacketRTP(h264Media, h264Format, func(pkt *rtp.Packet) {
		stats.RtpRecv.Add(1)
		// 单写者场景：rtpLastTimeNs 由本回调独占，非 atomic
		nowNs := time.Now().UnixNano()
		if rtpLastTimeNs > 0 {
			if gap := nowNs - rtpLastTimeNs; gap > stats.RtpMaxGapNs.Load() {
				stats.RtpMaxGapNs.Store(gap)
			}
		}
		rtpLastTimeNs = nowNs

		nalus, err := rtpDec.Decode(pkt)
		if err != nil {
			// ErrMorePacketsNeeded/ErrNonStartingPacketAndNoPrevious 是 FU-A 正常状态
			if err != rtph264.ErrMorePacketsNeeded && err != rtph264.ErrNonStartingPacketAndNoPrevious {
				stats.RtpErrors.Add(1)
				// 丢包后 P 帧依赖的参考帧缺失，必须等下一个 IDR 否则会产生残影
				if !h264Writer.streamNeedKeyframe.Load() {
					log.Printf("RTP层丢包/错误，等待下一个关键帧: %v", err)
					h264Writer.streamNeedKeyframe.Store(true)
					// 主动发 PLI 请求关键帧，避免等整个 GOP
					c.WritePacketRTCP(h264Media, &rtcp.PictureLossIndication{
						MediaSSRC: pkt.SSRC,
					})
				}
			}
			return
		}

		frame := h264Writer.buildFrameFromNALUs(nalus)
		if frame != nil {
			stats.FrameBuilt.Add(1)
			h264Writer.broadcastFrame(frame)
		}
	})

	// RTCP回调 (用于QoS监控)
	c.OnPacketRTCPAny(func(medi *description.Media, pkt rtcp.Packet) {
		log.Printf("RTCP packet from media %v, type %T\n", medi, pkt)
	})

	// 开始播放RTSP流
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	// 等待直到连接断开
	panic(c.Wait())
}

// =============================================================================
// 关闭信号处理 (未实现)
// =============================================================================

var shutdownCtx, shutdownCancel = context.WithCancel(context.Background())

func init() {
	go func() {
		<-shutdownCtx.Done()
	}()
}
