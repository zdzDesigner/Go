package main

import (
	"encoding/json"
	"fmt"
)

type Article struct {
	Id   string  `json:"id"`
	Name *string `json:"name,omitempty"`
	Desc *string `json:"desc,omitempty"`
}

func main() {
	Test_JSON_Empty()
	Test_JSON_Nil()
}

func Test_JSON_Empty() {
	jsonData := `{"id":"1234","name":"xyz","desc":""}`
	req := Article{}
	_ = json.Unmarshal([]byte(jsonData), &req)
	fmt.Printf("%+v\n", req)
	fmt.Printf("name:%s\n", *req.Name)
	fmt.Printf("desc:%s\n", *req.Desc)
}
func Test_JSON_Nil() {
	jsonData := `{"id":"1234","name":"xyz"}`
	req := Article{}
	_ = json.Unmarshal([]byte(jsonData), &req)
	fmt.Printf("%+v\n", req)
	fmt.Printf("%s\n", *req.Name)
}
