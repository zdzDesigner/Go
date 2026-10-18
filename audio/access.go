// [输入]: 共享访问密钥、会话有效期和 HTTP 请求中的 Cookie。
// [输出]: 访问会话 Cookie 及会话校验中间件。
// [定位]: 无帐号访问控制模块，只维护内存中的匿名会话。
// [同步]: CONTEXT.md、README.md、docs/adr/0001-device-local-playback-progress.md。
package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const sessionCookieName = "audio_session"

// AccessManager 管理共享密钥创建的匿名会话。
type AccessManager struct {
	enabled  bool
	key      string
	ttl      time.Duration
	secure   bool
	mu       sync.Mutex
	sessions map[string]time.Time
}

// NewAccessManager 创建访问控制器。
func NewAccessManager(enabled bool, key string, ttl time.Duration, secure bool) (*AccessManager, error) {
	if !enabled {
		return &AccessManager{}, nil
	}
	if key == "" {
		return nil, fmt.Errorf("共享访问密钥不能为空")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("会话有效期必须为正数，实际值=%s", ttl)
	}
	return &AccessManager{
		enabled:  true,
		key:      key,
		ttl:      ttl,
		secure:   secure,
		sessions: make(map[string]time.Time),
	}, nil
}

// Enabled 返回是否启用了共享访问密钥。
func (a *AccessManager) Enabled() bool {
	return a.enabled
}

// VerifyKey 使用常量时间比较校验共享访问密钥。
func (a *AccessManager) VerifyKey(key string) bool {
	return subtle.ConstantTimeCompare([]byte(a.key), []byte(key)) == 1
}

// CreateSession 创建匿名访问会话并写入 Cookie。
func (a *AccessManager) CreateSession(w http.ResponseWriter) error {
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return fmt.Errorf("生成访问会话令牌失败: %w", err)
	}
	token := hex.EncodeToString(rawToken)
	expiresAt := time.Now().Add(a.ttl)

	a.mu.Lock()
	now := time.Now()
	for existingToken, expiry := range a.sessions {
		if !expiry.After(now) {
			delete(a.sessions, existingToken)
		}
	}
	a.sessions[token] = expiresAt
	a.mu.Unlock()

	maxAge := int(a.ttl.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   a.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// HasSession 判断请求是否携带有效会话。
func (a *AccessManager) HasSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	a.mu.Lock()
	expiresAt, ok := a.sessions[cookie.Value]
	if ok && !expiresAt.After(time.Now()) {
		delete(a.sessions, cookie.Value)
		ok = false
	}
	a.mu.Unlock()
	return ok
}

// RequireSession 为受保护的 HTTP handler 添加会话校验。
func (a *AccessManager) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}
		if !a.HasSession(r) {
			writeAPIError(w, http.StatusUnauthorized, "需要有效的访问会话")
			return
		}
		next.ServeHTTP(w, r)
	})
}
