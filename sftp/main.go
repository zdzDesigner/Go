package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

// 用户信息结构体
type User struct {
	Username  string
	Password  string
	PublicKey string
	RootDir   string
}

// 简单的用户数据库
var users = map[string]*User{
	"user": {
		Username:  "user",
		Password:  "123456",      // 简化密码以便测试
		PublicKey: "",            // 可以留空，使用密码认证
		RootDir:   "./sftp_home", // 用户的根目录
	},
}

// 连接计数器
var (
	activeConnections int32
	maxConnections    int32 = 10
	connectionsMutex  sync.Mutex
)

func main() {
	// 确保用户根目录存在
	for _, user := range users {
		if err := os.MkdirAll(user.RootDir, 0755); err != nil {
			log.Fatalf("无法创建用户根目录 %s: %v", user.RootDir, err)
		}
	}

	// 生成或加载主机密钥
	hostKey, err := generateOrLoadHostKey("host_key")
	if err != nil {
		log.Fatalf("无法生成或加载主机密钥: %v", err)
	}

	// 配置SSH服务器
	config := &ssh.ServerConfig{
		NoClientAuth: false,
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			username := conn.User()
			log.Printf("密码验证尝试 - 用户名: %s", username)
			user, exists := users[username]
			if !exists {
				log.Printf("用户不存在: %s", username)
				return nil, fmt.Errorf("无效的用户名或密码")
			}
			if user.Password != string(password) {
				log.Printf("密码不匹配 - 用户名: %s, 提供的密码: %s, 正确密码: %s", username, string(password), user.Password)
				return nil, fmt.Errorf("无效的用户名或密码")
			}
			log.Printf("密码验证成功 - 用户名: %s", username)
			return &ssh.Permissions{
				Extensions: map[string]string{
					"root_dir": user.RootDir,
				},
			}, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			user, exists := users[conn.User()]
			if !exists || user.PublicKey == "" {
				// 返回ssh.ErrNoAuth，这样SSH服务器会继续尝试其他认证方法
				return nil, ssh.ErrNoAuth
			}
			// 这里简化处理，实际应用中应该验证提供的公钥是否匹配用户存储的公钥
			return &ssh.Permissions{
				Extensions: map[string]string{
					"root_dir": user.RootDir,
				},
			}, nil
		},
	}
	config.AddHostKey(hostKey)

	// 监听TCP连接
	listener, err := net.Listen("tcp", ":2023")
	if err != nil {
		log.Fatalf("无法监听端口 2023: %v", err)
	}
	defer listener.Close()

	log.Println("SFTP服务器启动在端口 2023")

	// 接受连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}

		// 检查连接数限制
		connectionsMutex.Lock()
		if activeConnections >= maxConnections {
			connectionsMutex.Unlock()
			conn.Close()
			log.Println("拒绝连接: 已达到最大连接数")
			continue
		}
		activeConnections++
		connectionsMutex.Unlock()

		// 处理SSH连接
		go handleSSHConnection(conn, config)
	}
}

// 处理SSH连接
func handleSSHConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer func() {
		connectionsMutex.Lock()
		activeConnections--
		connectionsMutex.Unlock()
		conn.Close()
	}()

	// 进行SSH握手
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("SSH握手失败: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("用户 %s 从 %s 连接成功", sshConn.User(), sshConn.RemoteAddr())

	// 处理全局请求
	go ssh.DiscardRequests(reqs)

	// 处理通道请求
	handleSessionRequests(chans, sshConn.Permissions.Extensions["root_dir"])
}

// 处理会话请求
func handleSessionRequests(chans <-chan ssh.NewChannel, rootDir string) {
	for newChannel := range chans {
		// 只处理session类型的通道
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "未知的通道类型")
			continue
		}

		// 接受通道
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("无法接受通道: %v", err)
			continue
		}

		// 处理通道请求
		go func(in <-chan *ssh.Request) {
			for req := range in {
				done := func() {
					if req.WantReply {
						req.Reply(false, nil)
					}
				}

				// 只处理subsystem请求
				if req.Type != "subsystem" {
					done()
					continue
				}

				// 检查是否是sftp子系统
				if string(req.Payload[4:]) != "sftp" {
					done()
					continue
				}

				// 确保根目录存在
				if _, err := os.Stat(rootDir); os.IsNotExist(err) {
					if err := os.MkdirAll(rootDir, 0755); err != nil {
						log.Printf("创建根目录失败: %v", err)
						if req.WantReply {
							req.Reply(false, nil)
						}
						return
					}
				}

				// 发送成功响应
				if req.WantReply {
					req.Reply(true, nil)
				}

				// 处理SFTP消息
				handleSFTP(channel, rootDir)
			}
		}(requests)
	}
}

// 处理SFTP协议
func handleSFTP(channel ssh.Channel, rootDir string) {
	defer channel.Close()

	log.Printf("SFTP请求已接收，用户根目录: %s", rootDir)

	// 处理SFTP消息
	reader := bufio.NewReader(channel)
	for {
		// 读取消息长度
		lengthBytes := make([]byte, 4)
		n, err := io.ReadFull(reader, lengthBytes)
		if err != nil {
			if err != io.EOF {
				log.Printf("读取SFTP消息长度失败: %v", err)
			}
			break
		}
		if n != 4 {
			log.Printf("读取SFTP消息长度不完整: %d 字节", n)
			break
		}

		// 解析消息长度
		length := uint32(lengthBytes[0])<<24 | uint32(lengthBytes[1])<<16 | uint32(lengthBytes[2])<<8 | uint32(lengthBytes[3])

		// 读取消息内容
		buffer := make([]byte, length)
		n, err = io.ReadFull(reader, buffer)
		if err != nil {
			log.Printf("读取SFTP消息失败: %v", err)
			break
		}
		if n != int(length) {
			log.Printf("读取SFTP消息不完整: %d/%d 字节", n, length)
			break
		}

		log.Printf("收到SFTP消息 - 类型: %d 长度: %d 字节", buffer[0], length)

		// 解析消息类型
		msgType := buffer[0]

		// 解析请求ID (如果消息长度足够)
		var requestID uint32 = 0
		if length >= 5 {
			requestID = uint32(buffer[1])<<24 | uint32(buffer[2])<<16 | uint32(buffer[3])<<8 | uint32(buffer[4])
		}

		// 根据消息类型处理
		switch msgType {
		case 1: // SSH_FXP_INIT - 初始化
			log.Printf("处理SSH_FXP_INIT消息")
			// 发送版本响应
			response := []byte{
				0, 0, 0, 9, // 长度
				2,          // 类型 (SSH_FXP_VERSION)
				0, 0, 0, 3, // 版本号 (3)
			}
			channel.Write(response)

		case 16: // SSH_FXP_REALPATH - 获取真实路径
			log.Printf("处理SSH_FXP_REALPATH消息，请求ID: %d", requestID)
			// 解析请求的路径
			var path string
			if n >= 13 { // 至少需要13字节: 长度(4) + 类型(1) + 请求ID(4) + 路径长度(4)
				pathLen := uint32(buffer[9])<<24 | uint32(buffer[10])<<16 | uint32(buffer[11])<<8 | uint32(buffer[12])
				if n >= 13+int(pathLen) {
					path = string(buffer[13 : 13+int(pathLen)])
					log.Printf("REALPATH请求路径: %s", path)
				}
			}

			// 对于REALPATH，我们需要返回规范化的路径
			// 这里我们简单地返回根目录作为规范化路径 "/"
			canonicalPath := "/"
			pathLen := uint32(len(canonicalPath))
			log.Printf("REALPATH规范化路径: %s, 长度: %d", canonicalPath, pathLen)

			// 创建SSH_FXP_NAME响应
			// 响应格式: 长度(4) + 类型(1) + 请求ID(4) + 条目数量(4) + 条目列表
			// 每个条目: 文件名长度(4) + 文件名 + 长文件名长度(4) + 长文件名 + 属性(17)
			contentSize := uint32(1 + 4 + 4 + 4 + pathLen + 4 + pathLen + 17)
			response := make([]byte, 4+contentSize)
			log.Printf("REALPATH响应内容大小: %d, 总大小: %d", contentSize, 4+contentSize)

			// 设置长度字段
			response[0] = byte(contentSize >> 24)
			response[1] = byte(contentSize >> 16)
			response[2] = byte(contentSize >> 8)
			response[3] = byte(contentSize)
			log.Printf("REALPATH响应长度字段: %02X %02X %02X %02X", response[0], response[1], response[2], response[3])

			// 设置类型 (SSH_FXP_NAME = 104)
			response[4] = 104
			log.Printf("REALPATH响应类型字段: %02X", response[4])

			// 设置请求ID
			response[5] = byte(requestID >> 24)
			response[6] = byte(requestID >> 16)
			response[7] = byte(requestID >> 8)
			response[8] = byte(requestID)
			log.Printf("REALPATH响应请求ID: %02X %02X %02X %02X", response[5], response[6], response[7], response[8])

			// 设置条目数量 (1)
			response[9] = 0
			response[10] = 0
			response[11] = 0
			response[12] = 1
			log.Printf("REALPATH响应条目数量: %02X %02X %02X %02X", response[9], response[10], response[11], response[12])

			// 设置文件名长度
			response[13] = byte(pathLen >> 24)
			response[14] = byte(pathLen >> 16)
			response[15] = byte(pathLen >> 8)
			response[16] = byte(pathLen)
			log.Printf("REALPATH响应文件名长度: %02X %02X %02X %02X", response[13], response[14], response[15], response[16])

			// 设置文件名
			copy(response[17:17+len(canonicalPath)], canonicalPath)
			log.Printf("REALPATH响应文件名: %s", canonicalPath)

			// 设置长文件名长度
			longNameOffset := 17 + len(canonicalPath)
			response[longNameOffset] = byte(pathLen >> 24)
			response[longNameOffset+1] = byte(pathLen >> 16)
			response[longNameOffset+2] = byte(pathLen >> 8)
			response[longNameOffset+3] = byte(pathLen)
			log.Printf("REALPATH响应长文件名长度: %02X %02X %02X %02X", response[longNameOffset], response[longNameOffset+1], response[longNameOffset+2], response[longNameOffset+3])

			// 设置长文件名
			copy(response[longNameOffset+4:longNameOffset+4+len(canonicalPath)], canonicalPath)
			log.Printf("REALPATH响应长文件名: %s", canonicalPath)

			// 设置文件属性
			attrOffset := longNameOffset + 4 + len(canonicalPath)
			// 标志 (支持大小和权限)
			flags := uint32(0x00000003) // SSH_FILEXFER_ATTR_SIZE | SSH_FILEXFER_ATTR_PERMISSIONS
			response[attrOffset] = byte(flags >> 24)
			response[attrOffset+1] = byte(flags >> 16)
			response[attrOffset+2] = byte(flags >> 8)
			response[attrOffset+3] = byte(flags)
			log.Printf("REALPATH响应属性标志: %02X %02X %02X %02X", response[attrOffset], response[attrOffset+1], response[attrOffset+2], response[attrOffset+3])

			// 文件大小 (0)
			response[attrOffset+4] = 0
			response[attrOffset+5] = 0
			response[attrOffset+6] = 0
			response[attrOffset+7] = 0
			log.Printf("REALPATH响应文件大小: %02X %02X %02X %02X", response[attrOffset+4], response[attrOffset+5], response[attrOffset+6], response[attrOffset+7])

			// 权限 (0755)
			permissions := uint32(0755)
			response[attrOffset+8] = byte(permissions >> 24)
			response[attrOffset+9] = byte(permissions >> 16)
			response[attrOffset+10] = byte(permissions >> 8)
			response[attrOffset+11] = byte(permissions)
			log.Printf("REALPATH响应权限: %02X %02X %02X %02X", response[attrOffset+8], response[attrOffset+9], response[attrOffset+10], response[attrOffset+11])

			// 发送响应
			bytesWritten, err := channel.Write(response)
			if err != nil {
				log.Printf("发送REALPATH响应失败: %v", err)
			} else {
				log.Printf("成功发送REALPATH响应，写入字节数: %d", bytesWritten)
			}

		default:
			log.Printf("未处理的SFTP消息类型: %d", msgType)
		}
	}
}

// 发送句柄响应 (SSH_FXP_HANDLE)
func sendHandleResponse(channel ssh.Channel, requestID uint32, handle []byte) {
	response := make([]byte, 9+len(handle))
	// 长度
	length := uint32(5 + len(handle))
	response[0] = byte(length >> 24)
	response[1] = byte(length >> 16)
	response[2] = byte(length >> 8)
	response[3] = byte(length)
	// 类型 (SSH_FXP_HANDLE = 2)
	response[4] = 2
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 句柄
	copy(response[9:], handle)

	channel.Write(response)
}

// 发送状态响应 (SSH_FXP_STATUS)
func sendStatusResponse(channel ssh.Channel, requestID uint32, statusCode uint32, message string) {
	response := make([]byte, 17+len(message))
	// 长度
	length := uint32(13 + len(message))
	response[0] = byte(length >> 24)
	response[1] = byte(length >> 16)
	response[2] = byte(length >> 8)
	response[3] = byte(length)
	// 类型 (SSH_FXP_STATUS = 101)
	response[4] = 101
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 状态码
	response[9] = byte(statusCode >> 24)
	response[10] = byte(statusCode >> 16)
	response[11] = byte(statusCode >> 8)
	response[12] = byte(statusCode)
	// 消息长度
	msgLen := uint32(len(message))
	response[13] = byte(msgLen >> 24)
	response[14] = byte(msgLen >> 16)
	response[15] = byte(msgLen >> 8)
	response[16] = byte(msgLen)
	// 消息内容
	copy(response[17:], message)

	channel.Write(response)
}

// 发送数据响应 (SSH_FXP_DATA)
func sendDataResponse(channel ssh.Channel, requestID uint32, data []byte) {
	response := make([]byte, 9+len(data))
	// 长度
	length := uint32(5 + len(data))
	response[0] = byte(length >> 24)
	response[1] = byte(length >> 16)
	response[2] = byte(length >> 8)
	response[3] = byte(length)
	// 类型 (SSH_FXP_DATA = 102)
	response[4] = 102
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 数据
	copy(response[9:], data)

	channel.Write(response)
}

// 发送属性响应 (SSH_FXP_ATTRS)
func sendAttrsResponse(channel ssh.Channel, requestID uint32) {
	response := make([]byte, 21)
	// 长度
	length := uint32(17)
	response[0] = byte(length >> 24)
	response[1] = byte(length >> 16)
	response[2] = byte(length >> 8)
	response[3] = byte(length)
	// 类型 (SSH_FXP_ATTRS = 105)
	response[4] = 105
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 标志 (支持大小和权限)
	flags := uint32(0x00000003) // SSH_FILEXFER_ATTR_SIZE | SSH_FILEXFER_ATTR_PERMISSIONS
	response[9] = byte(flags >> 24)
	response[10] = byte(flags >> 16)
	response[11] = byte(flags >> 8)
	response[12] = byte(flags)
	// 文件大小 (0)
	response[13] = 0
	response[14] = 0
	response[15] = 0
	response[16] = 0
	// 权限 (0755)
	permissions := uint32(0755)
	response[17] = byte(permissions >> 24)
	response[18] = byte(permissions >> 16)
	response[19] = byte(permissions >> 8)
	response[20] = byte(permissions)

	channel.Write(response)
}

// 发送名称响应 (SSH_FXP_NAME)
func sendNameResponse(channel ssh.Channel, requestID uint32, rootDir string) {
	// 读取目录内容
	entries, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Printf("读取目录失败: %v", err)
		// 发送空列表响应
		sendEmptyNameResponse(channel, requestID)
		return
	}

	// 计算消息内容大小（不包括长度字段）
	var contentSize uint32 = 1 // 类型(1)
	contentSize += 4           // 请求ID(4)
	contentSize += 4           // 条目数量(4)

	// 为每个条目计算大小
	for _, entry := range entries {
		name := entry.Name()
		// 每个条目包含: 文件名长度(4) + 文件名 + 长文件名长度(4) + 长文件名 + 属性(17)
		contentSize += 4 + uint32(len(name)) + 4 + uint32(len(name)) + 17
	}

	// 创建响应缓冲区（包括长度字段）
	totalSize := 4 + contentSize
	response := make([]byte, totalSize)

	// 设置长度字段（只包含内容部分的大小）
	response[0] = byte(contentSize >> 24)
	response[1] = byte(contentSize >> 16)
	response[2] = byte(contentSize >> 8)
	response[3] = byte(contentSize)
	// 类型 (SSH_FXP_NAME = 104)
	response[4] = 104
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 条目数量
	entryCount := uint32(len(entries))
	response[9] = byte(entryCount >> 24)
	response[10] = byte(entryCount >> 16)
	response[11] = byte(entryCount >> 8)
	response[12] = byte(entryCount)

	// 填充每个条目
	offset := 13
	for _, entry := range entries {
		name := entry.Name()
		nameLen := uint32(len(name))

		// 文件名长度
		response[offset] = byte(nameLen >> 24)
		response[offset+1] = byte(nameLen >> 16)
		response[offset+2] = byte(nameLen >> 8)
		response[offset+3] = byte(nameLen)
		offset += 4

		// 文件名
		copy(response[offset:offset+len(name)], name)
		offset += len(name)

		// 长文件名长度 (与文件名相同)
		response[offset] = byte(nameLen >> 24)
		response[offset+1] = byte(nameLen >> 16)
		response[offset+2] = byte(nameLen >> 8)
		response[offset+3] = byte(nameLen)
		offset += 4

		// 长文件名 (与文件名相同)
		copy(response[offset:offset+len(name)], name)
		offset += len(name)

		// 文件属性
		// 标志 (支持大小和权限)
		flags := uint32(0x00000003) // SSH_FILEXFER_ATTR_SIZE | SSH_FILEXFER_ATTR_PERMISSIONS
		response[offset] = byte(flags >> 24)
		response[offset+1] = byte(flags >> 16)
		response[offset+2] = byte(flags >> 8)
		response[offset+3] = byte(flags)
		offset += 4

		// 文件大小
		size := uint32(entry.Size())
		response[offset] = byte(size >> 24)
		response[offset+1] = byte(size >> 16)
		response[offset+2] = byte(size >> 8)
		response[offset+3] = byte(size)
		offset += 4

		// 权限
		mode := entry.Mode()
		permissions := uint32(mode.Perm())
		response[offset] = byte(permissions >> 24)
		response[offset+1] = byte(permissions >> 16)
		response[offset+2] = byte(permissions >> 8)
		response[offset+3] = byte(permissions)
		offset += 4
	}

	channel.Write(response)
}

// 发送空名称响应
func sendEmptyNameResponse(channel ssh.Channel, requestID uint32) {
	// 计算消息内容大小（不包括长度字段）
	contentSize := uint32(1 + 4 + 4) // 类型(1) + 请求ID(4) + 条目数量(4)
	// 创建响应缓冲区（包括长度字段）
	totalSize := 4 + contentSize
	response := make([]byte, totalSize)

	// 设置长度字段（只包含内容部分的大小）
	response[0] = byte(contentSize >> 24)
	response[1] = byte(contentSize >> 16)
	response[2] = byte(contentSize >> 8)
	response[3] = byte(contentSize)
	// 类型 (SSH_FXP_NAME = 104)
	response[4] = 104
	// 请求ID
	response[5] = byte(requestID >> 24)
	response[6] = byte(requestID >> 16)
	response[7] = byte(requestID >> 8)
	response[8] = byte(requestID)
	// 条目数量 (0)
	response[9] = 0
	response[10] = 0
	response[11] = 0
	response[12] = 0

	channel.Write(response)
}

// 生成或加载主机密钥
func generateOrLoadHostKey(filePath string) (ssh.Signer, error) {
	// 检查密钥文件是否存在
	keyData, err := ioutil.ReadFile(filePath)
	if err == nil {
		// 加载现有密钥
		key, err := ssh.ParsePrivateKey(keyData)
		if err == nil {
			return key, nil
		}
		log.Printf("无法解析现有密钥，将生成新密钥: %v", err)
	}

	// 生成新的RSA密钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥失败: %v", err)
	}

	// 编码为PEM格式
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	// 保存密钥到文件
	if err := ioutil.WriteFile(filePath, privateKeyPEM, 0600); err != nil {
		log.Printf("无法保存主机密钥: %v", err)
	}

	// 创建签名器
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("创建签名器失败: %v", err)
	}

	return signer, nil
}
