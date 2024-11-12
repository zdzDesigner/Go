package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	// bufio.MaxScanTokenSize
	input := strings.Repeat("x", 10)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}
}
