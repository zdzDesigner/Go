package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-ch, "--")
		}
	}()
	for i := 0; i < 10; i++ {
		go do(i, ch)
	}

	time.Sleep(time.Second)

}

func do(i int, ch chan int) {
	ch <- i
}
