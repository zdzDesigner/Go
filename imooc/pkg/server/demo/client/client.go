package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

var port = "127.0.0.1:6003"

func main() {
	conn, err := net.Dial("tcp", port)
	// fmt.Println("Dial")
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	fmt.Println("waiting ...")
	userInput := bufio.NewReader(os.Stdin)
	response := bufio.NewReader(conn)
	for {

		// fmt.Println(ioutil.ReadAll(os.Stdin)) // channel 此处打印会阻塞
		userLine, err := userInput.ReadBytes(byte('\n'))
		// fmt.Println("userLine: ", userLine)
		switch err {
		case nil:
			conn.Write(userLine)
		case io.EOF:
			os.Exit(0)
		default:
			fmt.Println("ERROR", err)
			os.Exit(1)
		}

		serverLine, err := response.ReadBytes(byte('\n'))
		switch err {
		case nil:
			fmt.Print(string(serverLine))
		case io.EOF:
			os.Exit(0)
		default:
			fmt.Println("ERROR", err)
			os.Exit(2)
		}
	}
}
