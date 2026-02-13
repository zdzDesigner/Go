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

// WTWriter WebTransport专用的帧写入器
type WTWriter struct {
	mu       sync.Mutex
	sessions map[string]*WTSession
	sps      []byte
	pps      []byte
}

// WTSession 表示一个WebTransport会话
type WTSession struct {
	stream    webtransport.Stream
	session   *webtransport.Session
	sessionID string
	clientID  string
}

// NewWTWriter 创建新的WebTransport写入器
func NewWTWriter() *WTWriter {
	return &WTWriter{
		sessions: make(map[string]*WTSession),
	}
}

// BroadcastFrame 广播视频帧到所有WebTransport会话
func (w *WTWriter) BroadcastFrame(frame *H264Frame) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, session := range w.sessions {
		if session.stream != nil {
			data, err := json.Marshal(frame)
			if err != nil {
				log.Printf("WebTransport: Error marshaling frame to JSON for session %s: %v", session.sessionID, err)
				continue
			}

			_, err = session.stream.Write(data)
			if err != nil {
				log.Printf("WebTransport: Error sending frame to session %s: %v", session.sessionID, err)

				// 清理失败的会话
				session.stream.Close()
				delete(w.sessions, session.sessionID)
			}
		}
	}
}

// GenerateTLSCert 生成自签名TLS证书
func GenerateTLSCert() (tls.Certificate, error) {
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
func StartWebTransportServer(h264Writer *H264Writer, addr string) error {
	// 生成证书
	cert, err := GenerateTLSCert()
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// 创建HTTP/3服务器
	h3Server := &http3.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3"},
		},
	}

	// 创建WebTransport服务器实例
	wtServer := &webtransport.Server{
		H3: h3Server,
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源，生产环境下应更严格
			return true
		},
	}

	// 处理WebTransport连接
	http.HandleFunc("/wt", func(w http.ResponseWriter, r *http.Request) {
		// 升级到WebTransport连接
		session, err := wtServer.Upgrade(w, r)
		if err != nil {
			log.Printf("WebTransport: Failed to upgrade connection: %v", err)
			return
		}

		// 创建新会话
		sessionID := fmt.Sprintf("wt-%d", len(getGlobalWTWriter().sessions))
		clientID := fmt.Sprintf("wt-client-%d", len(getGlobalWTWriter().sessions))

		wtSession := &WTSession{
			session:   session,
			sessionID: sessionID,
			clientID:  clientID,
		}

		// 添加到全局会话映射
		globalWriter := getGlobalWTWriter()
		globalWriter.mu.Lock()
		globalWriter.sessions[sessionID] = wtSession
		globalWriter.mu.Unlock()

		log.Printf("WebTransport: New session connected: %s (client: %s)", sessionID, clientID)

		// 启动会话处理协程
		go handleWTSession(globalWriter, wtSession)
	})

	// 启动服务器
	go func() {
		log.Printf("WebTransport server listening on %s", addr)
		err := h3Server.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			log.Printf("WebTransport server error: %v", err)
		}
	}()

	return nil
}

// handleWTSession 处理会话
func handleWTSession(globalWriter *WTWriter, wtSession *WTSession) {
	// 创建数据流
	stream, err := wtSession.session.AcceptStream(context.Background())
	if err != nil {
		log.Printf("WebTransport: Failed to accept stream for session %s: %v", wtSession.sessionID, err)
		return
	}

	wtSession.stream = stream

	// 发送初始配置（SPS/PPS）
	sendInitialConfig(globalWriter, wtSession, stream)

	// 读取消息循环（保持连接活跃）
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			break
		}
	}

	// 清理会话
	globalWriter.mu.Lock()
	delete(globalWriter.sessions, wtSession.sessionID)
	globalWriter.mu.Unlock()

	stream.Close()
	log.Printf("WebTransport: Session disconnected: %s", wtSession.sessionID)
}

// sendInitialConfig 发送初始配置（SPS/PPS）
func sendInitialConfig(globalWriter *WTWriter, wtSession *WTSession, stream webtransport.Stream) {
	globalWriter.mu.Lock()
	defer globalWriter.mu.Unlock()

	currentTime := uint64(time.Now().UnixNano() / 1000)

	// 发送SPS
	if len(globalWriter.sps) > 0 {
		spsData := append([]byte{0x00, 0x00, 0x00, 0x01}, globalWriter.sps...)
		log.Printf("WebTransport: Sending SPS to session %s, length: %d", wtSession.sessionID, len(spsData))

		configMsg, err := json.Marshal(H264Frame{
			Data:      spsData,
			Timestamp: currentTime,
			IsKey:     true,
		})
		if err != nil {
			log.Printf("WebTransport: Error marshaling SPS for session %s: %v", wtSession.sessionID, err)
		} else {
			_, err := stream.Write(configMsg)
			if err != nil {
				log.Printf("WebTransport: Error sending SPS to session %s: %v", wtSession.sessionID, err)
			} else {
				log.Printf("WebTransport: Sent SPS to session %s", wtSession.sessionID)
			}
		}
	}

	// 发送PPS
	if len(globalWriter.pps) > 0 {
		ppsData := append([]byte{0x00, 0x00, 0x00, 0x01}, globalWriter.pps...)
		log.Printf("WebTransport: Sending PPS to session %s, length: %d", wtSession.sessionID, len(ppsData))

		configMsg, err := json.Marshal(H264Frame{
			Data:      ppsData,
			Timestamp: currentTime,
			IsKey:     true,
		})
		if err != nil {
			log.Printf("WebTransport: Error marshaling PPS for session %s: %v", wtSession.sessionID, err)
		} else {
			_, err := stream.Write(configMsg)
			if err != nil {
				log.Printf("WebTransport: Error sending PPS to session %s: %v", wtSession.sessionID, err)
			} else {
				log.Printf("WebTransport: Sent PPS to session %s", wtSession.sessionID)
			}
		}
	}
}

// 全局WebTransport写入器
var (
	globalWTWriter *WTWriter
	globalWTONCE   sync.Once
)

func getGlobalWTWriter() *WTWriter {
	globalWTONCE.Do(func() {
		globalWTWriter = NewWTWriter()
	})
	return globalWTWriter
}

// SyncSPSPPS 同步SPS/PPS到WebTransport写入器
func SyncSPSPPS(h264Writer *H264Writer) {
	globalWriter := getGlobalWTWriter()
	globalWriter.mu.Lock()
	defer globalWriter.mu.Unlock()

	if len(h264Writer.sps) > 0 {
		globalWriter.sps = make([]byte, len(h264Writer.sps))
		copy(globalWriter.sps, h264Writer.sps)
	}

	if len(h264Writer.pps) > 0 {
		globalWriter.pps = make([]byte, len(h264Writer.pps))
		copy(globalWriter.pps, h264Writer.pps)
	}

	log.Printf("WebTransport: Synchronized SPS (len=%d) and PPS (len=%d)", len(globalWriter.sps), len(globalWriter.pps))
}

// UpdateWTConnections 更新所有WebTransport连接
func UpdateWTConnections(frame *H264Frame) {
	getGlobalWTWriter().BroadcastFrame(frame)
}
