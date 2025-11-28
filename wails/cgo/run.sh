#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 设置库路径并运行应用程序
export LD_LIBRARY_PATH="${SCRIPT_DIR}/lib/ffmpeg_output/lib:$LD_LIBRARY_PATH"

# 运行应用程序
./cgo-app