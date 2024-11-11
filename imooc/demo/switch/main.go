package main

import (
	"errors"
)

func main() {
	err := condition("xxx")
	println(err.Error())
  condition("udp")

}

func condition(network string) error {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return errors.New("err")
	}

	println("pass")

	return nil

}
