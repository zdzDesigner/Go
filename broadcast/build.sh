#!/bin/bash

# 当任何命令失败时，脚本将立即退出。
set -e

echo "正在为 Windows (amd64) 进行交叉编译..."

# 定义输出的可执行文件名。
OUTPUT_NAME="broadcast.exe"

# 设置目标操作系统 (GOOS) 和架构 (GOARCH)，然后执行构建命令。
# -o 标志指定输出文件的名称。
# "." 表示编译当前目录中的包。
GOOS=windows GOARCH=amd64 go build -o ${OUTPUT_NAME} .

echo "编译成功！"
echo "生成的可执行文件: ${OUTPUT_NAME}"
