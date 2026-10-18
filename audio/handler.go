// [输入]: HTTP 请求、音频库、访问控制器和内嵌 H5 静态资源。
// [输出]: H5 页面、访问会话、音频列表、音频元数据和 Range 音频流。
// [定位]: HTTP 边界层，负责路由、协议转换和错误响应。
// [同步]: README.md、CONTEXT.md。
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed web/*
var embeddedWebFS embed.FS

// App 组装音频 server 的 HTTP 依赖。
type App struct {
	library *Library
	access  *AccessManager
	static  http.Handler
}

// NewApp 创建 HTTP 应用。
func NewApp(library *Library, access *AccessManager) (*App, error) {
	staticFS, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		return nil, fmt.Errorf("初始化内嵌 H5 文件系统失败: %w", err)
	}
	return &App{
		library: library,
		access:  access,
		static:  http.FileServer(http.FS(staticFS)),
	}, nil
}

// Handler 返回完整 HTTP 路由。
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/access", a.handleAccess)
	mux.Handle("/api/library", a.access.RequireSession(http.HandlerFunc(a.handleLibrary)))
	mux.Handle("/api/audio/", a.access.RequireSession(http.HandlerFunc(a.handleAudio)))
	mux.Handle("/", a.static)
	return requestLogger(mux)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeAPIError(w, http.StatusMethodNotAllowed, "健康检查只支持 GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type accessRequest struct {
	Key string `json:"key"`
}

func (a *App) handleAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeAPIError(w, http.StatusMethodNotAllowed, "访问认证只支持 POST")
		return
	}
	if !a.access.Enabled() {
		writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
		return
	}
	var request accessRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, fmt.Sprintf("访问密钥请求格式无效: %v", err))
		return
	}
	if !a.access.VerifyKey(request.Key) {
		writeAPIError(w, http.StatusUnauthorized, "访问密钥不正确")
		return
	}
	if err := a.access.CreateSession(w); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("创建访问会话失败: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeAPIError(w, http.StatusMethodNotAllowed, "音频列表只支持 GET")
		return
	}
	if err := a.library.Refresh(); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("刷新音频库失败: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": a.library.List()})
}

func (a *App) handleAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeAPIError(w, http.StatusMethodNotAllowed, "音频资源只支持 GET 或 HEAD")
		return
	}
	pathParts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/audio/"), "/"), "/")
	if len(pathParts) != 1 && !(len(pathParts) == 2 && pathParts[1] == "stream") {
		writeAPIError(w, http.StatusNotFound, "音频资源路径不存在")
		return
	}

	item, ok := a.library.Get(pathParts[0])
	if !ok {
		writeAPIError(w, http.StatusNotFound, fmt.Sprintf("音频资源 id=%s 不存在", pathParts[0]))
		return
	}
	if len(pathParts) == 1 {
		writeJSON(w, http.StatusOK, item)
		return
	}
	a.serveAudio(w, r, item)
}

func (a *App) serveAudio(w http.ResponseWriter, r *http.Request, item Audio) {
	path, err := a.library.PathFor(item)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, fmt.Sprintf("打开音频资源 id=%s 失败: %v", item.ID, err))
		return
	}
	file, err := os.Open(path)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, fmt.Sprintf("打开音频资源 id=%s path=%q 失败: %v", item.ID, item.RelativePath, err))
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("关闭音频文件失败: id=%s path=%q error=%v", item.ID, item.RelativePath, closeErr)
		}
	}()

	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Cache-Control", "private, no-store")
	http.ServeContent(w, r, item.Name, item.ModifiedAt, file)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("写入 JSON 响应失败: status=%d error=%v", status, err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("http request: method=%s path=%s", r.Method, r.URL.Path)
	})
}
