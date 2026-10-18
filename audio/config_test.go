package main

import "testing"

func TestLoadConfig(t *testing.T) {
	env := map[string]string{
		"AUDIO_ADDR":           "127.0.0.1:8080",
		"AUDIO_ROOT":           "/tmp/audio",
		"AUDIO_ACCESS_ENABLED": "true",
		"AUDIO_ACCESS_KEY":     "secret",
		"AUDIO_SESSION_TTL":    "12h",
		"AUDIO_COOKIE_SECURE":  "true",
	}
	for key, value := range env {
		t.Setenv(key, value)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if config.Addr != env["AUDIO_ADDR"] || config.AudioRoot != env["AUDIO_ROOT"] || !config.AccessEnabled || !config.CookieSecure {
		t.Fatalf("配置内容错误: %+v", config)
	}
}

func TestLoadConfigRequiresAllValues(t *testing.T) {
	keys := []string{"AUDIO_ADDR", "AUDIO_ROOT", "AUDIO_ACCESS_ENABLED", "AUDIO_ACCESS_KEY", "AUDIO_SESSION_TTL", "AUDIO_COOKIE_SECURE"}
	for _, key := range keys {
		t.Setenv(key, "")
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("缺失配置时应返回错误")
	}
}

func TestLoadConfigDisablesOptionalAccessConfig(t *testing.T) {
	t.Setenv("AUDIO_ADDR", "127.0.0.1:8080")
	t.Setenv("AUDIO_ROOT", "/tmp/audio")
	t.Setenv("AUDIO_ACCESS_ENABLED", "false")
	t.Setenv("AUDIO_ACCESS_KEY", "")
	t.Setenv("AUDIO_SESSION_TTL", "")
	t.Setenv("AUDIO_COOKIE_SECURE", "")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("关闭访问控制时配置不应失败: %v", err)
	}
	if config.AccessEnabled || config.AccessKey != "" || config.SessionTTL != 0 || config.CookieSecure {
		t.Fatalf("关闭访问控制时配置内容错误: %+v", config)
	}
}
