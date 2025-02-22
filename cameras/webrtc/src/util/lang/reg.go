package lang

import "regexp"

type Reger interface {
	Replace(reg, rep string) *Reg
	Val() string
}

func NewReg(str string) Reger { return &Reg{str} }

type Reg struct{ str string }

func (r *Reg) Val() string { return r.str }
func (r *Reg) Replace(reg, rep string) *Reg {
	r.str = regexp.MustCompile(reg).ReplaceAllString(r.str, rep)
	return r
}
