package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	hex := "FFFF16AC" // Example hexadecimal value

	// Convert hexadecimal to decimal bytes
	decBytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		byteVal, _ := strconv.ParseUint(hex[i:i+2], 16, 8)
		decBytes[i/2] = byte(byteVal)
	}

	// Create the IP address
	ip := net.IP(decBytes)

	// Print the IP address
	fmt.Println("IP address:", ip.String())
}
