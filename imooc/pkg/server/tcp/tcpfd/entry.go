package tcpfd

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
)

// fd: 资源描述符

func Entry() {
	var (
		port = "9090"
		err  error
		// 控制回复, 验证 fd 的创建和销毁
		chit = make(chan struct{})
	)
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
	if err != nil {
		return
	}
	fmt.Println("listen::", port, os.Getpid())

	// channel 下发 conn.Close 标志位
	go func() {
		var order string
		for {
			n, err := fmt.Scanln(&order)
			fmt.Println(n, err)
			if order == "close" {
				fmt.Println("close chan")
				chit <- struct{}{}
			}
		}
	}()

	// listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConn4(conn, chit)

		// go handleConnGetBreak(conn)

	}

}

func handleConn4(conn net.Conn, chit chan struct{}) {

	fmt.Println(conn.(*net.TCPConn))
	s, ok := conn.(*net.TCPConn)
	if !ok {
		return
	}

	f, err := s.File()
	if err != nil {
		return
	}
	fmt.Println("fd::", int(f.Fd()))

	fi, err := f.Stat()
	fmt.Println("fi.Name()::", fi.Name())
	fmt.Println("fi.Size()::", fi.Size())
	fmt.Println("fi.Sys()::", fi.Sys())
	fmt.Println(err)
	isend := false

	go func() {
		bts := make([]byte, 0)
		for {
			bt := make([]byte, 10)
			n, err := syscall.Read(int(f.Fd()), bt)
			if err != nil || isend {
				fmt.Println(n, err)
				break
			}
			bts = append(bts, bt...)
			fmt.Println("-------::", string(bts))
		}
	}()
	go func() {
		bts := make([]byte, 0)
		for {
			bt := make([]byte, 10)
			n, err := syscall.Read(int(f.Fd()), bt)
			if err != nil || isend {
				fmt.Println(n, err)
				break
			}
			bts = append(bts, bt...)
			fmt.Println("++++++++::", string(bts))
		}
	}()

	// tcp1.Echo2(conn)
	// return
	// reader := bufio.NewReader(conn)
	// contentLength := 0

	// header
	// for {
	// 	// 读取一行数据，交给后台处理
	// 	line, _, err := reader.ReadLine()
	// 	if err != nil {
	// 		println("err:", err.Error())
	// 		break
	// 	}
	// 	fmt.Println(string(line))
	// 	if strings.Contains(string(line), "content-length:") {
	// 		fmt.Println("Content-Length:", string(line))
	// 		contentLength, _ = strconv.Atoi(regexp.MustCompile(`.*?(\d+)`).ReplaceAllString(string(line), "$1"))
	// 		fmt.Println(contentLength)
	// 	}
	// 	if len(line) == 0 {
	// 		fmt.Println("header")
	// 		break
	// 	}
	// }

	// select {
	// case <-chit:
	// 	fmt.Println("close pre::")
	// 	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nServer:OK"))
	// 	conn.Close()
	// 	fmt.Println("close aft::")
	// }
	<-chit
	isend = true
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 9\r\n\r\nServer:OK"))
	conn.Close()
	fmt.Println("close aft::")
	runtime.GC()

}
