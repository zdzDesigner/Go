package main

import "fmt"

type V struct {
	name string
}

// copy
func main() {
	// v := make(V) // make  只能给内置类型 map slice channel 分配内存
}

func flat() {
	v := V{name: "ccccc"}
	call(v)
	fmt.Println(v) // ccccc
}

func newPointer() { // new 返回指针
	v := new(V)
	v.name = "ccccc"
	call(*v)
	fmt.Println(v) // ccccc
}

func call(v V) {
	v.name = "dddd"
}
