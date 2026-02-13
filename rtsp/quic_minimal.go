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

// WebTransportWriter 管理WebTransport连接的写入和分发
type WebTransportWriter struct {
	mu       sync.Mutex                      // 互斥锁，保护共享数据
	sessions map[string]*WebTransportSession // WebTransport会话映射
	sps      []byte                          // 序列参数集 (Sequence Parameter Set)
	pps      []byte                          // 图像参数集 (Picture Parameter Set)
}

// WebTransportSession 表示一个连接的WebTransport会话
type WebTransportSession struct {
	stream    webtransport.Stream   // WebTransport双向流
	session   *webtransport.Session // WebTransport会话引用
	sessionID string                // 会话标识符
}

// NewWebTransportWriter 创建新的WebTransportWriter实例
func NewWebTransportWriter() *WebTransportWriter {
	return &WebTransportWriter{
		sessions: make(map[string]*WebTransportSession),
	}
}

// broadcastFrame 广播视频帧到所有连接的会话
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

			_, err = session.stream.Write(data)
			if err != nil {
				log.Printf("Error sending frame to session %s: %v", session.sessionID, err)
				session.stream.Close()
				delete(w.sessions, session.sessionID)
			}
		}
	}
}

// generateSelfSignedCert 生成自签名证书用于开发
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

// startWebTransportServer 启动WebTransport服务器
func startWebTransportServer(h264Writer *H264Writer, addr string) {
	// 生成证书
	cert, err := generateSelfSignedCert()
	if err != nil {
		log.Printf("Failed to generate certificate for WebTransport: %v", err)
		return
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
	server := &webtransport.Server{
		H3: h3Server,
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

		// 发送SPS/PPS配置
		wtWriter.mu.Lock()
		currentTime := uint64(time.Now().UnixNano() / 1000)

		// 发送SPS (Sequence Parameter Set)
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

		// 保持流活跃
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
		err := h3Server.ListenAndServeTLS("", "") // 使用已配置的证书
		if err != nil && err != http.ErrServerClosed {
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
