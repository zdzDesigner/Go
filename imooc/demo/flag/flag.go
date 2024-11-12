package main

import (
	"flag"
	"fmt"
)

func main() {

	sourceVal := flag.String("source", "xxx", "source file path.")
	exportVal := flag.String("export", "xxx", "export file dir.")
	flag.Parse()
	fmt.Println(*sourceVal, *exportVal)
	fmt.Println(flag.Lookup("source").Value.String())
	fmt.Println(flag.Lookup("export").Value.String())
}
