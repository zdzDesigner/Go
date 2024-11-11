package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func Request() {
	req, err := http.NewRequest("GET", "api", strings.NewReader("body"))
	fmt.Println(req, err)
	// r2 := new(Request)
	// r2.ctx = ctx
	req2 := req.WithContext(context.Background()) //
	fmt.Println(req.Body == req2.Body)
	fmt.Println(req.Cancel == req2.Cancel)
	fmt.Println(req.MultipartForm == req2.MultipartForm)
}
