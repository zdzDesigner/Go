# 原生播放插件

- `src/lib.rs`：注册 `app.audio.playback.PlaybackPlugin`；`build.rs` 生成五个命令权限。
- `permissions/default.toml`：配置、请求、播放、控制、状态查询权限。
- `android/src/main/java/app/audio/playback/Server.kt`：校验 server 根地址，只允许音频列表及认证 API，内存 Cookie 和按 server 隔离的续播记录。
- `PlaybackPlugin.kt`：原生桥接，工作线程处理 HTTP，主线程通过 MediaController 控制服务。
- `PlaybackService.kt`：ExoPlayer 与 MediaSessionService、音频焦点、耳机断开处理、原生连播和后台续播。
- `android/src/main/AndroidManifest.xml`：媒体前台服务及网络、唤醒权限。

页面不得传入任意播放 URL，播放地址由 server 根地址与校验过的资源 ID 拼接。共享访问密钥不持久化。修改契约需同步 `../../../web/android.js` 与 `../../../README.md`。
