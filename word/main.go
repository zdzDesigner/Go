package main

import (
	"archive/zip"
	"bytes"
	// "encoding/xml"
	"fmt"
	"syscall/js"

	"word/lib"
)

func readFile2(this js.Value, args []js.Value) interface{} {
	// alert := js.Global().Get("alert")
	// alert.Invoke("Hello World!")

	fmt.Println(len(args))
	// 获取文件对象
	data := args[0]
	fmt.Println(data.Length())

	return nil
}

func readFile3(this js.Value, args []js.Value) interface{} {
	fileInput := js.Global().Call("getElementById", "fileInput")

	fileInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
		fileInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
			data := js.Global().Get("Uint8Array").New(x[0])
			dst := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(dst, data)
			// the data from the file is in dst - do what you want with it

			return nil
		}))

		return nil
	}))
	return nil
}

func reader(b []byte) (*zip.Reader, error) {
	reader := bytes.NewReader(b)
	return zip.NewReader(reader, int64(len(b)))
}

func parseFile(b []byte) {
	r, err := reader(b)
	if err != nil {
		return
	}

	lib.NewZf(r)
	// var buf bytes.Buffer
	// zf, err := lib.NewZf(r)
	// err = zf.Walk(&node, &buf)
	// if err != nil {
	// 	fmt.Println("parse md:", err.Error())
	// 	return
	// }
	// fmt.Print(buf.String())
}

func readFile(this js.Value, p []js.Value) interface{} {
	file := p[0]
	fileReader := js.Global().Get("FileReader").New()
	// fileReader.Call("readAsText", file)
	fileReader.Call("readAsArrayBuffer", file)
	fileReader.Call("addEventListener", "load", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		fmt.Println("p length:", len(p))
		// fileReader.Get("result").Call("readAsArrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
		// 	return nil
		// }))
		// length := data.Get("length").Int()
		// data := js.Global().Get("Uint8Array").New(x[0])
		// length := data.Get("length").Int()
		// fmt.Println("file data length:", length)
		// dst := make([]byte, length)
		// js.CopyBytesToGo(dst, data)
		// fmt.Println("File contents:", fileReader.Get("result").String())
		res := fileReader.Get("result")
		// res := p[0].Get("target").Get("result")
		data := js.Global().Get("Uint8Array").New(res)
		length := data.Get("length").Int()
		fmt.Println("file data length:", length)
		dst := make([]byte, length)
		js.CopyBytesToGo(dst, data)
		parseFile(dst)
		return nil
	}))
	return nil
}

func main() {
	c := make(chan struct{}, 0)

	// 在Go中注册函数
	js.Global().Set("readFile", js.FuncOf(readFile))

	<-c // 阻塞，保持程序运行
}
