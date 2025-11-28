#!/bin/bash

# MAIN_DIR=$(cd $(dirname "$0");cd ..;pwd)
MAIN_DIR=$(cd $(dirname "$0");pwd)
# echo $MAIN_DIR

FFMPEG_LIB=$MAIN_DIR/lib/ffmpeg_output
# FFMPEG_LIB=/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output

# 设置环境变量
# export CGO_CFLAGS="-I/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/include"
# export CGO_LDFLAGS="-L/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib"
# export PKG_CONFIG_PATH="/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib/pkgconfig"
# export LD_LIBRARY_PATH="/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib:$LD_LIBRARY_PATH"  # 运行时库路径


# echo $CGO_CFLAGS
export CGO_CFLAGS="-I$FFMPEG_LIB/include"
export CGO_LDFLAGS="-L$FFMPEG_LIB/lib"
export PKG_CONFIG_PATH="$FFMPEG_LIB/lib/pkgconfig"

# 确保FFmpeg库路径被添加到LD_LIBRARY_PATH
export LD_LIBRARY_PATH="$FFMPEG_LIB/lib:$LD_LIBRARY_PATH"  # 运行时库路径

# 检查LD_LIBRARY_PATH是否正确设置
echo "Current LD_LIBRARY_PATH: $LD_LIBRARY_PATH"

# 验证libpostproc.so.58是否存在
echo "Checking for libpostproc.so.58 in $FFMPEG_LIB/lib/"
if [ -f "$FFMPEG_LIB/lib/libpostproc.so.58" ]; then
    echo "✓ libpostproc.so.58 found"
else
    echo "✗ libpostproc.so.58 not found"
fi


# 在文件中添加拼接好的wav文件转码成mp3的功能
# go run .
# go run . -o output.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_019a8907.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_01bd5489.wav
# go run . -o output.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_019a8907.wav /home/zdz/Documents/Try/Python/course/http-server/go_client/server/output/1_01bd5489.wav
# go run . -o output.wav /home/zdz/Downloads/capgen_example.wav  /home/zdz/Downloads/capgen_example.wav

# go run . -o output.wav ./assets/capgen_example.wav  ./assets/aigei_com.wav
# go run . -o output.m4a ./assets/capgen_example.wav  ./assets/aigei_com.wav
# go run . -o output.mp3 ./assets/capgen_example.wav  ./assets/aigei_com.wav

# 构建后运行示例:
# ./ffmpeg-concat -o output.mp3 ./assets/capgen_example.wav ./assets/aigei_com.wav


## build ===============
# 编译cgo动态库版本
function build_dynamic() {
    echo "正在编译cgo动态库版本..."
    # 使用单引号包裹rpath参数，并用反斜杠转义$ORIGIN，确保它被正确传递给链接器
    CGO_ENABLED=1 go build -buildmode=c-shared -ldflags="-linkmode=external -extldflags='-Wl,-rpath,\$ORIGIN/lib/ffmpeg_output/lib'" -o libffmpegconcat.so .
    echo "cgo动态库版本编译完成: libffmpegconcat.so"
    
    # 验证生成的动态库依赖
    echo "检查动态库依赖..."
    ldd libffmpegconcat.so | grep libpostproc || echo "警告: libpostproc依赖未找到"
    
    # 显示动态库的rpath和runpath设置
    echo "检查动态库的rpath设置..."
    readelf -d libffmpegconcat.so | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
}

# 编译可执行文件版本（不依赖LD_LIBRARY_PATH）
function build_executable() {
    echo "正在编译可执行文件版本..."
    # 使用单引号包裹rpath参数，并用反斜杠转义$ORIGIN，确保它被正确传递给链接器
    # $ORIGIN表示可执行文件所在的目录
    CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags='-Wl,-rpath,\$ORIGIN/lib/ffmpeg_output/lib'" -o ffmpeg-concat .
    echo "可执行文件版本编译完成: ffmpeg-concat"
    
    # 验证生成的可执行文件依赖
    echo "检查可执行文件依赖..."
    ldd ffmpeg-concat | grep libpostproc || echo "警告: libpostproc依赖未找到"
    
    # 显示可执行文件的rpath和runpath设置
    echo "检查可执行文件的rpath设置..."
    readelf -d ffmpeg-concat | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
}

# 运行可执行文件（不依赖LD_LIBRARY_PATH）
function run_executable() {
    echo "正在运行可执行文件..."
    
    # 检查可执行文件是否存在
    if [ -f "./ffmpeg-concat" ]; then
        echo "执行: ./ffmpeg-concat (使用内置rpath查找库文件)"
        # 不设置LD_LIBRARY_PATH，让程序使用内置的rpath查找库文件
        ./ffmpeg-concat
    else
        echo "错误: 可执行文件 ./ffmpeg-concat 不存在，请先使用 build 命令编译"
        echo "使用方法: $0 build && $0 run"
    fi
}

# 使用RUNPATH替代RPATH的构建函数（现代Linux系统推荐）
function build_with_runpath() {
    echo "正在使用RUNPATH编译可执行文件版本..."
    # 使用RUNPATH替代RPATH，在现代Linux系统中优先级更高
    # 注意：使用双引号包裹整个ldflags参数，并用反斜杠转义内部的$ORIGIN，确保它被正确传递给链接器
    CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags='-Wl,--enable-new-dtags,-rpath,\$ORIGIN/lib/ffmpeg_output/lib'" -o ffmpeg-concat .
    echo "可执行文件版本编译完成: ffmpeg-concat"
    
    # 验证生成的可执行文件依赖
    echo "检查可执行文件依赖..."
    ldd ffmpeg-concat | grep libpostproc || echo "警告: libpostproc依赖未找到"
    
    # 显示可执行文件的rpath和runpath设置
    echo "检查可执行文件的rpath/runpath设置..."
    readelf -d ffmpeg-concat | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
}

# 使用patchelf工具修复已编译文件的rpath（备选方案）
function fix_rpath_with_patchelf() {
    echo "=== 使用patchelf工具修复rpath设置 ==="
    
    # 检查patchelf工具是否安装
    if command -v patchelf &> /dev/null; then
        echo "patchelf工具已安装"
        
        # 检查可执行文件是否存在
        if [ -f "./ffmpeg-concat" ]; then
            echo "修复可执行文件的rpath设置..."
            patchelf --set-rpath '$ORIGIN/lib/ffmpeg_output/lib' ./ffmpeg-concat
            echo "修复完成"
            
            # 显示修复后的rpath设置
            echo "检查修复后的rpath设置..."
            readelf -d ffmpeg-concat | grep -E 'RPATH|RUNPATH' || echo "警告: 未找到RPATH或RUNPATH设置"
        else
            echo "错误: 可执行文件 ./ffmpeg-concat 不存在"
        fi
    else
        echo "警告: patchelf工具未安装，请使用 'sudo apt-get install patchelf' 安装"
        echo "这是一个备选方案，如果前面的方法有效，可以忽略此警告"
    fi
}

# 测试不依赖LD_LIBRARY_PATH直接运行
function test_no_ld_library_path() {
    echo "=== 测试不依赖LD_LIBRARY_PATH直接运行 ==="
    
    # 首先编译可执行文件（使用RUNPATH方式）
    build_with_runpath
    
    # 可选：使用patchelf修复rpath（如果前面的方法失败）
    # fix_rpath_with_patchelf
    
    # 检查可执行文件是否存在
    if [ -f "./ffmpeg-concat" ]; then
        echo "\n=== 清除LD_LIBRARY_PATH并直接运行可执行文件 ==="
        # 保存当前LD_LIBRARY_PATH
        SAVED_LD_LIBRARY_PATH="$LD_LIBRARY_PATH"
        # 清除LD_LIBRARY_PATH
        export LD_LIBRARY_PATH=""
        echo "当前LD_LIBRARY_PATH已清空"
        
        # 直接运行可执行文件
        echo "执行: ./ffmpeg-concat"
        ./ffmpeg-concat
        
        # 恢复LD_LIBRARY_PATH
        export LD_LIBRARY_PATH="$SAVED_LD_LIBRARY_PATH"
        echo "LD_LIBRARY_PATH已恢复"
    else
        echo "错误: 可执行文件 ./ffmpeg-concat 不存在，编译失败"
    fi
}

# 默认运行模式
if [ "$1" = "build" ]; then
    build_executable
elif [ "$1" = "build-with-runpath" ]; then
    build_with_runpath
elif [ "$1" = "build-dynamic" ]; then
    build_dynamic
elif [ "$1" = "build-all" ]; then
    build_with_runpath  # 使用RUNPATH方式构建可执行文件
    build_dynamic
elif [ "$1" = "fix-rpath" ]; then
    fix_rpath_with_patchelf
elif [ "$1" = "run" ]; then
    run_executable
elif [ "$1" = "test-no-ld" ]; then
    test_no_ld_library_path
else
    # 默认运行程序
    echo "默认模式: 使用 go run 运行程序"
    go run .
    echo "提示: 可以使用 '$0 build && $0 run' 来编译并运行可执行文件"
    echo "提示: 可以使用 '$0 build-with-runpath' 来使用RUNPATH方式编译可执行文件"
    echo "提示: 可以使用 '$0 test-no-ld' 来测试不依赖LD_LIBRARY_PATH直接运行可执行文件"
    echo "提示: 如果rpath设置失败，可以尝试 '$0 fix-rpath' 使用patchelf工具修复"
fi














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