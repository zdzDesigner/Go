package main

import (
	"bytes"
	"fmt"
)

func main() {
	fmt.Println("buffer")
	base()
}

func base() {
	b := bytes.NewBuffer([]byte{})
	fmt.Println("\t cap::", b.Cap()) // 0
	b.WriteString("abc")
	fmt.Println("\t len::", b.Len()) // 3
	fmt.Println("\t cap::", b.Cap()) // 3
	b.Grow(10)                       // grow 10
	fmt.Println("\t len::", b.Len()) // 3
	fmt.Println("\t cap::", b.Cap()) // 16:cap*2 + 10
	for i := 0; i < 6; i++ {
		b.WriteString("a")
	}
	fmt.Println("\t len::", b.Len()) // 9
	fmt.Println("\t cap::", b.Cap()) // 16
	for i := 0; i < 3; i++ {
		b.WriteString("aaa")
		fmt.Println("\t len::", b.Len()) // 19
		fmt.Println("\t cap::", b.Cap()) // 33
	}
	// 16,16
	// 17,(2*len(16)+step)
	fmt.Println("\t len::", b.Len()) // 19
	fmt.Println("\t cap::", b.Cap()) // 33
}
