//go:build windows

package main

import (
	"context"
	"fmt"
	"net"
	"syscall"
)

// listenForBroadcasts 函数在 Windows 上设置一个UDP监听服务，用于接收网络中的广播消息。
// 这个版本通过设置 SO_REUSEADDR 套接字选项来实现端口复用。
func listenForBroadcasts() {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var controlErr error
			err := c.Control(func(fd uintptr) {
				// 在 Windows 上，设置 SO_REUSEADDR 允许多个套接字绑定到同一个地址和端口。
				// 这对于UDP广播监听来说，效果上类似于 *nix 系统上的 SO_REUSEPORT。
				controlErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
			if err != nil {
				return err
			}
			return controlErr
		},
	}

	conn, err := lc.ListenPacket(context.Background(), "udp", BROADCAST_PORT)
	if err != nil {
		fmt.Println("错误：使用ListenConfig监听失败:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("正在端口 %s 上监听广播 (已启用端口复用)...\n", conn.LocalAddr().String())

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Println("错误：读取数据失败:", err)
			continue
		}
		fmt.Printf("从 %s 接收到广播: %s\n", remoteAddr, string(buffer[:n]))
	}
}
