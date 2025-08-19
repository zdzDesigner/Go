package main

import (
	"container/list"
	"fmt"
)

type Element struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element

	// The list to which this element belongs.
	list *List

	// The value stored with this element.
	Value any
}
type List struct {
	root Element // sentinel list element, only &root, root.prev, and root.next are used
	len  int     // current list length excluding (this) sentinel element
}

func main() {
	pool := list.New()

	fmt.Println(pool.Front())
	pool.PushFront(1)
	fmt.Println(pool.Front().Value)

	fmt.Println(Element{})    // {<nil> <nil> <nil> <nil>}
	fmt.Println(new(Element)) // &{<nil> <nil> <nil> <nil>}
}
