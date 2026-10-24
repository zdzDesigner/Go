# 本地音频 H5 播放器

Go server 扫描服务器上的音频目录，并提供一个可在微信 WebView 中打开的 H5 播放页面。播放进度保存在当前设备的 `localStorage`，不建立帐号、不保存服务端用户进度。

Android App 使用 Tauri 2 将同一份 `web/` HTML、CSS、JS 嵌入 APK。App 连接 Go server 获取音频，Media3 原生服务负责后台播放、锁屏控制、自动连播及设备本地续播，不读取手机本地音频。

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

## Android App

要求 Node.js 20+、Rust、JDK 17+、Android SDK Platform 36、Build-Tools 36、NDK 和 `aarch64-linux-android` Rust 目标。Android 最低版本为 8.0（API 26）。

```bash
npm ci
rustup target add aarch64-linux-android
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
export ANDROID_HOME="$HOME/Android/Sdk"
export NDK_HOME="$ANDROID_HOME/ndk/27.0.12077973"
npm run android:build
```

脚本生成 ARM64 debug APK，使用 Android debug 签名，供直接安装验证。产物位于 `src-tauri/gen/android/app/build/outputs/apk/arm64/debug/app-arm64-debug.apk`。正式分发需单独配置 release 签名，不将密钥提交到仓库。

首次打开 App 时输入手机可访问的 Go server 根地址，如 `https://audio.example.com` 或局域网地址 `http://192.168.1.20:8080`。不支持带路径前缀的地址。`127.0.0.1` 指手机自身；局域网直连时，Go server 需监听手机可访问的网卡地址。App 明确允许 HTTP 局域网连接，公网部署应使用 HTTPS。

启用访问控制时，App 显示共享访问密钥表单。认证和音频流共用原生内存 Cookie，会话不写入网页存储；App 进程或 Go server 重启后可能需要重新认证。App 不保存共享访问密钥。

点击音频资源后，Media3 前台服务播放整个队列，支持通知栏及锁屏播放、暂停、切换音频和拖动进度。返回页面会读取服务状态，不重新启动正在播放的音频。进度约每 4 秒保存一次，并在暂停和切换时保存；记录按 server 地址及资源 ID 隔离。重新打开 App 后展示最后资源和进度，点击播放恢复，不自动开启后台播放。

`src-tauri/gen/android/` 已纳入源码，包含项目实际使用的 Gradle 8.13、AGP 8.13.2 和 SDK 36 配置；常规构建无需再次 `tauri android init`，重新初始化会覆盖以下手工修补：`gradle.properties` 中的 `android.useAndroidX=true`（缺失导致 AndroidX 依赖校验失败）、`buildSrc` 中 `BuildTask.kt` 改为用 node 直接加载 `@tauri-apps/cli/tauri.js`（生成值把 `tauri` 当模块名，报 `MODULE_NOT_FOUND`）、`RustPlugin.kt` 的 `projectDir` 改用绝对路径（相对路径按 Gradle 进程工作目录解析，导致构建目录错位）。NDK 27 的 ARM64 链接参数已配置 16 KB 页对齐。

```bash
node --test app_js_test.js
go test ./...
```

设备验收：连接 server、开启/关闭认证、拖动进度、锁屏持续播放、锁屏自动下一项、通知栏操作、耳机断开暂停、回到页面状态同步、重开 App 续播、切换 server 后进度隔离，以及会话过期后的重新认证。单元测试和 APK 编译不能替代这些设备验证。
