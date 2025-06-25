package main

import (
	"cmp"
	"fmt"
)

func Stringify[T fmt.Stringer](s []T) (ret []string) {
	for _, v := range s {
		ret = append(ret, v.String())
	}
	return ret
}

type Stringer interface {
	cmp.Ordered
	comparable
	fmt.Stringer
}

func StringifyWithoutZero[T Stringer](s []T) (ret []string) {
	var zero T
	for _, v := range s {
		if v == zero {
			continue
		}
		ret = append(ret, v.String())
	}
	return ret
}

func StringifyLessThan[T Stringer](s []T, max T) (ret []string) {
	var zero T
	for _, v := range s {
		if v == zero || v >= max {
			continue
		}
		ret = append(ret, v.String())
	}
	return ret
}

type MyString string

func (s MyString) String() string {
	return string(s)
}

func main() {
	sl := Stringify([]MyString{"I", "love", "", "golang"})
	fmt.Println(sl) // 输出：[I love golang]

	fmt.Println(StringifyWithoutZero([]MyString{"a", "", "b", "z"}))

	fmt.Println(StringifyLessThan([]MyString{"a", "", "b", "z"}, MyString("xx")))
}




type Ia interface {
  int | string  // 仅代表int和string
}

// underlying type 底层类型
type Ib interface {
  ~int | ~string  // 代表以int和string为底层类型的所有类型
}
