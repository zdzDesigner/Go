package utils

import (
	"sync"
)

// MemoryPool manages reusable buffers to reduce GC pressure
type MemoryPool struct {
	pool *sync.Pool
}

// NewMemoryPool creates a new memory pool with default buffer size
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Initial capacity of 1KB
			},
		},
	}
}

// Get returns a buffer from the pool
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
