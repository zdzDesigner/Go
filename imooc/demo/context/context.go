package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// conn()
	// time.Sleep(time.Second * 6)
	// keepAlive()
	withValue()
}
func withValue() {
	ctx, cancle := context.WithCancel(context.Background())

	ctxval := context.WithValue(ctx, "playurl", "aa")
	go func() {
		// fmt.Println(ctxval.Value("playurl"))
		select {
		case <-ctxval.Done():
			fmt.Println(ctxval.Value("playurl"))
		default:
			fmt.Println("pp")
			time.Sleep(1 * time.Second)
		}
	}()
	cancle()
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ctxval.Done():
	// 			fmt.Println(ctxval.Value("playurl"))
	// 		}
	// 	}
	// }()
	time.Sleep(3 * time.Second)

}

func keepAlive() {

	for {
		fmt.Println("---- before")
		for {

		}
		fmt.Println("----")
	}

}

func conn() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		fmt.Println("defer cancel")
		cancel()
		cancel()
		cancel()
		cancel()
	}()
	go func() {
		time.Sleep(time.Second * 4)
		fmt.Println("timeout")
		cancel()
	}()
	time.Sleep(time.Second * 3)
	fmt.Println(ctx)
}
