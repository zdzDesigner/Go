package main

import (
	"fmt"
	"testing"
)

func TestConnectPacket(t *testing.T) {
	client := MQTTClient{}

  packet := client.createConnectPacket("aaa")
  fmt.Println(packet)
}
