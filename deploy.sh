#!/bin/bash
set -e

DEFAULT_VERSION="unknown"
VERSION=${1:-$DEFAULT_VERSION}
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

echo "Запуск сборки версии: $VERSION ($BUILD_TIME)..."

VERSION=$VERSION BUILD_TIME=$BUILD_TIME docker compose up -d --build

echo "Проект успешно собран и запущен!!!"