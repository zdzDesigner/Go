// Package utils 提供实用程序结构和函数来支持MQTT代理，
// 包括内存管理实用程序，以减少垃圾回收压力
// 并在高吞吐量场景中提高性能。
package utils

import (
	"sync"
)

// MemoryPool 管理可重用字节缓冲区以减少垃圾回收压力。
// 在高吞吐量MQTT系统中，频繁分配/释放字节切片
// 可能导致性能下降。此池重用缓冲区以最小化分配。
type MemoryPool struct {
	// pool 使用sync.Pool实现缓冲区重用的底层同步
	pool *sync.Pool
}

// NewMemoryPool 创建一个具有默认缓冲区大小的新内存池。
// 使用New函数初始化sync.Pool，该函数创建长度为零但容量为1KB的字节切片
// 以便高效重用。
//
// 返回值：
//   - 指向新创建的MemoryPool实例的指针
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Initial capacity of 1KB
			},
		},
	}
}

// Get 从池中返回一个缓冲区，在保留容量的同时将其长度重置为0。
// 这样可以在不需要新分配的情况下高效重用已分配的内存。
// 使用后，应使用Put将缓冲区返回到池中。
//
// 返回值：
//   - 一个长度为零但保留容量以供重用的字节切片
func (mp *MemoryPool) Get() []byte {
	buf := mp.pool.Get().([]byte)
	return buf[:0] // Reset length to 0 while keeping capacity
}

// Put returns a buffer to the pool
func (mp *MemoryPool) Put(buf []byte) {
	mp.pool.Put(buf[:0]) // Keep capacity but reset length
}

// SizedBuffer wraps a byte slice and provides Reset functionality
type SizedBuffer struct {
	data []byte
	pool *MemoryPool
}

// NewSizedBuffer creates a buffer from the pool
func (mp *MemoryPool) NewSizedBuffer() *SizedBuffer {
	return &SizedBuffer{
		data: mp.Get(),
		pool: mp,
	}
}

// Append adds data to the buffer
func (sb *SizedBuffer) Append(data []byte) {
	sb.data = append(sb.data, data...)
}

// Bytes returns the underlying byte slice
func (sb *SizedBuffer) Bytes() []byte {
	return sb.data
}

// Reset clears the buffer to 0 length but retains capacity
func (sb *SizedBuffer) Reset() {
	sb.data = sb.data[:0]
}

// Release returns the buffer to the pool
func (sb *SizedBuffer) Release() {
	sb.pool.Put(sb.data)
	sb.data = nil
	sb.pool = nil
}
