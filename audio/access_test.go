package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAccessManagerSession(t *testing.T) {
	manager, err := NewAccessManager(true, "secret", time.Hour, true)
	if err != nil {
		t.Fatalf("创建访问控制器失败: %v", err)
	}
	if !manager.VerifyKey("secret") {
		t.Fatal("正确访问密钥校验失败")
	}
	if manager.VerifyKey("wrong") {
		t.Fatal("错误访问密钥校验通过")
	}

	response := httptest.NewRecorder()
	if err := manager.CreateSession(response); err != nil {
		t.Fatalf("创建访问会话失败: %v", err)
	}
	cookie := response.Result().Cookies()
	if len(cookie) != 1 {
		t.Fatalf("Cookie 数量=%d，期望=1", len(cookie))
	}
	if !cookie[0].HttpOnly || !cookie[0].Secure {
		t.Fatalf("会话 Cookie 安全属性错误: %+v", cookie[0])
	}

	request := httptest.NewRequest("GET", "/api/library", nil)
	request.AddCookie(cookie[0])
	if !manager.HasSession(request) {
		t.Fatal("有效访问会话校验失败")
	}
}

func TestAccessManagerExpiresSession(t *testing.T) {
	manager, err := NewAccessManager(true, "secret", time.Millisecond, false)
	if err != nil {
		t.Fatalf("创建访问控制器失败: %v", err)
	}
	response := httptest.NewRecorder()
	if err := manager.CreateSession(response); err != nil {
		t.Fatalf("创建访问会话失败: %v", err)
	}
	cookie := response.Result().Cookies()[0]
	time.Sleep(5 * time.Millisecond)
	request := httptest.NewRequest("GET", "/api/library", nil)
	request.AddCookie(cookie)
	if manager.HasSession(request) {
		t.Fatal("已过期访问会话仍然有效")
	}
}

func TestDisabledAccessManagerAllowsRequests(t *testing.T) {
	manager, err := NewAccessManager(false, "", 0, false)
	if err != nil {
		t.Fatalf("关闭访问控制时不应创建失败: %v", err)
	}
	if manager.Enabled() {
		t.Fatal("关闭访问控制后 Enabled 应为 false")
	}

	called := false
	handler := manager.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest("GET", "/api/library", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if !called || response.Code != http.StatusNoContent {
		t.Fatalf("关闭访问控制后请求未放行: called=%v status=%d", called, response.Code)
	}
}
