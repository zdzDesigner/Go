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
	mode     = flag.String("mode", "both", "Mode: pub (publish only), sub (subscribe only), both (default)")
	message  = flag.String("message", "Hello from Go MQTT client!", "Message to publish")
	qos      = flag.Int("qos", 0, "QoS level")
	retain   = flag.Bool("retain", false, "Retain message")
)

func main() {
	flag.Parse()

	fmt.Printf("Starting program with mode: %s, topic: %s\n", *mode, *topic)

	// 设置 MQTT 客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*broker)

	if *clientID == "" {
		// 根据模式添加不同的客户端ID后缀，以确保唯一性
		*clientID = fmt.Sprintf("go-mqtt-%s-%d", *mode, time.Now().Unix())
	}
	opts.SetClientID(*clientID)

	// 设置连接和消息处理回调
	opts.OnConnect = func(client mqtt.Client) {
		fmt.Printf("[%s] Connected to broker: %s with client ID: %s\n", *mode, *broker, *clientID)
	}

	opts.OnConnectionLost = func(client mqtt.Client, reason error) {
		log.Printf("[%s] Connection lost: %v", *mode, reason)
	}

	// 创建 MQTT 客户端
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[%s] Failed to connect to broker: %v", *mode, token.Error())
	}

	defer client.Disconnect(250)

	// 创建一个通道来接收系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	switch *mode {
	case "pub":
		// 发布模式 - 只发布消息
		fmt.Printf("[%s] Starting publisher...\n", *mode)
		go publishMessages(client, sigCh)
		<-sigCh // 等待信号

	case "sub":
		// 订阅模式 - 只订阅消息
		fmt.Printf("[%s] Starting subscriber...\n", *mode)
		subscribeMessages(client)
		<-sigCh // 等待信号

	case "both":
		// 默认模式 - 同时发布和订阅
		fmt.Printf("[%s] Starting both publisher and subscriber...\n", *mode)
		subscribeMessages(client)
		go publishMessages(client, sigCh)
		<-sigCh // 等待信号
	}

	fmt.Printf("[%s] Shutting down...\n", *mode)
}

// 发布消息
func publishMessages(client mqtt.Client, sigCh chan os.Signal) {
	fmt.Printf("[PUB] Starting to publish messages to topic: %s (QoS: %d, Retain: %t)\n", *topic, *qos, *retain)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			messageText := fmt.Sprintf("%s [%d]", *message, counter)

			token := client.Publish(*topic, byte(*qos), *retain, messageText)
			token.Wait()

			if token.Error() != nil {
				log.Printf("[PUB] Error publishing message: %v", token.Error())
			} else {
				fmt.Printf("[PUB] Published: %s\n", messageText)
			}
		case <-sigCh: // 接收外部信号
			fmt.Printf("[PUB] Publisher shutting down...\n")
			return
		}
	}
}

// 订阅消息
func subscribeMessages(client mqtt.Client) {
	// 设置消息处理回调
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("[SUB] Received message on topic '%s': %s\n", msg.Topic(), string(msg.Payload()))
	}

	// 订阅主题
	if token := client.Subscribe(*topic, byte(*qos), messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("[SUB] Failed to subscribe to topic %s: %v", *topic, token.Error())
	}

	fmt.Printf("[SUB] Subscribed to topic: %s\n", *topic)
}
