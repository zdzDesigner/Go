package main

import (
	"fmt"
)

// 通信

func main() {
	ch := make(chan string)
	ch2 := make(chan string)
	go handler2(ch, ch2)
	go handler1(ch)

	fmt.Println(<-ch2)
}

func handler1(ch chan string) {
	ch <- "aaa"
}
func handler2(ch, ch2 chan string) {
	ch2 <- (<-ch) + "bbb"
}
