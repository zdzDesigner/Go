package etcd

import (
	"fmt"
	"time"
)

func Entry() {
	fmt.Println("etcd start")
	Lock("lmy", func() {
		time.Sleep(time.Second * 10)
		fmt.Println("bns end")
	})
}
