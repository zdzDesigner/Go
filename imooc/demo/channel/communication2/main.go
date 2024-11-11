package main

import (
	"fmt"
)

func main() {
	ch := make(chan string)
	ch2 := make(chan string)
	go handler2(ch2)
	go handler1(ch)

	// fmt.Println(<-ch)
	for v := range ch {
		fmt.Println(v)
	}
	for v := range ch2 {
		fmt.Println(v)
	}
}

func handler1(ch chan string) {
	ch <- "aaa"
	ch <- "aaa"
	ch <- "aaa"
	close(ch)
}
func handler2(ch2 chan string) {
	ch2 <- "bbb"
	ch2 <- "bbb"
	close(ch2)
}
