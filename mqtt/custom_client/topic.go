package main

const (
	TOPIC_READY        string = "app/ready/notify"
	APP_CUSTOM_NOTIFY  string = "app/custom/notify" // 设备发送通知事件
	TOPIC_REMOTE_READY string = "remote/ready/notify"
	DEV_QUERY_VERSION  string = "dev.query.version"
)

type Topic struct {
	Name   string
	QOS    byte
	Dup    byte
	Retain byte
}
