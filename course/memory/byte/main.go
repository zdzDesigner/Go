package main

import (
	"bytes"
	"fmt"
)

func main() {
	t_SplitN()
	t_Split()
}

// 切分几片
func t_SplitN() {
	bts := []byte("2,xx,cc,xxxx:34")

	// list := bytes.Split(bts, []byte{','})
	// fmt.Println(string(list[0]))
	// fmt.Println(string(list[1]))
	list := bytes.SplitN(bts, []byte{','}, 4)
	fmt.Println(len(list)) // 4
	fmt.Println(string(list[0]))
	fmt.Println(string(list[1]))
}

// 切分所有
func t_Split() {
	bts := []byte("2,xx,cc,xx,cxx:34")

	list := bytes.Split(bts, []byte{','})
	fmt.Println(len(list)) // 5
	fmt.Println(string(list[0]))
	fmt.Println(string(list[1]))
}
