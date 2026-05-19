# RTSP-to-WebSocket H.264 流媒体系统设计文档

## 1. 系统概述

本系统实现了一个 RTSP 视频流到 WebSocket 的实时转发服务，将 IP 摄像头的 H.264 视频流通过浏览器 WebCodecs API 进行解码和渲染。

### 1.1 核心目标

- 从 RTSP 源（IP 摄像头）拉取 H.264 视频流（支持 4K 分辨率）
- 通过 WebSocket 实时推送到浏览器客户端
- 使用浏览器 WebCodecs API 硬件加速解码
- 保持端到端低延迟，长时间运行延迟不递增

### 1.2 技术栈

| 层级        | 技术                           | 作用                                         |
| ------      | ------                         | ------                                       |
| RTSP 客户端 | gortsplib/v5                   | 连接 RTSP 源、接收 RTP 包                    |
| RTP 解包    | gortsplib rtph264.Decoder      | H.264 RTP 解包（FU-A/STAP-A/丢包检测）       |
| RTP 底层    | pion/rtp, pion/rtcp            | RTP/RTCP 包定义                              |
| WebSocket   | gorilla/websocket              | 向浏览器推送视频帧                           |
| 前端解码    | WebCodecs VideoDecoder         | 浏览器端 H.264 硬件解码                      |
| 前端渲染    | Canvas 2D                      | 视频帧绘制                                   |

### 1.3 源文件

```
main.go       — Go 服务端（RTSP 客户端 + WebSocket 服务器）
index.html    — 浏览器前端（WebSocket 客户端 + WebCodecs 解码 + Canvas 渲染）
```

---

## 2. 整体架构

### 2.1 数据流

```
┌──────────────┐     RTSP/RTP      ┌──────────────────┐    WebSocket     ┌──────────────────┐
│  IP 摄像头   │ ───────────────►  │   Go 服务端      │ ──────────────►  │   浏览器客户端   │
│  (RTSP 源)   │   H.264 over RTP  │  (main.go)       │  二进制帧协议    │  (index.html)    │
└──────────────┘                   └──────────────────┘                  └──────────────────┘
```

### 2.2 服务端内部流程

```
gortsplib RTSP Client
        │
        │ OnPacketRTP 回调（按 H264 format 过滤）
        ▼
┌───────────────────────────┐
│  rtpDec.Decode(pkt)       │  gortsplib 内置 rtph264.Decoder
│                           │
│  内置处理:                │
│  - FU-A 分片重组          │
│  - STAP-A 聚合解析        │
│  - RTP 序列号校验(丢包)   │
│  - Marker bit 帧边界      │
└────────┬──────────────────┘
         │ 返回 [][]byte NALU 列表 或 error
         │
         │ error (非 ErrMorePacketsNeeded)
         │  → streamNeedKeyframe = true (层1 背压)
         │
         ▼
┌───────────────────────────┐
│  buildFrameFromNALUs()    │  持 mu 锁
│                           │
│  1. 扫描 NALU 检测 IDR    │
│  2. 缓存 SPS/PPS          │
│  3. streamNeedKeyframe    │
│     且无 IDR → 返回 nil   │
│  4. 拼接 Annex B 帧数据   │
└────────┬──────────────────┘
         │ *H264Frame 或 nil
         ▼
┌───────────────────────────┐
│  broadcastFrame()         │
│                           │
│  packBinaryFrame() (锁外) │
│  持 mu 锁: 遍历 clients   │
│  client.send() (非阻塞)   │
└────────┬──────────────────┘
         │
         ▼ 每个客户端独立
┌───────────────────────────┐
│  client.send()            │
│                           │
│  needKeyframe 且非IDR     │
│    → 丢弃 (层2 背压)      │
│  channel 满               │
│    → needKeyframe = true  │
│                           │
│  frameChan <- data        │
└────────┬──────────────────┘
         │
         ▼
┌───────────────────────────┐
│  writeLoop()              │  独立 goroutine
│                           │
│  从 frameChan 读          │
│  SetWriteDeadline(1s)     │
│  WriteMessage             │
│  超时则断开               │
└───────────────────────────┘
```

### 2.3 客户端内部流程

```
WebSocket onmessage
        │
        │ ArrayBuffer (二进制)
        ▼
┌─────────────────────┐
│  parseBinaryFrame() │  DataView 零拷贝解析
│                     │
│  提取: isKey,       │
│    timestamp,       │
│    h264Data         │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  NALU 类型判断      │
│                     │
│  type=7 → 保存 SPS  │
│  type=8 → 保存 PPS  │
│  其他 → 继续解码    │
└────────┬────────────┘
         │
         ▼
┌──────────────────────────┐
│  GOP-aware 背压 (层3)    │
│                          │
│  clientNeedKeyframe      │
│    非IDR → 丢弃          │
│    IDR → flush恢复       │
│                          │
│  decodeQueueSize > 3     │
│    → clientNeedKeyframe  │
│      = true              │
│    非IDR → 丢弃          │
│    IDR → flush + 解码    │
└────────┬─────────────────┘
         │
         ▼
┌─────────────────────┐
│  VideoDecoder.decode│
│                     │
│  output 回调:       │
│    drawImage→Canvas │
│    frame.close()    │
└─────────────────────┘
```

---

## 3. 核心数据结构

### 3.1 H264Frame（服务端）

```go
type H264Frame struct {
    Data      []byte  // H.264 Annex B 格式数据（含起始码 00 00 00 01）
    Timestamp uint64  // 相对时间戳（微秒），从第一帧开始计算
    IsKey     bool    // 是否为关键帧（IDR, NALU type=5）
}
```

关键帧的 Data 布局:
```
[00 00 00 01][SPS数据][00 00 00 01][PPS数据][00 00 00 01][IDR Slice数据]
```

非关键帧的 Data 布局:
```
[00 00 00 01][P-Slice数据]
```

### 3.2 H264Writer（服务端核心）

```go
type H264Writer struct {
    mu             sync.Mutex                  // 保护下面所有字段
    clients        map[string]*WebSocketClient // 活跃客户端
    pending        []byte                      // FU-A 分片累积缓冲区
    firstTimestamp uint32                      // 首帧 RTP 时间戳
    startTime      time.Time                   // 首帧墙钟时间
    sps            []byte                      // 缓存的 SPS（来自 SDP 或 RTP）
    pps            []byte                      // 缓存的 PPS（来自 SDP 或 RTP）
}
```

### 3.3 WebSocketClient（服务端）

```go
type WebSocketClient struct {
    conn      *websocket.Conn // WebSocket 连接
    writer    *H264Writer     // 所属 writer 引用
    clientID  string          // 唯一标识（如 "client-0"）
    frameChan chan []byte     // 帧发送缓冲通道（容量 3）
    closed    atomic.Bool    // 关闭标记（CAS 保证只关一次）
}
```

---

## 4. WebSocket 二进制传输协议

### 4.1 旧协议（已弃用）

```
JSON 文本消息:
{
  "data": "<Base64 编码的 H.264 数据>",
  "timestamp": <uint64 微秒>,
  "is_key": <bool>
}
```

**问题**: Base64 编码使数据膨胀约 33%；JSON 序列化和 Base64 编解码在 4K 分辨率下 CPU 开销显著。

### 4.2 新协议（当前实现）

```
二进制消息，总长度 = 9 + len(H264Data) 字节:

偏移    长度    字段              说明
────    ────    ────              ────
0       1       flags             bit0: isKey (1=关键帧, 0=非关键帧)
                                  bit1-7: 保留（置0）
1       8       timestamp         uint64 大端序，单位微秒
9       N       h264Data          H.264 Annex B 原始数据
```

**优势**:
- 零膨胀：直接传输二进制数据，无编码开销
- 零拷贝解析：前端使用 DataView 直接读取，h264Data 是 ArrayBuffer 的视图
- CPU 开销极低：Go 端 `binary.BigEndian.PutUint64` + `copy`；JS 端 `DataView.getUint32` x2

### 4.3 服务端打包

```go
func packBinaryFrame(frame *H264Frame) []byte {
    buf := make([]byte, 9+len(frame.Data))
    if frame.IsKey {
        buf[0] = 1
    }
    binary.BigEndian.PutUint64(buf[1:9], frame.Timestamp)
    copy(buf[9:], frame.Data)
    return buf
}
```

### 4.4 客户端解析

```javascript
function parseBinaryFrame(arrayBuffer) {
    const view = new DataView(arrayBuffer)
    const flags = view.getUint8(0)
    const isKey = (flags & 1) !== 0
    const timestampHi = view.getUint32(1, false)
    const timestampLo = view.getUint32(5, false)
    const timestamp = timestampHi * 0x100000000 + timestampLo
    const h264Data = new Uint8Array(arrayBuffer, 9)  // 零拷贝视图
    return { isKey, timestamp, h264Data }
}
```

---

## 5. RTP 处理与 FU-A 分片重组

### 5.1 RTP 负载中的 NALU 类型

```
RTP Payload 第一字节:
┌───┬───────┬───────────┐
│ F │  NRI  │   Type    │
│1位│ 2位   │   5位     │
└───┴───────┴───────────┘

Type 值:
  1     = 非IDR Slice (P帧)
  5     = IDR Slice (关键帧)
  7     = SPS (序列参数集)
  8     = PPS (图像参数集)
  28    = FU-A (分片单元)
```

### 5.2 FU-A 分片结构

```
FU-A Payload:
[FU-Indicator (1字节)] [FU-Header (1字节)] [分片数据...]

FU-Indicator = RTP payload[0]，其中 Type=28
FU-Header:
┌───┬───┬───┬───────────┐
│ S │ E │ R │   Type    │
│1位│1位│1位│   5位     │
└───┴───┴───┴───────────┘
S = 1: 第一个分片
E = 1: 最后一个分片
R: 保留位，必须为 0
Type: 原始 NALU 的类型
```

### 5.3 重组流程

```
processRTPPacket(pkt):

  naluType = payload[0] & 0x1F

  if naluType == 28 (FU-A):
      fuHeader = payload[1]
      isStart  = fuHeader & 0x80
      isEnd    = fuHeader & 0x40
      nalType  = fuHeader & 0x1F

      if isStart:
          pending = [NRI|nalType] + payload[2:]    // 重建 NALU 头 + 首段数据
          if isEnd:                                 // 单分片情况（S=1,E=1）
              → 完成（缓存 SPS/PPS 或 buildFrame）
      else if pending 非空:
          pending += payload[2:]                    // 追加中间/末尾段数据
          if isEnd:
              → 完成（缓存 SPS/PPS 或 buildFrame）

  if naluType == 7: 缓存 SPS
  if naluType == 8: 缓存 PPS
  if naluType 1-12: buildFrame(payload)
```

### 5.4 曾经存在的双重追加 Bug

**问题**: 原代码在 `isEnd` 分支中对 `payload[2:]` 追加了两次：
```go
// 第一次追加（L322，中间/末尾分片通用路径）
w.pending = append(w.pending, payload[2:]...)

if isEnd {
    // 第二次追加（L326，仅末尾分片路径）——这是 Bug！
    w.pending = append(w.pending, payload[2:]...)
```

**影响**: 每个 FU-A 分组的最后一段数据被重复拼入帧数据，导致 NALU 数据损坏、解码花屏或失败。4K 流的 FU-A 分片更多更频繁，问题更明显。

**修复**: 删除第二次追加，`isEnd` 分支仅做完成判断。

---

## 6. 并发模型与背压控制

### 6.1 服务端并发架构

```
                              ┌─────────────────┐
                              │  RTP 回调线程    │  gortsplib 内部线程
                              │                 │
                              │ processRTPPacket│
                              │ broadcastFrame  │
                              └────────┬────────┘
                                       │ client.send()（非阻塞）
                    ┌──────────────────┼──────────────────┐
                    ▼                  ▼                  ▼
            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
            │ frameChan[3] │  │ frameChan[3] │  │ frameChan[3] │
            │  client-0    │  │  client-1    │  │  client-2    │
            └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
                   ▼                 ▼                 ▼
            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
            │  writeLoop() │  │  writeLoop() │  │  writeLoop() │
            │  goroutine   │  │  goroutine   │  │  goroutine   │
            └──────────────┘  └──────────────┘  └──────────────┘
```

**设计要点**:

1. **每客户端一个 goroutine**: `writeLoop()` 独立运行，互不阻塞
2. **带缓冲 channel (容量=3)**: 在 RTP 回调和网络写入之间解耦
3. **非阻塞发送**: `send()` 使用 `select/default` 模式，永不阻塞 RTP 处理
4. **写超时**: `SetWriteDeadline(1s)`，慢客户端超时后断开而不是无限等待
5. **原子关闭标记**: `atomic.Bool` 避免重复关闭 channel 导致 panic

### 6.2 服务端丢帧策略

```go
func (c *WebSocketClient) send(data []byte) {
    select {
    case c.frameChan <- data:     // 正常发送
    default:
        // channel 满，丢弃队列中最旧的帧
        select {
        case <-c.frameChan:       // 弹出最旧帧
        default:
        }
        select {
        case c.frameChan <- data: // 放入最新帧
        default:
        }
    }
}
```

**策略**: 始终保留最新帧。当 channel 满时（消费端跟不上），丢弃队头（最旧的帧），插入队尾（最新的帧）。这确保了客户端收到的总是尽可能接近实时的帧。

**trade-off**: 丢弃的帧如果是某个 P 帧序列的依赖帧，可能导致后续 P 帧解码失败出现花屏，但这会在下一个关键帧到来时自动恢复。相比延迟无限递增，短暂花屏是更好的选择。

### 6.3 客户端背压策略

```javascript
if (currentDecoder.decodeQueueSize > MAX_DECODE_QUEUE_SIZE) {  // 阈值=3
    if (!isKey) {
        return  // 丢弃非关键帧
    }
    // 收到关键帧 → flush 队列 → 从此关键帧重新开始
    await currentDecoder.flush()
    baseTimestamp = null  // 重置时间基准
}
```

**分层控制**:
1. `decodeQueueSize <= 3`: 正常解码所有帧
2. `decodeQueueSize > 3` 且非关键帧: 丢弃（等待关键帧追赶）
3. `decodeQueueSize > 3` 且关键帧: flush 解码器 + 重置时间戳 + 从关键帧重新开始

**为什么阈值是 3**: 4K@30fps 的单帧解码时间约 10-30ms，3 帧的缓冲约 100ms，足以应对偶发的解码抖动，但不会让延迟持续积累。

### 6.4 锁粒度设计

```
processRTPPacket():  持 mu 锁  → 操作 pending/sps/pps（纯内存操作，快）
broadcastFrame():    持 mu 锁  → 遍历 clients map + 非阻塞 send（快）
writeLoop():         无锁       → 网络 I/O 在锁外（慢操作不影响其他路径）
```

**关键优化**: 序列化 (`packBinaryFrame`) 在加锁之前完成，锁内只做 map 遍历和 channel 非阻塞写入，持锁时间极短。

---

## 7. SPS/PPS 处理

### 7.1 获取来源

SPS/PPS 有两个获取来源（优先级从高到低）：

1. **SDP 协商阶段**: `c.Describe()` 返回的媒体描述中包含 SPS/PPS，在流开始前就可以获得
2. **RTP 流内**: 编码器会在关键帧前重新发送 SPS/PPS（作为独立 NALU 或 FU-A 分片）

两个来源都会更新 `H264Writer.sps` 和 `H264Writer.pps`。

### 7.2 发送时机

- **客户端连接时**: 立即发送缓存的 SPS 和 PPS（各一条二进制消息）
- **每个关键帧**: `buildFrame()` 在关键帧数据前拼接 SPS+PPS

### 7.3 前端处理

前端在 `onmessage` 中扫描整条 Annex B access unit：
- 任意位置的 type=7(SPS): 保存到 `savedSPS`，用于推导 codec 字符串（如 `avc1.640033`）
- 任意位置的 type=8(PPS): 保存到 `savedPPS`
- 只有包含 VCL NALU 的 access unit 才会进入解码路径

两者都收到后才配置 VideoDecoder。解码器使用 Annex B 模式（`optimizeForLatency: true`，无 `description` 属性），浏览器从码流中自动解析参数集。`configure()` 之后前端会等待下一个真实 IDR access unit，避免把仅含 SPS/PPS 的消息错误地作为 key chunk 送入解码器。

### 7.4 Codec 字符串推导

```javascript
// 从 SPS NALU 中提取 profile_idc、compatibility 和 level_idc
// SPS 结构: [起始码][NALU Header(0x67)][profile_idc][constraint_flags][level_idc]...
const profile_idc = sps[offset + 1]  // 如 0x64 = High Profile
const compatibility = sps[offset + 2]
const level_idc   = sps[offset + 3]  // 如 0x33 = Level 5.1 (4K)
codecString = `avc1.${profileHex}${compatHex}${levelHex}`  // 如 "avc1.640033"
```

---

## 8. 时间戳方案

### 8.1 服务端时间戳

```go
func (w *H264Writer) getTimestamp() uint64 {
    if w.firstTimestamp == 0 {
        w.startTime = time.Now()
    }
    elapsed := time.Since(w.startTime).Microseconds()
    return uint64(elapsed)  // 相对时间戳（微秒）
}
```

使用墙钟相对时间而非 RTP 时间戳，避免 RTP 时间戳回绕和不同流之间的时钟域问题。

### 8.2 客户端时间戳

```javascript
function getMicroseconds(ts) {
    // 自动适配单位：<10000 视为毫秒，>1e16 视为纳秒
    if (baseTimestamp === null) {
        baseTimestamp = ts
        return 0  // 首帧从 0 开始
    }
    const relativeTs = ts - baseTimestamp
    if (relativeTs < 0) {
        baseTimestamp = ts  // 时间戳回退时重置基准
        return 0
    }
    return relativeTs
}
```

**关键**: 背压 flush 时会重置 `baseTimestamp = null`，确保从关键帧重新开始时时间戳连续递增。

---

## 9. WebSocket 服务端

### 9.1 端点

- 地址: `:8080`
- 路径: `/ws`
- 协议: WebSocket Binary

### 9.2 连接生命周期

```
客户端 HTTP GET /ws
     │
     ▼
Upgrade 到 WebSocket
     │
     ▼
创建 WebSocketClient（分配 frameChan, 启动 writeLoop goroutine）
     │
     ▼
加入 H264Writer.clients 映射
     │
     ▼
发送缓存的 SPS（二进制帧消息）
     │
     ▼
发送缓存的 PPS（二进制帧消息）
     │
     ▼
进入读消息循环（保持连接活跃，处理 Ping/Pong）
     │
     ▼ （连接断开）
从 clients 映射中删除
     │
     ▼
client.close()（关闭 channel + 连接，writeLoop 自动退出）
```

### 9.3 曾经存在的 PPS 重复发送 Bug

**问题**: 原代码在发送 PPS 的 if 块内发送了一次，if 块外又发送了一次：

```go
if len(h264Writer.pps) > 0 {
    // ... 构建 configMsg ...
    conn.WriteMessage(websocket.BinaryMessage, configMsg)  // 第一次
}
conn.WriteMessage(websocket.BinaryMessage, configMsg)      // 第二次（Bug!）
```

**影响**: 客户端收到两条 PPS 消息，可能导致解码器状态混乱。

**修复**: 删除 if 块外的重复发送。

---

## 10. 前端 WebCodecs 解码器

### 10.1 配置

```javascript
const config = {
    codec: derivedCodec,        // 如 "avc1.640033"（从 SPS 推导）
    optimizeForLatency: true,   // 低延迟模式
    // 无 description 属性 → Annex B 自动模式
}
```

`optimizeForLatency: true` 告诉浏览器优先降低延迟而非优化吞吐量，这对实时视频流至关重要。

### 10.2 输出回调

```javascript
output: (frame) => {
    // 动态调整 Canvas 尺寸匹配视频分辨率
    if (canvas.width !== frame.displayWidth || canvas.height !== frame.displayHeight) {
        canvas.width = frame.displayWidth
        canvas.height = frame.displayHeight
    }
    ctx.drawImage(frame, 0, 0)
    frame.close()  // 必须关闭，否则内存泄漏
}
```

### 10.3 错误恢复

- VideoDecoder 内部错误 → 重置 `isDecoderConfigured`，并要求从下一个真实 IDR access unit 恢复
- Decoder 状态变为 `"closed"` → `ensureDecoderReady()` 创建新实例
- decode 调用异常 → 重置配置状态，等待重新配置
- 背压 `reset()` 之后，前端同样要求下一个 access unit 必须是 IDR，防止旧参考帧残留导致花屏/绿屏

---

## 11. 延迟递增问题根因分析与解决方案

### 11.1 原始问题

拉取 4K RTSP 流时，播放延迟随时间不断增加，从亚秒级逐渐增长到数秒甚至数十秒。

### 11.2 根因分析

| 层级 | 根因 | 严重度 | 说明 |
|------|------|--------|------|
| 服务端 | broadcastFrame 同步写 WebSocket | **致命** | 4K 帧大(IDR 帧可达数百 KB)，写操作可能耗时数十 ms。期间 RTP 回调被阻塞，新帧排队，形成正反馈的延迟递增 |
| 服务端 | 持锁时间包含网络 I/O | 高 | mu 锁在写完所有客户端后才释放，processRTPPacket 等待锁导致 RTP 包处理延迟 |
| 客户端 | 无 decodeQueueSize 检查 | **致命** | 所有帧无条件送入解码器，4K 解码速率低于帧到达速率时队列无限增长 |
| 服务端 | JSON+Base64 编码开销 | 中 | 4K IDR 帧约 200KB，Base64 后约 270KB + JSON 开销，增加 CPU 和带宽消耗 |
| 服务端 | FU-A 双重追加 Bug | 中 | 帧数据损坏导致解码失败/重试，间接加剧延迟 |
| 服务端 | PPS 重复发送 | 低 | 可能导致解码器状态异常 |

### 11.3 解决方案总结

| 问题 | 方案 | 效果 |
|------|------|------|
| 同步写阻塞 RTP | 每客户端 `frameChan` + 独立 `writeLoop` goroutine | RTP 回调不再被网络 I/O 阻塞 |
| 无背压控制（服务端） | `send()` 非阻塞写入，channel 满则丢旧帧 | 保证最新帧优先，延迟有界 |
| 无背压控制（客户端） | `decodeQueueSize > 3` 时丢弃非关键帧，关键帧到来时 flush | 解码延迟有界，关键帧自动恢复 |
| 持锁时间过长 | 序列化在锁外完成，锁内只做 map 遍历和非阻塞 channel 写入 | 锁持有时间从 ms 级降到 us 级 |
| JSON+Base64 开销 | 改用 9 字节头的二进制协议 | CPU 降低，带宽节省约 33% |
| FU-A 双重追加 | 删除重复 append | 帧数据正确 |
| PPS 重复发送 | 删除重复 WriteMessage | 解码器初始化正常 |

---

## 12. 4K 场景下的关键参数

| 参数 | 值 | 依据 |
|------|------|------|
| `frameChan` 容量 | 3 | 约 100ms 缓冲(30fps)，足够吸收网络抖动，但不会累积过多延迟 |
| `WriteDeadline` | 1 秒 | 4K IDR 帧约 200KB，1Gbps 网络写入约 2ms；1秒超时留足余量，超时说明客户端确实无法消费 |
| `MAX_DECODE_QUEUE_SIZE` | 3 | 4K@30fps 单帧解码约 10-30ms，3 帧缓冲约 100ms，平衡流畅度和延迟 |
| 起始码 | 4 字节 `[00 00 00 01]` | WebCodecs Annex B 模式要求；统一使用 4 字节避免歧义 |
| 时间戳单位 | 微秒 (uint64) | WebCodecs `EncodedVideoChunk.timestamp` 要求微秒；uint64 可表示约 58 万年 |

---

## 13. 已知限制与后续优化方向

### 13.1 当前限制

1. **RTSP URL 硬编码**: 需要修改代码才能切换视频源
2. **无认证**: WebSocket 和 RTSP 连接均无鉴权
3. **仅支持 H.264**: 不支持 H.265/HEVC（WebCodecs 支持有限）
4. **clientID 碰撞**: `fmt.Sprintf("client-%d", len(clients))` 在客户端断开重连后可能产生重复 ID
5. **无优雅关闭**: `shutdownCtx` 已定义但未实际使用，进程退出时客户端会异常断开
6. **单流**: 仅支持单路 RTSP 流，不支持多路复用

### 13.2 后续优化方向

1. **配置化**: RTSP URL、WebSocket 端口等通过命令行参数或配置文件指定
2. **客户端 ID**: 使用 UUID 或原子递增计数器避免碰撞
3. **优雅关闭**: 监听 SIGINT/SIGTERM，通知所有客户端后关闭
4. **多路流**: 支持多个 RTSP 源，每路流有独立的 H264Writer
5. **自适应质量**: 根据客户端网络状况动态调整推送策略（如只推关键帧）
6. **RTCP 反馈**: 利用 RTCP 包做 QoS 监控和丢包统计
7. **WebTransport**: 替代 WebSocket，利用 QUIC 的多流特性进一步降低队头阻塞

---

## 14. 构建与运行

### 14.1 依赖

```
go 1.24+
github.com/bluenviron/gortsplib/v5
github.com/gorilla/websocket
github.com/pion/rtp
github.com/pion/rtcp
```

### 14.2 启动

```bash
cd /home/zdz/Documents/Try/Go/rtsp
go run main.go
```

服务启动后:
- WebSocket 服务监听 `:8080`
- 自动连接配置的 RTSP 源

### 14.3 访问

浏览器打开 `index.html`（需支持 WebCodecs，Chrome 94+ / Edge 94+）。

### 14.4 验证

1. 画面正常渲染，无花屏
2. 控制台无持续的解码错误
3. 长时间运行（>5 分钟）延迟不递增
4. 关键帧到来时如有积压，控制台打印 flush 日志后迅速恢复
