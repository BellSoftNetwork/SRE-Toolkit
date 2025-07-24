#!/bin/bash

# 플랫폼별 바이너리 선택 및 실행
BINARY_DIR="bin"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# 아키텍처 매핑
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
esac

# macOS의 경우 darwin으로 변경
if [ "$OS" = "darwin" ]; then
    OS="darwin"
fi

BINARY_NAME="k8s-diff-$OS-$ARCH"

if [ "$OS" = "windows" ]; then
    BINARY_NAME+=".exe"
fi

BINARY_PATH="$BINARY_DIR/$BINARY_NAME"

# 바이너리가 없으면 빌드
if [ ! -f "$BINARY_PATH" ]; then
    echo "바이너리가 없습니다. 빌드를 실행합니다..."
    ./build.sh
fi

# 바이너리 실행
if [ -f "$BINARY_PATH" ]; then
    exec "$BINARY_PATH" "$@"
else
    echo "❌ 바이너리를 찾을 수 없습니다: $BINARY_PATH"
    echo "지원되는 플랫폼이 아닐 수 있습니다. go run을 사용합니다."
    go run main.go "$@"
fi
