package ch

import (
	"fmt"
	"sync"
)

type ChStore[T any] struct {
	val  T
	sign chan struct{}
}

func (c *ChStore[T]) Set(val T) {
	c.val = val
	c.sign <- struct{}{}
}

func (c *ChStore[T]) Get() T {
	return c.val
}

func (c *ChStore[T]) Done() chan struct{} {
	return c.sign
}

// func NewChKV() *ChStore {
// 	chkv := &ChStore{val: make(map[string]string), sign: make(chan struct{})}
// 	return chkv
// }

// Storer ..
type Storer[K comparable, T any] interface {
	Add(key K, val T)                // 添加
	Valid(func(K, T) bool) bool // 检测有效
	Data() map[K]T              // 获取数据
	Done() <-chan map[K]T       // 完成
	Close()                          // 关闭store
}

func (ch *store[K, T]) Add(key K, val T) {
	ch.m.Lock()
	if !ch.closed {
		ch.data[key] = val
		fmt.Println(len(ch.data), ch.max)
		if len(ch.data) == int(ch.max) {
			ch.signal <- ch.data
		}
	}
	ch.m.Unlock()
}

func (ch *store[K, T]) Data() map[K]T {
	ch.m.Lock()
	ch.closed = true
	ch.m.Unlock()
	return ch.data
}

func (ch *store[K, T]) Done() <-chan map[K]T {
	return ch.signal
}

func (ch *store[K, T]) Close() {
	ch.m.Lock()
	ch.closed = true
	ch.data = nil
	close(ch.signal)
	ch.m.Unlock()
}

func (ch *store[K, T]) Valid(fn func(K, T) bool) bool {
	if len(ch.data) < int(ch.max) {
		return false
	}
	for k, v := range ch.data {
		if !fn(k, v) {
			return false
		}
	}
	return true
}

// NewChan ..
func NewChan[K comparable, T any](max int) Storer[K, T] {
	return &store[K, T]{
		max:    max,
		data:   make(map[K]T),
		signal: make(chan map[K]T),
	}
}

type store[K comparable,T any] struct {
	m      sync.Mutex
	max    int
	closed bool
	data   map[K]T
	signal chan map[K]T
}
