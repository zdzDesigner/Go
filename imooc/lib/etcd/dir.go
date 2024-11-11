package etcd

import "fmt"

const (
	// WS 项目的etcd 目录
	WS = "/webhook-source"
	// TwWeather 台湾天气
	TwWeather = "tw"
)

// TwWeatherLeaseKey 租约key
func TwWeatherLeaseKey(key string) string {
	return fmt.Sprintf("%s/%s/%s", WS, TwWeather, key)
}
