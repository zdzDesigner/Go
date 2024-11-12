package main

import (
	"fmt"
	"io"
	"time"
)

func main() {
	// demo1()
	// demo2()
	demo3()
}

type st struct{}

func (s *st) Close() error {
	fmt.Println("close")
	return nil
}

func demo1() {
	ch := make(chan io.Closer)

	go func() {
		time.Sleep(time.Second * 3)
		ch <- &st{}
	}()

	select {
	case c := <-ch:
		fmt.Println("ppp", c)
		c.Close()
	case <-time.After(time.Second * 10):
		fmt.Println("time.After")
	}
}

func demo3() {
	ch := make(chan io.Closer, 2)
	c := make(chan io.Closer)

	go func() {
		time.Sleep(time.Second * 1)
		c <- &st{}
		time.Sleep(time.Second * 1)
		c <- &st{}
		time.Sleep(time.Second * 1)
		c <- &st{}
	}()
	for {
		select {
		case ch <- <-c:
			fmt.Println("ppp", len(ch))
		default:
			fmt.Println("default")
			close(c)
		}
	}
}

func demo2() {
	ch := make(chan io.Closer)

	go func() {
		fmt.Println("+++++1")
		time.Sleep(time.Second * 1)
		ch <- &st{}
		fmt.Println("+++++2")
		time.Sleep(time.Second * 1)
		ch <- &st{}
		fmt.Println("+++++3")
		time.Sleep(time.Second * 1)
		ch <- &st{}
		fmt.Println("+++++4")
		time.Sleep(time.Second * 1)
		ch <- &st{}
	}()

	go func() {
		time.Sleep(time.Second * 4)
		close(ch)
	}()

	for c := range ch {
		fmt.Println(c)
		c.Close()
	}
}
