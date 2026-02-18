// Package broker implements a high-performance MQTT broker capable of handling 100K+ concurrent connections.
// It follows MQTT 3.1.1 specification for packet format and protocol handling, with efficient routing
// and connection management for scalable IoT messaging.
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

// MQTT packet types as defined in the MQTT 3.1.1 specification
const (
	Connect     = 1  // Client request to connect to Server
	Connack     = 2  // Connect acknowledgment
	Publish     = 3  // Publish message
	Puback      = 4  // Publish acknowledgment
	Pubrec      = 5  // Publish received (assured delivery part 1)
	Pubrel      = 6  // Publish release (assured delivery part 2)
	Pubcomp     = 7  // Publish complete (assured delivery part 3)
	Subscribe   = 8  // Client subscribe request
	Suback      = 9  // Subscribe acknowledgment
	Unsubscribe = 10 // Unsubscribe request
	Unsuback    = 11 // Unsubscribe acknowledgment
	Pingreq     = 12 // PING request
	Pingresp    = 13 // PING response
	Disconnect  = 14 // Client is disconnecting
)

// Broker represents the main MQTT broker instance that handles client connections,
// message routing, and protocol compliance. It manages the lifecycle of MQTT sessions
// and coordinates with the connection manager and topic router.
type Broker struct {
	// listener accepts incoming network connections on the configured port
	listener net.Listener
	// clients maintains a thread-safe map of active client connections by client ID
	clients sync.Map
	// messages channel queues incoming messages for routing to subscribers
	messages chan *Message
	// ctx provides cancellation capability for graceful shutdown
	ctx context.Context
	// cancel function cancels the context to signal shutdown
	cancel context.CancelFunc
	// wg waits for all goroutines to finish during shutdown
	wg sync.WaitGroup
	// connManager handles connection limits and resource management
	connManager *connection.ConnectionManager
	// topicRouter manages topic subscriptions and matching
	topicRouter *router.TopicMatcher
}

// Message represents an MQTT message that contains topic, payload, and quality of service level
type Message struct {
	// Topic specifies the MQTT topic to which the message is published
	Topic string
	// Value contains the binary payload data of the message
	Value []byte
	// QoS defines the quality of service level for this message (0, 1, or 2)
	QoS byte
}

// ClientConnection encapsulates the information and state for a single MQTT client connection
type ClientConnection struct {
	// ID uniquely identifies the client in the broker's client registry
	ID string
	// Conn holds the underlying network connection to the client
	Conn net.Conn
	// CreatedAt records the timestamp when the connection was established
	CreatedAt time.Time
	// mu provides thread-safe access to connection state
	mu sync.RWMutex
}

// NewBroker creates and initializes a new MQTT broker instance with the specified configuration.
// It sets up network listening, message channels, and associated managers needed for operation.
//
// Parameters:
//   - address: Network address to bind the MQTT listener (e.g., ":1883" or "localhost:1883")
//   - maxConnections: Maximum number of concurrent client connections allowed
//
// Returns:
//   - A pointer to the initialized Broker instance
//   - An error if the network listener could not be created
func NewBroker(address string, maxConnections int) (*Broker, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := &Broker{
		listener:    listener,
		messages:    make(chan *Message, 10000), // Buffer up to 10,000 messages in queue
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

			// 计算剩余长度：主题长度(2字节) + 主题 + 有效载荷
			remainingLength := 2 + len(topicBytes) + len(payload)

			// 计算变长编码的剩余长度所需的字节数
			encodedLength := encodeVariableByteInteger(uint32(remainingLength))

			// 创建完整的包
			packet := make([]byte, 1+len(encodedLength)+2+len(topicBytes)+len(payload))

			// 固定头
			packet[0] = (Publish << 4) | 0 // QoS 0, DUP=0, RETAIN=0

			// 编码后的剩余长度
			copy(packet[1:], encodedLength)

			// 主题长度
			binary.BigEndian.PutUint16(packet[1+len(encodedLength):], uint16(len(topicBytes)))

			// 主题
			copy(packet[1+len(encodedLength)+2:], topicBytes)

			// 有效载荷
			copy(packet[1+len(encodedLength)+2+len(topicBytes):], payload)

			if _, err := client.Conn.Write(packet); err != nil {
				log.Printf("Send to client %s error: %v", cID, err)
			}
		}(clientID)
	}
}

// encodeVariableByteInteger 将整数编码为MQTT变长字节整数
func encodeVariableByteInteger(length uint32) []byte {
	var res []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 128
		}
		res = append(res, digit)
		if length == 0 {
			break
		}
	}
	return res
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
