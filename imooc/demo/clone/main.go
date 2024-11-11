package main

import (
	"fmt"
)

type Brander interface {
	GetName() string
}

type Brand struct {
	Name string
}

func (b *Brand) GetName() string {
	return b.Name
}

type Bag struct {
	Type  string
	brand Brander
}

func main() {
	bag1 := &Bag{brand: &Brand{"v1"}}
	bag1.Type = "computer"
	fmt.Println(bag1)

	bag2 := new(Bag)
	*bag2 = *bag1
	fmt.Println(bag2)
	(*bag2).Type = "xx"
	bag2.brand = &Brand{"v2"}
	fmt.Println(bag1, bag2)
}
