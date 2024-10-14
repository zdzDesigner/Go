package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	socket_path := "/tmp/unix_socket"

	// 如果套接字文件已经存在，先删除它
	if _, err := os.Stat(socket_path); err == nil {
		os.Remove(socket_path)
	}

	// 监听 Unix 域套接字
	listener, err := net.Listen("unix", socket_path)
	if err != nil {
		log.Fatal("监听错误:", err)
	}
	defer listener.Close()

	fmt.Println("服务器正在监听", socket_path)

	for {
		// 接受客户端连接
		conn, err := listener.Accept()
		if err != nil {
			log.Println("接受连接错误:", err)
			continue
		}

		// 处理客户端连接
		go handleConnection(conn)
	}
}

// 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 从客户端接收数据
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("读取数据错误:", err)
		return
	}

	fmt.Printf("收到数据: %s\n", string(buf[:n]))

	// 向客户端发送响应
	_, err = conn.Write([]byte("服务器的问候!"))
	if err != nil {
		log.Println("发送数据错误:", err)
		return
	}
}
