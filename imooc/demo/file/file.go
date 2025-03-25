package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// fmt.Println(filepath.Ext("aaa.b"))
	// fmt.Println(filepath.Base("/fff/dddd/aaa.b"))
	// fmt.Println(getName("/fff/dddd/aaa.b"))
	createTempFile(getAudio())
}

func getName(path string) (name, ext string) {
	ext = filepath.Ext(path)
	name = strings.Replace(filepath.Base(path), ext, "", -1)
	return
}

func getAudio() []byte {
	res, _ := http.Get("https://dict.youdao.com/dictvoice?type=0&audio=name")

	buf := make([]byte, res.ContentLength)
	fmt.Println(len(buf))

	for {

		n, err := res.Body.Read(buf)
		fmt.Println(n, err, res.ContentLength)
		if err != nil {
			break
		}
	}
	// fmt.Println(string(buf))
	return buf
}

func createTempFile(source []byte) {
	// dir(""): 默认文件
	// pattern(xxv*v):*被替换为随机数
	f, _ := os.CreateTemp("", "go_vvvv*.mp3")
	fmt.Println(f)
  // f.ReadDir()

	n, err := f.Write(source)
	fmt.Println(n, err)
}
