package main

import (
	"fmt"
	"net"
)

func main() {
	server, err := net.Listen("tcp", ":1208")
	// server, err := net.Listen("udp", ":1208")
	if err != nil {
		fmt.Println("net.Listen error: ", err)
		return
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("server.Accept error: ", err)
			return
		}
		go func(conn net.Conn) {
			buf := make([]byte, 100)
			fmt.Println("init buf")
			for {
				n, err := conn.Read(buf)
				fmt.Println(string(buf[:n]), "--")
				if err != nil {
					fmt.Println("conn.Read error: ", err)
					return
				}
				conn.Write(buf[:n])

			}
		}(conn)
	}
}
