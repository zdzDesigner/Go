package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type Packet struct{}

func (p *Packet) connect(client_id string) []byte {
	protocolName := "MQTT"
	flags := CLEAN_SESSION
	keepAlive := KEEP_ALIVE

	// 可变头
	var_header := []byte{
		byte(len(protocolName) >> 8), byte(len(protocolName)),
	}
	fmt.Println(var_header)
	var_header = append(var_header, []byte(protocolName)...)
	var_header = append(var_header, 0x04) // 协议版本
	var_header = append(var_header, byte(flags))
	var_header = append(var_header, byte(keepAlive>>8), byte(keepAlive))
	fmt.Println(var_header)

	payload := []byte{
		byte(len(client_id) >> 8), byte(len(client_id)),
	}
	fmt.Println("payload:", payload)
	payload = append(payload, []byte(client_id)...)

	full_packet := append(encodeLength(len(var_header)+len(payload)), var_header...)
	full_packet = append(full_packet, payload...)
	return append([]byte{CONNECT << 4}, full_packet...)
}

func (p *Packet) connectAck(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	if header[0]>>4 != CONNACK {
		return errors.New("invalid CONNACK packet")
	}

	if header[1] != 2 {
		return errors.New("invalid CONNACK remaining length")
	}

	var_header := make([]byte, 2)
	if _, err := io.ReadFull(conn, var_header); err != nil {
		return fmt.Errorf("read var_header: %w", err)
	}

	if var_header[1] != 0 {
		return fmt.Errorf("connection refused with code %d", var_header[1])
	}

	return nil
}

// packet 剩余长度 - 改进版，更安全的长度解码
func (p *Packet) remainLength(conn net.Conn) (int, error) {
	multiplier := 1
	length := 0
	bytesRead := 0

	for bytesRead < 4 { // MQTT长度最多4字节
		digitBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, digitBuf); err != nil {
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

// func (p *Packet) subcribe(packet_id uint16, topic string) []byte {
func (p *Packet) subcribe(packet_id uint16, topic Topic) []byte {
	topic_bytes := []byte(topic.Name)

	var_header := []byte{
		byte(packet_id >> 8), byte(packet_id), // README.md(## Packet ID)
	}

	payload := []byte{
		byte(len(topic_bytes) >> 8), byte(len(topic_bytes)),
	}
	payload = append(payload, topic_bytes...)
	payload = append(payload, topic.QOS) // QoS 0

	full_packet := append(encodeLength(len(var_header)+len(payload)), var_header...)
	full_packet = append(full_packet, payload...)
	return append([]byte{SUBSCRIBE<<4 | 0x02}, full_packet...)
}

func (p *Packet) subcribeAck(packet []byte, packet_id uint16) error {
	if len(packet) < 3 {
		return errors.New("invalid SUBACK packet")
	}

	packetType := packet[0] >> 4
	if packetType != SUBACK {
		return fmt.Errorf("expected SUBACK, got packet type %d", packetType)
	}

	id := binary.BigEndian.Uint16(packet[1:3])
	if id != packet_id {
		return fmt.Errorf("packet ID mismatch: expected %d, got %d", packet_id, id)
	}

	if len(packet) < 4 {
		return errors.New("missing return codes in SUBACK")
	}

	for i, code := range packet[3:] {
		switch code {
		case 0x00, 0x01, 0x02:
			// QoS等级有效
		case 0x80:
			return fmt.Errorf("subscription failed for topic #%d", i+1)
		default:
			return fmt.Errorf("invalid return code: 0x%x", code)
		}
	}

	return nil
}

func (p *Packet) unsubscribe(packet_id uint16, topic string) []byte {
	topicBytes := []byte(topic)

	var_header := []byte{
		byte(packet_id >> 8), byte(packet_id),
	}

	payload := []byte{
		byte(len(topicBytes) >> 8), byte(len(topicBytes)),
	}
	payload = append(payload, topicBytes...)

	full_packet := append(encodeLength(len(var_header)+len(payload)), var_header...)
	full_packet = append(full_packet, payload...)
	return append([]byte{UNSUBSCRIBE<<4 | 0x02}, full_packet...)
}

// 负载解析
func (p *Packet) parsePayload(header byte, payload []byte) (topic string, qos byte, start int, err error) {
	// 确保包格式正确
	if len(payload) < 2 {
		fmt.Println("Invalid PUBLISH packet - too short")
		err = errors.New("Invalid PUBLISH packet - too short")
		return
	}

	// 提取主题长度
	topic_len := binary.BigEndian.Uint16(payload[:2])
	if int(2+topic_len) > len(payload) {
		fmt.Println("Invalid topic length")
		err = errors.New("Invalid PUBLISH packet - too short")
		return
	}

	topic = string(payload[2 : 2+topic_len]) // 提取主题
	qos = (header & 0x06) >> 1               // 提取QoS等级
	start = 2 + int(binary.BigEndian.Uint16(payload[:2]))

	return
}

func (p *Packet) publishWithPacketID(topic Topic, message string, packet_id uint16) []byte {
	topic_bytes := []byte(topic.Name)
	msg_bytes := []byte(message)

	// 构建固定头部，包含QoS, DUP, RETAIN标志位
	var fixed_header byte = (PUBLISH << 4)  // 包类型
	fixed_header |= (topic.Dup & 0x01) << 3 // DUP标志 (bit 3)
	fixed_header |= (topic.QOS & 0x03) << 1 // QoS级别 (bits 2,1)
	fixed_header |= (topic.Retain & 0x01)   // RETAIN标志 (bit 0)

	var_header := []byte{
		byte(len(topic_bytes) >> 8), byte(len(topic_bytes)), // 长度, 因为占2个字节，所以>>8
	}
	var_header = append(var_header, topic_bytes...)

	// 如果QoS > 0，添加包ID
	var payload []byte
	if topic.QOS > 0 {
		var_header = append(var_header, byte(packet_id>>8), byte(packet_id))
		payload = msg_bytes
	} else {
		payload = msg_bytes
	}

	full_packet := append(encodeLength(len(var_header)+len(payload)), var_header...)
	return append([]byte{fixed_header}, full_packet...)
}

// 保留旧函数用于向后兼容
func (p *Packet) publish(topic Topic, message string) []byte {
	return p.publishWithPacketID(topic, message, 1) // 默认包ID为1，仅用于向后兼容QoS 0
}

func (p *Packet) publishAck(packet_id uint16) []byte {
	return []byte{
		PUBACK << 4,                           // PUBACK包类型和标志
		0x02,                                  // 剩余长度
		byte(packet_id >> 8), byte(packet_id), // 包ID
	}
}

// 创建PUBREC包 (用于QoS 2流程)
func (p *Packet) publishRec(packet_id uint16) []byte {
	return []byte{
		PUBREC << 4,                           // PUBREC包类型和标志
		0x02,                                  // 剩余长度
		byte(packet_id >> 8), byte(packet_id), // 包ID
	}
}

// 创建PUBREL包 (用于QoS 2流程)
func (p *Packet) publishRel(packet_id uint16) []byte {
	return []byte{
		PUBREL << 4,                           // PUBREL包类型和标志
		0x02,                                  // 剩余长度
		byte(packet_id >> 8), byte(packet_id), // 包ID
	}
}

// 创建PUBCOMP包 (用于QoS 2流程)
func (p *Packet) publishComp(packet_id uint16) []byte {
	return []byte{
		PUBCOMP << 4,                          // PUBCOMP包类型和标志
		0x02,                                  // 剩余长度
		byte(packet_id >> 8), byte(packet_id), // 包ID
	}
}
