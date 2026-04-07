package main

import (
	"fmt"
	"testing"
)

func TestConnectPacket(t *testing.T) {
	client := MQTTClient{packet: &Packet{}}

	packet := client.packet.connect("aaa")
	fmt.Println(packet)
}
