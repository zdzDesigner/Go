package main

import (
	"fmt"
	"log"

	"github.com/neovim/go-client/nvim"
)

func main() {
	// 连接到 Neovim 服务端
	conn, err := nvim.Dial("/tmp/nvim.sock")
	if err != nil {
		log.Fatalf("Failed to connect to Neovim: %v", err)
	}
	defer conn.Close()

	// 获取当前缓冲区信息
	buffer, err := conn.CurrentBuffer()
	if err != nil {
		log.Fatalf("Failed to get current buffer: %v", err)
	}

	// 获取缓冲区的内容
	lines, err := conn.BufferLines(buffer, 0, -1, false)
	if err != nil {
		log.Fatalf("Failed to get buffer lines: %v", err)
	}
	fmt.Println("Buffer content:")
	for _, line := range lines {
		fmt.Println(string(line))
	}

	// 在 Neovim 中执行命令
	if err := conn.Command("echo 'Hello from Golang'"); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}

	// 设置一个全局变量
	if err := conn.SetVar("golang_var", "Golang Connected!"); err != nil {
		log.Fatalf("Failed to set variable: %v", err)
	}

	// 获取全局变量的值
	var value string
	if err := conn.Var("golang_var", &value); err != nil {
		log.Fatalf("Failed to get variable: %v", err)
	}
	fmt.Println("Neovim Variable golang_var:", value)
}
