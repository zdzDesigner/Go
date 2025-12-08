package main

import (
	"fmt"
	"net"
	"time"
)

const (
	// 定义广播端口。选择一个不常用的端口以避免冲突。
	// BROADCAST_PORT = ":19996"
	BROADCAST_PORT = ":19993"
	// 定义广播地址。255.255.255.255 是一个特殊的地址，
	// 代表当前局域网的所有主机。
	BROADCAST_IP = "169.254.255.255"
	// BROADCAST_IP = "172.16.40.255"
	// BROADCAST_IP = "169.254.203.255"
	// BROADCAST_IP = "255.255.255.255"

)

func main() {
	// 在一个独立的 goroutine 中启动监听者，使其可以并发地接收消息。
	go listenForBroadcasts()

	// 在另一个独立的 goroutine 中启动广播者，使其可以并发地发送消息。
	go sendBroadcasts()

	// 使用一个空的 select 语句来永久阻塞 main goroutine。
	// 这可以防止 main 函数退出，从而保证其他的 goroutine 能持续运行。
	// 在实际应用中，可能会使用更优雅的方式（如 WaitGroup 或 channel）来管理生命周期。
	select {}
}

// sendBroadcasts 函数周期性地向局域网发送UDP广播消息。
func sendBroadcasts() {
	// 组合广播IP和端口。
	broadcastAddrStr := fmt.Sprintf("%s%s", BROADCAST_IP, BROADCAST_PORT)

	// 使用 "udp" 协议连接到广播地址。
	// 这并不会建立一个持久的连接，而是准备一个用于发送数据的套接字。
	conn, err := net.Dial("udp", broadcastAddrStr)
	if err != nil {
		fmt.Println("错误：连接到广播地址失败:", err)
		return
	}
	// 函数结束时关闭连接。
	defer conn.Close()

	fmt.Println("开始发送广播消息...")

	// 无限循环，周期性地发送消息。
	for {
		// 准备要发送的消息。
		message := "{\"request\":\"remotekey.discovery\"}"
		// message := "{\"request\":\"dev.discovery\"}"

		// 将消息写入连接，发送出去。
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("错误：发送广播失败:", err)
			// 如果发送失败，通常意味着网络有问题，这里选择直接返回。
			// return
		} else {
			fmt.Println("发送广播消息:", message)
		}

		// 等待2秒，然后进行下一次广播。
		time.Sleep(2 * time.Second)
	}
}

