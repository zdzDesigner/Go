// [输入]: AUDIO_ADDR、AUDIO_ROOT、AUDIO_ACCESS_ENABLED 及可选认证环境变量。
// [输出]: 已校验的 Config 配置对象。
// [定位]: 应用启动阶段的配置边界，拒绝缺失或非法配置。
// [同步]: README.md、CONTEXT.md。
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 保存音频 server 的运行配置。
type Config struct {
	Addr          string
	AudioRoot     string
	AccessEnabled bool
	AccessKey     string
	SessionTTL    time.Duration
	CookieSecure  bool
}

// LoadConfig 从环境变量读取并校验配置。
func LoadConfig() (Config, error) {
	addr, err := requiredEnv("AUDIO_ADDR")
	if err != nil {
		return Config{}, err
	}
	audioRoot, err := requiredEnv("AUDIO_ROOT")
	if err != nil {
		return Config{}, err
	}
	accessEnabledValue, err := requiredEnv("AUDIO_ACCESS_ENABLED")
	if err != nil {
		return Config{}, err
	}
	accessEnabled, err := strconv.ParseBool(accessEnabledValue)
	if err != nil {
		return Config{}, fmt.Errorf("AUDIO_ACCESS_ENABLED=%q 不是有效布尔值: %w", accessEnabledValue, err)
	}

	var accessKey string
	var sessionTTL time.Duration
	var cookieSecure bool
	if accessEnabled {
		accessKey, err = requiredEnv("AUDIO_ACCESS_KEY")
		if err != nil {
			return Config{}, err
		}
		ttlValue, err := requiredEnv("AUDIO_SESSION_TTL")
		if err != nil {
			return Config{}, err
		}
		sessionTTL, err = time.ParseDuration(ttlValue)
		if err != nil {
			return Config{}, fmt.Errorf("AUDIO_SESSION_TTL=%q 必须是正数时间长度: %w", ttlValue, err)
		}
		if sessionTTL <= 0 {
			return Config{}, fmt.Errorf("AUDIO_SESSION_TTL=%q 必须是正数时间长度", ttlValue)
		}

		secureValue, err := requiredEnv("AUDIO_COOKIE_SECURE")
		if err != nil {
			return Config{}, err
		}
		cookieSecure, err = strconv.ParseBool(secureValue)
		if err != nil {
			return Config{}, fmt.Errorf("AUDIO_COOKIE_SECURE=%q 不是有效布尔值: %w", secureValue, err)
		}
	}

	return Config{
		Addr:          addr,
		AudioRoot:     audioRoot,
		AccessEnabled: accessEnabled,
		AccessKey:     accessKey,
		SessionTTL:    sessionTTL,
		CookieSecure:  cookieSecure,
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("环境变量 %s 是必填项", name)
	}
	return value, nil
}
