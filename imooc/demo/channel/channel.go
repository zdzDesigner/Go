package main

import (
	"fmt"
	"time"
)

func main() {
	// base()
	// base2()
	base3()
	// base4()
}

func base() {
	var val = make(chan string)
	var val2 = make(chan string)

	go func() {
		time.Sleep(time.Second * 3)
		val <- "aa"
	}()

	go func() {
		time.Sleep(time.Second * 6)
		val2 <- "aaaaaa"
	}()

	v := <-val
	v2 := <-val2
	if v != "" && v2 != "" {
		fmt.Println(v, v2)
	}
}

func base2() {
	var val = map[string](chan string){
		"token":  make(chan string),
		"userid": make(chan string),
	}
	timeStart := time.Now()
	go func() {
		time.Sleep(time.Second * 3)
		val["token"] <- "aa"
	}()

	go func() {
		time.Sleep(time.Second * 6)
		val["userid"] <- "aaaaaa"
	}()

	v := <-val["token"]
	v2 := <-val["userid"]
	if v != "" && v2 != "" {
		fmt.Println(time.Since(timeStart))
		fmt.Println(v, v2)
	}
}

func base3() {
	ch := make(chan int, 0)
	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(time.Second * 2)
			fmt.Println("i:", i)
			ch <- i
		}(i)
	}
	fmt.Println("-------")
	// fmt.Println(<-ch)
	for v := range ch {
		fmt.Println(v)
	}

}

func base4() {
	limits := make(chan struct{}, 2)
	for i := 0; i < 10; i++ {
		go func(i int) {
			// 缓冲区满了就会阻塞在这
			limits <- struct{}{}
			time.Sleep(time.Second * 2)
			fmt.Println("i:", i)
			<-limits
		}(i)
	}
	fmt.Println("pp")
}
