# Android 壳子

- `tauri.conf.json`：将 `../web` 嵌入 APK，启用全局 Tauri API，CSP 限制页面网络访问。
- `src/lib.rs`：移动入口，注册 `playback` 插件。
- `capabilities/default.json`：仅本地 `main` 窗口获得 Android 播放插件权限。
- `plugins/playback/`：Rust 注册、命令权限和 Kotlin 原生实现。
- `gen/android/`：已初始化的 Android 工程。Gradle 8.13、AGP 8.13.2、JDK 17+、SDK 36、最低 API 26；支持 HTTP 局域网连接；含 `android.useAndroidX` 与 `buildSrc` 的手工修补，重新 `tauri android init` 会覆盖。
- `icons/source.svg`：应用图标源文件。

构建入口是 `../build/android.sh`。修改原生能力时同步 `../README.md` 和插件文档。
