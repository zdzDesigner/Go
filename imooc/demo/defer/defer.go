package main

import (
	"fmt"
)

/**
* defer 中唯一修改返回参数的方式: 使用命名返回参数
 */
func main() {
	// fmt.Println("return:", base())
	// defer..
	// return: bbbbbbbbbbb

	fmt.Println("return:", base3())
	// defer.. aa
	// defer after.. bbb
	// return: aa

	// fmt.Println("return:", base4())
	// defer.. bb
	// return: ddd

	// fmt.Println("return:", backName())
	// fmt.Println("return:", noBackName())
	// blockName()
}

func base() (a string) { // 指针被更改
	a = "aa"
	defer func() {
		fmt.Println("defer...")
		a = "bbbbbbbbbbb"
	}()
	a = "ccc"
	return "dd"
}

func base3() string {
	a := "aa"
	defer func() {
		fmt.Println("defer..", a)
		a = "bbb"
		fmt.Println("defer after..", a)
	}()
	// a = "ccc"
	return a
}

func base4() (r string) {
	a := "aa"
	defer func() {
		fmt.Println("defer..", a)
		r = "ddd"
	}()
	a = "bb"
	return a
}

func backName() (name string) {
	defer func() {
		name = "lalala"
	}()
	return name
}

func noBackName() string {
	var name string
	defer func() {
		name = "lalala"
	}()
	return name
}

func blockName() {
	name := "aaa"
	defer func() {
		fmt.Println("defer:", name)
	}()
	fmt.Println(name)

	// 对name的4种相同操作：赋值
	// 1
	// name, c := "bbb", "c"
	// fmt.Println(c)
	// 2
	// name, err := val()
	// fmt.Println(err)
	// 3
	// if name, _ = val(); name != "" {
	// 	return
	// }
	// 4
	// name = "bbb"

	//
	// 此时重新声明了name
	if name, err := val(); err != nil {
		return
	} else {
		fmt.Println("block scope:", name)
	}
	fmt.Println(name)
}

func val() (string, error) {
	return "vvv", nil
}
