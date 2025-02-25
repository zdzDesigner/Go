package main

import (
	"encoding/binary"
	// "fmt"
	"log"
	"net"
)

const (
	stunPort        = ":3478" // STUN 默认端口
	stunMagicCookie = 0x2112A442
	attrXorMapped   = 0x0020
)

func main() {
	// 监听 UDP 端口
	conn, err := net.ListenPacket("udp", stunPort)
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer conn.Close()
	log.Printf("STUN server running on %s", stunPort)

	buffer := make([]byte, 1024)
	for {
		// 读取客户端请求
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		// 解析 STUN 请求
		if isStunBindingRequest(buffer[:n]) {
			log.Printf("Received STUN request from %s", addr.String())
			response := buildStunResponse(buffer[:n], addr)
			// 发送响应
			_, err = conn.WriteTo(response, addr)
			if err != nil {
				log.Println("Write error:", err)
			}
		}
	}
}

// 判断是否为 Binding Request
func isStunBindingRequest(data []byte) bool {
	if len(data) < 20 {
		return false
	}
	// 检查消息类型是否为 0x0001（Binding Request）
	return binary.BigEndian.Uint16(data[0:2]) == 0x0001
}

// 构建 STUN 响应
func buildStunResponse(request []byte, addr net.Addr) []byte {
	// 提取请求中的事务 ID
	transactionID := request[4:20]

	// 构建响应头
	response := make([]byte, 0)
	response = append(response, 0x01, 0x01) // Binding Response (0x0101)
	response = append(response, 0x00, 0x00) // 初始 Length，后续修正

	// 添加 Magic Cookie 和事务 ID
	cookieBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cookieBytes, stunMagicCookie)
	response = append(response, cookieBytes...)
	response = append(response, transactionID...)

	// 构建 XOR-MAPPED-ADDRESS 属性
	udpAddr := addr.(*net.UDPAddr)
	xorPort := udpAddr.Port ^ (stunMagicCookie >> 16)
	xorIP := xorIP(udpAddr.IP, stunMagicCookie)

	// 属性类型和长度
	attrType := []byte{0x00, 0x20}   // XOR-MAPPED-ADDRESS
	attrLength := []byte{0x00, 0x08} // 8字节（IPv4）或 20字节（IPv6）
	response = append(response, attrType...)
	response = append(response, attrLength...)

	// 属性值
	response = append(response, 0x00, 0x01) // IPv4 family (0x01)
	response = append(response, byte(xorPort>>8), byte(xorPort))
	response = append(response, xorIP...)

	// 修正 Length 字段
	length := len(response) - 20 // 总长度减去头部的20字节
	binary.BigEndian.PutUint16(response[2:4], uint16(length))

	return response
}

// 计算 XOR IP 地址
func xorIP(ip net.IP, magicCookie uint32) []byte {
	ip = ip.To4()
	xor := make([]byte, 4)
	cookieBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cookieBytes, magicCookie)
	for i := 0; i < 4; i++ {
		xor[i] = ip[i] ^ cookieBytes[i]
	}
	return xor
}
