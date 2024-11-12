package main

import (
	"fmt"
	"go/types"
)

func main() {

	fmt.Println(int(types.Invalid))
	fmt.Println(int(types.Bool))
	fmt.Println(int(types.Float64))

}
