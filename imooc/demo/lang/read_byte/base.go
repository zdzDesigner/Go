package main

import (
	"fmt"
	"time"
)

func main() {
	// base1()
	base2()
}

func base2() {
	// s := "Clear is better than clever"
	b := make([]byte, 2, 3000)

	// for {
	// 	i := len(b)
	// 	b = b[:i]
	// 	read(b[i:cap(b)], s)

	// 	fmt.Println(b)
	// }

	b = b[:100]

	fmt.Println(len(b))

}

func base1() {
	read(make([]byte, 4), "Clear is better than clever")

}

func read(b []byte, s string) int {
	var (
		i int
	)
	for {
		time.Sleep(time.Second)
		fmt.Println(s[i:])
		if i >= len(s) {
			break
		}
		n := copy(b, s[i:])
		i = i + n
		fmt.Println(n, string(b[:n]))
	}
	return i
}
