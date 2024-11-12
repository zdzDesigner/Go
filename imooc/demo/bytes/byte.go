package main

import (
	"bytes"
	"fmt"
)

/*
	bytes.NewBuffer()
	Buffer{
		buf []byte
		off int
		bootstrap [64]byte
	}
	func (*Buffer) Bytes() []byte
	func (*Buffer) String() string
	func (*Buffer) Len() int
*/
func main() {
	// base()
	// base2()
	// base3()
	// base4()
	// base5()
	// base6()
	// base7()
	// base8()
	mqtt()
}
func mqtt() {
	fmt.Println(string([]byte{123, 34, 109, 101, 116, 104, 111, 100, 34, 58, 34, 116, 104, 105, 110, 103, 46, 115, 101, 114, 118, 105, 99, 101, 46, 112, 114, 111, 112, 101, 114, 116, 121, 46, 115, 101, 116, 34, 44, 34, 105, 100, 34, 58, 34, 49, 48, 48, 55, 51, 55, 48, 51, 48, 49, 34, 44, 34, 112, 97, 114, 97, 109, 115, 34, 58, 123, 34, 80, 111, 119, 101, 114, 83, 119, 105, 116, 99, 104, 34, 58, 50, 50, 50, 125, 44, 34, 118, 101, 114, 115, 105, 111, 110, 34, 58, 34, 49, 46, 48, 46, 48, 34, 125}))
}

func base8() {
	a := []byte{1, 3}
	b := []byte{1, 2}
	fmt.Println(bytes.Compare(a, b))
}

func base7() {
	contain := []byte{3, 4, 5, 4}
	delim := byte(4)
	fmt.Println(bytes.IndexByte(contain, delim))
}

// 清空
func base6() {

	a := []byte{3, 4, 5}
	b := []byte{225, 224, 223, 222}
	bufa := bytes.NewBuffer(a)
	// 读完, m= bufa.Len(), off 滑动下标
	// m==0 && off!=0 => 触发 bufa.Reset() 重置
	bufa.Read(make([]byte, 3))
	fmt.Println(bufa.Bytes())
	// l+n = len(a) + len(b) = 7 > cap(a)
	bufa.Write(b)
	fmt.Println(bufa.Bytes())
}

func base5() {
	var a []int
	fmt.Println(a, a == nil)
	// [] true
	fmt.Println(len(a), cap(a))
	// 0 0

	// fmt.Println(a[:2])
	// panic: runtime error: slice bounds out of range

}

func base4() {
	b := make([]byte, 0, 100)
	b = append(b, 3)

	fmt.Println(b)
	// [3]
	fmt.Println(b[:20])
	// [3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]

	// fmt.Println(b[:200])
	// panic: runtime error: slice bounds out of range
}

func base3() {
	b := []byte{0, 250, 3, 10, 1, 2, 3, 4, 5}
	fmt.Println(b[4:])
	// [1 2 3 4 5]
	fmt.Println(b)
	// [0 250 3 10 1 2 3 4 5]
	fmt.Println(b[:0])
	// []
}

func base2() {
	b := []byte{0, 250, 3, 10, 1, 2, 3, 4, 5}
	bf := bytes.NewBuffer(b)
	fmt.Println(bf.Len())
	// 9
	nb := make([]byte, 2)
	n, err := bf.Read(nb)
	fmt.Println(bf.Bytes())
	// [3 10 1 2 3 4 5]
	fmt.Println(bf.Len())
	// 7
	fmt.Println(n, err)
	// 2 nil
	fmt.Println(bf.Next(3))
	// [3 10 1]
	fmt.Println(bf.Bytes())
	// [2 3 4 5]
	bf.Reset()
	fmt.Println(bf.Bytes())
	// []
}

func base() {
	b := []byte{0, 250, 3, 10}
	fmt.Println(b[:2])
	fmt.Println(b[1:2])
	fmt.Println(int(^uint(0) >> 1))
}

func base1() {
	b := []byte{0, 250, 3, 10}

	// b_buf := bytes.NewBuffer(b)

	// var x int32

	// binary.Read(b_buf, binary.BigEndian, &x)

	// fmt.Println(x)
	var total = 0
	for _, v := range b {
		fmt.Println(int(v))
		total += int(v)
	}
	fmt.Println(total)

}
