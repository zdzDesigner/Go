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

// 客户端结构添加缓冲
type MQTTClient struct {
	conn       net.Conn
	packetID   uint16
	pendingAck chan []byte  // 用于缓冲非预期的控制包
	pendingPub chan mqttMsg // 用于缓冲PUBLISH消息
}

type mqttMsg struct {
	topic   string
	message string
}

func main() {
	client := &MQTTClient{
		packetID:   1,
		pendingAck: make(chan []byte, 10),
		pendingPub: make(chan mqttMsg, 10),
	}
	defer client.cleanup()

	// if err := client.connect("broker.emqx.io:1883", "go-optimized-client"); err != nil {
	if err := client.connect("172.16.40.51:19992", "go-optimized-client"); err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	fmt.Println("Connected to MQTT broker")

	// 启动消息处理器
	go client.messageProcessor()

	// 启动消息消费者
	go client.getMessages()

	messageChan := make(chan string, 10)
	go client.messageReceiver(messageChan)

	topic := "app/ready/notify"
	if err := client.subscribe(topic); err != nil {
		fmt.Println("Subscribe error:", err)
		return
	}
	fmt.Println("Subscribed to topic:", topic)

	// go func() {
	// 	for i := 0; i < 5; i++ {
	// 		msg := fmt.Sprintf("Message %d at %s", i, time.Now().Format("15:04:05"))
	// 		if err := client.publish(topic, msg); err != nil {
	// 			fmt.Println("Publish error:", err)
	// 		} else {
	// 			fmt.Println("Published:", msg)
	// 		}
	// 		time.Sleep(2 * time.Second)
	// 	}
	// }()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg := <-messageChan:
			fmt.Printf("\n> Received: [%s] %s\n", topic, msg)

		case <-time.After(30 * time.Second):
			if err := client.ping(); err != nil {
				fmt.Println("Ping error:", err)
				break
			}
			fmt.Println("Ping successful")

		case sig := <-exit:
			fmt.Printf("\nReceived signal: %s. Disconnecting...\n", sig)
			// if err := client.unsubscribe(topic); err != nil {
			// 	fmt.Println("Unsubscribe error:", err)
			// }
			// client.disconnect()
			return
		}
	}
}

// 连接到MQTT代理
func (c *MQTTClient) connect(broker, clientID string) error {
	// conn, err := net.Dial("tcp", broker)
	// if err != nil {
	// 	return err
	// }
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         "broker.emqx.io", // 匹配服务器证书域名
		InsecureSkipVerify: true,             // 仅限测试环境使用
	}

	conn, err := tls.Dial("tcp", broker, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection error: %w", err)
	}

	// 显式握手确保立即检测连接问题
	if err := conn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %w", err)
	}
	c.conn = conn

	connectPacket := c.createConnectPacket(clientID)
	if _, err := conn.Write(connectPacket); err != nil {
		return err
	}

	if err := c.awaitConnectResponse(); err != nil {
		return err
	}

	return nil
}

// 创建CONNECT包
func (c *MQTTClient) createConnectPacket(clientID string) []byte {
	protocolName := "MQTT"
	flags := CLEAN_SESSION
	keepAlive := KEEP_ALIVE

	vHeader := []byte{
		byte(len(protocolName) >> 8), byte(len(protocolName)),
	}
	vHeader = append(vHeader, []byte(protocolName)...)
	vHeader = append(vHeader, 0x04)
	vHeader = append(vHeader, byte(flags))
	vHeader = append(vHeader, byte(keepAlive>>8), byte(keepAlive))

	payload := []byte{
		byte(len(clientID) >> 8), byte(len(clientID)),
	}
	payload = append(payload, []byte(clientID)...)

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{CONNECT << 4}, fullPacket...)
}

// 等待连接响应
func (c *MQTTClient) awaitConnectResponse() error {
	// 确保读取完整2字节头
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	if header[0]>>4 != CONNACK {
		return errors.New("invalid CONNACK packet")
	}

	// CONNACK剩余长度固定为2
	if header[1] != 2 {
		return errors.New("invalid CONNACK remaining length")
	}

	vHeader := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, vHeader); err != nil {
		return fmt.Errorf("read vHeader: %w", err)
	}

	// 检查返回代码 (0 = 成功)
	if vHeader[1] != 0 {
		return fmt.Errorf("connection refused with code %d", vHeader[1])
	}

	return nil
}

// 订阅主题
func (c *MQTTClient) subscribe(topic string) error {
	packetID := c.nextPacketID()
	packet := c.createSubscribePacket(packetID, topic)

	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	return c.awaitSubAck(packetID)
}

// 创建SUBSCRIBE包
func (c *MQTTClient) createSubscribePacket(packetID uint16, topic string) []byte {
	topicBytes := []byte(topic)

	vHeader := []byte{
		byte(packetID >> 8), byte(packetID),
	}

	payload := []byte{
		byte(len(topicBytes) >> 8), byte(len(topicBytes)),
	}
	payload = append(payload, topicBytes...)
	payload = append(payload, 0)

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{SUBSCRIBE<<4 | 0x02}, fullPacket...)
}

// 等待订阅确认
// func (c *MQTTClient) awaitSubAck(expectedID uint16) error {
// 	// 读取包类型
// 	typeByte := make([]byte, 1)
// 	if _, err := io.ReadFull(c.conn, typeByte); err != nil {
// 		return err
// 	}
//
// 	fmt.Println(typeByte)
// 	if typeByte[0]>>4 != SUBACK {
// 		return errors.New("invalid SUBACK packet")
// 	}
//
// 	// 解码剩余长度
// 	remaining, err := c.decodeLengthBytes()
// 	if err != nil {
// 		return err
// 	}
//
// 	// 读取剩余部分
// 	payload := make([]byte, remaining)
// 	if _, err := io.ReadFull(c.conn, payload); err != nil {
// 		return err
// 	}
//
// 	if len(payload) < 2 {
// 		return errors.New("invalid SUBACK payload")
// 	}
//
// 	pktID := binary.BigEndian.Uint16(payload[:2])
// 	if pktID != expectedID {
// 		return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
// 	}
//
// 	return nil
// }

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
	packetID := c.nextPacketID()
	packet := c.createUnsubscribePacket(packetID, topic)

	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	return c.awaitUnsubAck(packetID)
}

// 创建UNSUBSCRIBE包
func (c *MQTTClient) createUnsubscribePacket(packetID uint16, topic string) []byte {
	topicBytes := []byte(topic)

	vHeader := []byte{
		byte(packetID >> 8), byte(packetID),
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
	typeByte := make([]byte, 1)
	if _, err := io.ReadFull(c.conn, typeByte); err != nil {
		return err
	}

	if typeByte[0]>>4 != UNSUBACK {
		return errors.New("invalid UNSUBACK packet")
	}

	remaining, err := c.decodeLengthBytes()
	if err != nil {
		return err
	}

	payload := make([]byte, remaining)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return err
	}

	if len(payload) < 2 {
		return errors.New("invalid UNSUBACK payload")
	}

	pktID := binary.BigEndian.Uint16(payload)
	if pktID != expectedID {
		return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
	}

	return nil
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

// 消息接收器 (彻底修复)
func (c *MQTTClient) messageReceiver(ch chan<- string) {
	for {
		// 读取固定头（包类型）
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

		packetType := header[0] >> 4

		// 解码剩余长度
		remaining, err := c.decodeLengthBytes()
		if err != nil {
			fmt.Println("Length decode error:", err)
			return
		}

		// 没有剩余数据时跳过读取
		if remaining == 0 {
			continue
		}

		// 读取完整包有效载荷
		payload := make([]byte, remaining)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			fmt.Println("Payload read error:", err)
			return
		}

		// 处理PUBLISH包
		if packetType == PUBLISH {
			// 确保足够数据提取主题
			if remaining < 2 {
				fmt.Println("Invalid PUBLISH packet - too short")
				continue
			}

			topicLen := binary.BigEndian.Uint16(payload[:2])
			payloadStart := 2 + int(topicLen)

			// 验证主题长度
			if payloadStart > remaining {
				fmt.Println("Invalid topic length")
				continue
			}

			// 提取消息
			message := string(payload[payloadStart:])
			ch <- message
		}
	}
}

// 从连接解码长度 (安全方式)
func (c *MQTTClient) decodeLengthBytes() (int, error) {
	multiplier := 1
	length := 0
	bytesRead := 0

	for bytesRead < 4 { // MQTT长度最多4字节
		digitBuf := make([]byte, 1)
		if _, err := io.ReadFull(c.conn, digitBuf); err != nil {
			// 处理EOF
			if err == io.EOF && bytesRead > 0 {
				return 0, errors.New("unexpected EOF while reading length")
			}
			return 0, err
		}

		bytesRead++
		digit := digitBuf[0]

		// 累计值计算
		length += int(digit&0x7F) * multiplier

		// 检查连续位
		if (digit & 0x80) == 0 {
			break
		}

		// 更新乘数
		multiplier *= 128
		// 防止整数溢出
		if multiplier > 128*128*128 {
			return 0, errors.New("length too large")
		}
	}

	return length, nil
}

// 获取下一个Packet ID
func (c *MQTTClient) nextPacketID() uint16 {
	id := c.packetID
	c.packetID++
	if c.packetID == 0 {
		c.packetID = 1
	}
	return id
}

// 核心消息处理循环
func (c *MQTTClient) messageProcessor() {
	for {
		// 读取固定头
		header := make([]byte, 1)
		if _, err := io.ReadFull(c.conn, header); err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				fmt.Println("Connection closed by peer")
				return
			}
			fmt.Println("Read header error:", err)
			return
		}

		packetType := header[0] >> 4

		// 解码剩余长度
		remaining, err := c.decodeLengthBytes()
		if err != nil {
			fmt.Println("Length decode error:", err)
			return
		}

		// 读取完整包有效载荷
		payload := make([]byte, remaining)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			fmt.Println("Payload read error:", err)
			return
		}

		// 包路由分发
		switch packetType {
		case PUBLISH:
			c.handlePublish(header[0], payload)
		case SUBACK, UNSUBACK:
			c.pendingAck <- append([]byte{header[0]}, payload...)
		case PINGRESP:
			// 心跳响应不处理
		default:
			fmt.Printf("Received unexpected packet type: %d\n", packetType)
		}
	}
}

// 处理PUBLISH消息
func (c *MQTTClient) handlePublish(header byte, payload []byte) {
	if len(payload) < 2 {
		fmt.Println("Invalid PUBLISH packet")
		return
	}

	topicLen := binary.BigEndian.Uint16(payload[:2])
	payloadStart := 2 + int(topicLen)

	if payloadStart > len(payload) {
		fmt.Println("Invalid topic length")
		return
	}

	// 提取主题和消息
	topic := string(payload[2:payloadStart])
	message := string(payload[payloadStart:])

	// QoS处理 (示例只处理QoS0)
	qos := (header & 0x06) >> 1 // 提取QoS位
	switch qos {
	case 0: // QoS 0
		c.pendingPub <- mqttMsg{topic: topic, message: message}
	case 1: // QoS 1 (至少一次)
		c.pendingPub <- mqttMsg{topic: topic, message: message}
		// 发送PUBACK确认
		// pktID := binary.BigEndian.Uint16(payload[2+int(topicLen) : 2+int(topicLen)+2])
		// c.sendPuback(pktID)
	case 2: // QoS 2 (确保一次) - 需要完整实现
		fmt.Println("QoS 2 not supported")
	}
}

// 等待SUBACK (适配新架构)
func (c *MQTTClient) awaitSubAck(expectedID uint16) error {
	timeout := time.After(5 * time.Second)

	for {
		select {
		case packet := <-c.pendingAck:
			packetType := packet[0] >> 4
			if packetType != SUBACK {
				continue
			}

			if len(packet) < 3 { // 1字节头 + 2字节包ID
				return errors.New("invalid SUBACK payload")
			}

			pktID := binary.BigEndian.Uint16(packet[1:3])
			if pktID != expectedID {
				return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
			}

			// 检查返回码
			if len(packet) < 4 {
				return errors.New("missing return codes in SUBACK")
			}

			for i, code := range packet[3:] {
				if code == 0x80 {
					return fmt.Errorf("subscription failed for topic #%d", i+1)
				} else if code > 0x02 {
					return fmt.Errorf("invalid return code: 0x%x", code)
				}
			}

			return nil

		case <-timeout:
			return errors.New("timeout waiting for SUBACK")
		}
	}
}

// 在订阅后获取消息
func (c *MQTTClient) getMessages() {
	for {
		select {
		case msg := <-c.pendingPub:
			fmt.Printf("Received: [%s] %s\n", msg.topic, msg.message)
		}
	}
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
