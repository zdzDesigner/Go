#!/bin/bash

# 设置 FFmpeg 库路径
FFMPEG_LIB=/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output

# 设置环境变量
export CGO_CFLAGS="-I$FFMPEG_LIB/include"
export CGO_LDFLAGS="-L$FFMPEG_LIB/lib"
export PKG_CONFIG_PATH="$FFMPEG_LIB/lib/pkgconfig"
export LD_LIBRARY_PATH="$FFMPEG_LIB/lib:$LD_LIBRARY_PATH"

echo "开始测试不同音频格式输出..."

# 测试 M4A 输出
echo "测试 M4A 输出..."
if go run . -o output_test.m4a ./assets/capgen_example.wav ./assets/aigei_com.wav; then
    echo "✓ M4A 输出成功"
    ls -lh output_test.m4a
else
    echo "✗ M4A 输出失败"
fi

# 测试 MP3 输出（会失败，因为编码器不可用）
echo -e "\n测试 MP3 输出（预计失败）..."
if go run . -o output_test.mp3 ./assets/capgen_example.wav ./assets/aigei_com.wav; then
    echo "✓ MP3 输出成功"
    ls -lh output_test.mp3
else
    echo "✗ MP3 输出失败（这在我们的构建中是预期的，因为MP3编码器未启用）"
fi

echo -e "\n测试完成！"