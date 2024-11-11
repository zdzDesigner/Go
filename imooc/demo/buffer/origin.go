package main

import (
	"errors"
	"fmt"
)

// Write
// 1. 查看cap够不够
// 2. 不够去自增长
// 3. 自增长:默认64, 赋值时 cap 会传染
// 4. 超过64 , 转半增算法
func main() {
	// base()
	base1()
}

func base() {
	var b []byte
	var bb = []byte{}
	fmt.Println(b == nil)
	fmt.Println(bb == nil)

	var bootstrap [10]byte
	fmt.Println(len(bootstrap), cap(bootstrap), bootstrap)
}

func base1() {
	bf := Buffer{}
	bf.WriteString("sss")
	fmt.Println(bf.buf, cap(bf.buf))
	bf.WriteString("bbb")
	fmt.Println(bf.buf, cap(bf.buf))
	// fmt.Println(bf.tryGrowByReslice(2))
}

// Buffer ..
type Buffer struct {
	buf       []byte
	off       int
	bootstrap [64]byte
	lastRead  readOp
}

var ErrTooLarge = errors.New("bytes.Buffer: too large")

type readOp int8

const (
	opInvalid readOp = 0 // Non-read operation.
	maxInt           = int(^uint(0) >> 1)
)

func (b *Buffer) Len() int { return len(b.buf) - b.off }
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.lastRead = opInvalid
}

func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	fmt.Println("cap(b.buf): ", cap(b.buf))
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		fmt.Println("b.buf:", b.buf)
		return l, true
	}
	return 0, false
}
func (b *Buffer) WriteString(s string) (n int, err error) {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	return copy(b.buf[m:], s), nil
}

func (b *Buffer) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	fmt.Println("b.buf == nil:", b.buf == nil)
	// Check if we can make use of bootstrap array.
	if b.buf == nil && n <= len(b.bootstrap) {
		b.buf = b.bootstrap[:n]
		fmt.Println("b.buf: ", b.buf)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}
