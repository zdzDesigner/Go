package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"cameras/src/service"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

var (
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	tracks     = make(map[*webrtc.TrackLocalStaticRTP]struct{})
	track_lock sync.RWMutex
)

func main() {
	// 启动 FFmpeg 生成 RTP 流
	cmd := exec.Command("ffmpeg",
		"-re",
		"-stream_loop", "-1",
		"-i", "/home/zdz/temp/video/SampleVideo_1280x720_5mb.mp4",
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-an",
		"-f", "rtp",
		"-sdp_file", "video.sdp", // 生成 SDP 文件用于解析参数
		"rtp://127.0.0.1:5004?pkt_size=1200",
	)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Process.Kill()

	// 启动广播服务
	go broadcastRTP()

	// 静态文件服务
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ws
	http.HandleFunc("/ws", websocketHandler)

	log.Println("Server running on :8777")
	log.Fatal(http.ListenAndServe(":8777", nil))
}

func broadcastRTP() {
	// 解析 SDP 获取 SSRC 和 PayloadType
	sdp, err := parseSDP("video.sdp")
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1500)
	sequenceNumber := uint16(0)
	timestamp := uint32(0)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取 RTP 失败: %v", err)
			continue
		}

		// fmt.Println(n, len(buffer[:n]))
		pkt := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    sdp.PayloadType,
				SequenceNumber: sequenceNumber,
				Timestamp:      timestamp,
				// SSRC:           sdp.SSRC,
			},
			Payload: buffer[:n],
		}

		track_lock.RLock()
		for track := range tracks {
			// fmt.Println("--------")
			if err := track.WriteRTP(pkt); err != nil {
				log.Printf("写入 RTP 失败: %v", err)
			}
		}
		track_lock.RUnlock()

		sequenceNumber++
		timestamp += 90000 / 30 // 假设 30fps
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer peerConnection.Close()

	// 创建视频轨道
	video_track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:  "video/h264",
			ClockRate: 90000,
		},
		"video",
		"streamxx",
	)
	if err != nil {
		log.Println(err)
		return
	}

	// 注册轨道
	track_lock.Lock()
	tracks[video_track] = struct{}{}
	track_lock.Unlock()
	defer func() {
		track_lock.Lock()
		delete(tracks, video_track)
		track_lock.Unlock()
	}()

	// 主动方 创建数据通道
	// dataChannel, err := peerConnection.CreateDataChannel("rtp-debug", nil)
	// if err != nil {
	// 	panic(err)
	// }

	// dataChannel.OnOpen(func() {
	// 	fmt.Println("dataChannel OPEN")
	// 	dataChannel.Send([]byte("xxxxxx"))
	// })
	// // fmt.Println("dataChannel:", dataChannel)
	// dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	fmt.Println(string(msg.Data))
	// })

  // 被动方
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Println(string(msg.Data))
			dc.Send([]byte("xxxxxx"))
		})
	})

	// 添加轨道
	if _, err = peerConnection.AddTrack(video_track); err != nil {
		log.Println(err)
		return
	}

	// 信令处理
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var signal map[string]interface{}
		if err := json.Unmarshal(msg, &signal); err != nil {
			continue
		}

		signal_type := signal["type"].(string)
		// fmt.Println("signal:", signal)
		switch signal_type {
		case "offer":
			// sdp := signal.(map[string]interface{})
			if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  signal["sdp"].(string),
			}); err != nil {
				log.Println(err)
				return
			}

			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Println(err)
				return
			}

			if err = peerConnection.SetLocalDescription(answer); err != nil {
				log.Println(err)
				return
			}

			conn.WriteJSON(answer)
		case "candidate":
			// fmt.Println("candidate:", signal["candidate"])
			sdpMid := signal["sdpMid"].(string)
			sdpMLineIndex := uint16(signal["sdpMLineIndex"].(float64))
			if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{
				Candidate:     signal["candidate"].(string),
				SDPMid:        &sdpMid,
				SDPMLineIndex: &sdpMLineIndex,
			}); err != nil {
				log.Println(err)
			}
		}
	}
}

// 解析 SDP 文件获取关键参数
type sdpInfo struct {
	SSRC        uint32
	PayloadType uint8
}

func parseSDP(filepath string) (*sdpInfo, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	sdp, err := service.ParseSDP(data)
	if err != nil {
		return nil, err
	}
	// fmt.Println(sdp.Media[0].PayloadType)
	payloadType := sdp.Media[0].PayloadType
	// 实现 SDP 解析逻辑（根据实际生成的 SDP 文件）
	return &sdpInfo{SSRC: 12345678, PayloadType: uint8(payloadType)}, nil
}
