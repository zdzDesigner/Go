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

	// 设置 MQTT 客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*broker)

	if *clientID == "" {
		pid := os.Getpid()
		*clientID = fmt.Sprintf("go-mqtt-client-%d-%d", pid, time.Now().UnixNano())
	}
	opts.SetClientID(*clientID)

	// 设置连接和消息处理回调
	opts.OnConnect = func(client mqtt.Client) {
		fmt.Printf("Connected to broker: %s\n", *broker)
	}

	opts.OnConnectionLost = func(client mqtt.Client, reason error) {
		log.Printf("Connection lost: %v", reason)
	}

	// 创建 MQTT 客户端
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}

	defer client.Disconnect(250)

	// 创建一个通道来接收系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	switch *mode {
	case "pub":
		// 发布模式 - 只发布消息
		go publishMessages(client, sigCh)
		<-sigCh // 等待信号

	case "sub":
		// 订阅模式 - 只订阅消息
		subscribeMessages(client)
		<-sigCh // 等待信号

	case "both":
		// 默认模式 - 同时发布和订阅
		subscribeMessages(client)
		go publishMessages(client, sigCh)
		<-sigCh // 等待信号
	}

	fmt.Println("\nShutting down...")
}

// 发布消息
func publishMessages(client mqtt.Client, sigCh chan os.Signal) {
	fmt.Printf("Starting to publish messages to topic: %s (QoS: %d, Retain: %t)\n", *topic, *qos, *retain)

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
				log.Printf("Error publishing message: %v", token.Error())
			} else {
				fmt.Printf("Published: %s\n", messageText)
			}
		case <-sigCh: // 接收外部信号
			fmt.Println("Publisher shutting down...")
			return
		}
	}
}

// 订阅消息
func subscribeMessages(client mqtt.Client) {
	// 设置消息处理回调
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message on topic '%s': %s\n", msg.Topic(), string(msg.Payload()))
	}

	// 订阅主题
	if token := client.Subscribe(*topic, byte(*qos), messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to subscribe to topic %s: %v", *topic, token.Error())
	}

	fmt.Printf("Subscribed to topic: %s\n", *topic)
}
