package main

import (
	"fmt"
	"unsafe"
)

type Pointer struct {
	P uintptr
}

type Man struct {
	Name string
	Age  int
}

type WoMan struct {
	Name string
	Age  int
}

type Work1 struct {
	Name string
}

type Work2 struct {
	Name string
}

type Man2 struct {
	Name string
	Age  int
	Work Work1
}

type WoMan2 struct {
	Name string
	Age  int
	Work Work2
}

func main() {
	// base()
	conv()
}

// 使用于: /home/zdz/Documents/Webhook/webhook/src/service/scp/guoxue/query.go
func conv() {
	man := Man{"zdz", 33}
	fmt.Println(man)

	woMan := *(*WoMan)(unsafe.Pointer(&man))
	fmt.Println(woMan, woMan.Age, woMan.Name)

	man2 := Man2{"zdz", 33, Work1{"aa"}}
	fmt.Println(man2)

	woMan2 := *(*WoMan2)(unsafe.Pointer(&man2))
	fmt.Println(woMan2, woMan.Age, woMan.Name, woMan2.Work.Name)
}

func base() {
	str := "aa"
	fmt.Println(&str, *&str)
	// *(*string)(xxx) 语法格式
	fmt.Println(*(*string)(unsafe.Pointer(&str)))
	fmt.Println(unsafe.Pointer(&str))
	fmt.Println(unsafe.Pointer(&str))
	fmt.Println(uintptr(123))

	fmt.Println(Pointer{P: uintptr(unsafe.Pointer(&str))})
	fmt.Println(uintptr(unsafe.Pointer(&str)), unsafe.Pointer(&str), &str)
	fmt.Println(uintptr(unsafe.Pointer(&str))+8, unsafe.Pointer(&str), &str)
	p := uintptr(unsafe.Pointer(&str))
	fmt.Println(unsafe.Pointer(p))

	fmt.Println(unsafe.Pointer(uintptr(unsafe.Pointer(&str))))
	fmt.Println(*(*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&str)))))
}
