// [输入]: 环境变量配置和音频根目录。
// [输出]: 可供微信 H5 访问的音频 HTTP server。
// [定位]: 应用启动入口，负责组装配置、领域模块和 HTTP 边界。
// [同步]: README.md、CONTEXT.md、docs/adr/0001-device-local-playback-progress.md。
package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	library, err := NewLibrary(config.AudioRoot)
	if err != nil {
		log.Fatalf("初始化音频库失败: %v", err)
	}
	if err := library.Refresh(); err != nil {
		log.Fatalf("首次扫描音频库失败: %v", err)
	}

	access, err := NewAccessManager(config.AccessEnabled, config.AccessKey, config.SessionTTL, config.CookieSecure)
	if err != nil {
		log.Fatalf("初始化访问控制失败: %v", err)
	}
	app, err := NewApp(library, access)
	if err != nil {
		log.Fatalf("初始化 HTTP 应用失败: %v", err)
	}

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	log.Printf("audio server listening: http://%s", config.Addr)
	log.Printf("audio root: %s", config.AudioRoot)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("audio server 运行失败: %v", err)
	}
}
