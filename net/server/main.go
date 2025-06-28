package main

import (
	"net"
	// "net/http"
)

func main() {
	// const server = http.Server
	listen, err := net.Listen("tcp", ":8999")
	println(err, listen)
	buf := [1024]byte{}
	for {
		// println("xxxx")
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		n, err := conn.Read(buf[0:])
		if err != nil {
			continue
		}
		println(string(buf[0:n]))
		conn.Write([]byte("HTTP/1.1 200 OK\n\nxxxxxx"))
		conn.Close()
	}
}
