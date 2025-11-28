#!/bin/bash

# 设置环境变量
export CGO_CFLAGS="-I/home/zdz/Documents/Try/Go/wails/cgo/lib/ffmpeg_output/include"
export CGO_LDFLAGS="-L/home/zdz/Documents/Try/Go/wails/cgo/lib/ffmpeg_output/lib"
export PKG_CONFIG_PATH="/home/zdz/Documents/Try/Go/wails/cgo/lib/ffmpeg_output/lib/pkgconfig"
export LD_LIBRARY_PATH="/home/zdz/Documents/Try/Go/wails/cgo/lib/ffmpeg_output/lib:$LD_LIBRARY_PATH"

# 构建项目（静态链接方式）
echo "Building project with static FFmpeg libraries..."
CGO_ENABLED=1 go build -ldflags '-extldflags "-static" -r /home/zdz/Documents/Try/Go/wails/cgo/lib/ffmpeg_output/lib' -o cgo-app .

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "To run the application, execute:"
    echo "  ./cgo-app"
else
    echo "Build failed!"
    exit 1
fi