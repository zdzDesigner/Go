# 音频库

- Go 标准库服务：`main.go` 组装配置、音频库、访问控制和 HTTP 路由；`web/` 同时嵌入 Go server 与 Android APK。
- `src-tauri/`：Android 壳子和 Media3 原生播放插件，连接已有 Go API。
- `build/`：Go 开发启动、ARMv7 server 编译、Android ARM64 APK 构建脚本。
- `app_js_test.js`：网页续播与 Android 桥接适配测试；`*_test.go`：Go 行为测试。
- 领域语言见 `CONTEXT.md`；运行方式见 `README.md`；续播决策见 `docs/adr/`。

修改子目录行为时同步对应 `folder.md`，运行契约变化同步 `README.md` 与 `CONTEXT.md`。
