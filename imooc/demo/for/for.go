package main

import (
	"fmt"
	"time"
)

func main() {
	// base1()
	// base2()
	// base3()
	base4()
}

func base4() {
	arr := make([]int, 0, 4)
	for i := 0; i < 10; i++ {
		arr = append(arr, i)
	}
	fmt.Println(arr)

}
func base3() {
	i := 0
	for i < 5 {
		i++
		fmt.Println(i)
	}
}

func base2() {
	for false {
		fmt.Println("for false")
	}
	for true {
		time.Sleep(time.Second)
		fmt.Println("for true")
	}
}

func base() {
	for {
		time.Sleep(time.Second)
		fmt.Println("e")
	}
}

func base1() {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if j == 3 {
				return
			}
			fmt.Println(i, j)
		}
	}
	fmt.Println("do not exec it")
}
