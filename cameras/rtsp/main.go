package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	// "strings"

	"cameras/src/config"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

const (
// rtspURL  = "rtsp://admin:password@192.168.1.100:554/stream" // RTSP源地址
// rtsp_url = "rtsp://localhost:8554/mystream"                 // RTSP源地址
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
	hls_dir := "./static/hls/"
	os.RemoveAll(hls_dir)
	uris := []string{
		"rtsp://localhost:8554/mystream1",
		"rtsp://localhost:8554/mystream2",
	}
	for index, uri := range uris {
		fmt.Println(uri, index)
		ctrl := make(chan string)
		go func() {
			time.Sleep(time.Second * 3)
			ctrl <- "reset"
		}()
		go convStream(uri, hls_dir+fmt.Sprintf("%d", index+1), ctrl)
	}

	// TODO:: 浏览器无法播放(检测ffmpeg pull错误, 重新执行ffmpeg 指令)

	app.Run(config.PORT)
	// 启动HTTP服务
	// http.Handle("/", http.FileServer(http.Dir(hls_dir)))
	// go func() {
	// 	log.Fatal(http.ListenAndServe(":8080", nil))
	// }()
}

func convStream(rtsp_url string, hls_dir string, ctrl <-chan string) {
	hls_segment := 5              // 切片时长(秒)
	hls_playlist := "stream.m3u8" // 播放列表名
	// 清理旧HLS文件
	os.MkdirAll(hls_dir, 0o755)

	// 启动FFmpeg转码进程
	cmd := exec.Command("ffmpeg",
		// "-rtsp_transport", "tcp", // 强制TCP传输
		// "-loglevel", "info",
		"-loglevel", "error",
		// "-loglevel", "fatal",
		// "-stimeout", "5000000",
		// "-reconnect", "1",
		// "-reconnect_streamed", "1",
		// "-reconnect_delay_max", "10",
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
	// ffmpeg_args := "ffmpeg -i rtsp://localhost:8554/mystream -c:v copy -c:a aac -f hls -hls_time 2 -hls_list_size 5 -hls_wrap 5 ./static/hls/output.m3u8"
	// cmd := exec.Command("ffmpeg", strings.Split(ffmpeg_args, " ")...)

	// 捕获FFmpeg日志
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	stderr_pipe, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	// 启动转码
	if err := cmd.Start(); err != nil {
		log.Fatalf("FFmpeg启动失败: %v", err)
	}
	// 实时读取stderr
	go func() {
		scanner := bufio.NewScanner(stderr_pipe)
		for scanner.Scan() {
			fmt.Printf("实时错误输出: %s\n", scanner.Text())
			cmd.Process.Kill()
			time.Sleep(time.Millisecond * 100)
			convStream(rtsp_url, hls_dir, ctrl)
			return
		}
	}()
	// 控制
	go func() {
		select {
		case res := <-ctrl:
			if res == "reset" {
				fmt.Println("reset")
			}
			return
		}
	}()

	fmt.Println("convert stream!")
}
