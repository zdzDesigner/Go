package main

import (
	"fmt"
	"os"
)

// `时间复杂度` 优先与 `空间复杂度`
// `时间`敏感
// 不带缓冲区的读写, 对内存`空间`无感知,时间要求快的

func main() {
	// os.Open => file => readAll(r,filesize)
	data, _ := os.ReadFile("./main.go")
	fmt.Println(string(data))

	fs, _ := os.ReadDir("./")
	fmt.Println(len(fs), fs)
}
