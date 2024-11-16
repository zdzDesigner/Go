package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func reuse(listener net.Listener) {
	// 强制获取底层的 *net.TCPListener
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		fmt.Println("Error: Unable to cast to *net.TCPListener")
		return
	}

	// 获取底层的文件描述符
	file, err := tcpListener.File()
	if err != nil {
		fmt.Println("Error getting file descriptor:", err)
		return
	}
	defer file.Close()

	// 设置 SO_REUSEADDR 选项
	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		fmt.Println("Error setting SO_REUSEADDR:", err)
		return
	}
}

// Entry ..
func Entry() {
	address := "127.0.0.1:6003"
	// address := "192.168.101.5:6002"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}

  // reuse(listener)
	fmt.Printf("start server: %s\n", address)

	// bts := make([]byte, 0, 4)
	for {
		// fmt.Println("for listener")
		// time.Sleep(time.Millisecond * 100)
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(conn.RemoteAddr())
		// bts, err := ioutil.ReadAll(conn)
		// fmt.Println(string(bts), err)
		conn.Write([]byte("hello"))

		// go echo(conn)
		// go Echo2(conn)
	}
}

func main() {
	Entry()
}

func echo(conn net.Conn) {
	r := bufio.NewReader(conn)
	b := make([]byte, 10000)
	r.Read(b)
	fmt.Println(string(b))
	for {
		_, err := r.ReadBytes(byte('\n'))
		// fmt.Println(r.Size())
		if err != nil {
			break
		}
		// switch err {
		// case nil:
		// 	fmt.Println("nil")
		// 	break
		// case io.EOF:
		// 	// fmt.Println("EOF")
		// default:
		// 	fmt.Println("ERROR", err)
		// }
		// fmt.Println(string(line))
		// conn.Write(line)
		// conn.Write([]byte(respRow(string(line))))

	}
	fmt.Println("close")
	conn.Close()
}

func Echo2(conn net.Conn) {
	r := bufio.NewReader(conn)
	buffer := bytes.NewBuffer([]byte{})
	headEnd := false
	isMethod := false
	contentLength := 0

	for {
		bit, err := r.ReadByte()
		if err != nil {
			break
		}
		// fmt.Println(string(bit))
		buffer.WriteByte(bit)
		// fmt.Println(string(buffer.Bytes()), buffer.Len())
		if bytes.IndexByte(buffer.Bytes(), byte('\n')) != -1 {
			line, err := buffer.ReadBytes(byte('\n'))
			if err != nil {
				fmt.Println("err:", err)
				continue
			}
			if !isMethod {
				fmt.Println(strings.Split(string(line), " ")[0])
				isMethod = true
			}
			// fmt.Print(string(line))
			if strings.Contains(string(line), "Content-Length:") {
				// fmt.Println(regexp.MustCompile(`.*?(\d+)\r\n`).ReplaceAllString(string(line), "$1"))
				contentLength, _ = strconv.Atoi(regexp.MustCompile(`.*?(\d+)\r\n`).ReplaceAllString(string(line), "$1"))
				// fmt.Println(contentLength)
			}
			// if bytes.Compare(line, []byte{13, 10}) == 0 {
			if string(line) == "\r\n" {
				headEnd = true
				// fmt.Println("head end")
			}
		}
		if headEnd {
			if buffer.Len() == contentLength {
				fmt.Println(string(buffer.Bytes()))
				// conn.Write(buffer.Bytes())
				conn.Write([]byte(respRow(string(buffer.Bytes()))))
				conn.Close()
			}
		}
	}
}

func respRow(line string) string {
	// fmt.Println(line)
	// return "HTTP/1.1 200 OK\r\n\r\n"
	return "HTTP/1.1 200 OK\r\n\r\nServer:OK"
	// return "HTTP/1.1 200 OK\r\nContent-Length: 81\r\n\r\nd" // transfer closed with 80 bytes remaining to read
	// return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-type: application/json\r\nContent-Length: %d\r\n\r\n%s", len([]byte(line)), line)
	// return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: 100000\r\n\r\n%s", len([]byte(line)), line)
	// return "aaa\r\naaa" // 少于4个字符被忽略
	// return "\r\naaa"
	// return "\naaa"
	// return "aaa"
}
