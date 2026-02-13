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

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

// WTClient represents a WebTransport client
type WTClient struct {
	stream   webtransport.Stream
	session  *webtransport.Session
	clientID string
}

// WTServer WebTransport服务器管理器
type WTServer struct {
	mu      sync.Mutex
	clients map[string]*WTClient
	sps     []byte
	pps     []byte
	addr    string
}

// NewWTServer 创建新的WebTransport服务器
func NewWTServer(addr string) *WTServer {
	return &WTServer{
		clients: make(map[string]*WTClient),
		addr:    addr,
	}
}

// BroadcastFrame 广播视频帧到所有WebTransport客户端
func (s *WTServer) BroadcastFrame(frame *H264Frame) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 将H264Frame转换为字节切片
	data, err := json.Marshal(frame)
	if err != nil {
		log.Printf("WebTransport: Error marshaling frame: %v", err)
		return
	}

	for _, client := range s.clients {
		if client.stream != nil {
			_, err = client.stream.Write(data)
			if err != nil {
				log.Printf("WebTransport: Error sending frame to client %s: %v", client.clientID, err)
				client.stream.Close()
				delete(s.clients, client.clientID)
			}
		}
	}
}

// GenerateWTTLSCert 生成TLS证书
func GenerateWTTLSCert() (tls.Certificate, error) {
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
			Organization: []string{"RTSP-WebTransport-Demo"},
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

// StartWebTransportServer 启动WebTransport服务器
func (s *WTServer) StartWebTransportServer(h264Writer *H264Writer) error {
	// 生成证书
	cert, err := GenerateTLSCert()
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// 创建HTTP/3服务器
	h3Server := &http3.Server{
		Addr: s.addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3"},
		},
	}

	// 创建WebTransport服务器
	wtServer := &webtransport.Server{
		H3: h3Server,
		CheckOrigin: func(r *http.Request) bool {
			return true // 简单的CORS检查，生产环境应更严格
		},
	}

	// 处理WebTransport连接
	http.HandleFunc("/wt", func(w http.ResponseWriter, r *http.Request) {
		session, err := wtServer.Upgrade(w, r)
		if err != nil {
			log.Printf("WebTransport: Upgrade failed: %v", err)
			return
		}

		clientID := fmt.Sprintf("wt-%d", len(s.clients))
		client := &WTClient{
			session:  session,
			clientID: clientID,
		}

		// 添加客户端
		s.mu.Lock()
		s.clients[clientID] = client
		s.mu.Unlock()

		log.Printf("WebTransport: New client connected: %s", clientID)

		// 处理客户端连接
		s.handleClient(client, h264Writer)
	})

	// 启动服务器
	go func() {
		log.Printf("WebTransport server listening on %s", s.addr)
		err := h3Server.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			log.Printf("WebTransport server error: %v", err)
		}
	}()

	return nil
}

// handleClient 处理WebTransport客户端
func (s *WTServer) handleClient(client *WTClient, h264Writer *H264Writer) {
	// 创建数据流
	stream, err := client.session.AcceptStream(context.Background())
	if err != nil {
		log.Printf("WebTransport: Failed to accept stream for client %s: %v", client.clientID, err)
		return
	}

	client.stream = &stream

	// 发送初始配置信息
	s.sendInitialConfig(&stream, h264Writer)

	// 保持连接活跃
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			break
		}
	}

	// 清理客户端
	s.mu.Lock()
	delete(s.clients, client.clientID)
	s.mu.Unlock()

	stream.Close()
	log.Printf("WebTransport: Client disconnected: %s", client.clientID)
}

// sendInitialConfig 发送初始配置信息
func (s *WTServer) sendInitialConfig(stream *webtransport.Stream, h264Writer *H264Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentTime := uint64(time.Now().UnixNano() / 1000)

	// 发送SPS
	if len(h264Writer.sps) > 0 {
		spsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.sps...)
		log.Printf("WebTransport: Sending SPS to client, length: %d", len(spsData))

		frame := H264Frame{
			Data:      spsData,
			Timestamp: currentTime,
			IsKey:     true,
		}
		configMsg, err := json.Marshal(frame)
		if err != nil {
			log.Printf("WebTransport: Error marshaling SPS: %v", err)
		} else {
			_, err := stream.Write(configMsg)
			if err != nil {
				log.Printf("WebTransport: Error sending SPS: %v", err)
			} else {
				log.Printf("WebTransport: SPS sent successfully")
			}
		}
	}

	// 发送PPS
	if len(h264Writer.pps) > 0 {
		ppsData := append([]byte{0x00, 0x00, 0x00, 0x01}, h264Writer.pps...)
		log.Printf("WebTransport: Sending PPS to client, length: %d", len(ppsData))

		frame := H264Frame{
			Data:      ppsData,
			Timestamp: currentTime,
			IsKey:     true,
		}
		configMsg, err := json.Marshal(frame)
		if err != nil {
			log.Printf("WebTransport: Error marshaling PPS: %v", err)
		} else {
			_, err := stream.Write(configMsg)
			if err != nil {
				log.Printf("WebTransport: Error sending PPS: %v", err)
			} else {
				log.Printf("WebTransport: PPS sent successfully")
			}
		}
	}
}

// SyncSPSPPS 同步SPS/PPS
func (s *WTServer) SyncSPSPPS(h264Writer *H264Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(h264Writer.sps) > 0 {
		s.sps = make([]byte, len(h264Writer.sps))
		copy(s.sps, h264Writer.sps)
	}

	if len(h264Writer.pps) > 0 {
		s.pps = make([]byte, len(h264Writer.pps))
		copy(s.pps, h264Writer.pps)
	}
}

var globalWTServer *WTServer
var wtOnce sync.Once

func GetWTServer() *WTServer {
	wtOnce.Do(func() {
		globalWTServer = NewWTServer(":4433")
	})
	return globalWTServer
}
