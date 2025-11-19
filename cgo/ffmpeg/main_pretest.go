package main

import (
	"fmt"
	"log"

	"github.com/asticode/go-astiav"
	// 根据需求添加更多，如 _ "github.com/asticode/go-astiav/swresample" 用于重采样
)

func main1() {
	// 设置日志（可选，调试用）

	// astiav.SetLogLevel(astiav.LogLevelDebug)
	astiav.SetLogLevel(astiav.LogLevelInfo)
	astiav.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, f, msg string) {
		if l <= astiav.LogLevelWarning {
			log.Printf("FFmpeg: %s (level: %d)\n", msg, l)
		}
	})

	// 打印 FFmpeg 版本（确认链接成功）
	// fmt.Printf("FFmpeg version: %s\n", astiav.VersionFFmpeg())

	// 测试分配对象（基本功能）
	fmtCtx := astiav.AllocFormatContext()
	if fmtCtx == nil {
		log.Fatal("Failed to allocate FormatContext")
	}
	defer fmtCtx.Free()

	fmt.Println("go-astiav v0.38.0 初始化成功！环境正常。")
}
