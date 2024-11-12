package main

import (
	"bytes"
	"fmt"
)

func main() {
	src := "adafasdfasfasf"
	dist := make([]byte, 3)
	n, err := bytes.NewReader([]byte(src)).Read(dist)
	fmt.Println(n, err)
	fmt.Println(string(dist))

}
