package a

import (
	"fmt"
	"imooc/pkg/cycle_import/b"
	"imooc/pkg/cycle_import/c"
)

type A struct {
	Pb *b.B
	Pc *c.C
}

func New(ic int) *A {
	a := &A{
		Pc: c.New(ic),
	}

	a.Pb = b.New(a)

	return a
}

func Printf(v int) {
	fmt.Printf("%v", v)
}
