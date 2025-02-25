package rtsplib

import (
	"fmt"
	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/pkg/url"
)

func Pull() {
	// 创建客户端
	client := &gortsplib.Client{}

	// 解析 RTSP URL
	u, err := url.Parse("rtsp://username:password@host:port/path")
	if err != nil {
		panic(err)
	}

	// 连接到服务器
	err = client.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 获取流描述（SDP）
	tracks, baseURL, _, err := client.Describe(u)
	if err != nil {
		panic(err)
	}

	// 设置所有轨道
	err = client.SetupAll(tracks, baseURL)
	if err != nil {
		panic(err)
	}

	// 开始播放
	_, err = client.Play(nil)
	if err != nil {
		panic(err)
	}

	// 接收帧数据
	for {
		trackID, streamType, payload, err := client.ReadFrame()
		if err != nil {
			break
		}
		fmt.Printf("Track %d, Type %v, Payload size: %d bytes\n", trackID, streamType, len(payload))
	}
}
