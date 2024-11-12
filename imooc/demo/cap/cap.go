package main

import (
	"fmt"
)

func main() {
	// base()
	// base1()
	base2()

}

func base() {
	a := [20]byte{}
	b := a[:4] // cap 传递了
	fmt.Printf("len(a): %d,cap(a): %d\n", len(a), cap(a))
	fmt.Printf("len(b): %d,cap(b): %d\n", len(b), cap(b))
}

func base1() {
	bs := []byte{}
	for i := 0; i < 100; i++ {
		bs = append(bs, 1)
		fmt.Println(len(bs), cap(bs))
	}
}

func base2() {
	buf := []byte{2, 3}
	fmt.Println(buf, len(buf), cap(buf))
	fmt.Println(buf[:2])
}
