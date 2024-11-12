package question

import (
	"fmt"
	"time"
)

func Slice() {
	arr := []int{1}
	go sliceHandler(&arr)
	time.Sleep(time.Second)
	arr = append(arr, 2)
	time.Sleep(time.Second)
	arr = append(arr, 2)
	time.Sleep(time.Second)
	arr = append(arr, 2)
	fmt.Println(arr)
}

func sliceHandler(arr *[]int) {
	for {
		fmt.Println(arr, len(*arr), cap(*arr))
		time.Sleep(time.Millisecond * 500)
	}
}
