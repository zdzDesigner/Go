package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	base()
}

func base() {
	filea, err := os.Open("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write_stream/a.text")
	fileb, err := os.OpenFile("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write_stream/b.text", os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer filea.Close()
	defer fileb.Close()

	stream := make([]byte, 4)
	for {
		time.Sleep(time.Millisecond * 100)
		n, err := filea.Read(stream)
		// 获取当前文件指针位置 io.SeekCurrent
		position_cur, _ := filea.Seek(0, io.SeekCurrent)
		fmt.Println("position_cur:", position_cur)
		// position_start, _ := filea.Seek(0, io.SeekStart)
		// fmt.Println("position_start:", position_start)
		// position_end, _ := filea.Seek(0, io.SeekEnd)
		// fmt.Println("position_end", position_end)

		if err != nil && err == io.EOF {
			fmt.Println("Read:", err)
			break
		}

		fmt.Println(string(stream), n)
		if _, err := fileb.Write(stream[:n]); err != nil {
			fmt.Println("Write:", err)
			break
		}

	}
}
