package main

import (
	"os"
	"syscall"
	"time"
)

func main() {
	inout()

	time.Sleep(time.Second * 10)
}

func inout() (err error) {
	_, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	// fmt.Println(in, out)
	return nil
}
