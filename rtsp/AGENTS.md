# AGENTS.md - RTSP to WebCodecs Streaming Project

## Project Overview

This project is a Go-based RTSP-to-WebSocket streaming server that broadcasts H.264 video frames to web browsers. The frontend uses WebCodecs API for hardware-accelerated video decoding.

**Key Technologies:**
- **Backend**: Go 1.24+, gortsplib/v5, gorilla/websocket, pion/rtp/rtcp
- **Frontend**: HTML5, vanilla JavaScript, WebCodecs API, Canvas rendering
- **Protocols**: RTSP, RTP, WebSocket, H.264 Annex-B

---

## Build Commands

### Go Backend

```bash
# Build the Go server
go build -o rtsp_demo .

# Build with race detector (for debugging concurrency issues)
go build -race -o rtsp_demo .

# Run the server
go run main.go

# Download dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph
```

### Web Frontend

The frontend is a single HTML file (`index.html`). No build step required.

```bash
# Serve locally (requires Node.js http-server or similar)
npx http-server -p 8081

# Or use Python
python3 -m http.server 8081
```

### Development Servers

```bash
# Start RTSP server and frontend (from start.sh)
./start.sh
```

---

## Testing

### Go Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test function
go test -v -run TestFunctionName

# Run tests with coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=.
```

### Frontend Testing

No automated tests exist for the frontend. Manual testing via browser console is required.

```bash
# Open browser DevTools and check:
# 1. Network tab for WebSocket messages
# 2. Console for SPS/PPS logs
# 3. Application tab for WebCodecs state
```

---

## Code Style Guidelines

### Go (main.go)

#### Imports
```go
// Standard library first, then third-party
import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/bluenviron/gortsplib/v5"
    "github.com/gorilla/websocket"
    "github.com/pion/rtcp"
    "github.com/pion/rtp"
)
```

#### Formatting
- Use `gofmt` (default Go formatter)
- Line length: no hard limit, use best judgment
- Group related code with blank lines
- Comment exported functions and types

#### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, short | `main` |
| Variables | camelCase | `frameData`, `naluType` |
| Constants | UPPER_SNAKE_CASE or camelCase | `maxPayloadSize` |
| Functions | PascalCase exported, camelCase unexported | `processRTPPacket`, `broadcastFrame` |
| Types | PascalCase | `H264Frame`, `H264Writer` |
| Interfaces | PascalCase with "er" suffix | `Reader`, `Writer` |

#### Error Handling
```go
// Good: Check errors immediately
if err != nil {
    return nil, fmt.Errorf("failed to parse frame: %w", err)
}

// Good: Log and continue for non-critical errors
if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
    log.Printf("Error sending frame: %v", err)
    // Don't return - frame broadcast is best-effort
}

// Avoid: Ignoring errors with _
data, _ := json.Marshal(frame)  // ❌
```

#### Concurrency
- Use `sync.Mutex` for shared state protection
- Lock only the necessary section (avoid `defer` for short operations)
- Document goroutine lifecycles

```go
func (w *H264Writer) processRTPPacket(pkt *rtp.Packet) *H264Frame {
    w.mu.Lock()
    defer w.mu.Unlock()  // Or lock/unlock for short operations
    // ... processing code
}
```

#### Comments
- Use Chinese comments (matching existing codebase)
- Comment exported functions with `// FunctionName does X`
- Document complex logic inline

```go
// processRTPPacket 处理RTP数据包，提取H.264帧
// Input: RTP包 (来自gortsplib的回调)
// Output: H264Frame (用于广播) 或 nil
```

---

### JavaScript (index.html)

#### General Style
- Use `const` by default, `let` when reassignment needed
- Use template literals for string interpolation
- Use arrow functions for callbacks
- No semicolons required (matching existing code)

#### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Variables | camelCase | `frameData`, `naluType` |
| Constants | UPPER_SNAKE_CASE or const camelCase | `MAX_FRAME_SIZE` |
| Classes | PascalCase | `H264WebSocketPlayer` |
| Methods | camelCase | `processMessage`, `configureDecoder` |

#### Error Handling
```javascript
// Good: Try-catch for async operations
try {
    await this.decoder.decode(chunk);
    console.log('Decoded successfully');
} catch (error) {
    console.error('Decode error:', error);
}

// Good: Null coalescing for optional values
const timestamp = (frame.timestamp || 0) * 1000;
```

#### Comments
- Use Chinese comments (matching existing codebase)
- Document function purpose at the start

```javascript
// processMessage - 处理WebSocket接收的视频帧数据
// 数据流程:
// 1. WebSocket接收JSON格式数据
// 2. 将Base64字符串解码为Uint8Array
```

---

## Project Structure

```
rtsp/
├── main.go          # Go backend (RTSP client, WebSocket server)
├── index.html       # Frontend (WebCodecs player)
├── index2.html      # Alternative frontend
├── go.mod           # Go module dependencies
├── go.sum           # Go module checksums
├── start.sh         # Development startup script
├── server.log       # Server logs
└── README.md        # Project documentation
```

---

## Common Development Tasks

### Adding a New RTSP Source
Edit `main.go` line ~265:
```go
rtspURL := "rtsp://your-camera-ip:554/stream"
```

### Adding Frontend Logging
Add to `ws.onmessage` handler:
```javascript
console.log('=== WebSocket.onmessage START ===');
// ... logging code
```

### Modifying WebSocket Protocol
The message format is:
```json
{
  "data": "base64-encoded-H264-NALU",
  "timestamp": 1234567890,
  "is_key": true
}
```

### Testing Changes
1. Rebuild Go server: `go build -o rtsp_demo .`
2. Restart servers: `./start.sh`
3. Open browser console (F12)
4. Check for SPS/PPS logs
5. Verify video playback

---

## Debugging Tips

### Go Backend
```bash
# Verbose logging
go run main.go 2>&1 | grep -E "SPS|PPS|Sent|Error"

# Check WebSocket client connections
grep "Client connected" server.log
```

### Frontend
- Open DevTools Console (F12)
- Filter logs by "processMessage" or "configureDecoder"
- Check Network tab for WebSocket frames
- Monitor Performance tab for decoding metrics

---

## Notes for Agents

1. **No existing linting tools** - Code style is manual
2. **Single HTML file** - No build system for frontend
3. **Chinese comments** - Maintain consistency with existing codebase
4. **WebCodecs API** - Requires Chrome 94+ or Edge 94+
5. **RTSP dependencies** - Requires running RTSP server for full testing
