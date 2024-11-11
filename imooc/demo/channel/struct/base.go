package main

import (
	"fmt"
	"time"
)

func main() {
	// base()

}

//
//
//
//
func base2() {

	str1, str2 := make(chan string), make(chan string)
	defer func() {
		close(str1)
		close(str2)
	}()
	go func() {
		time.Sleep(time.Second * 2)
		str1 <- "str1"
	}()

	go func() {
		time.Sleep(time.Second * 1)
		str1 <- "str2"
	}()

	select {
	// case r1, r2 := <-str1,<-str1: // 失败 val ,ok:= <-v
	// fmt.Println(r1, r2)
	case r1, r2 := <-str1:
		fmt.Println(r1, r2)
	case <-time.After(time.Second * 10):
		fmt.Println("1 second")
	}
}

func base() {
	type zdz struct {
		name string
		work string
	}
	resch := make(chan zdz)

	go func() {
		time.Sleep(time.Second * 2)
		resch <- zdz{name: "zdz"}
	}()

	go func() {
		time.Sleep(time.Second * 1)
		close(resch)
	}()

	select {
	case res := <-resch:
		fmt.Println(res)
	case <-time.After(time.Second * 10):
		fmt.Println("time.After")
	}
}
