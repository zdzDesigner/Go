package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	ctx, cancel := do(ch)

	select {
	case <-ctx.Done():
		fmt.Println("done", <-ch)
	case <-time.After(3 * time.Second):

		// fmt.Println("timeout", <-ch)
		fmt.Println("timeout")
		cancel()
	}
}

func do(ch chan int) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.TODO())

	go func() {
		time.Sleep(time.Second * 12)
		cancel()
		ch <- 3
	}()

	// go func() {
	// 	time.Sleep(time.Second * 12)
	// 	fmt.Println("执行内容")
	// 	cancel()
	// }()

	return ctx, cancel
}
