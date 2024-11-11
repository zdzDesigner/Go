package main

import (
	"fmt"
	"time"
)

type V struct {
	name string
}

func main() {
	arr := []V{{"aaa"}, {"bbb"}, {"ccc"}}

	// go func(arrnew []V) {
	// 	arrnew = append(arrnew, V{})
	// 	arrnew[0].name = "zzz"
	// 	fmt.Println(arrnew)
	// }(arr)
	// go func(arrnew []V) {
	// 	arrnew = append(arrnew, V{})
	// 	arrnew[0] = V{"ccc"}
	// 	fmt.Println(arrnew)
	// }(arr)
	arr = append(arr, V{"ddd"})
	arg(arr)

	time.Sleep(time.Second)
	fmt.Println(arr, len(arr), cap(arr))
	fmt.Println(arr[0:cap(arr)])
	fmt.Printf("%p\n", arr)
}

func arg(arr []V) {
	arr[0] = V{"zzz"}
	arr = append(arr, V{"arg"})
	fmt.Println(arr, len(arr), cap(arr))
	fmt.Printf("%p\n", arr)
}
