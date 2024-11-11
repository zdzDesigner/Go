package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	// base()
	// base2()
	// base3()
	// tee()
	base4()
}

// cap 区也可以写, 参考 demo/slice/args/args.go
func base4() {
	r := strings.NewReader("zdz todo")
	buf := make([]byte, 3, 6)
	fmt.Println(buf, len(buf), cap(buf)) // [0 0 0] 3 6
	n, _ := r.Read(buf)
	fmt.Println(n, string(buf)) // 3 zdz
	n, _ = r.Read(buf[len(buf):cap(buf)])

	fmt.Println(n, len(buf), string(buf), string(buf[0:cap(buf)])) // 3 3 zdz zdz to
	buf = buf[0:cap(buf)]
	fmt.Println(string(buf), cap(buf)) // zdz to 6
}

func base3() {
	b := make([]byte, 6) // 申请
	r := strings.NewReader("zdz todo")
	// 滑动窗口
	n, err := r.Read(b[:1]) // n:读取个数
	fmt.Println(n, err, b, r.Len(), r.Size())
	n, err = r.Read(b[1:2]) // n:读取个数
	fmt.Println(n, err, b, r.Len(), r.Size())
	n, err = r.Read(b[2:3]) // n:读取个数
	fmt.Println(n, err, b, r.Len(), r.Size())

}

func base2() {
	b := make([]byte, 2) // 申请
	r := strings.NewReader("zdz todo")
	n, err := r.Read(b) // n:读取个数
	fmt.Println(n, err, b, string(b), r.Len(), r.Size())
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读出
	fmt.Println(r.ReadByte()) // 读完
	fmt.Println(r.ReadByte()) // 读完

	// ioutil.ReadAll()

}
func base1() {
	r := strings.NewReader("zdz todo")
	b, err := ioutil.ReadAll(r)
	fmt.Println(string(b), err)
}
func tee() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	fmt.Println(tee, buf)
	printall := func(r io.Reader) {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s", b)
	}

	printall(tee)
	fmt.Println(buf)
	printall(&buf)

}
