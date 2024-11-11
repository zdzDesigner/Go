package main

import (
	"bytes"
	"fmt"
)

func main() {
	// buf := bufio.NewReader()
	_byte()
	// _writeTo()
}

func _byte() {
	buf := bytes.NewBuffer([]byte{})
	fmt.Println(buf.Bytes())
	buf.Write([]byte{1, 2, 3, 4, 5, 6})
	fmt.Println(buf.Bytes())
	fmt.Println(buf.ReadByte())
	fmt.Println(buf.Bytes())
}

func _writeTo() {
	bufSource := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6})
	buf := bytes.NewBuffer([]byte{})
	fmt.Println("buf:", buf.Bytes())
	fmt.Println("bufSource:", bufSource.Bytes())
	fmt.Println(bufSource.WriteTo(buf))
	fmt.Println("buf:", buf.Bytes())
	fmt.Println("bufSource:", bufSource.Bytes())
}
