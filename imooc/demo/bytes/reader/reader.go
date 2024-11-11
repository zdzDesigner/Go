package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"
)

func main() {

	// ImmRead()
	_read()
}

func _read() {
	rw := &RW{}
	buf := bytes.NewBuffer([]byte{3, 2, 1})
	buf.WriteTo(rw)
}

type RW struct{}

func (r *RW) Read(p []byte) (n int, err error) {
	return 0, errors.New("")
}
func (r *RW) Write(p []byte) (n int, err error) {
	fmt.Println("write args:", p)
	return 0, errors.New("")
}

func ImmRead() {
	reader := NewReader([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	go func() {
		fmt.Println("pp")
		time.Sleep(time.Second * 5)
		reader.WriteBytes([]byte{7, 8})
	}()
	fmt.Println(reader.ReadBytes(byte(4)))
	fmt.Println(reader.ReadBytes(byte(8)))
	fmt.Println(reader.ReadBytes(byte(8)))
}

// Reader ..
type Reader interface {
	ReadBytes(byte) ([]byte, error)
	WriteBytes([]byte) error
}

func (b *reader) ReadBytes(delim byte) (part []byte, err error) {
	for {
		time.Sleep(time.Millisecond * 1000)
		if i := bytes.IndexByte(b.buf[b.r:b.w], delim); i >= 1 {
			part = b.buf[b.r : b.r+i+1]
			b.r = b.r + i + 1
			break
		}

		if b.w-b.r >= len(b.buf) {
			break
		}

		if b.r >= 0 {
			copy(b.buf, b.buf[b.r:b.w])
			b.w = b.w - b.r
			b.r = 0
		}

		fmt.Println(b.buf, len(b.buf))

	}
	return
}

func (b *reader) WriteBytes(comein []byte) error {
	copy(b.buf[b.w:], comein)
	b.w = b.w + len(comein)
	return nil
}

// NewReader ..
func NewReader(bs []byte) Reader {

	return &reader{buf: bs, w: len(bs)}
}

type reader struct {
	buf []byte // 容器
	r   int    // 读取位置
	w   int    // 些入位置
}
