#!/bin/bash

# MAIN_DIR=$(cd $(dirname "$0");cd ..;pwd)
MAIN_DIR=$(cd $(dirname "$0");pwd)
# echo $MAIN_DIR

FFMPEG_LIB=$MAIN_DIR/lib/ffmpeg_output



# echo $CGO_CFLAGS
export CGO_CFLAGS="-I$FFMPEG_LIB/include"
export CGO_LDFLAGS="-L$FFMPEG_LIB/lib"
export PKG_CONFIG_PATH="$FFMPEG_LIB/lib/pkgconfig"

# 确保FFmpeg库路径被添加到LD_LIBRARY_PATH
export LD_LIBRARY_PATH="$FFMPEG_LIB/lib:$LD_LIBRARY_PATH"  # 运行时库路径

# 检查LD_LIBRARY_PATH是否正确设置
echo "Current LD_LIBRARY_PATH: $LD_LIBRARY_PATH"

# 注意：此项目是音频处理工具，不使用libpostproc（视频后处理库）
# 因此不需要检查libpostproc.so.58是否存在
# ===================================================================
# 内部库依赖无(RUNPATH) =============================================
# ===================================================================
# ➜ readelf -d ./lib/ffmpeg_output/lib/libavfilter.so.10
# Dynamic section at offset 0x45fd28 contains 37 entries:
#   标记        类型                         名称/值
#  0x0000000000000001 (NEEDED)             共享库：[libswscale.so.8]
#  0x0000000000000001 (NEEDED)             共享库：[libpostproc.so.58]
#  0x0000000000000001 (NEEDED)             共享库：[libavformat.so.61]
#  0x0000000000000001 (NEEDED)             共享库：[libavcodec.so.61]
#  0x0000000000000001 (NEEDED)             共享库：[libswresample.so.5]
#  0x0000000000000001 (NEEDED)             共享库：[libavutil.so.59]
#  0x0000000000000001 (NEEDED)             共享库：[libm.so.6]
#  0x0000000000000001 (NEEDED)             共享库：[libpthread.so.0]
#  0x0000000000000001 (NEEDED)             共享库：[libc.so.6]
#  0x000000000000000e (SONAME)             Library soname: [libavfilter.so.10]
#  0x0000000000000010 (SYMBOLIC)           0x0
#  0x000000000000000c (INIT)               0x84000

# ➜ readelf -d ffmpeg-concat
#
# Dynamic section at offset 0x2b4d70 contains 34 entries:
#   标记        类型                         名称/值
#  0x0000000000000001 (NEEDED)             共享库：[libavcodec.so.61]
#  0x0000000000000001 (NEEDED)             共享库：[libavdevice.so.61]
#  0x0000000000000001 (NEEDED)             共享库：[libavfilter.so.10]
#  0x0000000000000001 (NEEDED)             共享库：[libavformat.so.61]
#  0x0000000000000001 (NEEDED)             共享库：[libswresample.so.5]
#  0x0000000000000001 (NEEDED)             共享库：[libswscale.so.8]
#  0x0000000000000001 (NEEDED)             共享库：[libavutil.so.59]
#  0x0000000000000001 (NEEDED)             共享库：[libpthread.so.0]
#  0x0000000000000001 (NEEDED)             共享库：[libresolv.so.2]
#  0x0000000000000001 (NEEDED)             共享库：[libc.so.6]
#  0x000000000000001d (RUNPATH)            Library runpath: [$ORIGIN/lib/ffmpeg_output/lib]
#  0x000000000000000c (INIT)               0x40a000
#  0x000000000000000d (FINI)               0x5404c8


## build ===============
# 编译cgo动态库版本
function build_dynamic() {
    echo "正在编译cgo动态库版本..."
    # 使用单引号包裹rpath参数，并用反斜杠转义$ORIGIN，确保它被正确传递给链接器
    CGO_ENABLED=1 go build -buildmode=c-shared -ldflags="-linkmode=external -extldflags='-Wl,-rpath,\$ORIGIN/lib/ffmpeg_output/lib'" -o libffmpegconcat.so .
    echo "cgo动态库版本编译完成: libffmpegconcat.so"
    
    # 验证生成的动态库依赖
    echo "检查动态库依赖..."
    ldd libffmpegconcat.so | head -10  # 显示前10个依赖
    
    # 显示动态库的rpath和runpath设置
    echo "检查动态库的rpath/runpath设置..."
    readelf -d libffmpegconcat.so | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
}

# 编译可执行文件版本（不依赖LD_LIBRARY_PATH）
function build_executable() {
    echo "正在编译可执行文件版本..."
    # # 使用单引号包裹rpath参数，并用反斜杠转义$ORIGIN，确保它被正确传递给链接器
    # # $ORIGIN表示可执行文件所在的目录
    CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags='-Wl,-rpath,\$ORIGIN/lib/ffmpeg_output/lib'" -o ffmpeg-concat .
    # echo "可执行文件版本编译完成: ffmpeg-concat"
    # 
    # # 验证生成的可执行文件依赖
    # echo "检查可执行文件依赖..."
    # ldd ffmpeg-concat | head -10  # 显示前10个依赖
    # 
    # # 显示可执行文件的rpath/runpath设置
    # echo "检查可执行文件的rpath/runpath设置..."
    # readelf -d ffmpeg-concat | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
    # CGO_ENABLED=1 go build -ldflags="-linkmode=external" -o ffmpeg-concat .
}

function dist(){
  if [[ ! -d "./dist" ]]; then
     mkdir dist
  fi
  rm -r dist/*
  cp ./ffmpeg-concat ./dist/
  cp -r ./lib/ffmpeg_output/lib/* ./dist/
}

# 默认运行模式
if [ "$1" = "build" ]; then
    build_executable
    dist
elif [ "$1" = "build-dynamic" ]; then
    build_dynamic
fi
