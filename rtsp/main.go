// Package main contains an RTSP client with WebSocket/H264 streaming.
package main

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
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

type H264Frame struct {
	Data      []byte `json:"data"`
	Timestamp uint64 `json:"timestamp"`
	IsKey     bool   `json:"is_key"`
}

type H264Writer struct {
	mu             sync.Mutex
	clients        map[string]*WebSocketClient
	pending        []byte
	firstTimestamp uint32
	startTime      time.Time
	sps            []byte
	pps            []byte
}

type WebSocketClient struct {
	conn     *websocket.Conn
	writer   *H264Writer
	clientID string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewH264Writer() *H264Writer {
	return &H264Writer{
		clients: make(map[string]*WebSocketClient),
	}
}

func (w *H264Writer) broadcastFrame(frame *H264Frame) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, client := range w.clients {
		if client.conn != nil {
			data, _ := json.Marshal(frame)
			if err := client.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Printf("Error sending frame to client %s: %v", client.clientID, err)
				client.conn.Close()
				delete(w.clients, client.clientID)
			}
		}
	}
}

func (w *H264Writer) processRTPPacket(pkt *rtp.Packet) *H264Frame {
	w.mu.Lock()
	defer w.mu.Unlock()

	payload := pkt.Payload

	if len(payload) < 2 {
		return nil
	}

	naluType := payload[0] & 0x1F
	nal := payload[0] & 0x60

	if naluType == 28 {
		if len(payload) < 2 {
			return nil
		}

		fuHeader := payload[1]
		isStart := fuHeader&0x80 != 0
		isEnd := fuHeader&0x40 != 0
		nalType := fuHeader & 0x1F

		reconstructed := []byte{nal | nalType}

		if isStart {
			w.pending = append(reconstructed, payload[2:]...)
		} else if len(w.pending) > 0 {
			w.pending = append(w.pending, payload[2:]...)
		}

		if isEnd && len(w.pending) > 0 {
			w.pending = append(w.pending, payload[2:]...)
			frame := w.buildFrame(w.pending)
			w.pending = nil
			return frame
		}

		return nil
	}

	if naluType >= 1 && naluType <= 12 {
		return w.buildFrame(payload)
	}

	if naluType == 7 {
		w.sps = payload
	}
	if naluType == 8 {
		w.pps = payload
	}

	return nil
}

func (w *H264Writer) buildFrame(data []byte) *H264Frame {
	nalType := data[0] & 0x1F
	isKey := nalType == 5

	timestamp := w.getTimestamp()

	frameData := append([]byte{0x00, 0x00, 0x00, 0x01}, data...)

	return &H264Frame{
		Data:      frameData,
		Timestamp: timestamp,
		IsKey:     isKey,
	}
}

func (w *H264Writer) getTimestamp() uint64 {
	if w.firstTimestamp == 0 {
		w.firstTimestamp = 0
		w.startTime = time.Now()
	}

	elapsed := time.Since(w.startTime).Microseconds()
	return uint64(elapsed)
}

func startWebSocketServer(h264Writer *H264Writer) {
	httpMux := http.NewServeMux()

	httpMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		clientID := fmt.Sprintf("client-%d", len(h264Writer.clients))
		client := &WebSocketClient{
			conn:     conn,
			writer:   h264Writer,
			clientID: clientID,
		}

		h264Writer.mu.Lock()
		h264Writer.clients[clientID] = client
		h264Writer.mu.Unlock()

		log.Printf("Client connected: %s", clientID)

		// Send config (SPS/PPS)
		h264Writer.mu.Lock()
		currentTime := uint64(time.Now().UnixNano() / 1000)
		if len(h264Writer.sps) > 0 {
			spsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.sps...)
			configMsg, _ := json.Marshal(H264Frame{
				Data:      spsData,
				Timestamp: currentTime,
				IsKey:     true,
			})
			conn.WriteMessage(websocket.BinaryMessage, configMsg)
			log.Printf("Sent SPS to client %s, length: %d", clientID, len(spsData))
		}
		if len(h264Writer.pps) > 0 {
			ppsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.pps...)
			configMsg, _ := json.Marshal(H264Frame{
				Data:      ppsData,
				Timestamp: currentTime,
				IsKey:     true,
			})
			conn.WriteMessage(websocket.BinaryMessage, configMsg)
			log.Printf("Sent PPS to client %s, length: %d", clientID, len(ppsData))
		}
		h264Writer.mu.Unlock()

		// Handle client messages (ping/pong, etc.)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		h264Writer.mu.Lock()
		delete(h264Writer.clients, clientID)
		h264Writer.mu.Unlock()

		conn.Close()
		log.Printf("Client disconnected: %s", clientID)
	})

	go func() {
		log.Printf("WebSocket server listening on :8080")
		if err := http.ListenAndServe(":8080", httpMux); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()
}

func parseFramesFromRTP(pkt *rtp.Packet) [][]byte {
	payload := pkt.Payload
	var frames [][]byte
	offset := 0

	for offset < len(payload) {
		if offset+4 > len(payload) {
			break
		}

		naluLen := int(payload[offset])<<24 | int(payload[offset+1])<<16 |
			int(payload[offset+2])<<8 | int(payload[offset+3])

		offset += 4

		if offset+naluLen > len(payload) {
			break
		}

		nalu := make([]byte, naluLen+4)
		copy(nalu, []byte{0x00, 0x00, 0x00, 0x01})
		copy(nalu[4:], payload[offset:offset+naluLen])

		frames = append(frames, nalu)
		offset += naluLen
	}

	return frames
}

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

func main() {
	h264Writer := NewH264Writer()

	startWebSocketServer(h264Writer)

	// rtspURL := "rtsp://172.16.40.9:554" // Adjust this to your actual RTSP stream URL
	rtspURL := "rtsp://localhost:8554/live" // Adjust this to your actual RTSP stream URL
	// Common formats: "rtsp://ip:port/", "rtsp://ip:port/stream", "rtsp://ip:port/live.sdp"

	u, err := base.ParseURL(rtspURL)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", rtspURL, err)
		panic(err)
	}

	log.Printf("Parsed URL - Scheme: %s, Host: %s, Path: %s", u.Scheme, u.Host, u.Path)

	c := gortsplib.Client{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	err = c.SetupAll(desc.BaseURL, desc.Medias)
	if err != nil {
		panic(err)
	}

	c.OnPacketRTPAny(func(medi *description.Media, ffmt format.Format, pkt *rtp.Packet) {
		frame := h264Writer.processRTPPacket(pkt)
		if frame != nil {
			// fmt.Println(frame)
			h264Writer.broadcastFrame(frame)
		}
	})

	c.OnPacketRTCPAny(func(medi *description.Media, pkt rtcp.Packet) {
		log.Printf("RTCP packet from media %v, type %T\n", medi, pkt)
	})

	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	panic(c.Wait())
}

var shutdownCtx, shutdownCancel = context.WithCancel(context.Background())

func init() {
	go func() {
		<-shutdownCtx.Done()
	}()
}
