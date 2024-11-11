package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Second)
			defer wg.Done()
		}()
	}
	// wg.Wait()此时也要go出去,防止在wg.Wait()出堵住
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	// 正常结束完成
	case <-done:
		fmt.Println("done")
	// 超时
	case <-time.After(500 * time.Millisecond):
		fmt.Println("timeout")
	}
}
