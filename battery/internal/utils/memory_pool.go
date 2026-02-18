// Package utils provides utility structures and functions to support the MQTT broker,
// including memory management utilities to reduce garbage collection pressure
// and improve performance in high-throughput scenarios.
package utils

import (
	"sync"
)

// MemoryPool manages reusable byte buffers to reduce garbage collection pressure.
// In high-throughput MQTT systems, frequent allocation/deallocation of byte slices
// can cause performance degradation. This pool reuses buffers to minimize allocations.
type MemoryPool struct {
	// pool implements the underlying synchronization for buffer reuse using sync.Pool
	pool *sync.Pool
}

// NewMemoryPool creates a new memory pool with default buffer size.
// Initializes the sync.Pool with a New function that creates byte slices
// with zero length but 1KB capacity for efficient reuse.
//
// Returns:
//   - A pointer to the newly created MemoryPool instance
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Initial capacity of 1KB
			},
		},
	}
}

// Get returns a buffer from the pool, resetting its length to 0 while preserving capacity.
// This allows efficient reuse of allocated memory without requiring new allocations.
// After use, the buffer should be returned to the pool using Put.
//
// Returns:
//   - A byte slice with zero length but preserved capacity for reuse
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
