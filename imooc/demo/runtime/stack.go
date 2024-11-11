package main

import (
	"fmt"
	"runtime"
)

func main() {
	env()
	// stack()
}

func env() {
	fmt.Println(runtime.GOOS)
}

func stack() {
	fmt.Println(runtime.Caller(1))
	pc, _, _, _ := runtime.Caller(1)
	fmt.Println(runtime.FuncForPC(pc).Name())
}
