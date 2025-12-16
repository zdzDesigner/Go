package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
)

// 服务器配置常量
const (
	HOST      = "0.0.0.0"     // 监听地址，0.0.0.0 表示监听所有网络接口
	PORT      = 2222          // SFTP 服务端口
	USERNAME  = "user"        // 登录用户名
	PASSWORD  = "pass"        // 登录密码
	SFTP_ROOT = "./sftp_data" // SFTP 根目录，所有文件操作限制在此目录内
	// RATE_LIMIT_KB = 512          // 传输速率限制（KB/s），1024 = 1MB/s
	RATE_LIMIT_KB = 2 << 20 // 传输速率限制（KB/s），1024 = 1MB/s
)

func main() {
	// 创建 SFTP 根目录，权限 0755 (rwxr-xr-x)
	if err := os.MkdirAll(SFTP_ROOT, 0o755); err != nil {
		log.Fatal("Failed to create SFTP root directory:", err)
	}

	// 配置 SSH 服务器，设置密码认证回调函数
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// 验证用户名和密码
			if c.User() == USERNAME && string(pass) == PASSWORD {
				return nil, nil // 认证成功
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	// 加载或生成 SSH 主机密钥
	privateKey, err := generateHostKey()
	if err != nil {
		log.Fatal("Failed to generate host key:", err)
	}
	config.AddHostKey(privateKey)

	// 创建 TCP 监听器
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	defer listener.Close()

	log.Printf("SFTP server listening on %s:%d (user: %s, pass: %s)", HOST, PORT, USERNAME, PASSWORD)
	log.Printf("SFTP root directory: %s", SFTP_ROOT)

	// 主循环：接受并处理客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		// 每个连接在独立的 goroutine 中处理，支持并发
		go handleConnection(conn, config)
	}
}

// handleConnection 处理单个客户端连接的完整生命周期
func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// 执行 SSH 握手，建立加密通道
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Println("SSH handshake failed:", err)
		return
	}
	defer sshConn.Close()

	log.Printf("New connection from %s (%s)", sshConn.RemoteAddr(), sshConn.User())

	// 丢弃全局 SSH 请求（如 keepalive）
	go ssh.DiscardRequests(reqs)

	// 处理 SSH 通道请求（每个 SFTP 会话使用一个通道）
	for newChannel := range chans {
		// 只接受 "session" 类型的通道
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println("Could not accept channel:", err)
			continue
		}

		// 处理通道内的子系统请求
		go func(in <-chan *ssh.Request) {
			defer channel.Close()
			
			for req := range in {
				// 检查是否为 SFTP 子系统请求
				// Payload[4:] 跳过前 4 字节的长度字段
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					req.Reply(true, nil) // 接受 SFTP 子系统请求

					// 获取 SFTP 根目录的绝对路径
					absRoot, err := filepath.Abs(SFTP_ROOT)
					if err != nil {
						log.Println("Failed to get absolute path:", err)
						return
					}

					// 创建 SFTP 处理器（包含速率限制器）
					handler := newSftpHandler(absRoot)
					// 注册文件操作处理器
					rootFS := sftp.NewRequestServer(channel, sftp.Handlers{
						FileGet:  handler, // 处理文件读取（下载）
						FilePut:  handler, // 处理文件写入（上传）
						FileCmd:  handler, // 处理文件命令（删除、重命名等）
						FileList: handler, // 处理目录列表
					})

					// 启动 SFTP 服务循环，处理客户端请求
					if err := rootFS.Serve(); err != nil && err != io.EOF {
						log.Println("SFTP server error:", err)
					}
					return
				}
				req.Reply(false, nil) // 拒绝其他类型的请求
			}
		}(requests)
	}

	log.Printf("Connection closed from %s", sshConn.RemoteAddr())
}

// generateHostKey 从文件加载 SSH 主机私钥
func generateHostKey() (ssh.Signer, error) {
	keyPath := "host_key"

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key: %v", err)
	}

	return ssh.ParsePrivateKey(keyBytes)
}

// sftpHandler 实现 SFTP 文件操作接口，并集成速率限制
type sftpHandler struct {
	root    string        // SFTP 根目录的绝对路径
	limiter *rate.Limiter // 速率限制器（Token Bucket 算法）
}

// newSftpHandler 创建新的 SFTP 处理器实例
func newSftpHandler(root string) *sftpHandler {
	return &sftpHandler{
		root: root,
		// 初始化速率限制器：
		// - 令牌生成速率：RATE_LIMIT_KB * 1024 bytes/s
		// - 桶容量（burst）：RATE_LIMIT_KB * 1024 bytes，允许突发传输
		limiter: rate.NewLimiter(rate.Limit(RATE_LIMIT_KB*1024), RATE_LIMIT_KB*1024),
	}
}

// resolve 将客户端请求的路径映射到服务器文件系统路径
// 例如：客户端请求 "/file.txt" -> 服务器 "./sftp_data/file.txt"
func (h *sftpHandler) resolve(path string) string {
	path = filepath.Clean(path) // 清理路径，移除 ".." 等不安全元素
	if !strings.HasPrefix(path, "/") {
		path = "/" + path // 确保路径以 "/" 开头
	}
	return filepath.Join(h.root, path) // 拼接根目录和请求路径
}

// Fileread 实现文件读取接口（下载）
func (h *sftpHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	path := h.resolve(r.Filepath)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// 返回带速率限制的 ReaderAt
	return &rateLimitedReaderAt{f: f, limiter: h.limiter}, nil
}

// Filewrite 实现文件写入接口（上传）
func (h *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	path := h.resolve(r.Filepath)
	// 创建或截断文件，权限 0644 (rw-r--r--)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	// 返回带速率限制的 WriterAt
	return &rateLimitedWriterAt{f: f, limiter: h.limiter}, nil
}

// Filecmd 实现文件命令接口（删除、重命名、创建目录等）
func (h *sftpHandler) Filecmd(r *sftp.Request) error {
	path := h.resolve(r.Filepath)

	switch r.Method {
	case "Setstat":
		// 设置文件属性（权限、时间戳等），暂不实现
		return nil
	case "Rename":
		target := h.resolve(r.Target)
		return os.Rename(path, target)
	case "Rmdir":
		return os.Remove(path) // 删除空目录
	case "Remove":
		return os.Remove(path) // 删除文件
	case "Mkdir":
		return os.MkdirAll(path, 0o755) // 递归创建目录
	case "Link":
		target := h.resolve(r.Target)
		return os.Link(path, target) // 创建硬链接
	case "Symlink":
		target := h.resolve(r.Target)
		return os.Symlink(path, target) // 创建符号链接
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

// Filelist 实现目录列表接口
func (h *sftpHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	path := h.resolve(r.Filepath)

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// 将 DirEntry 转换为 FileInfo
	var fileInfos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的条目
		}
		fileInfos = append(fileInfos, info)
	}

	return listerat(fileInfos), nil
}

// listerat 实现 ListerAt 接口，用于分页返回目录列表
type listerat []os.FileInfo

func (l listerat) ListAt(f []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF // 偏移超出范围
	}
	n := copy(f, l[offset:]) // 复制数据到目标切片
	if n < len(f) {
		return n, io.EOF // 数据不足，返回 EOF
	}
	return n, nil
}

// rateLimitedReaderAt 带速率限制的文件读取器
type rateLimitedReaderAt struct {
	f       *os.File      // 底层文件对象
	limiter *rate.Limiter // 速率限制器
}

// ReadAt 实现 io.ReaderAt 接口，在读取前进行速率限制
func (r *rateLimitedReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	chunkSize := len(p)
	// 限制单次请求的令牌数为 32KB，避免长时间阻塞
	// 32KB 在 1MB/s 下等待约 32ms，用户无感知
	if chunkSize > 32*1024 {
		chunkSize = 32 * 1024
	}

	// 等待并消费令牌，令牌不足时阻塞
	// 使用 context.Background() 表示无超时限制
	if err := r.limiter.WaitN(context.Background(), chunkSize); err != nil {
		return 0, err
	}

	// 令牌获取成功，执行实际的文件读取
	return r.f.ReadAt(p, off)
}

// rateLimitedWriterAt 带速率限制的文件写入器
type rateLimitedWriterAt struct {
	f       *os.File      // 底层文件对象
	limiter *rate.Limiter // 速率限制器
}

// WriteAt 实现 io.WriterAt 接口，在写入前进行速率限制
func (w *rateLimitedWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	chunkSize := len(p)
	// 限制单次请求的令牌数为 32KB
	if chunkSize > 32*1024 {
		chunkSize = 32 * 1024
	}

	// 等待并消费令牌
	if err := w.limiter.WaitN(context.Background(), chunkSize); err != nil {
		return 0, err
	}

	// 执行实际的文件写入
	return w.f.WriteAt(p, off)
}
