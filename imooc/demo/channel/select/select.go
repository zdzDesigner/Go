package main

import (
	"fmt"
	"time"
)

func main() {
	// demo_for()
	// nobuf_call1()
	nobuf_call2()
	// clog_call()
}

func clog_call() {

	// clog(ch)
}

func clog(ch chan string) {
	v := <-ch
	fmt.Println(v)
}

func nobuf_call1() {
	ch := make(chan string, 3) // 缓冲区打印了三个, 缓冲区中装载了部分数据
	nobuf_call(ch)
// 2
// 0
// 1
// 3333
// receive 3
// 1111
// 2222

}
func nobuf_call2() {
	ch := make(chan string) // 非缓冲区, 一个后卡住
	nobuf_call(ch)
// 2
// 0
// 1
// 3333
// receive 3

}
func nobuf_call(ch chan string) {
	for i, _ := range []int{1, 2, 3} {
		go func(i int) {
			fmt.Println(i)
			if i == 0 {
				time.Sleep(time.Second * 2)
				ch <- "1"
				fmt.Println("1111")
			}
			if i == 1 {
				time.Sleep(time.Second * 3)
				ch <- "2"
				fmt.Println("2222")
			}
			if i == 2 {
				time.Sleep(time.Second * 1)
				ch <- "3"
				fmt.Println("3333")
			}
		}(i)
	}
	nobuf(ch)

	time.Sleep(time.Second * 100)
}

func nobuf(ch chan string) {

	select {
	case v := <-ch:
		fmt.Println("receive", v)
	case <-time.After(time.Second * 2):
		fmt.Println("time out")
	}
}

func demo_for() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	ch3 := make(chan int)

	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("ch going")
		ch1 <- 1
	}()
	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("ch2 going")
		ch2 <- 2
	}()
	go func() {
		time.Sleep(time.Second)
		fmt.Println("ch3 going")
		ch3 <- 3
	}()

	for {
		select {
		case v, ok := <-ch3:
			fmt.Println("ch3", ok, v)
		case <-ch2:
			fmt.Println("ch2")
		case <-ch1:
			fmt.Println("ch1")
		}

		fmt.Println("----- code after select -----")
	}

}
