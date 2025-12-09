#!/bin/bash

set -e

echo "Building add executable..."
go build -o add add.go

echo "Build successful!"
echo "Run with: ./add"
