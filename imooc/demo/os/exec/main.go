package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	var (
		err error
		nc  = strings.Fields("nc -u -l 8888")
	)

	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	cmd := exec.Command(nc[0], nc[1:]...)
	if err = cmd.Start(); err != nil {
		return
	}
	fmt.Println("process pid:", cmd.Process.Pid)

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Child process %d exit with err: %v\n", cmd.Process.Pid, err)
		}
	}()
	// After five second, kill cmd's process
	time.Sleep(5 * time.Second)
	err = cmd.Process.Kill()

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// <-c

}
