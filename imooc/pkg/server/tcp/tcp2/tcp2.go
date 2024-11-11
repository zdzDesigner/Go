package tcp2

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

//
//	curl 'http://127.0.0.1:6002?aa=bb' -d '{aaa}'
//	POST /?aa=bb HTTP/1.1
//	Host: 127.0.0.1:6002
//	User-Agent: curl/7.58.0
//	Accept: */*
//	Content-Length: 5
//	Content-Type: application/x-www-form-urlencoded
//
// 根据Content-Length: 5检测传输完成 http协议内容

// Entry ..
func Entry() {
	listener, err := net.Listen("tcp", "127.0.0.1:6002")
	if err != nil {
		fmt.Println(err)
		return
	}

	// bts := make([]byte, 0, 4)
	for {
		// time.Sleep(time.Millisecond * 100)
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Println(conn.RemoteAddr())
		// bts, err := ioutil.ReadAll(conn)
		// fmt.Println(string(bts), err)
		// conn.Write([]byte("hello"))

		// go echo(conn)
		// go echo2(conn)
		// go doConn(conn)
		go handleConn(conn)
	}

}

func echo(conn net.Conn) {

	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadBytes(byte('\n'))
		switch err {
		case nil:
			fmt.Println(string(line))
			break
		case io.EOF:
			// conn.Write([]byte(respRow()))
			// fmt.Println("EOF")
		default:
			fmt.Println("ERROR", err)
		}

		// conn.Write([]byte(resp(string(line))))
		// conn.Close()
	}

}

func echo2(conn net.Conn) {

	r := bufio.NewReader(conn)
	// bs, _ := ioutil.ReadAll(r)
	// fmt.Println(string(bs))
	// conn.Write([]byte(resp(string(""))))
	// conn.Close()
	chunk := make([]byte, 4)
	buffer := bytes.NewBuffer(chunk)
	checkLen := false
	for {
		bit, err := r.ReadByte()
		if err == nil {
			// fmt.Println(string(bit))
			buffer.WriteByte(bit)
			// fmt.Println(string(buffer.Bytes()), buffer.Len())
			if bytes.IndexByte(buffer.Bytes(), byte('\n')) != -1 {
				line, err := buffer.ReadBytes(byte('\n'))
				if err != nil {
					fmt.Println(err)

					// continue
				} else {
					fmt.Println("line:", string(line), line, []byte{13, 0})
					if strings.Contains(string(line), "Content-Length:") {
						fmt.Println(string(line))
					}
					if bytes.Compare(line, []byte{13, 10}) == 0 {
						checkLen = true
						fmt.Println("head end")
					}

				}
			}
			if checkLen {
				if buffer.Len() == 5 {
					fmt.Println(string(buffer.Bytes()))
					conn.Write([]byte(resp(string(""))))
					conn.Close()
				}
			}
		}

	}

}

func resp(line string) string {
	fmt.Println(line)
	return "HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\ndddd"
	// return "HTTP/1.1 200 OK\r\nContent-Length: 81\r\n\r\nd" // transfer closed with 80 bytes remaining to read
	// return "aaa\r\naaa" // 少于4个字符被忽略
	// return "\r\naaa"
	// return "\naaa"
	// return "aaa"
}

// 自定义协议,约定了head大小, head中有 body大小
func doConn(conn net.Conn) {
	var (
		BUF_SIZE  = 1024
		HEAD_SIZE = 1024
		buffer    = bytes.NewBuffer(make([]byte, 0, BUF_SIZE)) //buffer用来缓存读取到的数据
		readBytes = make([]byte, BUF_SIZE)                     //readBytes用来接收每次读取的数据，每次读取完成之后将readBytes添加到buffer中
		isHead    = true                                       //用来标识当前的状态：正在处理size部分还是body部分
		bodyLen   = 0                                          //表示body的长度
	)

	for {
		//首先读取数据
		readByteNum, err := conn.Read(readBytes)
		if err != nil {
			log.Fatal(err)
			return
		}
		buffer.Write(readBytes[0:readByteNum]) //将读取到的数据放到buffer中

		// 然后处理数据
		for {
			if isHead {
				if buffer.Len() >= HEAD_SIZE {
					isHead = false
					head := make([]byte, HEAD_SIZE)
					_, err = buffer.Read(head)
					if err != nil {
						log.Fatal(err)
						return
					}
					bodyLen = int(binary.BigEndian.Uint16(head))
				} else {
					break
				}
			}

			if !isHead {
				if buffer.Len() >= bodyLen {
					body := make([]byte, bodyLen)
					_, err = buffer.Read(body[:bodyLen])
					if err != nil {
						log.Fatal(err)
						return
					}
					fmt.Println("received body: " + string(body[:bodyLen]))
					isHead = true
				} else {
					break
				}
			}
		}
	}
}

func handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	var jsonBuf bytes.Buffer
	for {
		// 读取一行数据，交给后台处理
		line, isPrefix, err := reader.ReadLine()
		fmt.Println("line: ", string(line))
		if len(line) > 0 {
			jsonBuf.Write(line)
			if !isPrefix {
				fmt.Println(string(jsonBuf.Bytes()))
				jsonBuf.Reset()
			}
		}
		if err != nil {
			break
		}
	}
	conn.Write([]byte("ok"))
	conn.Close()
}
