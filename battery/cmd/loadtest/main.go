// Package main implements a high-scale MQTT load testing tool.
// This command-line application simulates thousands of concurrent MQTT clients
// to stress-test the MQTT broker's performance and capacity.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

// MQTT packet types as defined in the MQTT 3.1.1 specification
const (
	Connect     = 1  // Client request to connect to Server
	Connack     = 2  // Connect acknowledgment
	Publish     = 3  // Publish message
	Subscribe   = 8  // Client subscribe request
	Suback      = 9  // Subscribe acknowledgment
	Unsubscribe = 10 // Unsubscribe request
	Pingreq     = 12 // PING request
	Pingresp    = 13 // PING response
	Disconnect  = 14 // Client is disconnecting
)

var (
	// brokerAddr specifies the MQTT broker address to connect to
	// Default value is "localhost:1883"
	brokerAddr = flag.String("addr", "localhost:1883", "MQTT broker address")
	// numClients defines the number of concurrent MQTT clients to simulate
	// Default value is 100,000 to test high-scale scenarios
	numClients = flag.Int("clients", 100000, "Number of concurrent clients to simulate")
	// testTime specifies the duration for which the load test will run
	// Default value is 30 seconds
	testTime = flag.Duration("time", 30*time.Second, "Duration of the test")
)

type Client struct {
	id     string
	conn   net.Conn
	stopCh chan struct{}
	wg     *sync.WaitGroup
}

func NewClient(id string) *Client {
	return &Client{
		id:     id,
		stopCh: make(chan struct{}),
	}
}

func (c *Client) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c.conn = conn

	// Send CONNECT packet
	connectPacket := buildConnectPacket(c.id)
	_, err = c.conn.Write(connectPacket)
	if err != nil {
		return err
	}

	// Read CONNACK
	buf := make([]byte, 4)
	_, err = c.conn.Read(buf)
	if err != nil {
		return err
	}

	if buf[0]>>4 != Connack {
		return fmt.Errorf("expected CONNACK, got %d", buf[0]>>4)
	}

	return nil
}

func buildConnectPacket(clientID string) []byte {
	// Protocol name: "MQTT"
	protoName := []byte("MQTT")

	// Calculate remaining length:
	// 2 bytes for protocol name length +
	// len(protocol name) +
	// 1 byte for protocol level +
	// 1 byte for connect flags +
	// 2 bytes for keep alive +
	// 2 bytes for client ID length +
	// len(client ID)
	remainingLength := 2 + len(protoName) + 1 + 1 + 2 + 2 + len(clientID)

	var packet []byte
	packet = append(packet, Connect<<4)                                    // MQTT Control Packet type
	packet = append(packet, encodeVariableByteInteger(remainingLength)...) // Remaining Length

	// Protocol Name
	packet = append(packet, byte(len(protoName)>>8), byte(len(protoName)&0xFF))
	packet = append(packet, protoName...)

	// Protocol Level
	packet = append(packet, 4) // MQTT v3.1.1

	// Connect Flags
	packet = append(packet, byte(0)) // Clean session, no other flags

	// Keep Alive (30 seconds)
	packet = append(packet, byte(0), byte(30))

	// Client Identifier
	packet = append(packet, byte(len(clientID)>>8), byte(len(clientID)&0xFF))
	packet = append(packet, []byte(clientID)...)

	return packet
}

// encodeVariableByteInteger encodes an integer into variable byte integer format
func encodeVariableByteInteger(length int) []byte {
	var result []byte
	encodedByte := byte(0)

	for {
		encodedByte = byte(length % 128)
		length = length / 128
		if length > 0 {
			encodedByte |= 128
		}
		result = append(result, encodedByte)
		if length == 0 {
			break
		}
	}

	return result
}

func (c *Client) Subscribe(topic string) error {
	packetID := uint16(rand.Intn(65535) + 1)

	packet := make([]byte, 4+len(topic))
	packet[0] = Subscribe<<4 | 2     // packet ID required
	packet[1] = byte(2 + len(topic)) // remaining length
	packet[2] = byte(packetID >> 8)
	packet[3] = byte(packetID & 0xFF)

	// Topic
	packet[4] = byte(len(topic) >> 8)
	packet[5] = byte(len(topic) & 0xFF)
	copy(packet[6:], topic)

	// QoS level
	packet = append(packet, 0) // QoS 0

	packet[1] = byte(len(packet) - 2) // update remaining length

	_, err := c.conn.Write(packet)
	if err != nil {
		return err
	}

	// Read SUBACK
	buf := make([]byte, 5)
	_, err = c.conn.Read(buf)
	if err != nil {
		return err
	}

	if buf[0]>>4 != Suback {
		return fmt.Errorf("expected SUBACK, got %d", buf[0]>>4)
	}

	return nil
}

func (c *Client) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	defer c.conn.Close()

	ticker := time.NewTicker(20 * time.Second) // Send PINGREQ every 20 sec to keep alive
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			// Send DISCONNECT before closing
			disconnectPacket := []byte{Disconnect << 4, 0}
			c.conn.Write(disconnectPacket)
			return
		case <-ticker.C:
			// Send PINGREQ to keep connection alive
			pingreqPacket := []byte{Pingreq << 4, 0}
			c.conn.Write(pingreqPacket)

			// Read PINGRESP
			buf := make([]byte, 2)
			c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			c.conn.Read(buf)
		}
	}
}

func (c *Client) Stop() {
	close(c.stopCh)
}

func main() {
	flag.Parse()

	log.Printf("Starting load test with %d clients for %v", *numClients, *testTime)

	// Create clients
	clients := make([]*Client, *numClients)
	var wg sync.WaitGroup

	// Connect all clients
	for i := 0; i < *numClients; i++ {
		clientID := "client_" + strconv.Itoa(i)
		client := NewClient(clientID)
		clients[i] = client

		err := client.Connect(*brokerAddr)
		if err != nil {
			log.Printf("Failed to connect client %s: %v", clientID, err)
			continue
		}

		// Subscribe to a topic
		topic := fmt.Sprintf("loadtest/topic%d", rand.Intn(1000))
		err = client.Subscribe(topic)
		if err != nil {
			log.Printf("Failed to subscribe client %s to topic %s: %v", clientID, topic, err)
		}

		wg.Add(1)
		go client.Start(&wg)

		// Print progress every 1000 connections
		if (i+1)%1000 == 0 {
			log.Printf("Connected %d/%d clients", i+1, *numClients)
		}

		// Small delay to avoid overwhelming the broker
		time.Sleep(10 * time.Microsecond)
	}

	log.Printf("All %d clients connected and subscribed", *numClients)

	// Sleep for test duration
	time.Sleep(*testTime)

	log.Println("Stopping all clients...")

	// Stop all clients
	for _, client := range clients {
		if client != nil {
			client.Stop()
		}
	}

	// Wait for all clients to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All clients stopped")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for clients to stop")
	}

	log.Println("Load test completed")
}
