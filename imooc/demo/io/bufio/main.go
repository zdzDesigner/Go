package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// 空间敏感
// bufio 带有缓冲区的读写, 内存小,时间无感知(速度慢), 一般处理大文件

func main() {
	f, err := os.Open("./main.go")
	defer f.Close()

	if err != nil {
		log.Fatal(err)
		return
	}

	buf := bufio.NewReader(f)
	count := 0
	for {
		count += 1
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			return
		}
		fmt.Println("line", line)
		// 这里是避免全部打印
		if count > 100 {
			break
		}
	}

}
