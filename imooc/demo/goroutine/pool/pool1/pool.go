package main

import "sync"

func main() {

}

// GoroutinePool pool
type GoroutinePool struct {
	c  chan struct{}
	wg *sync.WaitGroup
}

// NewGoroutinePool 采用有缓冲channel实现,当channel满的时候阻塞
func NewGoroutinePool(maxSize uint) *GoroutinePool {
	if maxSize <= 0 {
		panic("max size too small")
	}
	return &GoroutinePool{
		c:  make(chan struct{}, maxSize),
		wg: new(sync.WaitGroup),
	}
}

// Add ..
func (g *GoroutinePool) Add(delta int) {
	g.wg.Add(delta)
	for i := 0; i < delta; i++ {
		g.c <- struct{}{}
	}

}

// Done ..
func (g *GoroutinePool) Done() {
	<-g.c
	g.wg.Done()
}

// Wait ..
func (g *GoroutinePool) Wait() {
	g.wg.Wait()
}
