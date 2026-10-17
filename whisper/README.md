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

启用确定性中文纠错词表：

```bash
export WHIPER_CORRECTIONS="向木=项目，衣赖=依赖，从新=重新，单元侧试=单元测试，全线=权限，权现效验=权限校验，带办=待办"
go run .
```

启用 CPU 中文短句纠错：

```bash
export WHIPER_CORRECTOR_ENABLED="true"
export WHIPER_CORRECTOR_CLI="/path/to/llama.cpp/build/bin/llama-cli"
export WHIPER_CORRECTOR_MODEL="/path/to/Qwen2.5-0.5B-Instruct-Q4_K_M.gguf"
export WHIPER_CORRECTOR_HOTWORDS="欧喷扣=OpenCode，欧盆扣=OpenCode，扣德=code"
export WHIPER_CORRECTOR_TIMEOUT="5s"
go run .
```

确定性词表优先用于开发领域词纠错；LLM 纠错层用于 50 字以内中文短句，只修正明显语音识别错误，模型输出异常时保留原始 STT 文本。

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
- `WHIPER_CORRECTIONS`：可选，确定性纠错词表，格式为 `错词=正确词`，多个词用中文逗号、英文逗号或换行分隔
- `WHIPER_CORRECTOR_ENABLED`：可选，启用中文语义纠错，例如 `true`、`false`
- `WHIPER_CORRECTOR_CLI`：启用纠错时必填，`llama-cli` 绝对路径
- `WHIPER_CORRECTOR_MODEL`：启用纠错时必填，推荐 `Qwen2.5-0.5B-Instruct GGUF Q4_K_M`
- `WHIPER_CORRECTOR_HOTWORDS`：可选，中文热词表，默认包含 `OpenCode` 相关纠错
- `WHIPER_CORRECTOR_TIMEOUT`：可选，纠错超时，默认 `5s`
