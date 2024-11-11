package tcp3

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Entry() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// go handleConn(conn)
		// go handleConn2(conn)
		// go handleConn3(conn)
		go handleConn4(conn)

		// go handleConnGetBreak(conn)

	}

}

func handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	buf := bytes.NewBuffer(make([]byte, 0))
	_, err := buf.ReadFrom(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println("val", string(buf.Bytes()))
	conn.Close()
}

func handleConn2(conn net.Conn) {
	reader := bufio.NewReader(conn)
	// var jsonBuf bytes.Buffer
	for {
		// 读取一行数据，交给后台处理
		line, isPrefix, err := reader.ReadLine()
		fmt.Println(string(line), isPrefix, err)
		// if len(line) > 0 {
		// jsonBuf.Write(line)
		// 	if !isPrefix {
		// 		fmt.Println(jsonBuf.String())
		// 		jsonBuf.Reset()
		// 	}
		// }
		if err != nil {
			break
		}
	}
	conn.Close()
}

// !只读取了一行
func handleConnErr(conn net.Conn) {
	for {
		// 读取一行数据，交给后台处理
		line, _, err := bufio.NewReader(conn).ReadLine()

		fmt.Println(string(line))
		if err != nil {
			println("err:", err.Error())
			break
		}
	}

	conn.Close()
}

func handleConn3(conn net.Conn) {
	// tcp1.Echo2(conn)
	// return
	reader := bufio.NewReader(conn)
	contentLength := 0
	count := 0
	end := false
	line := make([]byte, 0)
	body := make([]byte, 0)
	var err error
	for {
		time.Sleep(time.Millisecond * 100)
		// 读取一行数据，交给后台处理
		if end {
			b, err := reader.ReadByte()
			if err != nil {
				break
			}
			body = append(body, b)
			if len(body) == contentLength {
				fmt.Println(string(body))
				conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nServer:OK"))
				conn.Close()
			}

			line = append(line, 'a')
		} else {
			line, _, err = reader.ReadLine()
			fmt.Println(string(line))
		}

		if err != nil {
			println("err:", err.Error())
			break
		}
		if strings.Contains(string(line), "Content-Length:") {
			// fmt.Println(regexp.MustCompile(`.*?(\d+)\r\n`).ReplaceAllString(string(line), "$1"))
			contentLength, _ = strconv.Atoi(regexp.MustCompile(`.*?(\d+)`).ReplaceAllString(string(line), "$1"))
			fmt.Println(contentLength)
		}

		if len(line) == 0 {
			end = true
			fmt.Println("end", contentLength, count)
		}

	}
	fmt.Println(end, contentLength)

}

func handleConn4(conn net.Conn) {
	reader := bufio.NewReader(conn)
	contentLength := 0

	// header
	for {
		// 读取一行数据，交给后台处理
		line, _, err := reader.ReadLine()
		if err != nil {
			println("err:", err.Error())
			break
		}
		fmt.Println(string(line))
		if strings.Contains(string(line), "Content-Length:") {
			contentLength, _ = strconv.Atoi(regexp.MustCompile(`.*?(\d+)`).ReplaceAllString(string(line), "$1"))
			fmt.Println(contentLength)
		}
		if len(line) == 0 {
			fmt.Println("header")
			break
		}
	}

	// body
	body := make([]byte, 0)
	for {
		fmt.Println("body")
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
			break
		}
		body = append(body, b)
		if len(body) == contentLength {
			fmt.Println(string(body))
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nServer:OK"))
			// conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 9\r\n\r\nServer:OK"))
			conn.Close()
			break // read tcp 127.0.0.1:8080->127.0.0.1:39248: use of closed network connection

			// time.Sleep(time.Second * 3)
		}
	}

}

func handleConnGetBreak(conn net.Conn) {
	// tcp1.Echo2(conn)
	// return
	reader := bufio.NewReader(conn)
	end := false
	for {
		// time.Sleep(time.Millisecond * 300)
		// 读取一行数据，交给后台处理
		line, _, err := reader.ReadLine()
		fmt.Println(string(line))
		if err != nil {
			println("err:", err.Error())
			break
		}

		if len(line) == 0 {
			end = true
			fmt.Println("end")

		}
		if end {
			fmt.Println("send")
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nServer:OK"))
			conn.Close()
			break
		}

	}
	fmt.Println(end)

}
