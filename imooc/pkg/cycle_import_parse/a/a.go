package a

import (
	"imooc/pkg/cycle_import_parse/b"
	"imooc/pkg/cycle_import_parse/c"
)

type Generatec interface { // 硬依赖通过接口解决
	GetC() *c.C
}

type A struct {
	Pb *b.B
	Pc *c.C
}

func (a *A) GetC() *c.C {
	return a.Pc
}

func New(ic int) Generatec {
	a := &A{
		Pc: c.New(ic),
	}

	a.Pb = b.New(a)

	return a
}
