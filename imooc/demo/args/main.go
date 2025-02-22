package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	args := os.Args
	fmt.Println(args)
	// tpl = args[]
	arg := parseArgs(args[1:])
	fmt.Println(arg)

	pointer()
	structer()
}

type Arg struct {
	TplPath string // 模板路径
	ValPath string // 模板中的值对应路径
}

func parseArgs(args []string) Arg {
	return Arg{
		TplPath: argVal(strings.Join(args, " "), "tpl"),
		ValPath: "",
	}
}

func argVal(source string, key string) string {
	reg := regexp.MustCompile(fmt.Sprintf(`\s*%s=(\w+)`, key))
	val := reg.FindAllStringSubmatch(source, -1)
	if val == nil {
		return ""
	}
	if len(val[0]) == 0 {
		return ""
	}
	return val[0][1]
}

func update(names []string) {
	names[0] = "xxx"
}

func pointer() {
	names := []string{"ccc"}
	update(names)

	time.Sleep(time.Second)

	fmt.Println(names)
}

type Val struct {
	Name string
}

func updateStruct(val Val) {
	val.Name = "ccc"
}

func structer() {
	val := Val{Name: "aaa"}
	updateStruct(val)
	time.Sleep(time.Second)
	fmt.Println(val.Name)
}

