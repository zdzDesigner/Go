package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	base()
	// test()
	buf()
}

func base() {
	ch := make(chan string, 2)
	ch <- "naveen"
	ch <- "paul"
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

// 2个缓冲， 发送满阻塞，发，读非同步
func buf() {
	ch := make(chan int, 2)
	go bufHandler(ch)
	time.Sleep(2 * time.Second)
	for v := range ch {
		fmt.Println("read value", v, "from ch")
		time.Sleep(2 * time.Second)

	}
}

func bufHandler(ch chan int) {
	for i := 0; i < 5; i++ {
		ch <- i
		fmt.Println("successfully wrote", i, "to ch")
	}
	close(ch)
}

func test() {
	inCh := generator(20)
	outCh := make(chan int, 10) // 使用5个`do`协程同时处理输入数据
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go do(inCh, outCh, &wg)
	}
	go func() {
		wg.Wait()
		fmt.Println("+++++++++++++++")
		close(outCh)
	}()
	for r := range outCh {
		fmt.Println(r)
	}
}
func generator(n int) <-chan int {
	outCh := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			outCh <- i
		}
		close(outCh)
		println("---------")
	}()

	return outCh
}

func do(inCh <-chan int, outCh chan<- int, wg *sync.WaitGroup) {
	for v := range inCh {
		outCh <- v * v
	}
	wg.Done()
}
