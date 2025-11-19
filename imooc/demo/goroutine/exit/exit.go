package main

import (
	"fmt"
	"time"
)

func main() {
	lkarCh := make(chan interface{})
	// go func() {
	// 	for {
	// 		select {
	// 		case lkar := <-lkarCh:
	// 			if lkar == nil {
	// 				fmt.Println("租约结束")
	// 				return
	// 			}
	// 			fmt.Println("续租:")
	// 		}
	//
	// 		// fmt.Println("Hello gorotine")
	// 		// time.Sleep(time.Second)
	// 	}
	// }()
	go func() {
		for lkar := range lkarCh {
			if lkar == nil {
				fmt.Println("租约结束")
				return
			}
			fmt.Println("续租:")
		}

		// fmt.Println("Hello gorotine")
		// time.Sleep(time.Second)
	}()
	// lkarCh <- "aa"
	lkarCh <- nil
	// context.TODO()
	time.Sleep(time.Second)
}

