# Go MQTT Broker 示例

本目录包含各种使用 Go MQTT Broker 的示例应用程序。

## 示例列表

### 1. Simple Pub/Sub (简单发布/订阅)

位于 `simple-pub-sub/` 目录，展示了基本的 MQTT 发布和订阅功能。

#### 特性
- 支持发布模式、订阅模式或两者结合
- 可配置的 MQTT 代理 URL、主题、QoS 等参数
- 命令行参数控制运行模式

#### 如何运行

1. 确保 MQTT Broker 正在运行:
   ```bash
   cd ../..  # 回到项目根目录
   ./server
   ```

2. 进入示例目录:
   ```bash
   cd simple-pub-sub
   ```

3. 初始化 Go 模块 (首次运行):
   ```bash
   go mod init example.com/pubsub
   go get github.com/eclipse/paho.mqtt.golang
   ```

4. 运行示例:

   - **仅发布模式**:
     ```bash
     go run main.go --mode pub --topic "test/topic" --message "Hello World"
     ```

   - **仅订阅模式**:
     ```bash
     go run main.go --mode sub --topic "test/topic"
     ```

   - **发布和订阅模式**:
     ```bash
     go run main.go --mode both --topic "test/topic" --message "Hello from Go"
     ```

5. 停止程序:
   - 使用 `Ctrl+C` 发送中断信号，程序会优雅地断开连接并退出
   - 如果程序未响应，可以使用 `Ctrl+Z` 暂停然后 `kill %` 或直接发送 `kill` 信号

6. 可用参数:
   - `--broker`: MQTT 代理 URL (默认: "tcp://localhost:1883")
   - `--clientid`: MQTT 客户端 ID (默认: 自动生成)
   - `--topic`: 要发布/订阅的主题 (默认: "test/topic")
   - `--mode`: 运行模式 ("pub", "sub", "both") (默认: "both")
   - `--message`: 要发布的消息 (默认: "Hello from Go MQTT client!")
   - `--qos`: QoS 级别 (默认: 0)
   - `--retain`: 是否保留消息 (默认: false)

## 贡献

欢迎贡献更多有用的示例！请确保您的示例:

1. 包含清晰的 README 说明
2. 展示特定的功能或用例
3. 遵循良好的 Go 语言编程实践
4. 包含适当的错误处理