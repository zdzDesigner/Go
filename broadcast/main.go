package main

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// 定义广播端口。选择一个不常用的端口以避免冲突。
	// BROADCAST_PORT = ":19996"
	BROADCAST_PORT = ":19993"
	// 定义广播地址。255.255.255.255 是一个特殊的地址，
	// 代表当前局域网的所有主机。
	BROADCAST_IP = "255.255.255.255"
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

// listenForBroadcasts 函数设置一个UDP监听服务，用于接收网络中的广播消息。
// 这个版本通过设置 SO_REUSEPORT 套接字选项来实现端口复用。
func listenForBroadcasts() {
	// ListenConfig 提供了对网络监听过程的更多控制。
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var controlErr error
			// c.Control() 接受一个回调函数，该函数接收套接字的底层文件描述符（fd）。
			err := c.Control(func(fd uintptr) {
				// 使用 unix 包来设置套接字选项，以获得更好的兼容性和功能。
				// 设置 SO_REUSEADDR 允许将套接字绑定到已在使用中的地址。
				controlErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if controlErr != nil {
					return
				}

				// 设置 SO_REUSEPORT 允许多个套接字绑定到完全相同的IP地址和端口。
				controlErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			if err != nil {
				return err
			}
			return controlErr
		},
	}

	// 使用配置好的 ListenConfig 来创建一个 PacketConn，监听UDP流量。
	conn, err := lc.ListenPacket(context.Background(), "udp", BROADCAST_PORT)
	if err != nil {
		fmt.Println("错误：使用ListenConfig监听失败:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("正在端口 %s 上监听广播 (已启用端口复用)...\n", conn.LocalAddr().String())

	// 创建一个缓冲区来存储接收到的数据。
	buffer := make([]byte, 1024)
	for {
		// ReadFrom 是 PacketConn 接口的方法，用于读取数据包。
		n, remoteAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("错误：读取数据失败:", err)
			continue
		}
		fmt.Printf("从 %s 接收到广播: %s\n", remoteAddr, string(buffer[:n]))
	}
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
		message := "\"request\": \"remotekey.discovery\""

		// 将消息写入连接，发送出去。
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("错误：发送广播失败:", err)
			// 如果发送失败，通常意味着网络有问题，这里选择直接返回。
			return
		}

		// 等待2秒，然后进行下一次广播。
		time.Sleep(2 * time.Second)
	}
}
