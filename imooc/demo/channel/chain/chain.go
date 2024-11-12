package main

import (
	"fmt"
	"time"
)

func main() {
	// base2()
	base3()
}

func base3() {
	a := make(chan string)
	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("go routine")
		a <- "aa"
		a <- "bb"
		close(a) // 这里重要，丢失：fatal error: all goroutines are asleep - deadlock!
	}()
	for v := range a { // 锁住
		fmt.Println(v)
	}

}

func base2() {
	a := make(chan string)
	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("go routine")
		a <- "aa"
		close(a)
	}()
	<-a // 锁住
	fmt.Println("chan end")

}

// 包装
func base() {
	chain(do())
	chain(chain(do()))
}

func chain(c chan int) chan int {
	return c
}

func do() chan int {
	return make(chan int)
}
