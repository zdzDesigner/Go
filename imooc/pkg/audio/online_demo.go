package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func demo() {
	buf := make([]byte, 4)
	bufcopy := make([]byte, 2)
	file, err := os.Open("test.pcm")
	if err != nil {
		print(err.Error())
	}
	out1, err := create("l.pcm")
	out2, err2 := os.Create("r.pcm")
	if err != nil || err2 != nil {
		return
	}
	for {
		n, err := file.Read(buf)
		if (err != nil && err != io.EOF) || n == 0 {
			println(err.Error())
			break
		}
		copy(bufcopy, buf[:2])
		out1.Write(bufcopy)
		copy(bufcopy, buf[2:])
		out2.Write(bufcopy)
	}
	out1.Close()
	file.Close()
}

func create(file string) (*os.File, error) {
	out, err := os.Create(file)
	if err != nil {
		log.Print(err.Error())
	}
	return out, err
}

func base() {
	// buf := make([]byte, 4)
	// bufcopy := make([]byte, 2)
	file, err := os.Open("/home/zdz/Documents/Try/Go/imooc/lib/audio/test.wav")
	if err != nil {
		print(err.Error())
	}

	newfile, err := createFile(fmt.Sprintf("/home/zdz/Documents/Try/Go/imooc/lib/audio/test-%s.wav", time.Now().Format("2006-01-2-15-04-05")))
	defer func() {
		file.Close()
		newfile.Close()
	}()
	// limit := 125
	// count := 0

	bslen := 24
	bs := make([]byte, 0, bslen)
	for {
		buf := make([]byte, 4)
		n, err := file.Read(buf)
		if (err != nil && err != io.EOF) || n == 0 {
			println(err.Error())
			break
		}
		// fmt.Println(buf[0] == 0)
		newfile.Write(buf)
		// if info, _ := newfile.Stat(); info.Size() > 100000 {
		// 	break
		// }
		// if buf[0] == 0 {
		// 	count++
		// }
		// if count > limit {
		// 	break
		// }
		if buf[0] == 0 {
			bs = append(bs, buf...)
		} else {
			bs = make([]byte, 0, bslen)
		}

		if len(bs) >= bslen {
			break
		}

	}
}

func base1() {
	file, err := os.Open("/home/zdz/Documents/Try/Go/imooc/lib/audio/test.wav")
	if err != nil {
		print(err.Error())
	}

	newfile, err := createFile(fmt.Sprintf("/home/zdz/Documents/Try/Go/imooc/lib/audio/test-%s.wav", time.Now().Format("2006-01-2-15-04-05")))
	defer func() {
		file.Close()
		newfile.Close()
	}()

	bslen := 24
	bs := make([]byte, 0, bslen)
	b := false
	header := make([]byte, 0, 44)
	for {
		buf := make([]byte, 4)
		n, err := file.Read(buf)
		if (err != nil && err != io.EOF) || n == 0 {
			println(err.Error())
			break
		}
		// fmt.Println(buf)
		if len(header) <= 44 {
			header = append(header, buf...)
			newfile.Write(buf)
		}
		// newfile.Write(buf)
		if !b {
			if buf[0] == 0 {
				bs = append(bs, buf...)
			} else {
				bs = make([]byte, 0, bslen)
			}

		} else {
			newfile.Write(buf)
		}

		if len(bs) >= bslen {
			b = true

			// fmt.Println(len(bs))
			// time.Sleep(time.Second)
			// newfile, err = createFile(fmt.Sprintf("/home/zdz/Documents/Try/Go/imooc/lib/audio/test-%s.wav", time.Now().Format("2006-01-2-15-04-05")))
			// bs = make([]byte, 0, bslen)
			// break
		}

	}
}
