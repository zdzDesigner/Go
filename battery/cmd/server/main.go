// Package main 实现MQTT代理服务器的入口点。
// 此命令行应用程序创建和管理MQTT代理实例，
// 处理配置、启动和优雅关闭程序。
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"battery/internal/broker"
)

var (
	// address 指定MQTT代理侦听连接的网络地址
	// 默认值为":1883"，绑定到端口1883上的所有接口
	address = flag.String("address", ":1883", "Address to bind the MQTT broker")
	// maxConnections 设置允许的最大并发客户端连接数
	// 默认值为100,000以支持大规模部署
	maxConnections = flag.Int("max-connections", 100000, "Maximum number of concurrent connections")
)

// main 是MQTT代理服务器应用程序的入口点。
// 它处理命令行参数解析、代理初始化、
// 优雅关闭的信号处理和代理生命周期管理。
func main() {
	// 解析命令行标志以配置代理
	flag.Parse()

	// Log startup information with configured parameters
	log.Printf("Starting MQTT Broker with max %d connections on %s", *maxConnections, *address)

	// Create a new broker instance with the specified configuration
	b, err := broker.NewBroker(*address, *maxConnections)
	if err != nil {
		// Fatal error if broker creation fails (usually due to network binding issues)
		log.Fatal("Failed to create broker: ", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// Notify channel on receiving interrupt (Ctrl+C) or termination signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the broker in a separate goroutine to allow for graceful shutdown
	go func() {
		b.Start()
	}()

	// Block until a signal is received
	<-sigChan
	// Log shutdown initiation and stop the broker gracefully
	log.Println("Shutdown signal received, stopping broker...")
	b.Stop()
	log.Println("Broker stopped")
}
