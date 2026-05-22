# RTSP 到 WebCodecs 播放器

一个高性能的 Go 应用程序，使用 WebCodecs API 将 RTSP 视频流传输到 Web 浏览器，实现硬件加速的 H264 解码，延迟极低。

## 特性

- **低延迟流媒体**：通过 WebSocket 直接传输 H264 NALU
- **RTP 数据包处理**：完整的 RTP 数据包解析和 H264 NALU 提取
- **FU-A 分片处理**：正确重组分片的 NALU 数据包
- **多客户端支持**：同时广播到多个浏览器客户端
- **硬件加速解码**：WebCodecs API 实现 GPU 加速的 H264 解码
- **Canvas 渲染**：在 HTML5 Canvas 上直接渲染视频帧
- **自动编解码器配置**：SPS/PPS 参数提取和解码器自动配置
- **实时分辨率选择**：根据视频流参数动态调整 Canvas 大小

## 项目架构

### 高层数据流

```
RTSP 服务器 → RTSP 客户端 → RTP 解析器 → H264 NALU 处理 → WebSocket 广播 → 浏览器 WebCodecs → Canvas 渲染
```

### 详细组件架构

```
┌─────────────────┐
│   RTSP 源头     │  IP 摄像机、媒体服务器等
└────────┬────────┘
         │ RTSP 协议
┌────────▼────────┐
│   RTSP 客户端   │  gortsplib v5
│   - Connect     │  - URL 解析和连接管理
│   - Describe    │  - 流描述和媒体发现
│   - SetupAll    │  - 媒体轨道设置和参数协商
│   - Play        │  - 开始流传输
└────────┬────────┘
         │ RTP 数据包
┌────────▼────────┐
│  RTP 处理器     │  数据包处理流水线
│   - OnPacketRTP │  - 每个 RTP 数据包的回调
│   - OnPacketRTCP│  - RTCP 数据包处理（统计/反馈）
└────────┬────────┘
         │ 原始 RTP 载荷
┌────────▼────────┐
│  H264 解析器    │  NALU 提取和重组
│   - NALU 检测   │  - 单个 NALU vs FU-A 分片识别
│   - FU-A 重组    │  - 分片数据包重建
│   - SPS/PPS 缓存│  - 序列和图像参数存储
└────────┬────────┘
         │ H264 NALU 单元
┌────────▼────────┐
│  帧构建器       │  帧创建和时间戳
│   - buildFrame  │  - NALU 到帧的转换
│   - Timestamp   │  - 相对时间戳计算（90kHz → µs）
│   - 关键帧标识   │  - IDR 帧检测用于同步点
└────────┬────────┘
         │ JSON 编码的帧
┌────────▼────────┐
│  WebSocket      │  广播引擎
│  - broadcastFrame│  - 多客户端帧分发
│  - 客户端管理    │  - 连接生命周期管理
│  - SPS/PPS 同步  │  - 向新客户端传递参数
└────────┬────────┘
         │ WebSocket 二进制消息
┌────────▼────────┐
│  WebCodecs 播放器│  浏览器端处理
│   - SPS/PPS 解析 │  - 参数提取用于解码器配置
│   - 解码器初始化 │  - VideoDecoder 初始化
│   - 块解码       │  - EncodedVideoChunk 处理
│   - 帧输出       │  - VideoFrame 渲染到 canvas
└────────┬────────┘
         │ 视频帧
┌────────▼────────┐
│   HTML5 Canvas  │  最终显示
│   - 实时显示    │  - 硬件加速渲染
│   - 缩放        │  - Canvas 尺寸和宽高比
└─────────────────┘
```

## 关键技术点

### 1. RTP 到 H264 NALU 转换

**NALU 类型检测：**
- **单个 NALU (1-12)**：完整 H264 NALU 在一个 RTP 数据包中
- **FU-A 分片 (28)**：大型 NALU 分割到多个 RTP 数据包中
  - **起始位 (0x80)**：第一个分片包含 NALU 头部
  - **结束位 (0x40)**：最后一个分片完成 NALU
  - **重建**：使用正确的 NALU 头部组合分片

**Start Code 插入：**
- Annex-B 格式：`00 00 00 01` + NALU 载荷
- 确保与 WebCodecs 的浏览器兼容性

### 2. WebCodecs 集成

**解码器配置：**
```javascript
await decoder.configure({
  codec: 'avc1.42E01E',  // 从 SPS 参数派生
  codedWidth: 1920,      // 流分辨率
  codedHeight: 1080
});
```

**帧处理：**
```javascript
const chunk = new EncodedVideoChunk({
  type: 'key' | 'delta',  // 基于 NALU 类型（5 = 关键帧）
  timestamp: µs,          // 从 90kHz RTP 时间戳转换
  data: H264_NALU         // 带 start code 的原始 NALU
});
```

### 3. WebSocket 协议设计

**二进制消息格式：**
```json
{
  "data": [00, 00, 00, 01, 0x67, ...SPS...],  // 带 start code 的 H264 NALU
  "timestamp": 1234567890,                     // 微秒
  "is_key": true                              // IDR 帧标志
}
```

**初始化序列：**
1. 客户端连接 → WebSocket 握手
2. 服务器发送缓存的 SPS/PPS → 解码器初始化
3. 服务器开始流传输 → 正常视频帧

### 4. 性能优化

**服务端：**
- **无锁帧广播**：最小化互斥锁争用
- **二进制协议**：WebSocket 上的 JSON（高效传输）
- **SPS/PPS 缓存**：新客户端的参数存储
- **时间戳转换**：预计算的相对时间戳

**客户端：**
- **硬件解码**：WebCodecs GPU 加速
- **帧回收**：适当的 VideoFrame.close() 内存管理
- **块处理**：高效的 EncodedVideoChunk 创建
- **Canvas 优化**：直接帧到 canvas 渲染

### 5. 错误处理与容错

**连接管理：**
- WebSocket 断开时自动重连
- 客户端连接生命周期跟踪
- 死客户端清理

**解码器恢复：**
- SPS/PPS 接收检测
- 参数变化时的解码器重新配置
- 解码错误时丢弃帧

**流质量：**
- RTSP 连接超时处理
- RTP 数据包丢失容忍
- 分片重组验证

## 环境要求

- **Go 1.19+** - RTSP 客户端和 WebSocket 服务器
- **Chrome 94+ 或 Edge 94+** - WebCodecs 支持
- **RTSP 服务器** - H264 视频流源
- **网络** - 对服务器的 WebSocket 访问

## 安装

```bash
git clone <repository>
cd rtsp-webcodecs
go mod tidy
go build -o rtsp-service .
go build -tags ui -o rtsp-ui .
./build.sh all
```

## 使用方法

### 1. 配置 RTSP 源头

通过启动参数传入 RTSP 流地址：

```bash
./rtsp-service -url "rtsp://192.168.1.100:554/live"
```

常见格式：

- `rtsp://ip:554/`
- `rtsp://ip:554/stream`
- `rtsp://ip:554/live.sdp`

### 2. 启动服务器

```bash
./rtsp-service -url "rtsp://192.168.1.100:554/live"
```

服务器将启动：
- WebSocket 服务器在 `:8080`
- RTSP 客户端连接到配置的源
- 日志输出显示连接状态

也可以直接使用脚本构建：

```bash
./build.sh service
./build.sh ui
./build.sh all
```

### 3. 打开 Web 播放器

内置 UI 仅在 UI 构建中可用，并且需要显式开启 `-ui`：

```bash
./rtsp-ui -url "rtsp://192.168.1.100:554/live" -ui
./build.sh run-ui -- -url "rtsp://192.168.1.100:554/live" -port 8080
```

然后访问：

- `http://127.0.0.1:8080/`

如果使用独立 `index.html`，页面默认按当前来源自动连接，也可以通过查询参数覆盖：

- `file:///path/to/index.html?host=127.0.0.1&port=8080`
- `http://localhost:3000/index.html?host=127.0.0.1&port=8080`

### 4. 配置并连接

1. **选择分辨率**：选择匹配的视频分辨率
2. **点击连接**：建立 WebSocket 连接
3. **监控**：观察控制台日志中的 SPS/PPS 接收情况
4. **查看**：视频在 canvas 上显示

## 故障排除

### 连接问题

**RTSP 连接超时：**
- 验证 RTSP URL 正确
- 检查到 RTSP 服务器的网络连接
- 确认 RTSP 服务器接受连接
- 使用 VLC Media Player 测试：`vlc rtsp://ip:554/stream`

**WebSocket 连接失败：**
- 确保端口 8080 未被防火墙阻止
- 检查浏览器控制台是否有 WebSocket 错误
- 验证服务器正在运行并监听

### 视频无法播放

**"Decoder not configured yet" 错误：**
- 等待浏览器控制台中的 SPS/PPS 日志
- 验证 RTSP 流包含有效的 H264 参数
- 检查分辨率选择是否匹配流

**解码错误：**
- 确认 H264 配置文件兼容性（Baseline/Main/High）
- 验证 NALU 结构带有 start code
- 检查浏览器 WebCodecs 支持：`VideoDecoder in window`

**性能问题：**
- 如果 CPU/GPU 受限，降低分辨率
- 高分辨率流检查网络带宽
- 监控浏览器控制台是否有帧丢失

### 网络优化

**减少延迟：**
- 使用有线连接而不是 WiFi
- 将服务器放置在 RTSP 源附近
- 最小化网络跳数
- 考虑 UDP WebSocket 替代方案

**提高质量：**
- 确保足够带宽（>2 倍流比特率）
- 使用适当的路由器 QoS 设置
- 检查网络拥塞

## 技术实现细节

### H264 NALU 类型参考

| NALU 类型 | 描述 | 用途 |
|-----------|-------------|---------|
| 1         | 非 IDR Slice | P/B 帧 |
| 5         | IDR Slice | 关键帧（I 帧）|
| 6         | SEI | 补充增强信息 |
| 7         | SPS | 序列参数集 |
| 8         | PPS | 图像参数集 |
| 28        | FU-A | 分片单元 A |

### RTP 时间戳处理

- **时钟频率**：90 kHz（视频常用）
- **计算**：`(current_timestamp - first_timestamp) / 90000 = 秒`
- **WebCodecs**：转换为微秒 `* 1000000`

### 浏览器兼容性

**WebCodecs 支持：**
- Chrome 94+（稳定版）
- Edge 94+（稳定版）
- Firefox（实验性 - 需要标志）
- Safari（有限支持）

**功能检测：**
```javascript
if ('VideoDecoder' in window) {
    // 支持 WebCodecs
}
```

## API 参考

### WebSocket 协议

**连接：**
- URL：`ws://server:8080/ws`
- 协议：二进制（JSON 编码帧）

**消息格式：**
```json
{
  "data": "base64 编码的 H264 NALU",
  "timestamp": 1234567890,
  "is_key": true
}
```

**消息类型：**
- **SPS/PPS**：配置消息（连接时发送）
- **关键帧**：IDR 帧（类型：5）
- **增量帧**：P/B 帧（类型：1）

## 性能指标

**预期延迟：**
- RTSP 到 WebSocket：~10-50ms
- WebSocket 传输：~1-10ms（本地）
- WebCodecs 解码：~5-20ms
- Canvas 渲染：~1-5ms
- **总计**：~17-85ms（典型：30-50ms）

**带宽需求：**
- 640x480@30fps：~1-2 Mbps
- 1280x720@30fps：~2-5 Mbps
- 1920x1080@30fps：~4-8 Mbps

**CPU/GPU 使用：**
- 服务器：低（每个流 5-15% CPU）
- 浏览器：中等（10-30% CPU，取决于硬件）

## 许可证

MIT License - 详见 LICENSE 文件
