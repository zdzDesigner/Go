package main

import "fmt"

type Model struct{}

func main() {
	// nilMap(nil)
	// nilSlice(nil)
	nilStruct(nil)
	nilStruct(&Model{})
}

func nilMap(obj map[string]string) {
	fmt.Println(obj, len(obj))
}

func nilSlice(obj []string) {
	fmt.Println(obj, len(obj))
}

func nilStruct(obj *Model) {
	fmt.Println(obj == nil)
}
