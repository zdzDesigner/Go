//go:build !ui

package main

import (
	"log"
	"net/http"
)

// registerUIRoutes 在纯服务构建中保留空实现，避免将调试页面打进二进制。
func registerUIRoutes(_ *http.ServeMux, enableUI bool) {
	if enableUI {
		log.Printf("内置 UI 不可用：当前构建未包含 -tags ui")
	}
}
