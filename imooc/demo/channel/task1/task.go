package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	ch := make(chan int)
	for i := 0; i < 10; i++ {
		go do(i, ch)
		fmt.Println("NumGoroutine::", runtime.NumGoroutine())
	}

	for {
		time.Sleep(time.Microsecond * 100) // 任务的间隔时间
		select {
		case v := <-ch:
			fmt.Println(v)
		default:
			fmt.Println("default")
			goto OVER
		}
		fmt.Println("--")
	}
OVER:
	fmt.Println("OVER")
}

func do(i int, ch chan int) {
	fmt.Println("do::", i)
	ch <- i
}
