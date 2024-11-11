package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry ..
func Entry() {

	sourcefile := "/home/zdz/Documents/Try/Go/imooc/lib/audio/10.wav"
	exportdir := "/home/zdz/Documents/Try/Go/imooc/lib/audio/"
	file, err := os.Open(sourcefile)
	if err != nil {
		print(err.Error())
	}
	defer func() {
		file.Close()
	}()

	var (
		bslen     = 16
		bs        = make([]byte, 0, bslen)
		header    = make([]byte, 0, 44)
		content   = make([][]byte, 0)
		temp      = make([]byte, 0)
		i         = 0
		name, ext = getName(sourcefile)
	)

	for {
		buf := make([]byte, 4)
		n, err := file.Read(buf)
		if (err != nil && err != io.EOF) || n == 0 {
			println(err.Error())
			break
		}
		// fmt.Println(buf)
		// 音频头
		if len(header) <= 44 {
			header = append(header, buf...)
			continue
		}
		temp = append(temp, buf...)
		if buf[0] == 0 {
			bs = append(bs, buf...)
		} else {
			bs = make([]byte, 0, bslen)
		}

		if len(bs) >= bslen {
			content = append(content, temp)
			bs = make([]byte, 0, bslen)
			temp = make([]byte, 0)
		}

	}

	fmt.Println(len(content))
	for _, v := range content {
		if (len(header) + len(v)) < 100000 {
			continue
		}
		fmt.Println(len(header), len(v))
		i++
		newfile, _ := createFile(fmt.Sprintf("%s%s-%s-%d%s", exportdir, name, time.Now().Format("2006-01-2-15-04-05"), i, ext))
		newfile.Write(header)
		newfile.Write(v)
		newfile.Close()

	}

}

func createFile(file string) (*os.File, error) {
	out, err := os.Create(file)
	if err != nil {
		log.Print(err.Error())
	}
	return out, err
}
func getName(path string) (name, ext string) {
	ext = filepath.Ext(path)
	name = strings.Replace(filepath.Base(path), ext, "", -1)
	return
}
