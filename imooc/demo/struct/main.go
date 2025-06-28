package main

import "fmt"

type User struct {
	Name string
}

func (u *User) name() {
	fmt.Println(u.Name)
}

func main() {
	u := User{Name: "zdz"}
	u.name()
}
