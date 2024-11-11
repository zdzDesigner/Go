package main

import (
	"fmt"
	"runtime"
	"time"
)

// 14000
// 13844
// 14268
// 10000 => 385M
// 100000 => 1106M
func main() {
	time.Sleep(time.Second * 5)
	for i := 0; i < 5000; i++ {
		go func() {
			a := make([]byte, 100000)
			// for j := 0; j < len(a); j++ {
			// 	a
			// }
			// fmt.Println(len(a))
			time.Sleep(time.Second * time.Duration(len(a)))
		}()
	}
	time.Sleep(time.Second * 10)
	fmt.Println(runtime.NumGoroutine())
	time.Sleep(time.Second * 130)
}
