package main

import (
	"fmt"
)

func main() {
	chs := make(chan int, 4)
	chs <- 3
	fmt.Println(chs, len(chs), cap(chs))
	val := <-chs
	fmt.Println(val, len(chs), cap(chs))
	close(chs)
	for val := range chs {
		fmt.Println(val)
	}
	fmt.Println("end")
}
