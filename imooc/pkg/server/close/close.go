package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
)

var (
	wg = sync.WaitGroup{} // 等待各个socket连接处理
)

func main() {

	stopChan := make(chan os.Signal) // 接收系统中断信号
	signal.Notify(stopChan, os.Interrupt)

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		<-stopChan
		fmt.Println("Get Stop Command. Now Stoping...")
		if err = listen.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Start listen :8080 ... ")
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			fmt.Println(err)
			continue
		}
		fmt.Println("Accept ", conn.RemoteAddr())
		wg.Add(1)
		go Handler(conn)
	}

	wg.Wait() // 等待是否有未处理完socket处理
	fmt.Println("over")
}

// Handler ..
func Handler(conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	// time.Sleep(5 * time.Second)

	conn.Write([]byte("Hello!"))
	fmt.Println("Send hello")
}
