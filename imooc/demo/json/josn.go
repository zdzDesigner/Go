package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// V ..
type V struct {
	Ab  string `json:"aa.bb"`
	cd  bool
	Dd  bool     `json:"dd"`
	Arr []string `json:"arr"`
	FF  *int     `json:"ff,string"` // 从string解析
}

func main() {
	// base()
	// base2()
	// marshal()
	unmap()
	// less()
}

func marshal() {

	strs := []string{"aa", "bb"}
	v, _ := json.Marshal(strs)
	fmt.Println("v:", v, string(v))
	strs1 := []string{"aa", "bb"}
	v1, _ := json.Marshal(strs1)
	fmt.Println("v1:", v1, string(v1))
	fmt.Println(string(v) == string(v1))
}

func base() {
	str := `{"aa.bb":"a•aa","cd":true,"ff":"43"}`
	var v V
	json.Unmarshal([]byte(str), &v)
	fmt.Println(v, *v.FF)
	fmt.Println(reflect.TypeOf(*v.FF))
	fmt.Println(v.Arr, len(v.Arr))

}
func base2() {
	str := `{"aa.bb":"a•aa","cd":true,"ff":43}`
	var v V
	json.Unmarshal([]byte(str), &v)
	fmt.Println(v, v.FF)
	fmt.Println(v.Arr, len(v.Arr))

}

func less() {
	lessUtil1()
	lessUtil2()

}
func lessUtil1() {
	str := `{"aa.bb":"a•aa","cd":true,"ff":"43"}`
	var v V
	json.Unmarshal([]byte(str), &v)
	fmt.Println(v, *v.FF)

}
func lessUtil2() {
	str := `{"aa.bb":"a•aa","cd":true}`
	var v V
	json.Unmarshal([]byte(str), &v)
	fmt.Println(v, v.FF)
	// fmt.Println(v.FF == 43)

}

func unmap() {
	str := `{"cmd":-1}`
	var v map[string]int
	json.Unmarshal([]byte(str), &v)
	fmt.Println(v["cmd"])

	str1 := `{"err_code": -1,"msg": "decode err"}`
	var v1 map[string]any
	json.Unmarshal([]byte(str1), &v1)
	fmt.Println(v1)

}
