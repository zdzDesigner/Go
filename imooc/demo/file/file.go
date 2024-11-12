package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println(filepath.Ext("aaa.b"))
	fmt.Println(filepath.Base("/fff/dddd/aaa.b"))
	fmt.Println(getName("/fff/dddd/aaa.b"))
}

func getName(path string) (name, ext string) {
	ext = filepath.Ext(path)
	name = strings.Replace(filepath.Base(path), ext, "", -1)
	return
}
