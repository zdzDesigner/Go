package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func main() {
	// readLine()
	// readFirstLine()
	// readFrom()
	bufioPeek()
	// bufioDiscard()
}

func bufioDiscard() {
	reader := strings.NewReader(`zdz todoadweradsaased`)

	br := bufio.NewReader(reader)
	// 丢弃 返回丢弃的个数
	br.Discard(4)
	buf := make([]byte, 4)
	_, err := br.Read(buf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(buf))
}

func bufioPeek() {
	reader := strings.NewReader(`zdz todoadweradsaased`)

	br := bufio.NewReader(reader)
	fmt.Println("br.Size:", br.Size())
	for {
		buf := make([]byte, 4)
		_, err := br.Read(buf)
		if err != nil {
			break
		}
		// bufio.Peek 窥探
		preBuf, _ := br.Peek(4)
		fmt.Println(string(buf))
		fmt.Println("preBuf::", string(preBuf))

	}
}

func readFirstLine() {
	count := 0
	reader := strings.NewReader(getBigStr())
	for {
		fmt.Println(reader.Len())
		time.Sleep(time.Millisecond * 100)
		// 每次都生成新的buf所以保存不了当前状态
		brd := bufio.NewReader(reader) // 每次读入 1<<12 个数据， 遇到换行，进入下一个循环（读出4096每组的第一行数据）
		line, isPrefix, err := brd.ReadLine()
		fmt.Println(string(line), isPrefix, err)
		count++
		if err != nil {
			break
		}
	}
	fmt.Println(count)
}

func readLine() {
	count := 0
	reader := strings.NewReader(getBigStr()) // reader 是 pool 池塘
	r := bufio.NewReader(reader)             // 每次 无业务数据（ReadLine时是无换行）都要从reader中取, 存有r,w

	fmt.Println(r.Size())

	for {
		// fmt.Println(reader.Len())
		time.Sleep(time.Millisecond * 100)
		line, isPrefix, err := r.ReadLine()
		fmt.Println(string(line), isPrefix, err)
		count++
		if err != nil {
			break
		}
	}
	fmt.Println(count)
}

func readFrom() {
	reader := strings.NewReader(`zdz todo
	adwer
	adsa
	ased
	`)

	// buf 的扩容, 使用了 bytes 中的 buf.ReadFrom, grow 自增策略
	buf1, _ := io.ReadAll(reader)
	fmt.Println(string(buf1))
	return
	buf := bytes.NewBuffer(make([]byte, 0, 10))
	_, err := buf.ReadFrom(reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf.Bytes()))
}

func getBigStr() (str string) {
	r := 0
	for i := 0; i < 1<<11; i++ {
		if len(str)-r > 100 {
			r = len(str)
			str += "\n"
		}
		str += strconv.Itoa(i)
	}
	// fmt.Println(str)
	return
}
