// [输入]: 内嵌 web 页面与 Android 播放插件。
// [输出]: Android 音频库 App。
// [定位]: Tauri 移动入口。
// [同步]: ../folder.md、../../README.md。
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_playback::init())
        .run(tauri::generate_context!())
        .expect("启动音频库 App 失败");
}
