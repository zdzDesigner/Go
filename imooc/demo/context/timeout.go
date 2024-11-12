package main

import (
	"context"
	"fmt"
	"time"
)

// !!!!!!!!!!!!!!!!!!!!!!
// !! 取消 !!!!!!!!!!!!!!
// !! 超时 !!!!!!!!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!!

func main() {
	// 创建一个带有截止时间的context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go processRequest(ctx)

	select {
	case <-ctx.Done():
		fmt.Println("Request canceled or timed out")
	}
}

func processRequest(ctx context.Context) {
	// 模拟耗时操作
	time.Sleep(5 * time.Second)
  fmt.Println("not execute the line")

	// 检查context是否被取消
	select {
	case <-ctx.Done():
		fmt.Println("Processing canceled")
	default:
		fmt.Println("Processing complete")
	}
}
