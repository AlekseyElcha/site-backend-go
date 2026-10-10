#!/bin/bash
set -e
COMPILED_NAME="compiled-project"

echo "Compiling project for macOS (Apple Silicon) for local use..."

mkdir -p ./bin

env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/${COMPILED_NAME}-mac-arm64 ./cmd/main.go

echo "./bin/$COMPILED_NAME"