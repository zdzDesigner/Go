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
	vHeader := []byte{
		byte(len(protocolName) >> 8), byte(len(protocolName)),
	}
	fmt.Println(vHeader)
	vHeader = append(vHeader, []byte(protocolName)...)
	vHeader = append(vHeader, 0x04) // 协议版本
	vHeader = append(vHeader, byte(flags))
	vHeader = append(vHeader, byte(keepAlive>>8), byte(keepAlive))
	fmt.Println(vHeader)

	payload := []byte{
		byte(len(client_id) >> 8), byte(len(client_id)),
	}
	fmt.Println("payload:", payload)
	payload = append(payload, []byte(client_id)...)

	fullPacket := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
	fullPacket = append(fullPacket, payload...)
	return append([]byte{CONNECT << 4}, fullPacket...)
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

	vHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, vHeader); err != nil {
		return fmt.Errorf("read vHeader: %w", err)
	}

	if vHeader[1] != 0 {
		return fmt.Errorf("connection refused with code %d", vHeader[1])
	}

	return nil
}

// packet 剩余长度
func (p *Packet) remainLength(conn net.Conn) (int, error) {
	extend := 1
	length := 0
	bytes := 0

	// 最大4字节
	for bytes < 4 {
		buf := make([]byte, 1)
		if _, err := io.ReadFull(conn, buf); err != nil {
			if err == io.EOF && bytes > 0 {
				return 0, errors.New("unexpected EOF while reading length")
			}
			return 0, err
		}

		bytes++
		digit := buf[0]
		length += int(digit&0x7F) * extend

		if digit&0x80 == 0 {
			break
		}

		extend *= 128
		if extend > 128*128*128 {
			return 0, errors.New("length too large")
		}
	}

	return length, nil
}

func (p *Packet) subcribe(packet_id uint16, topic string) []byte {
	topic_bytes := []byte(topic)

	vHeader := []byte{
		byte(packet_id >> 8), byte(packet_id),
	}

	payload := []byte{
		byte(len(topic_bytes) >> 8), byte(len(topic_bytes)),
	}
	payload = append(payload, topic_bytes...)
	payload = append(payload, 0) // QoS 0

	full_packet := append(encodeLength(len(vHeader)+len(payload)), vHeader...)
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

func (p *Packet) publish(topic, message string) []byte {
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

func (p *Packet) publishAck(packet_id uint16) []byte {
	return []byte{
		0x40,                                  // PUBACK包类型和标志
		0x02,                                  // 剩余长度
		byte(packet_id >> 8), byte(packet_id), // 包ID
	}
}
