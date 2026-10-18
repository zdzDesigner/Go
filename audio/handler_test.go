package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPAccessLibraryAndRangeStream(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "sample.mp3")
	content := []byte("0123456789")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("创建测试音频失败: %v", err)
	}

	library, err := NewLibrary(root)
	if err != nil {
		t.Fatalf("创建音频库失败: %v", err)
	}
	if err := library.Refresh(); err != nil {
		t.Fatalf("扫描音频库失败: %v", err)
	}
	access, err := NewAccessManager(true, "secret", time.Hour, false)
	if err != nil {
		t.Fatalf("创建访问控制器失败: %v", err)
	}
	app, err := NewApp(library, access)
	if err != nil {
		t.Fatalf("创建 HTTP 应用失败: %v", err)
	}
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	unauthorizedResponse, err := http.Get(server.URL + "/api/library")
	if err != nil {
		t.Fatalf("请求未认证音频列表失败: %v", err)
	}
	defer unauthorizedResponse.Body.Close()
	if unauthorizedResponse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("未认证状态码=%d，期望=%d", unauthorizedResponse.StatusCode, http.StatusUnauthorized)
	}

	accessResponse, err := http.Post(server.URL+"/api/access", "application/json", strings.NewReader(`{"key":"secret"}`))
	if err != nil {
		t.Fatalf("请求访问认证失败: %v", err)
	}
	defer accessResponse.Body.Close()
	if accessResponse.StatusCode != http.StatusOK {
		t.Fatalf("认证状态码=%d，期望=%d", accessResponse.StatusCode, http.StatusOK)
	}
	cookies := accessResponse.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("认证 Cookie 数量=%d，期望=1", len(cookies))
	}

	client := &http.Client{}
	listRequest, err := http.NewRequest(http.MethodGet, server.URL+"/api/library", nil)
	if err != nil {
		t.Fatalf("创建音频列表请求失败: %v", err)
	}
	listRequest.AddCookie(cookies[0])
	listResponse, err := client.Do(listRequest)
	if err != nil {
		t.Fatalf("请求已认证音频列表失败: %v", err)
	}
	listBody, err := io.ReadAll(listResponse.Body)
	closeErr := listResponse.Body.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("读取音频列表失败: read=%v close=%v", err, closeErr)
	}
	if listResponse.StatusCode != http.StatusOK {
		t.Fatalf("已认证列表状态码=%d，响应=%s", listResponse.StatusCode, listBody)
	}
	item := library.List()[0]
	if !strings.Contains(string(listBody), item.ID) {
		t.Fatalf("音频列表缺少资源 ID=%s，响应=%s", item.ID, listBody)
	}

	streamRequest, err := http.NewRequest(http.MethodGet, server.URL+"/api/audio/"+item.ID+"/stream", nil)
	if err != nil {
		t.Fatalf("创建音频流请求失败: %v", err)
	}
	streamRequest.Header.Set("Range", "bytes=2-5")
	streamRequest.AddCookie(cookies[0])
	streamResponse, err := client.Do(streamRequest)
	if err != nil {
		t.Fatalf("请求 Range 音频流失败: %v", err)
	}
	streamBody, err := io.ReadAll(streamResponse.Body)
	closeErr = streamResponse.Body.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("读取 Range 音频流失败: read=%v close=%v", err, closeErr)
	}
	if streamResponse.StatusCode != http.StatusPartialContent {
		t.Fatalf("Range 状态码=%d，期望=%d", streamResponse.StatusCode, http.StatusPartialContent)
	}
	if string(streamBody) != "2345" {
		t.Fatalf("Range 响应=%q，期望=%q", streamBody, "2345")
	}
}
