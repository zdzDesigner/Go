package broker

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"battery/internal/connection"
	"battery/internal/router"
)

const (
	Connect     = 1
	Connack     = 2
	Publish     = 3
	Puback      = 4
	Pubrec      = 5
	Pubrel      = 6
	Pubcomp     = 7
	Subscribe   = 8
	Suback      = 9
	Unsubscribe = 10
	Unsuback    = 11
	Pingreq     = 12
	Pingresp    = 13
	Disconnect  = 14
)

type Broker struct {
	listener    net.Listener
	clients     sync.Map
	messages    chan *Message
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	connManager *connection.ConnectionManager
	topicRouter *router.TopicMatcher
}

type Message struct {
	Topic string
	Value []byte
	QoS   byte
}

type ClientConnection struct {
	ID        string
	Conn      net.Conn
	CreatedAt time.Time
	mu        sync.RWMutex
}

func NewBroker(address string, maxConnections int) (*Broker, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := &Broker{
		listener:    listener,
		messages:    make(chan *Message, 10000),
		ctx:         ctx,
		cancel:      cancel,
		connManager: connection.NewConnectionManager(maxConnections),
		topicRouter: router.NewTopicMatcher(),
	}

	return broker, nil
}

func (b *Broker) Start() {
	log.Printf("MQTT Broker starting on %s", b.listener.Addr().String())

	go b.connManager.Monitor(10 * time.Second)

	b.wg.Add(1)
	go b.processMessages()

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			conn, err := b.listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				continue
			}

			b.wg.Add(1)
			go b.handleConnection(conn)
		}
	}
}

func (b *Broker) handleConnection(conn net.Conn) {
	defer b.wg.Done()
	defer conn.Close()

	packet, err := b.readPacket(conn)
	if err != nil {
		log.Printf("Read packet error: %v", err)
		return
	}

	if packet.Type != Connect {
		log.Printf("Expected CONNECT packet, got type %d", packet.Type)
		return
	}

	clientID, err := parseClientID(packet.Data)
	if err != nil {
		log.Printf("Parse client ID error: %v", err)
		return
	}

	if clientID == "" {
		clientID = generateClientID()
	}

	if err := b.connManager.RegisterConnection(clientID); err != nil {
		log.Printf("Connection refused for %s: %v", clientID, err)
		conn.Write([]byte{Connack<<4 | 0, 2, 0, 5}) // Connection Refused: not authorized
		return
	}
	defer func() {
		b.connManager.DeregisterConnection(clientID)
	}()

	conn.Write([]byte{Connack<<4 | 0, 2, 0, 0}) // Connection Accepted

	client := &ClientConnection{
		ID:        clientID,
		Conn:      conn,
		CreatedAt: time.Now(),
	}

	b.clients.Store(clientID, client)
	defer func() {
		b.clients.Delete(clientID)
	}()

	log.Printf("Client connected: %s. Active connections: %d",
		clientID, b.connManager.GetActiveConnections())

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			packet, err := b.readPacket(conn)
			if err != nil {
				if err != io.EOF {
					log.Printf("Client %s read error: %v", clientID, err)
				}
				return
			}

			if err := b.handlePacket(client, packet); err != nil {
				log.Printf("Handle packet error for client %s: %v", clientID, err)
				return
			}
		}
	}
}

func (b *Broker) readPacket(conn net.Conn) (packet struct {
	Type byte
	Data []byte
}, err error,
) {
	reader := bufio.NewReader(conn)

	firstByte, err := reader.ReadByte()
	if err != nil {
		return packet, err
	}

	packet.Type = firstByte >> 4
	_ = firstByte & 0x0F // typeFlags - unused

	multiplier := uint32(1)
	length := uint32(0)
	var digit byte

	for {
		digit, err = reader.ReadByte()
		if err != nil {
			return packet, err
		}

		length += uint32(digit&127) * multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier *= 128

		if multiplier > 128*128*128 {
			return packet, fmt.Errorf("malformed variable length")
		}
	}

	packet.Data = make([]byte, length)
	if length > 0 {
		_, err = io.ReadFull(reader, packet.Data)
		if err != nil {
			return packet, err
		}
	}

	packet.Type = firstByte >> 4

	return packet, nil
}

func parseClientID(data []byte) (string, error) {
	if len(data) < 10 {
		return "", fmt.Errorf("CONNECT packet too short")
	}

	protoNameLen := binary.BigEndian.Uint16(data[0:2])
	if protoNameLen != 4 || string(data[2:6]) != "MQTT" {
		return "", fmt.Errorf("invalid protocol name")
	}

	clientIDOffset := 6 + 4 // protocol name + protocol level + connect flags + keep alive
	if clientIDOffset+2 > len(data) {
		return "", fmt.Errorf("CONNECT packet too short for client ID")
	}

	clientIDLen := binary.BigEndian.Uint16(data[clientIDOffset : clientIDOffset+2])
	if clientIDOffset+2+int(clientIDLen) > len(data) {
		return "", fmt.Errorf("CONNECT packet too short for client ID data")
	}

	clientID := string(data[clientIDOffset+2 : clientIDOffset+2+int(clientIDLen)])

	return clientID, nil
}

func (b *Broker) handlePacket(client *ClientConnection, packet interface{}) error {
	pkt, ok := packet.(struct {
		Type byte
		Data []byte
	})
	if !ok {
		log.Printf("Invalid packet type: %T", packet)
		return nil
	}

	switch pkt.Type {
	case Publish:
		b.connManager.IncrementMessageCounter()
		return b.handlePublish(client, pkt.Data)
	case Subscribe:
		return b.handleSubscribe(client, pkt.Data)
	case Unsubscribe:
		return b.handleUnsubscribe(client, pkt.Data)
	case Pingreq:
		client.Conn.Write([]byte{Pingresp<<4 | 0, 0})
		return nil
	case Disconnect:
		log.Printf("Client %s disconnected", client.ID)
		return nil
	default:
		log.Printf("Unknown packet type: %d", pkt.Type)
		return nil
	}
}

func (b *Broker) handlePublish(client *ClientConnection, data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("publish packet too short")
	}

	topicLength := int(binary.BigEndian.Uint16(data[0:2]))
	if len(data) < 2+topicLength {
		return fmt.Errorf("publish packet too short for topic")
	}

	topic := string(data[2 : 2+topicLength])
	payload := data[2+topicLength:]

	msg := &Message{
		Topic: topic,
		Value: payload,
	}

	select {
	case b.messages <- msg:
	default:
		log.Printf("Message buffer full, dropping message for topic %s", topic)
	}

	return nil
}

func (b *Broker) handleSubscribe(client *ClientConnection, data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("subscribe packet too short")
	}

	packetID := binary.BigEndian.Uint16(data[0:2])
	offset := 2

	var topics []string
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}
		topicLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2

		if offset+topicLength > len(data) {
			break
		}

		topic := string(data[offset : offset+topicLength])
		topics = append(topics, topic)
		offset += topicLength

		offset++ // Skip QoS byte
	}

	for _, topic := range topics {
		b.topicRouter.Subscribe(topic, client.ID)
	}

	response := make([]byte, 4+len(topics))
	response[0] = Suback<<4 | 0
	response[1] = byte(2 + len(topics))
	binary.BigEndian.PutUint16(response[2:4], packetID)

	for i := range topics {
		response[4+i] = 0 // QoS granted
	}

	client.Conn.Write(response)

	return nil
}

func (b *Broker) handleUnsubscribe(client *ClientConnection, data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("unsubscribe packet too short")
	}

	packetID := binary.BigEndian.Uint16(data[0:2])
	offset := 2

	var topics []string
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}
		topicLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2

		if offset+topicLength > len(data) {
			break
		}

		topic := string(data[offset : offset+topicLength])
		topics = append(topics, topic)
		offset += topicLength
	}

	for _, topic := range topics {
		b.topicRouter.Unsubscribe(topic, client.ID)
	}

	response := make([]byte, 4)
	response[0] = Unsuback<<4 | 0
	response[1] = 2
	binary.BigEndian.PutUint16(response[2:4], packetID)

	client.Conn.Write(response)

	return nil
}

func (b *Broker) processMessages() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		case msg := <-b.messages:
			b.routeMessage(msg)
		}
	}
}

func (b *Broker) routeMessage(msg *Message) {
	clientIDs := b.topicRouter.GetClientsForTopic(msg.Topic)

	for _, clientID := range clientIDs {
		go func(cID string) {
			value, ok := b.clients.Load(cID)
			if !ok {
				return
			}

			client := value.(*ClientConnection)

			topicBytes := []byte(msg.Topic)
			payload := msg.Value

			packet := make([]byte, 2+len(topicBytes)+len(payload))

			packet[0] = (Publish << 4) | 0
			remainingLength := len(topicBytes) + len(payload)

			binary.BigEndian.PutUint16(packet[1:3], uint16(len(topicBytes)))
			copy(packet[3:3+len(topicBytes)], topicBytes)
			copy(packet[3+len(topicBytes):], payload)

			packetWithLength := make([]byte, 0)
			packetWithLength = append(packetWithLength, Publish<<4|0)

			rl := remainingLength
			for {
				digit := byte(rl % 128)
				rl /= 128
				if rl > 0 {
					digit |= 128
				}
				packetWithLength = append(packetWithLength, digit)
				if rl == 0 {
					break
				}
			}

			packetWithLength = append(packetWithLength, packet[1:]...)

			if _, err := client.Conn.Write(packetWithLength); err != nil {
				log.Printf("Send to client %s error: %v", cID, err)
			}
		}(clientID)
	}
}

func (b *Broker) Stop() {
	log.Println("Stopping MQTT Broker...")
	b.connManager.Close()
	b.cancel()
	b.listener.Close()

	b.wg.Wait()
	log.Println("MQTT Broker stopped")
}

func generateClientID() string {
	return "gen_" + time.Now().Format("20060102150405") + "_" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}
