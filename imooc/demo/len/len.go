package main

import "fmt"

func main() {
	var arr []int
	fmt.Println(arr == nil)
	arr = nil
	fmt.Println(len(arr))
}
