package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	broker   = flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	clientID = flag.String("clientid", "", "MQTT client ID")
	topic    = flag.String("topic", "test/topic", "MQTT topic to publish/subscribe")
	qos      = flag.Int("qos", 0, "QoS level")
)

func main() {
	flag.Parse()

	// 设置 MQTT 客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*broker)

	if *clientID == "" {
		// 使用独特客户端ID以方便调试
		pid := os.Getpid()
		*clientID = fmt.Sprintf("debug-subscriber-%d-%d", pid, time.Now().UnixNano())
	}
	opts.SetClientID(*clientID)

	fmt.Printf("Initializing subscriber with client ID: %s\n", *clientID)

	// 设置连接回调
	opts.OnConnect = func(client mqtt.Client) {
		fmt.Printf("Subscriber connected to broker: %s with client ID: %s\n", *broker, *clientID)
	}

	opts.OnConnectionLost = func(client mqtt.Client, reason error) {
		log.Printf("Subscriber %s connection lost: %v", *clientID, reason)
	}

	// 创建 MQTT 客户端
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect subscriber: %v", token.Error())
	}

	defer client.Disconnect(250)

	// 创建一个通道来接收系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 设置消息处理回调
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		// 添加时间戳以便区分消息
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] Subscriber %s - Received message on topic '%s': %s\n", timestamp, *clientID, msg.Topic(), string(msg.Payload()))
	}

	// 订阅主题
	fmt.Printf("Attempting to subscribe to topic: %s\n", *topic)
	if token := client.Subscribe(*topic, byte(*qos), messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to subscribe to topic %s: %v", *topic, token.Error())
	}

	fmt.Printf("Successfully subscribed to topic: %s with client ID: %s\n", *topic, *clientID)

	// 等待中断信号
	<-sigCh
	fmt.Printf("Shutting down subscriber %s...\n", *clientID)
}
