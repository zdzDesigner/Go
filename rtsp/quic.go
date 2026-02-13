package main

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/webtransport-go"
	"github.com/quic-go/quic-go/http3"
)

// generateSelfSignedCert generates a self-signed certificate for development
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"RTSP Demo"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}, nil
}

// WebTransportWriter 管理WebTransport连接的写入和分发
// =============================================================================
// 核心数据结构，负责:
//
// 1. 维护WebTransport会话列表
// 2. 缓存SPS/PPS参数集
// 3. 广播视频帧到所有连接的会话
//
// 线程安全: 使用sync.Mutex保护所有成员变量
// =============================================================================
type WebTransportWriter struct {
	mu       sync.Mutex                      // 互斥锁，保护共享数据
	sessions map[string]*WebTransportSession // WebTransport会话映射
	sps      []byte                          // 序列参数集 (Sequence Parameter Set)
	pps      []byte                          // 图像参数集 (Picture Parameter Set)
}

// WebTransportSession 表示一个连接的WebTransport会话
// =============================================================================
// 会话结构包含连接和引用信息
// 实际管理通过WebTransportWriter.sessions映射进行
// =============================================================================
type WebTransportSession struct {
	stream    *webtransport.Stream  // WebTransport双向流
	session   *webtransport.Session // WebTransport会话引用
	sessionID string                // 会话标识符
}

// NewWebTransportWriter 创建新的WebTransportWriter实例
// =============================================================================
// 初始化会话映射
// 其他字段使用零值，后续会填充SPS/PPS
// =============================================================================
func NewWebTransportWriter() *WebTransportWriter {
	return &WebTransportWriter{
		sessions: make(map[string]*WebTransportSession),
	}
}

// broadcastFrame 广播视频帧到所有连接的会话
// =============================================================================
// 工作流程:
// 1. 获取互斥锁，确保线程安全
// 2. 遍历所有会话
// 3. 将H264Frame序列化为JSON
// 4. 通过WebTransport流发送数据
// 5. 处理发送失败，关闭会话并移除
//
// 注意: 序列化的JSON包含Base64编码的数据，因为[]byte默认会Base64编码
// =============================================================================
func (w *WebTransportWriter) broadcastFrame(frame *H264Frame) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, session := range w.sessions {
		if session.stream != nil {
			data, err := json.Marshal(frame)
			if err != nil {
				log.Printf("Error marshaling frame to JSON for session %s: %v", session.sessionID, err)
				continue
			}

			if len(frame.Data) > 4 {
				// naluType := frame.Data[4] & 0x1F
				// log.Printf("Sending frame via WebTransport: NALU type=%d, isKey=%v, dataLen=%d", naluType, frame.IsKey, len(frame.Data))
			}

			_, err = session.stream.Write(data)
			if err != nil {
				log.Printf("Error sending frame to session %s: %v", session.sessionID, err)
				session.stream.Close()
				delete(w.sessions, session.sessionID)
			}
		}
	}
}

// startWebTransportServer 启动WebTransport服务器
// =============================================================================
// 在端口4433监听WebTransport连接
//
// 功能:
// 1. HTTP处理器: /wt 端点用于WebTransport升级
// 2. 会话管理: 记录连接/断开，维护会话列表
// 3. 配置发送: 会话连接时发送SPS/PPS
// 4. 消息处理: 处理双向数据流
//
// 会话连接流程:
// 1. 客户端发起WebTransport连接到 /wt
// 2. 服务器接受连接
// 3. 分配sessionID，加入会话列表
// 4. 发送SPS/PPS配置数据
// 5. 会话断开时清理
//
// 注意:
// - SPS/PPS在会话连接时发送一次
// - 后续的视频帧通过broadcastFrame持续发送
// =============================================================================
func startWebTransportServer(h264Writer *H264Writer, addr string) {
	// 创建WebTransport服务器实例
	server := &webtransport.Server{
		H3: http3.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{}, // 需要实际证书，开发时可自动生成
				NextProtos:   []string{"h3"},
			},
		},
		CheckOrigin: func(r *http.Request) bool {
			// 在生产环境中应该有更安全的检查
			return true
		},
	}

	// WebTransport端点
	http.HandleFunc("/wt", func(w http.ResponseWriter, r *http.Request) {
		// 接受WebTransport连接
		session, err := server.Upgrade(w, r)
		if err != nil {
			log.Printf("Failed to upgrade to WebTransport: %v", err)
			return
		}

		// 生成会话ID
		sessionID := fmt.Sprintf("wt-%d", len(getWebTransportWriterInstance().sessions))

		// 创建会话实例
		webTransportSession := &WebTransportSession{
			session:   session,
			sessionID: sessionID,
		}

		// 获取WebTransportWriter实例并添加会话
		wtWriter := getWebTransportWriterInstance()
		wtWriter.mu.Lock()
		wtWriter.sessions[sessionID] = webTransportSession
		wtWriter.mu.Unlock()

		log.Printf("WebTransport session connected: %s", sessionID)

		// 创建双向流
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Failed to accept stream from session %s: %v", sessionID, err)
			return
		}

		// 将流分配给会话
		webTransportSession.stream = stream

		// =================================================================
		// 发送SPS/PPS配置
		// =================================================================
		// 重要: 必须在任何视频帧之前发送SPS/PPS
		// 解码器需要这些参数才能正确解码视频
		// =================================================================
		wtWriter.mu.Lock()
		currentTime := uint64(time.Now().UnixNano() / 1000)

		// 发送SPS (Sequence Parameter Set)
		// =================================================================
		// SPS包含视频的基本编码参数
		// 需要在PPS之前发送
		// =================================================================
		if len(wtWriter.sps) > 0 {
			// 格式: [00 00 00 01] + SPS数据
			spsData := append([]byte{0x00, 0x00, 0x00, 0x01}, wtWriter.sps...)
			log.Printf("Preparing SPS for session %s: total length=%d, first 10 bytes=%v", sessionID, len(spsData), spsData[:minInt(10, len(spsData))])

			// 构建配置消息
			configMsg, err := json.Marshal(H264Frame{
				Data:      spsData,
				Timestamp: currentTime,
				IsKey:     true, // 标记为关键帧
			})
			if err != nil {
				log.Printf("Error marshaling SPS to JSON for session %s: %v", sessionID, err)
			} else {
				log.Printf("SPS JSON length for session %s: %d", sessionID, len(configMsg))

				// 发送
				_, err := stream.Write(configMsg)
				if err != nil {
					log.Printf("Error sending SPS to session %s: %v", sessionID, err)
				} else {
					log.Printf("Sent SPS to session %s, data length: %d", sessionID, len(spsData))
				}
			}
		}

		// 发送PPS (Picture Parameter Set)
		// =================================================================
		// PPS包含图像的具体编码参数
		// 需要在SPS之后发送
		// =================================================================
		if len(wtWriter.pps) > 0 {
			ppsData := append([]byte{0x00, 0x00, 0x00, 0x01}, wtWriter.pps...)
			log.Printf("Preparing PPS for session %s: total length=%d, first 10 bytes=%v", sessionID, len(ppsData), ppsData[:minInt(10, len(ppsData))])

			configMsg, err := json.Marshal(H264Frame{
				Data:      ppsData,
				Timestamp: currentTime,
				IsKey:     true,
			})
			if err != nil {
				log.Printf("Error marshaling PPS to JSON for session %s: %v", sessionID, err)
			} else {
				log.Printf("PPS JSON length for session %s: %d", sessionID, len(configMsg))

				_, err := stream.Write(configMsg)
				if err != nil {
					log.Printf("Error sending PPS to session %s: %v", sessionID, err)
				} else {
					log.Printf("Sent PPS to session %s, data length: %d", sessionID, len(ppsData))
				}
			}
		}
		wtWriter.mu.Unlock()

		// =================================================================
		// 保持流活跃
		// =================================================================
		// 读取会话流数据(用于保持连接活跃)
		// 会话断开时退出循环
		// =================================================================
		buf := make([]byte, 1024)
		for {
			_, err := stream.Read(buf)
			if err != nil {
				break
			}
		}

		// 清理断开连接的会话
		wtWriter.mu.Lock()
		delete(wtWriter.sessions, sessionID)
		wtWriter.mu.Unlock()

		stream.Close()
		log.Printf("WebTransport session disconnected: %s", sessionID)
	})

	// 启动HTTP/3服务器
	go func() {
		log.Printf("WebTransport server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebTransport server error: %v", err)
		}
	}()

	// 存储引用以便其他地方访问
	setWebTransportWriterInstance(getWebTransportWriterInstance())
}

// 全局变量保存WebTransportWriter实例
var wtWriterInstance *WebTransportWriter
var wtWriterOnce sync.Once

func getWebTransportWriterInstance() *WebTransportWriter {
	wtWriterOnce.Do(func() {
		wtWriterInstance = NewWebTransportWriter()
	})
	return wtWriterInstance
}

func setWebTransportWriterInstance(writer *WebTransportWriter) {
	wtWriterInstance = writer
}

// copySPSToWebTransport 将H264Writer中的SPS/PPS复制到WebTransportWriter
// =============================================================================
// 由于RTSP客户端从SDP获取SPS/PPS，需要同步到WebTransportWriter
// =============================================================================
func copySPSToWebTransport(h264Writer *H264Writer) {
	wtWriter := getWebTransportWriterInstance()
	wtWriter.mu.Lock()
	defer wtWriter.mu.Unlock()

	if len(h264Writer.sps) > 0 {
		wtWriter.sps = make([]byte, len(h264Writer.sps))
		copy(wtWriter.sps, h264Writer.sps)
	}

	if len(h264Writer.pps) > 0 {
		wtWriter.pps = make([]byte, len(h264Writer.pps))
		copy(wtWriter.pps, h264Writer.pps)
	}
}
