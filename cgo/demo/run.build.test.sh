#!/bin/bash

MAIN_DIR=$(cd $(dirname "$0");pwd)

export CGO_CFLAGS="-I$MAIN_DIR"
export CGO_LDFLAGS="-L$MAIN_DIR"
export LD_LIBRARY_PATH="$MAIN_DIR:$LD_LIBRARY_PATH"  # 运行时库路径


echo $CGO_LDFLAGS

set -e

echo "Building add executable..."

# go build -o add add.go
# CGO_ENABLED=1 go build -o add add.go
CGO_ENABLED=1 go build -ldflags="-linkmode=external -extldflags=-Wl,-rpath,'\$ORIGIN/'"  -o add add.go

echo "Build successful!"
echo "Run with: ./add"
