package main

import (
	"errors"
	"fmt"
	"time"
)

func func1() error {
	respC := make(chan int)
	// 处理逻辑
	go func() {
		time.Sleep(time.Second * 3)
		respC <- 10
		fmt.Println("go route") // 不会被执行
		close(respC)
	}()

	// 超时逻辑
	select {
	case r := <-respC:
		fmt.Printf("Resp: %d\n", r)
		return nil
	case <-time.After(time.Second * 2):
		fmt.Println("catch timeout")
		return errors.New("timeout")
	}
}

func main() {
	err := func1()
	fmt.Printf("func1 error: %v\n", err)
	time.Sleep(time.Second * 20)
}
