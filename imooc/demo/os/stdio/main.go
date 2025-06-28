package main

import (
	// "fmt"
	"os"
	// "time"
)

func main() {
	buf := make([]byte, 100)
	for {
		n, _ := os.Stdin.Read(buf)
		if n > 0 {
			// fmt.Println(n, string(buf))
			// os.Stdout.Write([]byte("sss"))
			os.Stdout.Write(buf)
		}
		// time.Sleep(time.Second * 10)
	}
}
