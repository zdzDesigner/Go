package main

import (
	"fmt"
	"time"
)

func main() {
	// base()
	loop()
	// loop2()
}

func base() {
	timer := time.NewTimer(time.Second * 3)
	fmt.Println(timer)
	defer timer.Stop()
	for {
		select {
		case t := <-timer.C:
			fmt.Println("timer.C", t)
		case <-time.After(time.Second * 3):
			fmt.Println("1 second")
		}
	}
}

func loop() {

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Println("1 second")
			// default:
			// 	fmt.Println("default")
		}
	}
}

func loop2() {
	for {
		select {
		case <-time.After(time.Second * 3):
			fmt.Println("1 second")
		}
	}
}
