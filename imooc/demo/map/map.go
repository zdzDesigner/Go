package main

import (
	"fmt"
)

type V struct {
	status bool
}

func main() {
	// base()
	// base1()
	// base2()
	// fmt.Println(base3())
	// fmt.Println(base4())
	// base5()
	// base6()
	// base7()
	// base8()
	// args()
	// _delete()
	// link()
	// isnil()
	// noInit()
	channel()
}
func channel() {

	v := make(map[int]chan string)
	fmt.Println(v[1] == nil)
}

func isnil() {
	var v1 map[string]interface{}
	fmt.Println(v1 == nil)

	v := map[string]interface{}{}
	fmt.Println(v == nil)

	fmt.Printf("m:%+v", []string{"ii", "vv"})
	fmt.Printf("m:%+v", nil)
}

func link() {
	v := V{status: false}
	if v.status && linkHandler() {
		fmt.Println("true 11")
	}

	var vv V
	if !vv.status {
		fmt.Println("true")
	}
}

func linkHandler() bool {
	fmt.Println("===linkHandler===")
	return true
}

func _delete() {
	obj := map[string]string{"a": "aaa"}
	fmt.Println(obj)
	delete(obj, "a")
	delete(obj, "ddd")
	fmt.Println(len(obj))
	fmt.Println(obj)

	memo := map[string]string{"malloc": "1", "mmap": "3"}
	for k, v := range memo {
		fmt.Println(k, v)
		delete(memo, k)
	}
	fmt.Println(memo, len(memo))
}

func args() {
	obj := map[string]string{"a": "aaa"}
	argsHandler(obj)
	fmt.Println(obj)
	argsHandler(make(map[string]string))
	fmt.Println(obj)
}

func argsHandler(obj map[string]string) {
	obj["b"] = "bbb"
}

func base8() {
	obj := make([]string, 0)
	fmt.Println(obj)
	fmt.Println(len(obj))

	var obj1 []string
	fmt.Println(obj1)
	fmt.Println(len(obj1))
	obj1 = append(obj1, "cc")
	fmt.Println(obj1)
}

func base7() {
	var (
		key = "aa"
		val = "bb"
	)
	data := map[string]string{
		key: val,
	}
	fmt.Println(data)
	fmt.Println(data[key])
	fmt.Println(data["aa"])
	fmt.Println(len(data))
}

func base6() {
	data := map[int]interface{}{
		1: "zdz",
		2: "lmy",
	}

	fmt.Println(data)
	fmt.Println(data[3], data[3] == nil)
}

func base5() {
	data := map[int]string{
		1: "zdz",
		2: "lmy",
	}

	fmt.Println(data)
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

func base2() {
	type zdz struct {
		name string
	}
	val := map[string]struct{}{}
	// val["a"] = zdz{} // cannot use zdz literal (type zdz) as type struct {} in assignmentgo

	fmt.Println(val)
}

func base1() {
	counts := map[int]int{
		0:      401100,
		100300: 401300,
		100301: 401301,
	}
	fmt.Println(counts[9])
	fmt.Println(len(counts))
}

func base() {

	build := make(map[string]string, 2)
	fmt.Println("undefined::", build["1"] == "")
	build["1"] = "11"
	build["2"] = "11"
	build["3"] = "11"
	build["4"] = "11"
	fmt.Println(build)

	alias := map[string]string{
		"jp": "ja", // 日文
		"kr": "ko", // 韩文
	}

	newalias := make(map[string]string, len(alias))
	for k, v := range alias {
		newalias[v] = k
	}
	alias = newalias
	fmt.Println(alias)

	var v map[string]string
	fmt.Println("-----------", len(v), v["aa"], v["aa"] == "")

	v1 := map[string]string{}
	fmt.Println(len(v1))
}

func noInit() {
	var obj map[string]string
	// var obj = map[string]string{}
	// var obj = map[string]string{"name": "aaa"}
	fmt.Println(obj["name"], len(obj))
	if name, ok := obj["name"]; ok {
		fmt.Println("name:", name)
	}
}
