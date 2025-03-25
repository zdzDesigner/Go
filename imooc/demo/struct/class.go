package main

import "fmt"

// User ..
type User struct {
	Name string
}

// Admin 包含了 User 的所有内容
type Admin struct {
	User
}

// sup ..
type SUP struct {
	user *User
}

type Common struct {
	*User
}
type Common2 struct {
	User
}

// Get ..
func (u *User) Get() {
	fmt.Println("user")
  fmt.Println(u.Name)
}

// 紧急包含了User struct内容
type Person User

func main() {
	common := Common{User: &User{Name: "xxx"}}
	fmt.Println(common.Name)
  common.Get()
	common2 := Common2{User: User{Name: "xxx"}}
	fmt.Println(common.Name)
  common2.Get()

	// (User{}).Get() // cannot call pointer method on User literalgo
	u := User{}
	u.Get()

	p := Person{Name: "aaa"}
	fmt.Println("type struct::", p.Name)
	// p.Get() // undefined type Person has no field or method Get

	a := Admin{User{Name: "bbb"}}
	fmt.Println(a.Name)
	a.Get()

	sp := SUP{user: &User{}}
	sp.user.Name = "aa"
	fmt.Println(sp)

	sp2 := SUP{}
	sp2.user = &User{}
	sp2.user.Name = "bb"
	fmt.Println(sp2)
	sp2.user.Get()
}
