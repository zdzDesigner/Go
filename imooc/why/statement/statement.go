package main

import (
	"fmt"
)

func main() {
	fmt.Println(base3())
	fmt.Println(base4())

}
func base4() (data []string) {
	data = append(data, "ss")
	return data
}

func base3() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["aa"] = "aa"
	return data
}
