package notify

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type IsWatch struct {
	m  sync.Mutex
	is bool
}

func DevDirFile() {
	// wait := make(chan bool)
	go watch()
	// <-wait
	for {
		// time.Sleep(time.Second)
		// fmt.Println("---------")
	}
}

func watch() {
	wait := make(chan bool)
	fmt.Println("listen file change")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	dir, _ := os.Getwd()
	dir = fmt.Sprintf("%s/pkg/notify/gomusic", dir)
	count := 0

	go func() {
		iswatch := IsWatch{is: false}
		for {
			select {
			case event, ok := <-watcher.Events:
				val, err := ioutil.ReadFile(dir)

				iswatch.m.Lock()
				if iswatch.is {
					iswatch.m.Unlock()
					break
				}
				iswatch.is = true

				count = count + 1
				fmt.Println(event, ok, string(val), err, count)

				iswatch.m.Unlock()
				go func() {
					time.Sleep(time.Microsecond * 10)
					iswatch.m.Lock()
					iswatch.is = false
					iswatch.m.Unlock()
				}()

			case _, err := <-watcher.Errors:
				fmt.Println(err)
			}
		}
	}()

	err = watcher.Add(dir)
	fmt.Println("watcher.Add:", err)
	<-wait
}
