package main

import (
	"fmt"
)

// 组合接口
type Computer interface {
	Memory()
	Run()
	Ctrl()
}

type ThinkPad struct {
	Computer
	name string
}

// func (tp *ThinkPad) Ctrl() {
// 	fmt.Println("ctrl")
// }

type Mac struct{}

func (m *Mac) Ctrl()   { fmt.Println("ctrl mac") }
func (m *Mac) Memory() { fmt.Println("Memory mac") }
func (m *Mac) Run()    { fmt.Println("Run mac") }

func main() {
	tp := ThinkPad{Computer: &Mac{}}
	tp.Ctrl()
}
