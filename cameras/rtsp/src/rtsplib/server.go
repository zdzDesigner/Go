package rtsplib

import (
	"log"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/rtph264"
)

func Server() {
	// 创建服务器
	server := &gortsplib.Server{
		RTSPAddress: ":8554", // 监听地址
	}

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	// 创建轨道（例如 H.264 视频）
	track := &gortsplib.TrackH264{
		PayloadType: 96,
	}

	// 创建会话
	session := gortsplib.NewServerSession(server, []*gortsplib.Track{track}, "/mystream")

	// 处理连接
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			session.HandleConn(conn)
		}
	}()

	log.Println("RTSP server running at rtsp://localhost:8554/mystream")
	select {} // 保持程序运行
}
