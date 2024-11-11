package main

import (
	"sync"
)

var m sync.Mutex

func main() {
	lock()
}

func lock() {
	m.Lock()
	defer func() {
		m.Unlock()
	}()
	panic("sss")
}
