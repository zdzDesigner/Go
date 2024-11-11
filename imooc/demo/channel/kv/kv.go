package main

import (
	"fmt"
	"time"
)

type Chkv struct {
	kv   map[string]string
	done chan struct{}
}

func main() {

	chkv := &Chkv{kv: make(map[string]string), done: make(chan struct{})}
	go source(chkv)
	sourceHandler(chkv)
	go source(chkv)
	sourceHandler(chkv)
	// time.Sleep(time.Second * 5)
}

func source(chkv *Chkv) {
	time.Sleep(time.Second * 3)
	chkv.kv["key"] = "va"
	chkv.done <- struct{}{}
}

func sourceHandler(chkv *Chkv) {
	select {
	case <-chkv.done:
		fmt.Println(chkv.kv["key"])
	}
	fmt.Println("000000")
}
