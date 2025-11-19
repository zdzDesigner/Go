#!/bin/bash

FFMPEG_LIB=/home/zdz/Documents/Try/Go/Go/cgo/ffmpeg/lib/ffmpeg_output

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
go run . -o output.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_019a8907.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_01bd5489.wav
