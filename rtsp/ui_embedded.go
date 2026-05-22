//go:build ui

package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed index.html
var embeddedUIFS embed.FS

// registerUIRoutes 在 UI 构建中按需暴露内置页面。
func registerUIRoutes(httpMux *http.ServeMux, enableUI bool) {
	if !enableUI {
		return
	}

	uiFS, err := fs.Sub(embeddedUIFS, ".")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(uiFS))
	httpMux.Handle("/", fileServer)
	log.Printf("内置 UI 已启用: /")
}
