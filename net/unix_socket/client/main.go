package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	socket_path := "/tmp/unix_socket"

	// 连接到Unix域套接字服务器
	conn, err := net.Dial("unix", socket_path)
	if err != nil {
		log.Fatal("连接错误:", err)
	}
	defer conn.Close()

	// 向服务器发送数据
	_, err = conn.Write([]byte("客户端的问候!"))
	if err != nil {
		log.Fatal("发送数据错误:", err)
	}

	// 接收服务器的响应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal("读取数据错误:", err)
	}

	fmt.Printf("服务器响应: %s\n", string(buf[:n]))
}
