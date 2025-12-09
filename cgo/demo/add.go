package main

// #cgo LDFLAGS: -L. -ladd
// #include "add.h"
import "C"
import "fmt"

func main() {
	result := C.add(10, 20)
	fmt.Printf("10 + 20 = %d\n", result)
}
