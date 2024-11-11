package main

import "fmt"

func main() {

	fn := func() {
		fmt.Println("inner fn")
	}

	fn()

}
