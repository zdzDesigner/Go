package errs

import (
	"fmt"
	"runtime"

	"battery/src/util/lang"
)

type Error int

var (
	USER_NOTFOUND    int = 44004
	USER_TOKEN_ERROR int = 40003
)

// ErrStack ..
func ErrStack(err interface{}) string {
	buf := make([]byte, 1<<11)
	buf = buf[:runtime.Stack(buf, false)]
	// fmt.Println("ErrStack:", string(buf))
	trace := lang.NewReg(string(buf)).
		Replace(`[\s\S]+runtime\/panic.go.+`, "").
		Replace(`.+gin-gonic\/gin.+|.+\/src|\+(.+)`, "").
		Replace(`\(0x.+\)`, "()").
		Val()
	errmsg := fmt.Sprintf("panic msg:%+v ; runtime.Stack:%s", err, trace[:300])
	return errmsg
}
