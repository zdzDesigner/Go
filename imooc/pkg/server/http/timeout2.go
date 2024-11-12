package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTP2 ..
func HTTP2() {
	ctx, cancel := context.WithCancel(context.TODO())
	timedead := 5 * time.Second
	// timedead := 550 * time.Millisecond
	timer := time.AfterFunc(timedead, func() {
		cancel()
	})

	req, err := http.NewRequest("GET", "http://httpbin.org/range/2048?duration=8&chunk_size=256", nil)
	// req, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	req = req.WithContext(ctx)

	log.Println("Sending request...")
	timeSince := time.Now()
	resp, err := http.DefaultClient.Do(req)
	fmt.Println(resp.Status, time.Since(timeSince))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bs := make([]byte, 0, 1024)
	resp.Body.Read(bs)

	log.Println("Reading body.x..", string(bs))

	for {
		timer.Reset(2 * time.Second)
		// timer.Reset(2 * time.Millisecond)
		// Try instead: timer.Reset(50 * time.Millisecond)
		b, err := io.CopyN(&W{}, resp.Body, 256)
		fmt.Println(b)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}
}

type W struct{}

func (w *W) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))

	return len(p), nil
}
