package main

import (
	"fmt"
)

func fromCharCode(codePoints ...int) string {
	var result string
	for _, code := range codePoints {
		result += string(rune(code))
	}
	return result
}

func main() {
	// 使用 fromCharCode 函数
	fmt.Println(fromCharCode(65, 66, 67)) // 输出 "ABC"
	fmt.Println(fromCharCode(36895))      // 输出 "速"
}
