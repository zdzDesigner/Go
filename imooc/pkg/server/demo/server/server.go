package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

var port = "127.0.0.1:6003"

func echo(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadBytes(byte('\n'))
		switch err {
		case nil:
			break
		case io.EOF:
		default:
			fmt.Println("ERROR", err)
		}

		conn.Write(append([]byte("respone:"), line...))
	}
}

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("ERROR", err)
			continue
		}
		go echo(conn)
	}
}
