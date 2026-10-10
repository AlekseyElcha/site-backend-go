#!/bin/bash
set -e
COMPILED_NAME="compiled-project"
DEFAULT_VERSION="unknown"

VERSION=${1:-$DEFAULT_VERSION}
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

mkdir -p ./bin
env GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
  -o ./bin/$COMPILED_NAME ./cmd/main.go

echo "./bin/$COMPILED_NAME"