package mqtt

import "fmt"

func MqttDemo() {
	address := "mqtt://127.0.0.1:1883"
	mqtt, err := NewMqtt("", address, nil)
	if err != nil {
		panic(err)
	}

	listen(mqtt, "topic_get_val", func(topic string, payload []byte) {
		fmt.Println("topic::", topic)
		fmt.Println("payload:", string(payload))
	})
}
