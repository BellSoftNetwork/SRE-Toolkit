#!/bin/bash

set -e

BINARY_NAME="k8s-diff"
OUTPUT_DIR="bin"
MAIN_FILE="main.go"

# 색상 정의
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 출력 디렉토리 생성
mkdir -p $OUTPUT_DIR

echo -e "${BLUE}🔧 K8s-Diff 멀티 플랫폼 빌드 시작...${NC}"

# Go 모듈 정리
echo -e "${YELLOW}📦 의존성 정리 중...${NC}"
cd ../.. && go mod tidy && cd - > /dev/null

# 빌드 대상 플랫폼
platforms=(
    "linux/amd64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# 각 플랫폼별로 빌드
for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"

    output_name="$OUTPUT_DIR/$BINARY_NAME-$GOOS-$GOARCH"

    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi

    echo -e "${YELLOW}🔨 빌드 중: $GOOS/$GOARCH${NC}"

    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$output_name" "$MAIN_FILE"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 성공: $output_name${NC}"
        ls -lh "$output_name" | awk '{print "   크기:", $5}'
    else
        echo -e "${RED}❌ 실패: $GOOS/$GOARCH${NC}"
    fi
done

echo -e "\n${GREEN}✨ 빌드 완료!${NC}"
echo -e "${BLUE}📁 빌드된 파일:${NC}"
ls -la $OUTPUT_DIR/
