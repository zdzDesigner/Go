package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

func main() {
	// Rand2()
	Shuffle()
}

func Shuffle() {
	// 创建一个切片
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 使用 Shuffle 函数对切片进行随机重排
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	fmt.Println(slice)
}

func Rand() {
	pool := []int{}
	newpool := []int{}
	length := 4
	for i := 0; i < length; i++ {
		pool = append(pool, i)
	}

	for i := 0; i < length; i++ {
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(len(pool))
		newpool = append(newpool, pool[r])
		pool = append(pool[:r], pool[r+1:]...)
	}
	// fmt.Println(newpool)
}

func Rand2() {
	pool := []int{}
	count := 4

	for {
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(count)
		if IndexOf(pool, r) == -1 {
			pool = append(pool, r)
		}
		if len(pool) == count {
			break
		}
	}
	fmt.Println(pool)
}

// IndexOf ..
func IndexOf(in interface{}, elem interface{}) int {
	inValue := reflect.ValueOf(in)
	elemValue := reflect.ValueOf(elem)

	t := inValue.Type().Kind()

	if t == reflect.String {
		return strings.Index(inValue.String(), elemValue.String())
	}

	if t == reflect.Slice {
		for i := 0; i < inValue.Len(); i++ {
			if equal(inValue.Index(i).Interface(), elem) {
				return i
			}
		}
	}

	return -1
}

func equal(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	return reflect.DeepEqual(expected, actual)

}
