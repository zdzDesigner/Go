# whiper

本地 `whisper.cpp` HTTP 封装服务，提供 OpenAI-compatible 转写接口：

```text
POST /v1/audio/transcriptions
```

## 启动

```bash
export WHISPER_CPP_CLI="/path/to/whisper.cpp/build/bin/whisper-cli"
export WHISPER_CPP_MODEL="/path/to/whisper.cpp/models/ggml-base.bin"
go run .
```

默认监听：

```text
127.0.0.1:8318
```

## 测试

```bash
curl http://127.0.0.1:8318/v1/audio/transcriptions \
  -F file="@/path/to/whisper.cpp/samples/jfk.wav" \
  -F model="whisper.cpp"
```

返回：

```json
{"text":"..."}
```

## opencode-talk 配置

如果 `opencode-talk` 使用 OpenAI SDK 调用，可以配置：

```bash
export OPENAI_API_KEY="local"
export OPENAI_BASE_URL="http://127.0.0.1:8318/v1"
```

或在 `/voice-config` 中设置：

```text
Base URL: http://127.0.0.1:8318/v1
API Key: local
```

`OPENAI_API_KEY` 是 OpenAI SDK 客户端侧需要的占位值；服务端只有设置了 `WHIPER_API_KEY` 才会校验它。

## 环境变量

- `WHISPER_CPP_CLI`：必填，`whisper-cli` 绝对路径
- `WHISPER_CPP_MODEL`：必填，`ggml-*.bin` 模型绝对路径
- `WHIPER_ADDR`：监听地址，默认 `127.0.0.1:8318`
- `WHIPER_API_KEY`：可选；不设置则不校验鉴权，设置后校验 `Authorization: Bearer <key>`
- `WHIPER_LANGUAGE`：可选，例如 `zh`、`en`、`auto`
- `WHIPER_TIMEOUT`：可选，例如 `120s`、`3m`
