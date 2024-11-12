package c

import "imooc/pkg/cycle_import/a"

type C struct {
	Vc int
}

func New(i int) *C {
	return &C{
		Vc: i,
	}
}

func (c *C) Show() {
	a.Printf(c.Vc)
}
