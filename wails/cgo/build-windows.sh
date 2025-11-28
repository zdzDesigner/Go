#!/bin/bash

# Windows cross-compilation script for cgo application
# This script builds a Windows executable from Linux

echo "Building Windows executable..."

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set environment variables for Windows cross-compilation
# Note: This requires MinGW-w64 toolchain to be installed
export CGO_ENABLED=1
export GOOS=windows
export GOARCH=amd64

# Set paths to FFmpeg libraries (adjust these paths as needed)
export CGO_CFLAGS="-I${SCRIPT_DIR}/lib/ffmpeg_output/include"
export CGO_LDFLAGS="-L${SCRIPT_DIR}/lib/ffmpeg_output/lib -lavcodec -lavformat -lavutil -lswresample -lswscale"

# Build for Windows
echo "Compiling for Windows..."
go build -o cgo-app.exe .

if [ $? -eq 0 ]; then
    echo "Windows build successful!"
    echo "Executable: cgo-app.exe"
    echo ""
    echo "IMPORTANT: This cross-compilation requires MinGW-w64 toolchain."
    echo "If you get errors, you may need to install it with:"
    echo "Ubuntu/Debian: sudo apt-get install gcc-mingw-w64-x86-64"
    echo "Fedora/RHEL: sudo dnf install mingw64-gcc"
    echo ""
    echo "Alternatively, you can build directly on Windows:"
    echo "1. Install Go for Windows"
    echo "2. Set up FFmpeg libraries for Windows"
    echo "3. Run: go build -o cgo-app.exe ."
else
    echo "Windows build failed!"
    echo "You may need to install the MinGW-w64 toolchain for cross-compilation."
    exit 1
fi