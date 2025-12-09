#!/bin/bash

# MAIN_DIR=$(cd $(dirname "$0");cd ..;pwd)
MAIN_DIR=$(cd $(dirname "$0");pwd)
# echo $MAIN_DIR

FFMPEG_LIB=$MAIN_DIR/lib/ffmpeg_output
# FFMPEG_LIB=./lib/ffmpeg_output
# FFMPEG_LIB=/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output

# export CGO_CFLAGS="-I/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/include"
# export CGO_LDFLAGS="-L/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib"
# export PKG_CONFIG_PATH="/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib/pkgconfig"
# export LD_LIBRARY_PATH="/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib:$LD_LIBRARY_PATH"  # 运行时库路径


# echo $CGO_CFLAGS
export CGO_CFLAGS="-I$FFMPEG_LIB/include"
export CGO_LDFLAGS="-L$FFMPEG_LIB/lib"
export PKG_CONFIG_PATH="$FFMPEG_LIB/lib/pkgconfig"
export LD_LIBRARY_PATH="$FFMPEG_LIB/lib:$LD_LIBRARY_PATH"  # 运行时库路径


# 在文件中添加拼接好的wav文件转码成mp3的功能
# go run .
# go run . -o output.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_019a8907.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_01bd5489.wav
# go run . -o output.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_019a8907.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_01bd5489.wav
# go run . -o output.wav /home/zdz/Downloads/capgen_example.wav  /home/zdz/Downloads/capgen_example.wav

# go run . -o output.wav ./assets/capgen_example.wav  ./assets/aigei_com.wav
# go run . -o output.m4a ./assets/capgen_example.wav  ./assets/aigei_com.wav
# go run . -o output.mp3 ./assets/capgen_example.wav  ./assets/aigei_com.wav
#
# -o /home/zdz/Documents/Try/Go/cgo/ffmpeg/assets/capgen_example.wav /home/zdz/Documents/Try/Go/cgo/ffmpeg/assets/aigei_com.wav


# 使用 rpath 构建可执行文件（使用绝对路径进行最终诊断）
echo "正在构建（使用绝对路径 rpath）..."
CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags=-Wl,-rpath,'\$ORIGIN/lib/ffmpeg_output/lib'" -o ffmpeg-concat .
# CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags=-Wl,-rpath,'/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'" -o ffmpeg-concat .

echo ""
echo "构建完成！"
echo "这一个版本使用了绝对路径rpath。请尝试直接运行 ./ffmpeg-concat"

# 构建后运行示例:
# ./ffmpeg-concat -o output.mp3 ./assets/capgen_example.wav ./assets/aigei_com.wav

# CGO_ENABLED=1 go build -ldflags="-linkmode external -extldflags '-static'"
# CGO_ENABLED=1 go build .


# ffmpeg -f concat -safe 0 -i ./assets/capgen_example.wav  -i ./assets/aigei_com.wav  -c copy output_ff.wav
# ffmpeg  concat -safe 0 -i ./assets/aigei_com.wav  -i ./assets/aigei_com.wav  -c copy output_ff.wav
#
# ffmpeg -i ./assets/capgen_example.wav -i ./assets/aigei_com.wav -filter_complex "[0:a][1:a]concat=n=2:v=0:a=1[outa]" -map "[outa]" output_ff.wav
# 
# ffmpeg -i ./assets/capgen_example.wav -i ./assets/aigei_com.wav -filter_complex "[0:a][1:a]concat=n=2:v=0:a=1[a];[a]aformat=channel_layouts=mono[outa]" -map "[outa]" output_ff.wav


# ffmpeg -i ./assets/capgen_example.wav -i ./assets/aigei_com.wav \
# -filter_complex "[0:a][1:a]concat=n=2:v=0:a=1[a];[a]aformat=channel_layouts=mono[outa]" \
# -map "[outa]" output_ff.wav



# ./gemini/tmp/df1c52d5454219c65e54dd32d59518c3a7ad2f42c4f2f65e422d14dd8a9131f9
#
# --enable-libmp3lame
#
#
# readelf -d ./ffmpeg-concat
