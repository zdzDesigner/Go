package main

import "fmt"

func main() {
	base("aa---", "bb")
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
	if str != "" {
		return true
	}
	return false
}
