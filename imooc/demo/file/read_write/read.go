package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fmt.Println("pid::", os.Getpid())
	chwt := make(chan struct{})
	go func() {
		for {

		}
	}()
	fileReadBase()
	<-chwt

}

func fileReadBase() {
	// f, err := os.Open("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write/a.text")
	f, err := os.OpenFile("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write/a.text", os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	// defer f.Close()
	// 查看文件描述符, f.Close() 销毁文件描述符

	txt, _ := ioutil.ReadAll(f)
	fmt.Println("ioutil.ReadAll::", string(txt))
	// 读完了

	bio := bufio.NewReader(f)
	for {
		v, err := bio.ReadString('\n')
		fmt.Println("line::", v)
		if err != nil {
			fmt.Println(err)
			break
		}

	}
	fmt.Println("write::")
	f.WriteString("zdzzdz")
	// fmt.Println(f.Read(make([]byte, 3000))) // 23 <nil>
}

func fileReadWrite() {
	filea, err := os.Open("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write/a.text")
	fileb, err := os.OpenFile("/home/zdz/Documents/Try/Go/imooc/demo/file/read_write/b.text", os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer filea.Close()
	defer fileb.Close()

	buf := bytes.NewBuffer(make([]byte, 0))

	n, err := buf.ReadFrom(filea)
	if err != nil {
		panic(err)
	}
	// fmt.Println("filea", buf.Bytes())

	fmt.Println(buf.String(), n)
	n, err = buf.WriteTo(fileb)
	fmt.Println(err, n)
}
