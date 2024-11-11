package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	block()
}

func block() {
	ch := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		fmt.Println("go recv")
		fmt.Println(<-ch) // 锁住

		for i := 0; i < 100; i++ {
			fmt.Println("go recv end", i)
		}

		fmt.Println("go recv end3")
		wg.Done()
	}()

	go func() {
		fmt.Println("go shend pre ")
		time.Sleep(time.Second * 3)
		ch <- 111
		fmt.Println("go shend ")
		wg.Done()
	}()

	wg.Wait()
}
