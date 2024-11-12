package main

import (
	"fmt"
	"reflect"
	"strconv"
)

type Smart struct {
	Name string `json:"name" sql:"sql-name"`
	Size int    `json:"size" sql:"sql-size"`
}

func main() {
	// base()
	// _reflect()
	var v any = 43
	vv, _ := v.(int)
	fmt.Println(strconv.Itoa(vv))
	fmt.Println(fmt.Sprintf("code:%d", v))
}

func base() {
	fmt.Println("reflect")
	smart := &Smart{Name: "chuizi"}
	fmt.Printf("smart: %p\n", smart)
	fmt.Printf("smart.Name: %p\n", &smart.Name)
	fmt.Printf("smart.Size: %p\n", &smart.Size)
	typeFunc(smart)
}

func typeFunc(elemt interface{}) {
	value := reflect.ValueOf(elemt)

	if value.Kind() == reflect.Ptr {
		fmt.Println(value.Elem())
		fmt.Println(value.Elem().Addr())
		fmt.Println(value.Elem().Addr().Interface())
		fmt.Printf("smart: %p\n", value.Elem().Addr().Interface())
		fmt.Printf("smart.Name: %p\n", value.Elem().Field(0).Addr().Interface())
		fmt.Printf("smart.Size: %p\n", value.Elem().Field(1).Addr().Interface())

	}

}

func _reflect() {

	reflectParse(&Smart{Name: "zdz"})
	// getFieldName(&Smart{Name: "zdz"})
	// getFieldName2(Smart{Name: "zdz"})
	// getFieldName3(&Smart{Name: "zdz"})
	// getFieldName3(Smart{Name: "zdz"})
}

func reflectParse(s *Smart) {
	rt := reflect.TypeOf(*s)
	rv := reflect.ValueOf(*s)
	rv2 := reflect.ValueOf(s)

	fmt.Println("rt:", rt.Kind(), int(rt.Kind()))
	fmt.Println(rv2.Elem())
	fmt.Println(rt.String(), rt.Name(), rv.String(), rv.Type())
	for i := 0; i < rt.NumField(); i++ {
		tf := rt.Field(i)
		vf := rv.Field(i)

		fmt.Printf("%-20s:%s\n", "tf.Tag", tf.Tag)
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("json")`, tf.Tag.Get("json"))
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("sql")`, tf.Tag.Get("sql"))
		fmt.Printf("%-20s:%s\n", `tf.Name`, tf.Name)
		fmt.Printf("%-20s:%s\n", `tf.Type`, tf.Type)
		fmt.Printf("%-20s:%d\n", `tf.Offset`, tf.Offset)
		fmt.Printf("%-20s:%t\n", `vf.CanInterface()`, vf.CanInterface())

		if vf.CanInterface() {

			fmt.Println(vf.Interface())
		}
	}
}

func getFieldName(s interface{}) {
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	fmt.Println(rt, rv)
	// ptr => element : 方法：Elem()
	fmt.Println(rt.Kind(), rt.NumMethod())
	fmt.Println(rv.Elem(), rv.Elem().CanInterface(), rv.Elem().Interface())
	fmt.Println(rt.Elem().Kind(), rt.Elem().NumField())
	fmt.Println(rv.Elem().Kind(), rv.Elem().NumField())

	rtt := rt.Elem()
	rvv := rv.Elem()
	for i := 0; i < rv.Elem().NumField(); i++ {
		tf := rtt.Field(i)
		vf := rvv.Field(i)

		fmt.Printf("%-20s:%s\n", "tf.Tag", tf.Tag)
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("json")`, tf.Tag.Get("json"))
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("sql")`, tf.Tag.Get("sql"))
		fmt.Printf("%-20s:%s\n", `tf.Name`, tf.Name)
		fmt.Printf("%-20s:%s\n", `tf.Type`, tf.Type)
		fmt.Printf("%-20s:%d\n", `tf.Offset`, tf.Offset)
		fmt.Printf("%-20s:%t\n", `vf.CanInterface()`, vf.CanInterface())

		if vf.CanInterface() {

			fmt.Println(vf.Interface())
		}
	}
}

func getFieldName2(s interface{}) {
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	fmt.Println(rt, rv)
	fmt.Println(rt.Kind(), rt.NumField(), rt.NumMethod())

	for i := 0; i < rv.NumField(); i++ {
		tf := rt.Field(i)
		vf := rv.Field(i)

		fmt.Printf("%-20s:%s\n", "tf.Tag", tf.Tag)
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("json")`, tf.Tag.Get("json"))
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("sql")`, tf.Tag.Get("sql"))
		fmt.Printf("%-20s:%s\n", `tf.Name`, tf.Name)
		fmt.Printf("%-20s:%s\n", `tf.Type`, tf.Type)
		fmt.Printf("%-20s:%d\n", `tf.Offset`, tf.Offset)
		fmt.Printf("%-20s:%t\n", `vf.CanInterface()`, vf.CanInterface())

		if vf.CanInterface() {

			fmt.Println(vf.Interface())
		}
	}
}

func getFieldName3(s interface{}) {
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	fmt.Println(rt, rv)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	fmt.Println(rt.Kind(), rt.NumField(), rt.NumMethod())

	for i := 0; i < rv.NumField(); i++ {
		tf := rt.Field(i)
		vf := rv.Field(i)

		fmt.Printf("%-20s:%s\n", "tf.Tag", tf.Tag)
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("json")`, tf.Tag.Get("json"))
		fmt.Printf("%-20s:%s\n", `tf.Tag.Get("sql")`, tf.Tag.Get("sql"))
		fmt.Printf("%-20s:%s\n", `tf.Name`, tf.Name)
		fmt.Printf("%-20s:%s\n", `tf.Type`, tf.Type)
		fmt.Printf("%-20s:%d\n", `tf.Offset`, tf.Offset)
		fmt.Printf("%-20s:%t\n", `vf.CanInterface()`, vf.CanInterface())

		if vf.CanInterface() {

			fmt.Println(vf.Interface())
		}
	}
}
