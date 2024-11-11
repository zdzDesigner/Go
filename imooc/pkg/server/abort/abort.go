package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	base()
}

func base() {
	req, _ := http.NewRequest("GET", "http://google.com", nil)
	tr := &http.Transport{} // TODO: copy defaults from http.DefaultTransport
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() {
		resp, err := client.Do(req)
		// handle response ...
		_ = resp
		c <- err
	}()

	// Simulating user cancel request channel
	user := make(chan struct{}, 0)
	go func() {
		time.Sleep(100 * time.Millisecond)
		user <- struct{}{}
	}()

	for {
		select {
		case <-user:
			log.Println("Cancelling request")
			tr.CancelRequest(req)
		case err := <-c:
			log.Println("Client finished:", err)
			return
		}
	}
}

func base2() {
	cx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequest("GET", "http://google.com", nil)
	req = req.WithContext(cx)
	ch := make(chan error)

	go func() {
		_, err := http.DefaultClient.Do(req)
		select {
		case <-cx.Done():
			// Already timedout
		default:
			ch <- err
		}
	}()

	// Simulating user cancel request
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	select {
	case err := <-ch:
		if err != nil {
			// HTTP error
			panic(err)
		}
		print("no error")
	case <-cx.Done():
		panic(cx.Err())
	}

}

func base3() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequest("GET", "http://google.com", nil)
	httpDo(ctx, req, func(resp *http.Response, err error) error {
		return nil
	})
}
func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req)) }()
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
