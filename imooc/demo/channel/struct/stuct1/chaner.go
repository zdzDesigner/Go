package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/*
未初始的 chan
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [select (no cases)]:

goroutine 18 [chan send (nil chan)]:
*/

func main() {
	// base()
	base3()
	// base4()
}

func base4() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go func() {
		time.Sleep(time.Second * 2)
		context.WithValue(ctx, "1", "1")
		cancel()
	}()

	go func() {
		time.Sleep(time.Second * 1)
		context.WithValue(ctx, "2", "1")
	}()

	select {
	case <-ctx.Done():
		fmt.Println("r1", ctx.Value("1"))
	}
}

func base3() {

	ch := NewChan(2)
	defer ch.Close()

	go func() {
		time.Sleep(time.Second * 2)
		ch.Add("1", "1")
	}()

	go func() {
		time.Sleep(time.Second * 3)
		ch.Add("2", "2")
	}()

	select {
	case val := <-ch.Done():
		fmt.Println("r1", val)
	case <-time.After(time.Second * 2):
		fmt.Println("time.After", ch.GetData())
	}
}

// Chaner ..
type Chaner interface {
	Add(key string, val interface{})           // 添加
	Done() <-chan map[string]interface{}       // 完成
	Valid(func(string, interface{}) bool) bool // 检测有效
	Close()                                    // 关闭channel
	GetData() map[string]interface{}           // 获取数据
}

func (ch *channel) Add(key string, val interface{}) {
	// if val != "" {}
	ch.mu.Lock()
	if !ch.closed {
		ch.data[key] = val
		if len(ch.data) == int(ch.max) {
			ch.signal <- ch.data
		}
	}
	ch.mu.Unlock()

}

func (ch *channel) GetData() map[string]interface{} {
	ch.mu.Lock()
	ch.closed = true
	ch.mu.Unlock()
	return ch.data
}

func (ch *channel) Done() <-chan map[string]interface{} {
	return ch.signal
}

func (ch *channel) Close() {
	ch.mu.Lock()
	ch.closed = true
	ch.data = nil
	close(ch.signal)
	ch.mu.Unlock()
}

func (ch *channel) Valid(fn func(string, interface{}) bool) bool {
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
func NewChan(max uint8) Chaner {
	return &channel{
		max:    max,
		data:   make(map[string]interface{}),
		signal: make(chan map[string]interface{}),
	}
}

type channel struct {
	mu     sync.Mutex
	max    uint8
	closed bool
	data   map[string]interface{}
	signal chan map[string]interface{}
}
