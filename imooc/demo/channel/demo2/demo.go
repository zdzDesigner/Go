package main

import (
	"fmt"
	"time"
)

func main() {
	// bufCh()
	// bufCh2()
	mutilSend()
}

func bufCh() {
	ch := make(chan int, 10)
	go func() {
		for i := 0; i < 3; i++ {
			ch <- i
		}
		close(ch)
	}()
	for v := range ch {
		fmt.Println(v)
	}
}

func bufCh2() {
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 4
	}()
	go func() {
		time.Sleep(time.Second)
		ch <- 2
	}()

	select {
	case v := <-ch:
		fmt.Println("select:", v)
		close(ch)
	case <-time.After(3 * time.Second):
		fmt.Println("default")
	}
	go func() {
		time.Sleep(time.Second)
		ch <- 3
	}()

	fmt.Println("close:", <-ch)
}

func mutilSend() {
	ch := mutilSendHandler()
	fmt.Println(<-ch) // 3
	fmt.Println("first ch")
	// fmt.Println(<-ch) // 4
	time.Sleep(time.Second * 2)
}

func mutilSendHandler() chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(time.Second)
		ch <- 3
		ch <- 4
	}()
	return ch
}
