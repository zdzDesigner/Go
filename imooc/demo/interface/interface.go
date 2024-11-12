package main

import (
	"fmt"
)

func main() {
	// receive(0)
	// receive("0")
	receive(nil)
	// isnil(nil)
	// isnil(0)
	// isnil("")
}

func receive(v interface{}) {
	if _, ok := v.(int); ok {
		fmt.Println("int")
	} else {
		fmt.Println("no int")
	}
}

func isnil(v interface{}) {
	fmt.Println(nil == v)
}
