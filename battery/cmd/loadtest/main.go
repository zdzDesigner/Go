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

const (
	Connect     = 1
	Connack     = 2
	Publish     = 3
	Subscribe   = 8
	Suback      = 9
	Unsubscribe = 10
	Pingreq     = 12
	Pingresp    = 13
	Disconnect  = 14
)

var (
	brokerAddr = flag.String("addr", "localhost:1883", "MQTT broker address")
	numClients = flag.Int("clients", 100000, "Number of concurrent clients to simulate")
	testTime   = flag.Duration("time", 30*time.Second, "Duration of the test")
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
	// Fixed header
	var packet []byte

	// Protocol name: "MQTT"
	protoName := []byte("MQTT")
	packet = append(packet, Connect<<4)
	packet = append(packet, byte(10+len(clientID))) // remaining length
	packet = append(packet, byte(len(protoName)>>8), byte(len(protoName)&0xFF))
	packet = append(packet, protoName...)
	packet = append(packet, 4)                 // protocol level
	packet = append(packet, byte(0))           // connect flags
	packet = append(packet, byte(0), byte(30)) // keep alive (30 seconds)

	// Client ID
	packet = append(packet, byte(len(clientID)>>8), byte(len(clientID)&0xFF))
	packet = append(packet, []byte(clientID)...)

	return packet
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
