// [输入]: 本地页面的播放及 API 指令。
// [输出]: Android 原生插件注册。
// [定位]: Tauri 与 Kotlin 的桥接入口。
// [同步]: ../../../folder.md。
use tauri::{plugin::{Builder, TauriPlugin}, Runtime};

pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::new("playback")
        .setup(|_app, api| {
            #[cfg(target_os = "android")]
            api.register_android_plugin("app.audio.playback", "PlaybackPlugin")?;
            Ok(())
        })
        .build()
}
