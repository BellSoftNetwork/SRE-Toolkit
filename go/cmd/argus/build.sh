#!/bin/bash

# argus 멀티 플랫폼 빌드 스크립트

# 색상 정의
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 프로젝트 정보
APP_NAME="argus"

# Go 모듈 루트로 이동
cd "$(dirname "$0")/../.." || exit 1

# bin 디렉토리 생성
echo -e "${YELLOW}📁 Creating bin directory structure...${NC}"
mkdir -p cmd/argus/bin/{windows,linux,darwin/{amd64,arm64}}

# 빌드 함수
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT_DIR=$3
    local OUTPUT_NAME=$4

    echo -e "${YELLOW}🔨 Building for ${GOOS}/${GOARCH}...${NC}"

    GOOS=$GOOS GOARCH=$GOARCH go build -o "cmd/argus/$OUTPUT_DIR/$OUTPUT_NAME" -ldflags="-s -w" "./cmd/argus"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built: cmd/argus/$OUTPUT_DIR/$OUTPUT_NAME${NC}"
        # 파일 크기 표시
        size=$(du -h "cmd/argus/$OUTPUT_DIR/$OUTPUT_NAME" | cut -f1)
        echo -e "   Size: ${size}"
    else
        echo -e "${RED}✗ Failed to build for ${GOOS}/${GOARCH}${NC}"
        exit 1
    fi
}

echo -e "${GREEN}🚀 Starting multi-platform build for $APP_NAME${NC}"
echo ""

# Windows (amd64)
build "windows" "amd64" "bin/windows" "${APP_NAME}.exe"

# Linux (amd64)
build "linux" "amd64" "bin/linux" "$APP_NAME"

# macOS Intel (amd64)
build "darwin" "amd64" "bin/darwin/amd64" "$APP_NAME"

# macOS Apple Silicon (arm64)
build "darwin" "arm64" "bin/darwin/arm64" "$APP_NAME"

echo ""
echo -e "${GREEN}✅ Build completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📦 Build artifacts:${NC}"
echo "  • Windows (x64):        cmd/argus/bin/windows/${APP_NAME}.exe"
echo "  • Linux (x64):          cmd/argus/bin/linux/$APP_NAME"
echo "  • macOS (Intel):        cmd/argus/bin/darwin/amd64/$APP_NAME"
echo "  • macOS (Apple Silicon): cmd/argus/bin/darwin/arm64/$APP_NAME"
echo ""

# 빌드 완료 메시지 (루트 디렉토리에는 바이너리를 생성하지 않음)
echo -e "${YELLOW}💡 실행 방법:${NC}"
echo "   ./run.sh [옵션]"
echo ""
echo "   run.sh 스크립트가 자동으로 적절한 바이너리를 선택합니다."

echo ""
echo -e "${GREEN}Done! 🎉${NC}"
