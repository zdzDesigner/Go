package mqtt

import (
	// "encoding/json"
	"errors"
	"fmt"
	// "log"
	// "modbus/backend/service/transceiver"
	"net/url"
	"time"

	// _ "modbus/backend/util/throttle"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 接收数据的处理
type (
	receiveFunc           func(topic string, payload []byte)
	hookClientOptionsFunc func(*mqtt.ClientOptions) *mqtt.ClientOptions
)

type Mqtter struct {
	client mqtt.Client
	// transceiver *transceiver.Transceive[int]
}

// // mqtt
func NewMqtter(session string, address string, hookClientOptions hookClientOptionsFunc) (*Mqtter, error) {
	client, err := NewMqtt(session, address, hookClientOptions)
	if err != nil {
		return nil, err
	}

	// transceiver := transceiver.NewTransceive([]int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 20}, session)
	// mqtter := &Mqtter{client: client, transceiver: transceiver}
	mqtter := &Mqtter{client: client}
	return mqtter, nil
}

func (m *Mqtter) Disconnect(delay uint) error {
	m.client.Disconnect(delay)
	return nil
}

// 延迟内算一组
// func (m *Mqtter) Listen(topic string) {
// 	fmt.Println("sub topic::", topic)
// 	listen(m.client, topic, m.transceiver.Listen)
// }

// func (m *Mqtter) Request(topic string, payload map[string]any) (res string, err error) {
// 	fmt.Println("topic::", topic)
// 	return m.transceiver.Request(topic, payload, m.Send)
// }

//	func (m *Mqtter) Send(topic string, data any) error {
//		bts, err := json.Marshal(data)
//		if err != nil {
//			return err
//		}
//		// send(m.uri, topic, fmt.Sprintf(`{"cmd":5}`))
//		send(m.client, topic, string(bts))
//
//		return nil
//	}
func (m *Mqtter) ListenTopic(topic string, receive receiveFunc) { listen(m.client, topic, receive) }

// (address URL) mqtt://<user>:<pass>@<server>.cloudmqtt.com:<port>/<topic>
func NewMqtt(clientId string, address string, hookClientOptions hookClientOptionsFunc) (mqtt.Client, error) {
	uri, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	// fmt.Println("uri::", uri)
	client := connect(clientId, uri, hookClientOptions)
	if client == nil {
		return nil, errors.New("连接错误, 打开panic查看详情")
	}
	return client, nil
}

func listen(client mqtt.Client, topic string, hook receiveFunc) {
	// fmt.Println(client, topic)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("%+v\n", msg)
		// fmt.Printf("%s, %s\n", msg.Topic(), string(msg.Payload()))
		hook(msg.Topic(), msg.Payload())
	})
}

func send(client mqtt.Client, topic string, payload string) {
	// var temp int = 30
	// fmt.Scanln(&temp)
	client.Publish(topic, 1, false, payload)
	fmt.Println("/xxxx:", topic, payload)
	// time.Sleep(3 * time.Second)
	// send(uri, topic)
}

func connect(clientId string, uri *url.URL, hookClientOptions hookClientOptionsFunc) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	if hookClientOptions != nil {
		opts = hookClientOptions(opts)
	}
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(30 * time.Second) {
		// panic("超时")
		return nil
	}
	if err := token.Error(); err != nil {
		// panic(err)
		return nil
	}
	// <-token.Done() // 连接成功
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	var (
		username    = uri.User.Username()
		password, _ = uri.User.Password()
	)
	// fmt.Println("host:", uri.Host, "username:", username, "password:", password, "clientId:", clientId)
	fmt.Printf("%+v\n", *uri)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	opts.AutoReconnect = true
	opts.MaxReconnectInterval = 10 * time.Second // 设置最大重连间隔时间
	// 连接断开时触发的回调函数，可以用来记录日志或执行其他操作。
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		fmt.Println("SetConnectionLostHandler::", c, err)
	})
	opts.SetKeepAlive(120 * time.Second)   // 每 60 秒发送一次心跳包
	opts.SetPingTimeout(20 * time.Second) // 等待心跳响应的超时时间为 10 秒
	// opts.SetKeepAlive(time.Second * 240)
	return opts
}
