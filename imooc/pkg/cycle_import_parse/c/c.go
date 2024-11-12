package c

import "imooc/pkg/cycle_import_parse/f"

type C struct {
	Vc int
}

func New(i int) *C {
	return &C{
		Vc: i,
	}
}

func (c *C) Show() {
	f.Printf(c.Vc) // 软依赖通过分包解决
}
