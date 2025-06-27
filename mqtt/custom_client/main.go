package main

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	SUBSCRIBE   = 8
	PUBACK      = 5
	SUBACK      = 9
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

const (
	CLEAN_SESSION = 1 << 1
	KEEP_ALIVE    = 60
)

type MQTTMsg struct {
	topic   string
	message string
}

type MQTTClient struct {
	conn      net.Conn
	packet    *Packet
	packet_id uint16
	rx_buf    chan []byte
	tx_buf    chan MQTTMsg
}

// ack_buf
// pub_buf
// read_buf
// send_buf
// rx_buf
// tx_buf

func main() {
	client := &MQTTClient{
		packet_id: 1,
		packet:    &Packet{},
		rx_buf:    make(chan []byte, 10),
		tx_buf:    make(chan MQTTMsg, 10),
	}
	defer client.cleanup()

	// 连接到MQTTS代理
	// if err := client.connect("broker.emqx.io:8883", "go-optimized-client"); err != nil {
	if err := client.connect("172.16.40.51:19992", "go-optimized-client"); err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	fmt.Println("Connected to MQTT broker via TLS")

	go client.producer() // 启动消息接收器
	go client.consumer() // 启动消息消费者

	// topic := "opt/test"
	topic := "app/ready/notify"
	if err := client.subscribe(topic); err != nil {
		fmt.Println("Subscribe error:", err)
		return
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)

	// 发布测试消息
	go func() {
		for i := 0; i < 5; i++ {
			msg := fmt.Sprintf("Message %d at %s", i, time.Now().Format("15:04:05"))
			if err := client.publish(topic, msg); err != nil {
				fmt.Println("Publish error:", err)
			} else {
				fmt.Println("Published:", msg)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// 处理退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\nDisconnecting...")
	if err := client.unsubscribe(topic); err != nil {
		fmt.Println("Unsubscribe error:", err)
	}
	client.disconnect()
}

// 使用TLS连接到MQTT代理
func (c *MQTTClient) connect(broker, clientID string) error {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         "broker.emqx.io",
		InsecureSkipVerify: true, // 仅用于测试
	}

	conn, err := tls.Dial("tcp", broker, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection error: %w", err)
	}

	if err := conn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %w", err)
	}

	c.conn = conn

	// 构建并发送连接包
	// connectPacket := c.createConnectPacket(clientID)
	connectPacket := c.packet.connect(clientID)
	if _, err := c.conn.Write(connectPacket); err != nil {
		return err
	}
	return c.packet.connectAck(c.conn)

	// return c.awaitConnectResponse()
}

// 订阅主题
func (c *MQTTClient) subscribe(topic string) error {
	packet_id := c.nextPacketID()
	// packet := c.createSubscribePacket(packet_id, topic)
	packet := c.packet.subcribe(packet_id, topic)

	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	timeout := time.After(5 * time.Second)
	select {
	case packet := <-c.rx_buf:
		return c.packet.subcribeAck(packet, packet_id)

	case <-timeout:
		return errors.New("timeout waiting for SUBACK")
	}
	// return c.awaitSubAck(packet_id)
}

// 创建SUBSCRIBE包
func (c *MQTTClient) createSubscribePacket(packet_id uint16, topic string) []byte {
	topicBytes := []byte(topic)

	vHeader := []byte{
		byte(packet_id >> 8), byte(packet_id),
	}

	payload := []byte{
		byte(len(topicBytes) >> 8), byte(len(topicBytes)),
	}
	payload = append(payload, topicBytes...)
	payload = append(payload, 0) // QoS 0

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{SUBSCRIBE<<4 | 0x02}, fullPacket...)
}

// 发布消息
func (c *MQTTClient) publish(topic, message string) error {
	packet := c.createPublishPacket(topic, message)
	_, err := c.conn.Write(packet)
	return err
}

// 创建PUBLISH包
func (c *MQTTClient) createPublishPacket(topic, message string) []byte {
	topicBytes := []byte(topic)
	msgBytes := []byte(message)

	vHeader := []byte{
		byte(len(topicBytes) >> 8), byte(len(topicBytes)),
	}
	vHeader = append(vHeader, topicBytes...)

	payload := msgBytes

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{PUBLISH << 4}, fullPacket...)
}

// 发送PUBACK (QoS 1)
// func (c *MQTTClient) sendPuback(packet_id uint16) error {
// 	puback := []byte{
// 		0x40,                                  // PUBACK包类型和标志
// 		0x02,                                  // 剩余长度
// 		byte(packet_id >> 8), byte(packet_id), // 包ID
// 	}
// 	_, err := c.conn.Write(puback)
// 	return err
// }

// Ping检查
func (c *MQTTClient) ping() error {
	packet := []byte{PINGREQ << 4, 0}
	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, resp); err != nil {
		return err
	}

	if resp[0]>>4 != PINGRESP || resp[1] != 0 {
		return errors.New("invalid PINGRESP")
	}

	return nil
}

// 取消订阅
func (c *MQTTClient) unsubscribe(topic string) error {
	packet_id := c.nextPacketID()
	packet := c.createUnsubscribePacket(packet_id, topic)

	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	return c.awaitUnsubAck(packet_id)
}

// 创建UNSUBSCRIBE包
func (c *MQTTClient) createUnsubscribePacket(packet_id uint16, topic string) []byte {
	topicBytes := []byte(topic)

	vHeader := []byte{
		byte(packet_id >> 8), byte(packet_id),
	}

	payload := []byte{
		byte(len(topicBytes) >> 8), byte(len(topicBytes)),
	}
	payload = append(payload, topicBytes...)

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{UNSUBSCRIBE<<4 | 0x02}, fullPacket...)
}

// 等待取消订阅确认
func (c *MQTTClient) awaitUnsubAck(expectedID uint16) error {
	timeout := time.After(5 * time.Second)

	select {
	case packet := <-c.rx_buf:
		if len(packet) < 3 {
			return errors.New("invalid UNSUBACK packet")
		}

		packetType := packet[0] >> 4
		if packetType != UNSUBACK {
			return fmt.Errorf("expected UNSUBACK, got packet type %d", packetType)
		}

		pktID := binary.BigEndian.Uint16(packet[1:3])
		if pktID != expectedID {
			return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
		}

		return nil

	case <-timeout:
		return errors.New("timeout waiting for UNSUBACK")
	}
}

// 断开连接
func (c *MQTTClient) disconnect() {
	c.conn.Write([]byte{DISCONNECT << 4, 0})
}

// 资源清理
func (c *MQTTClient) cleanup() {
	if c.conn != nil {
		c.conn.Close()
		fmt.Println("Connection closed")
	}
}

// 消息处理器 - 核心路由
func (c *MQTTClient) producer() {
	for {
		header := make([]byte, 1)
		_, err := io.ReadFull(c.conn, header)
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				fmt.Println("Connection closed by peer")
				return
			}
			fmt.Println("Read header error:", err)
			return
		}

		// remaining, err := c.decodeLengthBytes()
		remaining, err := c.packet.remainLength(c.conn)
		if err != nil {
			fmt.Println("Length decode error:", err)
			return
		}

		packet_type := header[0] >> 4
		payload := make([]byte, remaining)
		if remaining > 0 {
			if _, err := io.ReadFull(c.conn, payload); err != nil {
				fmt.Println("Payload read error:", err)
				return
			}
		}

		switch packet_type {
		case PUBLISH:
			c.publishAck(header[0], payload)
		case SUBACK, UNSUBACK:
			c.rx_buf <- append([]byte{header[0]}, payload...)
		case PINGRESP:
			// 不做特殊处理
		case DISCONNECT:
			fmt.Println("Server requested disconnect")
			return
		default:
			fmt.Printf("Received unexpected packet type: %d\n", packet_type)
		}
	}
}

// 处理PUBLISH消息
func (c *MQTTClient) publishAck(header byte, payload []byte) {
	topic, qos, start, err := c.packet.payload(header, payload)
	if err != nil {
		return
	}

	// 根据QoS级别处理
	switch qos {
	case 0: // QoS 0
		if start > len(payload) {
			fmt.Println("Invalid message start for QoS 0")
			return
		}
		message := string(payload[start:])
		c.tx_buf <- MQTTMsg{topic: topic, message: message}

	case 1: // QoS 1
		if len(payload) < start+2 {
			fmt.Println("Invalid QoS 1 packet")
			return
		}
		message := string(payload[start+2:])
		c.tx_buf <- MQTTMsg{topic: topic, message: message}

		id := binary.BigEndian.Uint16(payload[start:])
		// 发送PUBACK响应
		// if err := c.sendPuback(id); err != nil {
		if _, err := c.conn.Write(c.packet.publishAck(id)); err != nil {
			fmt.Printf("Failed to send PUBACK: %v\n", err)
		} else {
			fmt.Printf("Sent PUBACK for packet ID: %d\n", id)
		}

	case 2: // QoS 2
		fmt.Println("QoS 2 not supported in this implementation")
		return

	default:
		fmt.Printf("Invalid QoS level: %d\n", qos)
		return
	}
}

// 消息消费者
func (c *MQTTClient) consumer() {
	for msg := range c.tx_buf {
		fmt.Printf("> Received [%s]: %s\n", msg.topic, msg.message)
	}
}

// 获取下一个Packet ID
func (c *MQTTClient) nextPacketID() uint16 {
	id := c.packet_id
	c.packet_id++
	if c.packet_id == 0 { // 避免零值
		c.packet_id = 1
	}
	return id
}

// 创建CONNECT包
// func (c *MQTTClient) createConnectPacket(clientID string) []byte {
// 	protocolName := "MQTT"
// 	flags := CLEAN_SESSION
// 	keepAlive := KEEP_ALIVE
//
// 	// 可变头
// 	vHeader := []byte{
// 		byte(len(protocolName) >> 8), byte(len(protocolName)),
// 	}
// 	fmt.Println(vHeader)
// 	vHeader = append(vHeader, []byte(protocolName)...)
// 	vHeader = append(vHeader, 0x04) // 协议版本
// 	vHeader = append(vHeader, byte(flags))
// 	vHeader = append(vHeader, byte(keepAlive>>8), byte(keepAlive))
// 	fmt.Println(vHeader)
//
// 	payload := []byte{
// 		byte(len(clientID) >> 8), byte(len(clientID)),
// 	}
// 	fmt.Println("payload:", payload)
// 	payload = append(payload, []byte(clientID)...)
//
// 	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
// 	fullPacket = append(fullPacket, payload...)
// 	return append([]byte{CONNECT << 4}, fullPacket...)
// }

// // 等待连接响应
// func (c *MQTTClient) awaitConnectResponse() error {
// 	fmt.Println(time.Now().UnixMilli())
// 	header := make([]byte, 2)
// 	if _, err := io.ReadFull(c.conn, header); err != nil {
// 		return fmt.Errorf("read header: %w", err)
// 	}
// 	fmt.Println(time.Now().UnixMilli(), header)
//
// 	if header[0]>>4 != CONNACK {
// 		return errors.New("invalid CONNACK packet")
// 	}
//
// 	if header[1] != 2 {
// 		return errors.New("invalid CONNACK remaining length")
// 	}
//
// 	vHeader := make([]byte, 2)
// 	if _, err := io.ReadFull(c.conn, vHeader); err != nil {
// 		return fmt.Errorf("read vHeader: %w", err)
// 	}
//
// 	if vHeader[1] != 0 {
// 		return fmt.Errorf("connection refused with code %d", vHeader[1])
// 	}
//
// 	return nil
// }

// // 从连接解码长度
// func (c *MQTTClient) decodeLengthBytes() (int, error) {
// 	multiplier := 1
// 	length := 0
// 	bytesRead := 0
//
// 	for bytesRead < 4 {
// 		digitBuf := make([]byte, 1)
// 		if _, err := io.ReadFull(c.conn, digitBuf); err != nil {
// 			if err == io.EOF && bytesRead > 0 {
// 				return 0, errors.New("unexpected EOF while reading length")
// 			}
// 			return 0, err
// 		}
//
// 		bytesRead++
// 		digit := digitBuf[0]
//
// 		length += int(digit&0x7F) * multiplier
//
// 		if digit&0x80 == 0 {
// 			break
// 		}
//
// 		multiplier *= 128
// 		if multiplier > 128*128*128 {
// 			return 0, errors.New("length too large")
// 		}
// 	}
//
// 	return length, nil
// }
//
// // 等待订阅确认
// func (c *MQTTClient) awaitSubAck(expectedID uint16) error {
// 	timeout := time.After(5 * time.Second)
//
// 	select {
// 	case packet := <-c.rx_buf:
// 		if len(packet) < 3 {
// 			return errors.New("invalid SUBACK packet")
// 		}
//
// 		packetType := packet[0] >> 4
// 		if packetType != SUBACK {
// 			return fmt.Errorf("expected SUBACK, got packet type %d", packetType)
// 		}
//
// 		pktID := binary.BigEndian.Uint16(packet[1:3])
// 		if pktID != expectedID {
// 			return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
// 		}
//
// 		if len(packet) < 4 {
// 			return errors.New("missing return codes in SUBACK")
// 		}
//
// 		for i, code := range packet[3:] {
// 			switch code {
// 			case 0x00, 0x01, 0x02:
// 				// QoS等级有效
// 			case 0x80:
// 				return fmt.Errorf("subscription failed for topic #%d", i+1)
// 			default:
// 				return fmt.Errorf("invalid return code: 0x%x", code)
// 			}
// 		}
//
// 		return nil
//
// 	case <-timeout:
// 		return errors.New("timeout waiting for SUBACK")
// 	}
// }

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
