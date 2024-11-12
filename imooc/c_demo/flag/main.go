package main

import (
	"bytes"
	"fmt"
)

func main() {
	for {
		// var str string
		// fmt.Scanf("%s", &str)
		// fmt.Println(str)

		var c byte
		i, err := fmt.Scanf("%c", &c)
		if err != nil || c == 10 {
			continue
		}
		if c == 'q' {
			fmt.Println("bye")
			break
		}

		fmt.Println(i, err, c, byte(c), string(c))
		fmt.Println([]byte{c}, bytes.NewBuffer([]byte{c, c}))
		fmt.Println(bytes.NewBuffer([]byte{c}).String() == string(c))
		fmt.Printf("%4d %4d", 5, 6)
	}

}
