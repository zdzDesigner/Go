package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	count := 0
	for i := 0; i < 10; i++ {
		go done(ch, i, &count)
	}
	for v := range ch {
		fmt.Println(v)
	}
}

func done(ch chan int, i int, count *int) {
	ctx := do()

	select {
	case <-ctx.Done():
		fmt.Println("done")
		ch <- i
		*count++
		if *count >= 10 {
			close(ch)
		}
	case <-time.After(3 * time.Second):
		fmt.Println("timeout")
	}
}

func do() context.Context {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	time.Sleep(time.Second * 15)
	return ctx
}
