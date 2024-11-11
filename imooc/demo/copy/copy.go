package main

import (
	"fmt"
)

// copy 有坑占坑(坑是len 长度), 无坑忽略; 返回占用的坑数
// 赋值时cap会传染
func main() {
	// base()
	// base1()
	// base2()
	// base31()
	// base32()
	// base4()
	// base5()
	// base6()
	// marge()
	justCopy()
	// clear()

}

func clear() {
	arr := []byte{'a', 'b'}
	fmt.Println(arr[1:1])
	copy(arr, arr[1:1])
	fmt.Println(arr, len(arr))
}

func justCopy() {
	arr := []string{"aa", "bb"}
	arr2 := []string{"cc", "dd", "e", "f"}
	copy(arr, arr2)
	fmt.Println(arr, len(arr), cap(arr))
	fmt.Println(arr2, len(arr2), cap(arr2))
}

func marge() {
	arr := []string{"aa", "bb"}
	arr2 := []string{"cc", "dd", "e", "f"}
	pool := make([]string, len(arr)+len(arr2)) // new
	copy(pool, arr)                            // merge arr
	copy(pool[len(arr):], arr2)                // merge arr2
	fmt.Println(pool)
}

func base5() {
	var (
		a = make([]byte, 0, 20)
		b = make([]byte, 0, 10)
		c = make([]byte, 2)
	)
	a = append(a, 3, 4, 5, 6, 7)
	b = append(b, 225, 224, 223, 222, 221)
	b = append(b, 225, 224, 223, 222, 221)
	fmt.Println("b:", b) // b: [225 224 223 222 221 225 224 223 222 221]

	fmt.Println("a:", a) // a: [3 4 5 6 7]
	copy(a[2:], b)
	fmt.Println("a:", a) // a: [3 4 225 224 223]
	// copy
	n := copy(c, b)
	fmt.Println(c, n) // [225 224] 2

}

func base4() {
	var (
		a = make([]byte, 0, 20)
	)
	a = append(a, 3, 4, 5, 6, 7)
	b := a[:] // 赋值后 cap 大小传染
	fmt.Println(b, len(b), cap(b))
}

// 模拟 tryGrowByReslice
func base31() {
	var (
		a = make([]byte, 0, 20)
		b = make([]byte, 0, 10)
	)
	a = append(a, 3, 4, 5, 6, 7)
	b = append(b, 225, 224, 223, 222, 221)

	fmt.Println(a[:15])
	// [3 4 5 6 7 0 0 0 0 0 0 0 0 0 0]

	// off 3 => a[3:]
	// tryGrowByReslice => a: [3, 4, 5, 6, 7, 0, 0, 0, 0, 0]
	// copy(a[5:],b)
	a = a[:len(a)+len(b)] // tryGrowByReslice
	fmt.Println(copy(a[5:], b))
	fmt.Println(a)
}

// 模拟grow n <= c/2-m
func base32() {
	var (
		a = make([]byte, 0, 20)
		b = make([]byte, 0, 10)
	)
	a = append(a, 3, 4, 5, 6, 7)
	b = append(b, 225, 224, 223, 222, 221, 220, 219)

	// off 3 => a[3:]
	// copy(b.buf, b.buf[b.off:]) => a:[6, 7, 5, 6, 7]
	// b.off = 0, b.buf = b.buf[:m+n] => a:[6, 7, 5, 6, 7, 0, 0, 0, 0, 0]
	// m = len(a) => 5

	a = a[:len(a)+len(b)]       // tryGrowByReslice
	fmt.Println(copy(a[5:], b)) // 7
	fmt.Println(a)              // [3 4 5 6 7 225 224 223 222 221 220 219]

}

func base2() {
	a := make([]byte, 0, 20)
	a = append(a, 3, 4, 5, 6, 7)

	fmt.Println(a, len(a), cap(a))
	// [3 4 5  -  6 7] 5 20
	copy(a, a[3:])
	fmt.Println(a, cap(a))
	// [6 7  -  5 6 7] 20
}

// 无坑忽略
func base1() {
	var b []byte
	a := make([]byte, 20)
	a[4] = 44
	copy(b, a) // b当前无坑,copy无效

	fmt.Println(a)
	// [0 0 0 0 44 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	fmt.Println(b)
	// []

}

// 赋值时cap会传染
// copy 有坑占坑, 返回占用的坑数
func base() {
	a := [20]byte{}
	b := []byte{}
	fmt.Println(len(a), cap(a))
	fmt.Println(len(b), cap(b))
	b = a[:4]                       // a的cap传给了b
	fmt.Println(len(b), cap(b))     // 4 20
	b[3] = 4                        // 传指针， a也改变了
	fmt.Println(a)                  // [0 0 0 4 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	fmt.Println(b)                  // [0 0 0 4]
	fmt.Println("cap(b): ", cap(b)) // cap(b):  20
	fmt.Println(copy(b, "aaaaaaa")) // 4
	fmt.Println(b, len(b))          // [97 97 97 97] 4

	fmt.Println(copy(b, []byte{9, 8, 7, 6, 5, 4}))
	// 4
	fmt.Println(b)
	// [9 8 7 6]

	// 赋值时传指针
	fmt.Println(a)
	// [9 8 7 6 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}
