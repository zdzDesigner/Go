package b

import "imooc/pkg/cycle_import_parse/c"

type generatec interface { // 硬依赖通过接口解决
	GetC() *c.C
}

type B struct {
	Pa generatec
}

func New(ar generatec) *B {
	return &B{
		Pa: ar,
	}
}

func (b *B) DisplayC() {
	b.Pa.GetC().Show()
}
