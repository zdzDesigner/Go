package main

import (
	"fmt"
	"net/url"
)

func main() {
	// base()
	// base1()
	base2()
}

func base() {
	up, _ := url.Parse("http://kong.dui.ai/apis/movienews/commingMovie?productId=userId123&deviceId=ec7d7f63d100a4aec0&a=")
	fmt.Println(up)
	fmt.Println(up.Query())
}

func base1() {
	up, _ := url.ParseQuery("http://kong.dui.ai/apis/movienews/commingMovie?productId=userId123&deviceId=ec7d7f63d100a4aec0")
	fmt.Println(up)
}

func base2() {
	fmt.Println(url.PathEscape("a"))
	fmt.Println(url.PathEscape("你"))
}
