package main

import (
	"errors"
	"fmt"
)

func main() {
	errs := []error{errors.New("err 1"), errors.New("err 2")}
	fmt.Println(fmt.Errorf("%+v", errs))
}
