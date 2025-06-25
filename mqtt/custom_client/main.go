package main

import (
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

type MQTTClient struct {
	conn     net.Conn
	packetID uint16
}

func main() {
	client := &MQTTClient{
		packetID: 1,
	}
	defer client.cleanup()

	if err := client.connect("broker.emqx.io:1883", "go-optimized-client"); err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	fmt.Println("Connected to MQTT broker")

	messageChan := make(chan string, 10)
	go client.messageReceiver(messageChan)

	topic := "opt/test"
	if err := client.subscribe(topic); err != nil {
		fmt.Println("Subscribe error:", err)
		return
	}
	fmt.Println("Subscribed to topic:", topic)

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

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg := <-messageChan:
			fmt.Printf("\n> Received: [%s] %s\n", topic, msg)

		// case <-time.After(30 * time.Second):
		// 	if err := client.ping(); err != nil {
		// 		fmt.Println("Ping error:", err)
		// 		break loop
		// 	}
		// 	fmt.Println("Ping successful")

		case sig := <-exit:
			fmt.Printf("\nReceived signal: %s. Disconnecting...\n", sig)
			// break loop
		}
	}

	if err := client.unsubscribe(topic); err != nil {
		fmt.Println("Unsubscribe error:", err)
	}
	client.disconnect()
}

// 连接到MQTT代理
func (c *MQTTClient) connect(broker, clientID string) error {
	conn, err := net.Dial("tcp", broker)
	if err != nil {
		return err
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
func (c *MQTTClient) awaitSubAck(expectedID uint16) error {
	// 读取包类型
	typeByte := make([]byte, 1)
	if _, err := io.ReadFull(c.conn, typeByte); err != nil {
		return err
	}

	if typeByte[0]>>4 != SUBACK {
		return errors.New("invalid SUBACK packet")
	}

	// 解码剩余长度
	remaining, err := c.decodeLengthBytes()
	if err != nil {
		return err
	}

	// 读取剩余部分
	payload := make([]byte, remaining)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return err
	}

	if len(payload) < 2 {
		return errors.New("invalid SUBACK payload")
	}

	pktID := binary.BigEndian.Uint16(payload[:2])
	if pktID != expectedID {
		return fmt.Errorf("packet ID mismatch: expected %d, got %d", expectedID, pktID)
	}

	return nil
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
