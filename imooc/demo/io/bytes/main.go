package main

import (
	"bytes"
)

func main() {
	bytes.NewReader([]byte{}).Read([]byte{})
}
