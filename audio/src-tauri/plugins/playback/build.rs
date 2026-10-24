fn main() {
    tauri_plugin::Builder::new(&["configure", "request", "play", "control", "state"])
        .android_path("android")
        .build();
}
