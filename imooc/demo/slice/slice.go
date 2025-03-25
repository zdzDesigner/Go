package main

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"
)

func main() {
	// fmt.Println(len(getnil()), getnil(), "aa")
	base()
	// base1()
	// arrnil()
	// append2()
	// append_cap()
	// base3()
	// fmt.Println(base4())
	// base5()
	// base6()
	// base7()
	// base8()
	// join()
	// limit()
	// _copy()

	// boot()
	// change()

	// _range()
	// _split()

	// define()
	// removeProgress()
	// _continue()
	// lennil()
	readSouce()
}

func lennil() {
	fmt.Println(len(getnil()) == 0)
}

// /go1.21.3/src/io/io.go
// b = append(b, 0)[:len(b)]
func readSouce() {
	b := make([]byte, 0, 4)
	fmt.Printf("%p\n", b) // 0xc000114010
	b = append(b, 'a')
	b = append(b, 'a')
	b = append(b, 'a')
	b = append(b, 'a')
	fmt.Printf("%p\n", b) // 0xc000114010

	b = append(b, 0) // 超过cap自动扩容, b 指针新地址
	b = append(b, 0)
	fmt.Printf("%p\n", b) // 0xc000114018
	fmt.Println(b)        // [97 97 97 97 0 0]
	b = append(b, 0)[:4]  // 扩容后填充原来值
	fmt.Printf("%p\n", b) // 0xc000114018
	fmt.Println(b)        // [97 97 97 97]
}

func append2() {
	arr := []int{1, 2, 3}
	// crr := []int{11, 22, 33, 44, 55, 66}
	crr := []int{}
	zrr := make([]int, 0, len(arr)+len(crr))
	zrr = append(zrr, arr...)
	zrr = append(zrr, crr...)

	fmt.Println(zrr)
}

func define() {
	err, names := _define()
	fmt.Println(err, names, len(names))
	names = append(names, "aaa")
	fmt.Println(names, len(names))
}

func _define() (err error, names []string) {
	return
}

func _split() {
	str := ""
	arr := strings.Split(str, "&")
	fmt.Println(arr, len(arr))
	fmt.Println(arr[len(arr)-1] == "")
}

func _range() {
	arr := []int{1, 2, 3}
	fmt.Println(arr[:0])

	for _, item := range _rangeHandler() {
		fmt.Println(item)
	}
}

func _rangeHandler() []string {
	fmt.Println("_rangeHandler")
	return []string{"3m", "ss"}
}

func change() {
	arr := []int{1, 2, 3}
	crr := []int{11, 22, 33, 44, 55, 66}
	brr := arr
	crr = arr

	fmt.Println(arr, brr, crr)
	arr[0] = 4
	arr = append(arr, 5)
	fmt.Println(arr, brr, crr)

	arr = []int{5, 5, 5}
	go changeHandler(arr)
	time.Sleep(time.Second)
	arr = append(arr, 1)
	time.Sleep(time.Second)
	arr = append(arr, 1)
	time.Sleep(time.Second)
	arr = append(arr, 1)
}

func changeHandler(arr []int) {
	for {
		time.Sleep(time.Millisecond * 500)
		fmt.Println(arr)
	}
}

type WList struct {
	PIDS []string
}

var wlist WList

func boot() {
	fmt.Println("wlist.PIDS:", wlist.PIDS)
	v := []string{}
	fmt.Println(append(v, wlist.PIDS...))
	bootHandler(wlist.PIDS...)
	bootHandler2(wlist.PIDS)
	bootHandler3()
}

func bootHandler(arr ...string) {
	fmt.Println(arr, len(arr), "bootHandler")
	for _, v := range arr {
		fmt.Println(v)
	}
}

func bootHandler2(arr []string) {
	fmt.Println(arr, len(arr), "bootHandler2")
	for _, v := range arr {
		fmt.Println(v)
	}
}

func bootHandler3() {
	var arr []string
	fmt.Println("bootHandler3")
	for _, v := range arr {
		fmt.Println(v)
	}
}

func _copy() {
	arr := []string{"a", "b", "c", "1"}
	arr2 := []string{"D", "E", "F"}
	fmt.Println(copy(arr, arr2))

	fmt.Println(arr)
	fmt.Println(arr2)

	temp := []string{}
	fmt.Println(copy(temp, arr2))
	fmt.Println(temp)
	fmt.Println(arr2)

	pretemp := make([]string, 10)
	fmt.Println(copy(pretemp, arr2))
	fmt.Println(pretemp)
	fmt.Println(arr2)
}

func limit() {
	str := "2019-09-20"
	ymd := strings.Split(str, "-")
	fmt.Println(ymd[0], ymd[1], ymd[2], ymd[3])
}

func join() {
	arr := make([]string, 0, 10)
	fmt.Println(strings.Join(arr, ",") == "")
}

func arrnil() {
	var names []string = nil
	fmt.Println(len(names))
	fmt.Println(names == nil)
	fmt.Printf("%+v\n", names)
	fmt.Printf("%+v\n", nil)
	for _, name := range names {
		fmt.Println(name)
	}

	fmt.Println(getnil() == nil)
	for _, item := range getnil() {
		fmt.Println(item)
	}

	fmt.Println(getnil3() == nil)
	for _, item := range getnil3() {
		fmt.Println(item)
	}

	arr := []int{1, 2, 4}
	fmt.Println("append nil", append(arr, getnil3()...))
}

func getnil3() []int {
	return nil
}

func getnil() []interface{} {
	return nil
}

func getnil2() interface{} {
	return nil
}

type Interface sort.Interface

type IntSlice struct {
	Interface
}

func (p IntSlice) Less(i, j int) bool {
	return p.Interface.Less(j, i)
}

func sort1() {
	keys := []int{4, 2, 7, 9, 1, 5}
	v := sort.IntSlice(keys)
	v.Sort()
	fmt.Println(keys)
}

func sort2() {
	keys := []int{4, 2, 7, 9, 1, 5}
	sort.Ints(keys)
	fmt.Println(keys)
}

func sort3() {
	keys := []int{4, 2, 7, 9, 1, 5}
	is := &IntSlice{sort.IntSlice(keys)}
	fmt.Println(is.Len())
	sort.Sort(is)
	// sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	fmt.Println(keys)
}

func base8() {
	sort1()
	sort2()
	sort3()
}

// CompareFunc ..
type CompareFunc func(interface{}, interface{}) int

func indexOf(in interface{}, e interface{}, cmp CompareFunc) int {
	var (
		i   int
		ins = reflect.ValueOf(in)
		n   = ins.Len()
	)
	for ; i < n; i++ {
		if cmp(e, ins.Index(i).Interface()) == 0 {
			return i
		}
	}
	return -1
}

func base7() {
	arr := []int{2, 5, 7, 8, 1}
	val := indexOf(arr, 2, func(a, b interface{}) int {
		return a.(int) - b.(int)
	})
	fmt.Println(val)
}

func base6() {
	arr := make([]string, 0, 10)
	for i := 0; i < cap(arr); i++ {
		arr = append(arr, fmt.Sprintf("%dxx", i))
	}
	fmt.Println(arr)
	dels := []int{1, 3, 8}
	newarr := removes(arr, dels)
	fmt.Println(newarr)
}

func removes(arr []string, dels []int) []string {
	newarr := make([]string, 0, len(arr))
	for i, v := range dels {
		if v >= len(arr) {
			break
		}
		if i == 0 && len(dels) > 1 {
			newarr = append(newarr, arr[:v]...)
			continue
		}
		if len(dels) > 1 {
			newarr = append(newarr, arr[dels[i-1]+1:v]...)
		}
		if i == len(dels)-1 {
			newarr = append(newarr, arr[v+1:]...)
		}
	}
	return newarr
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func base5() {
	arr := make([]string, 0, 10)
	for i := 0; i < 3; i++ {
		for j, v := range arr {
			fmt.Println(j, v)
		}
		arr = append(arr, fmt.Sprintf("aa%d", i))
	}
}

func base4() (list []map[string]string) {
	list = append(list, map[string]string{
		"name": "zdz",
	})
	return list
}

// append:cap变化=>指针变化, 操作变更cap, append后len>cap ,则 cap=2*len
func append_cap() {
	fmt.Println("append_cap() ========================")
	a := []byte{2, 3, 4}

	fmt.Printf("%p,len:%d, cap:%d\n", a, len(a), cap(a)) // 0xc00001a190, len 3, cap 3
	a = append(a, 5)                                     // len 4, cap 8
	a = append(a, 6)                                     // len 5, cap 8
	fmt.Printf("%p,cap:%d\n", a, cap(a))                 // 0xc00001a198, [2 3 4 5] 8
}

func clone() {
	a := []string{"aa", "bb"}
	fmt.Println(a)

	b := a
	b[0] = "ccc"
	fmt.Println(a, b)
}

func clone_deep() {
	a := []string{"aa", "bb"}
	fmt.Println(a)

	b := make([]string, len(a))
	copy(b, a)
	b[0] = "ccc"
	fmt.Println(a, b)
}

func base() {
	// clone()
	// clone_deep()
	// return

	// TO::未初始
	var arr []string
	fmt.Println(len(arr), cap(arr)) // 0, 0
	fmt.Println(arr == nil)         // true
	// arr2 := make([]string, 0)
	// fmt.Println("arr2 first item::", arr2[0]) // panic: runtime error: index out of range [0] with length 0

	// TO::初始
	bs := []byte{2, 3, 4, 5, 6, 8}
	// [) 左闭右开
	fmt.Println(bs[0:])     // 2, 3, 4, 5, 6, 8
	fmt.Println(bs[0:][2:]) // 4, 5, 6, 8
	fmt.Println(bs[:4])     // 2, 3, 4, 5
	fmt.Println(bs[0:2])    // 2, 3
	fmt.Println(bs[1:4])    // 3, 4, 5

	fmt.Println("半清空:")
	bs = bs[:0]
	fmt.Println(bs, len(bs), cap(bs)) // [] 0 6
	fmt.Println("bs[0:] 非标明区间 内容清空:")
	fmt.Println(bs[0:]) // []
	fmt.Println("bs[0:6] 标明区间 内容还在:")
	fmt.Println(bs[0:6]) // 2, 3, 4, 5, 6, 8
	fmt.Println(bs[1:4]) // 3, 4, 5
}

func base1() {
	var arr []string
	fmt.Println("len(nil)", len(arr))

	r := make([]bool, math.MaxInt32)
	fmt.Println("Size: ", len(r))

	r2 := make([]byte, 2147483648)
	fmt.Println("Size: ", len(r2))

	fmt.Println(int(^uint(0) >> 1))
}

func _continue() {
	arr := []string{"aa", "bb", "cc"}
	for i := range arr {

		for _, item := range arr {
			fmt.Println(item)
			if item == "bb" {
				break
			}
		}

		fmt.Println(i)
	}
}

func removeProgress() {
	arr := []string{"aa", "bb", "cc", "dd"}
	fmt.Println(arr)
	for i, item := range arr {
		fmt.Println(item, i, len(arr))
		if i+1 > len(arr) {
			continue
		}

		arr = append(arr[:i], arr[i+1:]...)
	}
	fmt.Println(arr)
}
