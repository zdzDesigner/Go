package main

import (
	"fmt"
	"reflect"
)

type Num struct {
	i string
	j int64
	N int
}

func main() {
	n := Num{i: "EDDYCJY", j: 1, N: 4}
	// nPointer := unsafe.Pointer(&n)
	// niPointer := (*string)(unsafe.Pointer(nPointer))
	// *niPointer = "煎鱼"
	// njPointer := (*int8)(unsafe.Pointer(uintptr(nPointer) + unsafe.Offsetof(n.j)))
	// *njPointer = 2
	// fmt.Printf("n.i: %s, n.j: %d\n", n.i, n.j)

	rt := reflect.TypeOf(n)
	rv := reflect.ValueOf(n)

	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		v := rv.Field(i)
		fmt.Println(t.Type, t.Name, t.Tag, v.CanInterface())
		if v.CanInterface() {
			fmt.Println(v.Interface())
		}
	}
}
