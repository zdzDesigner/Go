package main

import "fmt"

func main() {

	// base(3)
	base2()
}

func base2() {
	goto CC
CC:
	str := "strcc"
	if true {
		str, err := ret()
		fmt.Println(str, err)
	}
	fmt.Println(str)
}

func base(steps ...int) {
	fmt.Println("aa" == "" || "" == "")
	fmt.Println(append(steps, 1)[0])
}

func ret() (string, error) {
	return "aaa", nil
}
