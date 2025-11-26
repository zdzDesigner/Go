//go:build linux || darwin || freebsd || openbsd || netbsd

package main

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

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
	buffer := make([]byte, 10240)
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
