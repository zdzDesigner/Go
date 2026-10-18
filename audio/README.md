# 本地音频 H5 播放器

Go server 扫描服务器上的音频目录，并提供一个可在微信 WebView 中打开的 H5 播放页面。播放进度保存在当前设备的 `localStorage`，不建立帐号、不保存服务端用户进度。

## 支持范围

- 音频由 server 预先放置在配置目录中。
- 扫描会递归处理子目录。
- 首版支持 `aac`、`flac`、`m4a`、`mp3`、`oga`、`ogg`、`opus`、`wav` 和 `webm` 扩展名。
- 音频流支持 HTTP `Range`，可拖动播放进度。
- H5 提供上一个、下一个按钮，当前音频播放完成后自动播放下一个资源。
- H5 会记录最后播放的章节，并在重新打开页面时恢复章节和播放进度。
- 可通过 `AUDIO_ACCESS_ENABLED` 开关控制是否需要共享访问密钥，不提供帐号、上传、删除和重命名功能。
- 进度按设备和音频资源 ID 隔离保存。

## 启动

必须设置基础配置项；只有启用访问控制时才需要设置认证配置项：

```bash
export AUDIO_ADDR="127.0.0.1:8080"
export AUDIO_ROOT="/absolute/path/to/audio"
export AUDIO_ACCESS_ENABLED="true"
export AUDIO_ACCESS_KEY="replace-with-a-long-random-secret"
export AUDIO_SESSION_TTL="12h"
export AUDIO_COOKIE_SECURE="false"
go run .
```

关闭访问控制时：

```bash
export AUDIO_ADDR="127.0.0.1:8080"
export AUDIO_ROOT="/absolute/path/to/audio"
export AUDIO_ACCESS_ENABLED="false"
go run .
```

当 `AUDIO_ACCESS_ENABLED=false` 时，`AUDIO_ACCESS_KEY`、`AUDIO_SESSION_TTL` 和 `AUDIO_COOKIE_SECURE` 可以不设置，H5、音频列表和音频流直接开放。

公网部署时，建议：

- Go server 监听内网地址。
- 由反向代理提供 HTTPS 和公网域名。
- 将 `AUDIO_COOKIE_SECURE` 设置为 `true`。
- 不要把真实访问密钥提交到代码仓库。

配置模板见 `configs/config.example.env`。

## 接口

```text
GET  /
GET  /healthz
POST /api/access
GET  /api/library
GET  /api/audio/{id}
GET  /api/audio/{id}/stream
```

启用访问控制时，除 `/`、`/healthz` 和 `POST /api/access` 外，其余接口需要有效的访问会话 Cookie。关闭访问控制时，所有接口不需要访问会话。

访问认证请求：

```bash
curl -i -c cookies.txt \
  -H 'Content-Type: application/json' \
  -d '{"key":"replace-with-a-long-random-secret"}' \
  http://127.0.0.1:8080/api/access
```

访问音频列表：

```bash
curl -b cookies.txt http://127.0.0.1:8080/api/library
```

## 测试与构建

```bash
node --test app_js_test.js
go test ./...
go build -o audio-server .
```
