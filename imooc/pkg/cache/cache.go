package cache

import (
	"fmt"
	"time"
)

func Entry() {
	fmt.Println("redis-cli")
	Client().Set("name", "zdz", time.Second*100)
	cache := NewCache()
	cache.Sub("webhook.config.fetch.name", func(key, value string) {
		fmt.Println("=====Sub=====", key, value)
	})

	// go func() {
	// 	time.Sleep(time.Second * 5)
	// 	cache.Pub("name", "cccc")
	// }()
	for {
	}
}
