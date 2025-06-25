package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// MQTT控制包类型
const (
	CONNECT    = 1
	CONNACK    = 2
	PUBLISH    = 3
	SUBSCRIBE  = 8
	SUBACK     = 9
	DISCONNECT = 14
)

// MQTT连接标志
const (
	CLEAN_SESSION = 1 << 1
)

func main() {
	// 创建TCP连接
	conn, err := net.Dial("tcp", "broker.emqx.io:1883")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 发送CONNECT包
	connectPacket := createConnectPacket("go-raw-client", 60, CLEAN_SESSION)
	if _, err := conn.Write(connectPacket); err != nil {
		panic(err)
	}

	// 接收CONNACK响应
	response := make([]byte, 4)
	if _, err := conn.Read(response); err != nil {
		panic(err)
	}

	// 验证连接成功 (字节3应为0)
	if response[0] != (CONNACK<<4) || response[3] != 0 {
		panic("Connection refused by broker")
	}
	fmt.Println("Connected to MQTT broker")

	// 订阅主题
	subscribePacket := createSubscribePacket("raw/test")
	if _, err := conn.Write(subscribePacket); err != nil {
		panic(err)
	}
	fmt.Println("Subscribed to topic")

	// 启动接收循环
	go receiveMessages(conn)

	// 发布测试消息
	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("Raw message %d", i)
		publishPacket := createPublishPacket("raw/test", msg)
		if _, err := conn.Write(publishPacket); err != nil {
			fmt.Println("Publish error:", err)
		} else {
			fmt.Println("Published:", msg)
		}
		time.Sleep(1 * time.Second)
	}

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 发送DISCONNECT
	conn.Write([]byte{DISCONNECT << 4, 0})
	fmt.Println("Disconnected")
}

// 创建CONNECT包
func createConnectPacket(clientID string, keepAlive uint16, flags byte) []byte {
	protocol := "MQTT"
	packet := []byte{}

	// 可变头
	packet = append(packet, byte(len(protocol)>>8), byte(len(protocol)))
	packet = append(packet, protocol...)
	packet = append(packet, 0x04) // 协议版本 (4=MQTT 3.1.1)
	packet = append(packet, flags)
	packet = append(packet, byte(keepAlive>>8), byte(keepAlive))

	// 有效载荷
	idBytes := []byte(clientID)
	packet = append(packet, byte(len(idBytes)>>8), byte(len(idBytes)))
	packet = append(packet, idBytes...)

	// 固定头
	header := []byte{CONNECT << 4}
	header = append(header, encodeLength(len(packet))...)
	return append(header, packet...)
}

// 创建SUBSCRIBE包
func createSubscribePacket(topic string) []byte {
	// 固定头
	packet := []byte{SUBSCRIBE<<4 | 0x02} // QoS 1
	packet = append(packet, 0)            // 长度占位

	// 可变头（消息ID）
	packetID := uint16(1)
	packet = append(packet, byte(packetID>>8), byte(packetID))

	// 订阅内容
	topicBytes := []byte(topic)
	packet = append(packet, byte(len(topicBytes)>>8), byte(len(topicBytes)))
	packet = append(packet, topicBytes...)
	packet = append(packet, 0) // QoS 0

	// 更新长度
	packet[1] = encodeLength(len(packet[2:]))[0]
	return packet
}

// 创建PUBLISH包
func createPublishPacket(topic, message string) []byte {
	topicBytes := []byte(topic)
	msgBytes := []byte(message)

	// 可变头
	packet := []byte{}
	packet = append(packet, byte(len(topicBytes)>>8), byte(len(topicBytes)))
	packet = append(packet, topicBytes...)

	// 消息载荷
	packet = append(packet, msgBytes...)

	// 固定头 (QoS 0)
	header := []byte{PUBLISH << 4}
	header = append(header, encodeLength(len(packet))...)
	return append(header, packet...)
}

// MQTT长度编码
func encodeLength(length int) []byte {
	var encoded []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encoded = append(encoded, digit)
		if length == 0 {
			break
		}
	}
	return encoded
}

// 接收消息
func receiveMessages(conn net.Conn) {
	for {
		header := make([]byte, 1)
		if _, err := conn.Read(header); err != nil {
			fmt.Println("Read error:", err)
			return
		}

		packetType := header[0] >> 4
		length, _ := readLength(conn)
		payload := make([]byte, length)
		if _, err := conn.Read(payload); err != nil {
			fmt.Println("Payload read error:", err)
			return
		}

		switch packetType {
		case PUBLISH:
			// 提取主题和消息
			topicLen := binary.BigEndian.Uint16(payload[0:2])
			topic := string(payload[2 : 2+topicLen])
			message := string(payload[2+topicLen:])
			fmt.Printf("\nReceived: [%s] %s\n", topic, message)

		case SUBACK:
			fmt.Println("Subscription acknowledged")

		case DISCONNECT:
			fmt.Println("Broker requested disconnect")
			return
		}
	}
}

// MQTT长度解码
func readLength(conn net.Conn) (int, error) {
	multiplier := 1
	value := 0
	for {
		digit := make([]byte, 1)
		if _, err := conn.Read(digit); err != nil {
			return 0, err
		}
		value += int(digit[0]&0x7F) * multiplier
		if digit[0]&0x80 == 0 {
			break
		}
		multiplier *= 128
		if multiplier > 128*128*128 {
			return 0, fmt.Errorf("malformed length")
		}
	}
	return value, nil
}
