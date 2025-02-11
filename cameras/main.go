package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	// "strings"

	"cameras/src/config"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

const (
	// rtspURL     = "rtsp://admin:password@192.168.1.100:554/stream" // RTSP源地址
	rtsp_url     = "rtsp://localhost:8554/mystream" // RTSP源地址
	hls_dir      = "./static/hls"                   // HLS输出目录
	hls_segment  = 5                                // 切片时长(秒)
	hls_playlist = "stream.m3u8"                    // 播放列表名
)

func main() {
	app := gin.New()
	app.Use(func(ctx *gin.Context) {
		fmt.Println(ctx.Request.URL)
		ctx.Next()
	})
	app.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".pdf", ".mp4"})))
	app.StaticFS("./static", http.Dir("static"))
	app.StaticFS("./assets", http.Dir("assets"))
	app.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	go convStream()

	app.Run(config.PORT)
	// 启动HTTP服务
	// http.Handle("/", http.FileServer(http.Dir(hls_dir)))
	// go func() {
	// 	log.Fatal(http.ListenAndServe(":8080", nil))
	// }()
}

func convStream() {
	// 清理旧HLS文件
	os.RemoveAll(hls_dir)
	os.MkdirAll(hls_dir, 0o755)

	// 启动FFmpeg转码进程
	cmd := exec.Command("ffmpeg",
		// "-rtsp_transport", "tcp", // 强制TCP传输
		"-i", rtsp_url, // 输入源
		"-c:v", "libx264", // 视频编码
		"-crf", "23", // 质量参数
		// "-preset", "veryfast", // 编码速度
		"-preset", "ultrafast", // 最快编码速度
		"-g", fmt.Sprintf("%d", hls_segment*2), // GOP大小
		"-f", "hls", // 输出格式
		"-hls_time", fmt.Sprintf("%d", hls_segment),
		"-hls_list_size", "6", // 播放列表保留切片数
		// "-hls_flags", "delete_segments", // 自动删除旧切片
		"-hls_flags", "delete_segments+append_list", // 实时更新播放列表
		"-tune", "zerolatency", // 零延迟编码
		filepath.Join(hls_dir, hls_playlist),
	)
	// ffmpeg_args := "-i rtsp://localhost:8554/mystream -c:v copy -c:a aac -f hls -hls_time 2 -hls_list_size 5 -hls_wrap 5 ./static/hls/output.m3u8"
	// cmd := exec.Command("ffmpeg", strings.Split(ffmpeg_args, " ")...)

	// 捕获FFmpeg日志
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	// 启动转码
	if err := cmd.Start(); err != nil {
		log.Fatalf("FFmpeg启动失败: %v", err)
	}
	fmt.Println("convert stream!")
	// defer cmd.Process.Kill()
}
