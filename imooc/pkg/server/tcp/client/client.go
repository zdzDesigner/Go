package client

import (
	"fmt"
	"net"
	"os"
)

// 发起请求时同样会打开文件描述符
func NewClient() {
	chwt := make(chan struct{})
	strEcho := "Halo"
	servAddr := "localhost:8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	fmt.Println("pid::", os.Getpid())

	go func() {
		for {

		}
	}()

	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	println("write to server = ", strEcho)

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	println("reply from server=", string(reply))

	conn.Close()
	fmt.Println("conn close")
	<-chwt
}
