package base

import (
	"fmt"
	"net/http"
)

// MyHandler ..
type MyHandler struct{}

func (mh *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Println(w.Header())
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(w.Header())
	w.Write([]byte(`{"msg":"hello"}`))
}

// Entry ..
func Entry() {
	server := http.Server{
		Addr:    "127.0.0.1:6001",
		Handler: &MyHandler{},
	}
	server.ListenAndServe()
}
