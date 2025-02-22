package main

import "fmt"

func main() {
	// base("aa---", "bb")
	isNil(nil)
	isNil([]string{"aa"})
}

func base(args ...string) {
	if len(args) > 1 && isempty(args[1]) {
		fmt.Println(args[1])
	}
	baseGo(args...)
}

func baseGo(args ...string) {
	fmt.Println(args, len(args))
}

func isempty(str string) bool {
	fmt.Println("str:", str)
	return str != ""
}

func isNil(keys []string) {
	toArg(keys...)
}

func toArg(keys ...string) {
	fmt.Println("keys:", keys)
}
