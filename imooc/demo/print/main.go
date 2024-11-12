package main

import (
	"fmt"
	"os"
)

type write func(p []byte) (n int, err error)

func w(p []byte) (n int, err error) {
	return 0, nil
}

func main() {
	// fprintf()
	fmt.Println(w.(write))
}

func base() {
	fmt.Fprint(os.Stdout, "Stdout")
	fmt.Println("\n--------------")

	fmt.Println(fmt.Sprint(333) == "333")                      // true
	fmt.Println(fmt.Sprint(333, 444, 55))                      // 333 444 55
	fmt.Println(fmt.Sprint("aa", "bb", "cc"))                  // aabbcc
	fmt.Println(fmt.Sprint(33, "bb", "cc", 444, "dd", 55, 66)) // 33bbcc444dd55 66
	fmt.Println(fmt.Sprint(333, 444, "aaa"))                   // 333 444aaa
	fmt.Println(fmt.Sprint(333, 444, "aaa") == "333 444aaa")   // true

	fmt.Fprint(os.Stderr, "Stderr========")
	fmt.Println("\n--------------")
}

// 多行字符串 ``
