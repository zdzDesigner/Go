package pool

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

type st struct{}

func (s *st) Close() error {
	fmt.Println("close")
	return nil
}

type Pool struct {
	m       sync.Mutex                // 保证多个goroutine访问时候，closed的线程安全
	res     chan io.Closer            //连接存储的chan
	factory func() (io.Closer, error) //新建连接的工厂方法
	closed  bool                      //连接池关闭标志
}

//从资源池里获取一个资源
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r, ok := <-p.res:
		log.Println("Acquire:共享资源")
		if !ok {
			return nil, errors.New("ErrPoolClosed")
		}
		return r, nil
	default:
		log.Println("Acquire:新生成资源")
		return p.factory()
	}
}

//关闭资源池，释放资源
func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	//关闭通道，不让写入了
	close(p.res)

	//关闭通道里的资源
	for r := range p.res {
		r.Close()
	}
}

// 释放连接首先得有个前提，就是连接池还没有关闭
func (p *Pool) Release(r io.Closer) {
	//保证该操作和Close方法的操作是安全的
	p.m.Lock()
	defer p.m.Unlock()

	//资源池都关闭了，就剩下这一个没有释放的资源了，释放即可
	if p.closed {
		r.Close()
		return
	}

	select {
	case p.res <- r:
		log.Println("资源释放到池子里了")
	default:
		log.Println("资源池满了，释放这个资源吧")
		r.Close()
	}
}

func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("size的值太小了。")
	}
	return &Pool{
		factory: fn,
		res:     make(chan io.Closer, size),
	}, nil
}

func closer() (io.Closer, error) {
	return &st{}, nil
}

func implementation() {
	p, err := New(closer, 2)
	if err != nil {
		return
	}
	// 1
	// p.Release(&st{})
	// p.Release(&st{})
	// p.Close()

	// 2
	// p.Release(&st{}) // 塞入
	// p.Release(&st{}) // 塞入
	// p.Release(&st{}) // 塞入,满了

	// 3
	p.Release(&st{})         // 塞入
	p.Release(&st{})         // 塞入
	fmt.Println(p.Acquire()) // 取走
	fmt.Println(p.Acquire()) // 取走
	p.Release(&st{})         // 塞入
	p.Release(&st{})         // 塞入
}

func Entry() {
	implementation()
}
