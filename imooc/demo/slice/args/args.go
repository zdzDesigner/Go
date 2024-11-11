package main

/*
 * 值是非地址
 */

import (
	"fmt"
)

func main() {
	args()
	// args2()
	// argsPointer()
	// address()
	// grow()
	// dynamic()
}

type V struct {
	name string
}

// slice参数: 只会修改cap的长度内的下标数据, 其它改动都不会影响内部 (下面的爬虫 &nodes 指针slice传递)
// /home/zdz/Documents/Try/Go/music/source/server/collect/dp.go
func args() {
	arr := []int{1, 2, 3}
	arr = append(arr, 5)
	args_int(arr)
	fmt.Println(arr, len(arr), cap(arr)) // [444 2 3 4] 4 6
	fmt.Println(arr[0:cap(arr)])         // [444 2 3 4 4 0]

	arr2 := []V{{"aaa"}, {"bbb"}, {"ccc"}}
	arr2 = append(arr2, V{"ddd"})
	args_struct(arr2)
	fmt.Println(arr2, len(arr2), cap(arr2))
	fmt.Println(arr2[0:cap(arr2)])
} // 函数内的len,cap扩容操作都会被销毁;非扩容的更改会被保留
func args_int(arr []int) {
	arr[0] = 555
	arr = append(arr, 666, 777)
	fmt.Println("args_int2::", arr, len(arr), cap(arr)) //  [444 2 3 5 6] len:5 cap:6
}

func args_struct(arr []V) {
	arr[0] = V{"zzz"}
	arr = append(arr, V{"arg"})
	fmt.Println("args_struct::", arr, len(arr), cap(arr))
}

//
func args2() {
	arr := []int{1, 2, 3}
	fmt.Printf("%p\n", &arr)
	args2_hander(arr)
	fmt.Println(arr)
	fmt.Printf("%p\n", &arr)
}

func args2_hander(arr []int) {
	arr = append(arr, 5, 6, 7)
	fmt.Printf("%p\n", &arr)
	fmt.Println(arr)
}

func argsPointer() {
	arr := []int{1, 2, 3}
	argsPointerHandler(&arr)
	fmt.Println(arr)
}

func argsPointerHandler(arr *[]int) {
	(*arr)[0] = 555
}

func address() {
	list := []int{3, 4, 5}
	println(list, len(list))
	_address(list)
	println(list, len(list))

	_address2(&list)
	println(list, len(list))
}

func _address(list []int) {
	list = append(list, 8)
	println(list[3])
}

func _address2(list *[]int) {
	*list = append(*list, 8)
}

func grow() {
	list := []int{1}
	fmt.Printf("%p,%p\n", list, &list)
	list = append(list, 1)
	fmt.Printf("%p,%p\n", list, &list)
}

func dynamic(count ...int) {
	fmt.Println(count, len(count))
	// && 有阻止执行作用
	if len(count) > 0 && count[0] > 3 {
		fmt.Println("")
	}

	// 此时count[0]会先执行再当成参数传入
	// panic: runtime error: index out of range [0] with length 0
	// fmt.Println(if3(len(count) > 0, count[0], 4))
}

func if3(ok bool, val, def int) int {
	if ok {
		return val
	}
	return def
}
