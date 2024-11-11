package main

import (
	"fmt"
	"regexp"
)

type Reger interface {
	Replace(reg, rep string) *Reg
	Val() string
}

func (r *Reg) Val() string {
	return r.str
}

func (r *Reg) Replace(reg, rep string) *Reg {
	r.str = regexp.MustCompile(reg).ReplaceAllString(r.str, rep)
	return r
}

func NewReg(str string) Reger {
	return &Reg{str}
}

type Reg struct {
	str string
}

func main() {
	str := strGener()
	fmt.Println(NewReg(str).Replace(`[\s\S]+runtime\/panic.go.+`, "").
		Replace(`.+gin-gonic\/gin.+|.+\/src|\+(.+)`, "").
		Replace(`\(.+\)`, "()").
		Val())
}

func strGener() string {
	return `goroutine 146 [running]:
	webhook/src/util/helper/errs.ErrStack(0xe552c0, 0x1853bb0, 0x1853bb0, 0x0)
		/home/zdz/Documents/Webhook/webhook/src/util/helper/errs/recover.go:25 +0x82
	webhook/src/middleware/cerror.Entry.func1(0xc023a31554e5dcd8, 0x52c1571b, 0x186cdc0, 0xc0003f2b00, 0xc000480070, 0xf, 0xc000390000, 0x4, 0xc000390005, 0x18, ...)
		/home/zdz/Documents/Webhook/webhook/src/middleware/cerror/index.go:87 +0x961
	panic(0xe552c0, 0x1853bb0)
		/home/zdz/Application/Go/go1.11.4/src/runtime/panic.go:513 +0x1b9
	webhook/src/routes/movienews.imgConv(0x10d6200, 0xc0001f0a00, 0xc00047d800, 0x5d, 0xc0003ce6c0, 0x0)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/mixin.go:109 +0x2e2
	webhook/src/routes/movienews.Search(0x10e6300, 0xc0001f09c0)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/search.go:35 +0x1ea
	webhook/src/util/wh.IHander.func1(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/util/wh/ctx.go:73 +0x67
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/routes/movienews.glob..func2.1(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/search_test.go:26 +0x1c4
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.WListRule(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/wlist.go:33 +0x53
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.WListContainer(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/wlist.go:68 +0x2ef
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.Dmprotl(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/dmprotl.go:27 +0x39
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/h`
}
