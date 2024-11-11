package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	base("aaa/bbb.mp4?name=cc")
	base("https://m.ting22.com/Public/M/Js/jquery.jplayer.js?20191011")
	base1()
	base2()
	join()
}

func join() {
	dir, _ := os.Getwd()
	ph := filepath.Join(dir, "../../")
	fmt.Println(ph)
}

func base(str string) {
	fmt.Println(path.Ext(str))

	u, _ := url.Parse(str)
	fmt.Println(u.Path)
	fmt.Println("ext:", path.Ext(u.Path))
	fmt.Println(path.Ext(str))
	fmt.Println("dir::", filepath.Dir(str))
}

func base1() {
	str := "aaa/bbb.mp4?name=cc"
	fmt.Println(filepath.Ext(str))
	fmt.Println(filepath.Base(str))
	fmt.Println("dir::", filepath.Dir(str)) // aaa

	fmt.Println("文件名:", strings.Replace(filepath.Base(str), filepath.Ext(str), "", -1))
}

func base2() {
	dir := "/home/zdz/temp/lj-resource/"
	fs, _ := ioutil.ReadDir(dir)
	fmt.Println(fs[0].Name())
	fmt.Println(fs[1].Name())
	fmt.Println(fs[2].Name())
	fmt.Println(fs[3].Sys())
	fmt.Println(fs[3].Mode())
}
