package main

import (
	"fmt"
	"io/ioutil"
)

// 时间敏感
// 不带缓冲区的读写, 对内存空间无感知,时间要求快的

func main() {
	// os.Open => file => readAll(r,filesize)
	data, _ := ioutil.ReadFile("./main.go")
	fmt.Println(string(data))

	fs, _ := ioutil.ReadDir("./")
	fmt.Println(len(fs), fs)
}
