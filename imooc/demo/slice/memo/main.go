package main

import (
	"fmt"
	"time"
)

var ARR []int

func main() {
	// arr := make([]int, 0, 10)
	arr := []int{1, 2, 3}
	arr = append(arr, 4)
	fmt.Printf("\t %p, %d, %d , %+v, %+v\n", &arr, len(arr), cap(arr), arr, arr[0:cap(arr)])
	for i := 0; i < len(arr); i++ {
		fmt.Printf("%p\n", &arr[i])
	}

	arr = arg(arr)
	fmt.Printf("\t %p, %d, %d , %+v, %+v\n", &arr, len(arr), cap(arr), arr, arr[0:cap(arr)])
	arr = append(arr, 5)

	time.Sleep(time.Second * 2)
	fmt.Printf("\t %p, %d, %d , %+v, %+v\n", &ARR, len(ARR), cap(ARR), ARR, ARR[0:cap(ARR)])
	for i := 0; i < len(ARR); i++ {
		fmt.Printf("%p\n", &ARR[i])
	}
	fmt.Printf("\t %p, %d, %d , %+v, %+v\n", &arr, len(arr), cap(arr), arr, arr[0:cap(arr)])
}

func arg(arr []int) []int {
	// arr[0] = 999
	arr = append(arr, 100)
	fmt.Printf("\t \t %p, %d, %d, %+v, %+v\n", &arr, len(arr), cap(arr), arr, arr[0:cap(arr)])
	for i := 0; i < len(arr); i++ {
		fmt.Printf("%p\n", &arr[i])
	}
	go func() {
		/* code */
		time.Sleep(time.Second * 1)
		fmt.Printf("\t \t %p, %d, %d, %+v, %+v\n", &arr, len(arr), cap(arr), arr, arr[0:cap(arr)])
		for i := 0; i < len(arr); i++ {
			fmt.Printf("%p\n", &arr[i])
		}
	}()
	ARR = arr
	return arr
}

func onebyone() {
	// 地址不变
	arr := []int{}
	for i := 0; i < 10; i++ {
		arr = append(arr, i)
		fmt.Printf("%p, %d, %d\n", &arr, len(arr), cap(arr))
	}
	fmt.Println(arr)
}
