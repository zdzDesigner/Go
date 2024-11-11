package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {

	fmt.Println(os.Getpid())
	c := make(chan os.Signal)
	signal.Notify(c)
	s := <-c
	fmt.Println(s.String())
	if s.String() == "terminated" {
		time.Sleep(time.Second * 5)
	}

}
