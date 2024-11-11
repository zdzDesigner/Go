package main

import (
	"fmt"
	"time"
)

func main() {
	done := do()
	select {
	case <-done:
		fmt.Println("done")
		// logic
	case <-time.After(3 * time.Second):
		fmt.Println("timeout")
		// timeout
	}
}

func do() <-chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		// do something
		time.Sleep(time.Second * 3)
		// ...
		done <- struct{}{}
	}()
	return done
}
