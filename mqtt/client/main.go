package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	// "time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC_READY        string = "app/ready/notify"
	APP_CUSTOM_NOTIFY  string = "app/custom/notify" // 设备发送通知事件
	TOPIC_REMOTE_READY string = "remote/ready/notify"
	DEV_QUERY_VERSION  string = "dev.query.version"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println(client)
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")

	ids := make([]int, 10)
	for id := range ids {
		fmt.Println(id)
		go func(id int) {
			count := 0
			for {
				time.Sleep(time.Millisecond * (time.Duration(600 + random(2000))))
				count = count + 1
				// t := client.Publish("topic_get_val", 0, false, fmt.Sprintf("%d,xxxxxxxx:%d", id, count))
				t := client.Publish(APP_CUSTOM_NOTIFY, 1, false, fmt.Sprintf("%d,xxxxxxxx:%d", id, count))
				res := <-t.Done()
				fmt.Println("result:", res)
			}
		}(id)
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// broker := "0.0.0.0"
	broker := "127.0.0.1"
	port := 1883
	opts := mqtt.NewClientOptions()
	// opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	// opts.SetUsername("emqx")
	// opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	s := waitForSignal()
	fmt.Println(s)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}

func random(count int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(count)
}
