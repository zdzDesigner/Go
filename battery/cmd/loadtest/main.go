// Package main 实现了一个大规模MQTT负载测试工具。
// 此命令行应用程序模拟数千个并发MQTT客户端
// 来压力测试MQTT代理的性能和容量。
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

// MQTT包类型，按照MQTT 3.1.1规范定义
const (
	Connect     = 1  // 客户端请求连接到服务器
	Connack     = 2  // 连接确认
	Publish     = 3  // 发布消息
	Subscribe   = 8  // 客户端订阅请求
	Suback      = 9  // 订阅确认
	Unsubscribe = 10 // 取消订阅请求
	Pingreq     = 12 // PING请求
	Pingresp    = 13 // PING响应
	Disconnect  = 14 // 客户端断开连接
)

var (
	// brokerAddr 指定要连接的MQTT代理地址
	// 默认值为"localhost:1883"
	brokerAddr = flag.String("addr", "localhost:1883", "MQTT broker address")
	// numClients 定义要模拟的并发MQTT客户端数量
	// 默认值为100,000以测试大规模场景
	numClients = flag.Int("clients", 100000, "Number of concurrent clients to simulate")
	// testTime 指定负载测试运行的持续时间
	// 默认值为30秒
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

// encodeVariableByteInteger 将整数编码为可变字节整数格式
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
