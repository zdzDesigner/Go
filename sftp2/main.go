package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	HOST       = "0.0.0.0"
	PORT       = 2222
	USERNAME   = "user"
	PASSWORD   = "pass"
	SFTP_ROOT  = "./sftp_data"
)

func main() {
	if err := os.MkdirAll(SFTP_ROOT, 0755); err != nil {
		log.Fatal("Failed to create SFTP root directory:", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == USERNAME && string(pass) == PASSWORD {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateKey, err := generateHostKey()
	if err != nil {
		log.Fatal("Failed to generate host key:", err)
	}
	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	defer listener.Close()

	log.Printf("SFTP server listening on %s:%d (user: %s, pass: %s)", HOST, PORT, USERNAME, PASSWORD)
	log.Printf("SFTP root directory: %s", SFTP_ROOT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Println("SSH handshake failed:", err)
		return
	}
	defer sshConn.Close()

	log.Printf("New connection from %s (%s)", sshConn.RemoteAddr(), sshConn.User())

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println("Could not accept channel:", err)
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					req.Reply(true, nil)
					
					absRoot, err := filepath.Abs(SFTP_ROOT)
					if err != nil {
						log.Println("Failed to get absolute path:", err)
						return
					}
					
					handler := &sftpHandler{root: absRoot}
					rootFS := sftp.NewRequestServer(channel, sftp.Handlers{
						FileGet:  handler,
						FilePut:  handler,
						FileCmd:  handler,
						FileList: handler,
					})
					
					if err := rootFS.Serve(); err != nil && err != io.EOF {
						log.Println("SFTP server error:", err)
					}
					return
				}
				req.Reply(false, nil)
			}
		}(requests)
	}

	log.Printf("Connection closed from %s", sshConn.RemoteAddr())
}

func generateHostKey() (ssh.Signer, error) {
	keyPath := "host_key"
	
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key: %v", err)
	}
	
	return ssh.ParsePrivateKey(keyBytes)
}

type sftpHandler struct {
	root string
}

func (h *sftpHandler) resolve(path string) string {
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return filepath.Join(h.root, path)
}

func (h *sftpHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	path := h.resolve(r.Filepath)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (h *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	path := h.resolve(r.Filepath)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (h *sftpHandler) Filecmd(r *sftp.Request) error {
	path := h.resolve(r.Filepath)
	
	switch r.Method {
	case "Setstat":
		return nil
	case "Rename":
		target := h.resolve(r.Target)
		return os.Rename(path, target)
	case "Rmdir":
		return os.Remove(path)
	case "Remove":
		return os.Remove(path)
	case "Mkdir":
		return os.MkdirAll(path, 0755)
	case "Link":
		target := h.resolve(r.Target)
		return os.Link(path, target)
	case "Symlink":
		target := h.resolve(r.Target)
		return os.Symlink(path, target)
	}
	
	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (h *sftpHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	path := h.resolve(r.Filepath)
	
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	var fileInfos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}
	
	return listerat(fileInfos), nil
}

type listerat []os.FileInfo

func (l listerat) ListAt(f []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}
	n := copy(f, l[offset:])
	if n < len(f) {
		return n, io.EOF
	}
	return n, nil
}