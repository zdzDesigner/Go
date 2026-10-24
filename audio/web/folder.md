# H5 页面

- `index.html` 和 `style.css`：网页及 App 共享的音频库界面；server 配置与自定义播放控件仅在 App 启用。
- `app.js`：音频列表、认证与浏览器 `<audio>` 播放；检测 `window.audioNative` 后委托给 Android。
- `android.js`：先于 `app.js` 加载，通过 Tauri `playback` 插件请求 Go API、控制原生播放；页面可见时每秒同步状态。
- 普通网页使用同源 API 和 `localStorage`；App 使用原生 Cookie 与原生续播记录。

修改时同步 `../README.md`，并运行 `node --test app_js_test.js`。
