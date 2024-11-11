package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	ch3 := make(chan int)

	go func() {
		time.Sleep(time.Second * 3)
		ch1 <- <-ch2
	}()
	go func() {
		time.Sleep(time.Second * 2)
		ch2 <- <-ch3
	}()
	go func() {
		time.Sleep(time.Second)
		ch3 <- 3
	}()
	fmt.Println(<-ch1)
}
