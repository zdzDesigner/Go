package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	fmt.Println(utf8.DecodeRune([]byte{118, 250, 255, 120}))
	fmt.Println(utf8.DecodeRune([]byte{118, 120, 250, 255}))
	fmt.Println(utf8.DecodeRune([]byte{250, 120, 255}))
	fmt.Println(utf8.DecodeRune([]byte{120, 250, 255}))
	fmt.Println(utf8.DecodeRune([]byte{120, 250, 255}))
	fmt.Println(utf8.DecodeRune([]byte{120, 118}))
}
