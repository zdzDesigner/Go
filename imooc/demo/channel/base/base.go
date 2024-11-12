package main

import "fmt"

func main() {
	// base1()
	// base()
	// retain()
	run()
}

func run() {
	go func() {
		fmt.Println(1)
	}()
	go func() {
		fmt.Println(2)
	}()
	go func() {
		fmt.Println(33)
	}()
	go func() {
		fmt.Println(4)
	}()
	go func() {
		fmt.Println(5)
	}()
	go func() {
		fmt.Println(6)
	}()
	for {
	}

}

// 出现在net包中的，啥意思
func base1() {
	noCancel := (chan struct{})(nil)
	fmt.Println(noCancel, noCancel == nil) // nil, true
	// close(noCancel)                        // 非法关闭
}

func base() {
	noCancel := make(chan struct{})
	fmt.Println(noCancel)
	close(noCancel)
	fmt.Println(noCancel)
}

// 保持主程序
func retain() {
	ch := make(chan struct{})
	go func() {
		for {
			ch <- struct{}{}
		}
	}()
	fmt.Println("do program")
	<-ch
}

// 保持主程序
func retain2() {
	go func() {
		for {

		}

	}()
	fmt.Println("do program")
	// ch := make(chan struct{})
	// <-ch
	<-make(chan struct{})
}
