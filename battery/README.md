## MQTT Broker


```sh
# 开启服务
go run ./cmd/server/main.go

# cd ./examples/simple-pub-sub/ 
## 仅发布模式
go run main.go --mode pub --topic "test/topic" --message "Hello World"

## 仅订阅模式
go run main.go --mode sub --topic "test/topic"
```


