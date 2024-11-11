package etcddemo

import (
	"context"
	"fmt"
	"imooc/lib/etcd"
	"time"
)

// Entry ..
func Entry() {
	lock()
}

// Get ..
func get() {
	client := etcd.Client()
	client.Put(context.TODO(), "/temp/xx", "ddd")

	kv, err := client.Get(context.TODO(), "/temp/xx")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("kv :%+v\n", kv)
	fmt.Printf("kvs :%+v\n", kv.Kvs)
}

func lock() {
	etcd.Lock("/temp/name", func() {
		fmt.Println("lock start")
		time.Sleep(time.Second * 10)
		fmt.Println("lock end")
	})
}
