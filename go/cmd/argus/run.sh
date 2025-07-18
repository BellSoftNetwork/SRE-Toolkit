#!/bin/bash

# argus 자동 실행 스크립트
# OS와 아키텍처를 자동으로 감지하여 적절한 바이너리 실행

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 스크립트 디렉토리로 이동
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

# OS와 아키텍처 감지
detect_platform() {
    local os=""
    local arch=""

    # OS 감지
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)
            echo -e "${RED}❌ 지원하지 않는 운영체제입니다: $(uname -s)${NC}"
            exit 1
            ;;
    esac

    # 아키텍처 감지
    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64";;
        arm64|aarch64) arch="arm64";;
        *)
            echo -e "${RED}❌ 지원하지 않는 아키텍처입니다: $(uname -m)${NC}"
            exit 1
            ;;
    esac

    echo "$os:$arch"
}

# 바이너리 경로 결정
get_binary_path() {
    local platform=$1
    local os=$(echo $platform | cut -d: -f1)
    local arch=$(echo $platform | cut -d: -f2)

    case "$os" in
        "windows")
            echo "bin/windows/argus.exe"
            ;;
        "linux")
            echo "bin/linux/argus"
            ;;
        "darwin")
            echo "bin/darwin/$arch/argus"
            ;;
    esac
}

# 현재 플랫폼용 빌드 확인
check_current_platform_build() {
    if [ -f "bin/current/argus" ]; then
        echo "bin/current/argus"
        return 0
    fi
    return 1
}

# 메인 실행
main() {
    echo -e "${YELLOW}🔍 플랫폼 감지 중...${NC}"

    # 플랫폼 감지
    platform=$(detect_platform)
    os=$(echo $platform | cut -d: -f1)
    arch=$(echo $platform | cut -d: -f2)

    echo -e "${GREEN}✓ 감지된 플랫폼: $os/$arch${NC}"

    # 바이너리 경로 결정
    binary_path=$(get_binary_path $platform)

    # 현재 플랫폼용 빌드가 있는지 먼저 확인
    if current_build=$(check_current_platform_build); then
        binary_path="$current_build"
        echo -e "${GREEN}✓ 현재 플랫폼용 빌드 사용${NC}"
    fi

    # 바이너리 존재 확인
    if [ ! -f "$binary_path" ]; then
        echo -e "${RED}❌ 바이너리를 찾을 수 없습니다: $binary_path${NC}"
        echo -e "${YELLOW}💡 먼저 빌드를 실행해주세요:${NC}"
        echo "   ./build.sh"
        echo "   또는"
        echo "   make build"
        exit 1
    fi

    # 실행 권한 확인 (Windows가 아닌 경우)
    if [ "$os" != "windows" ] && [ ! -x "$binary_path" ]; then
        echo -e "${YELLOW}🔧 실행 권한 설정 중...${NC}"
        chmod +x "$binary_path"
    fi

    # 바이너리 실행
    echo -e "${GREEN}🚀 argus 실행${NC}"
    echo ""

    # 모든 인자를 그대로 전달
    exec "$binary_path" "$@"
}

# 스크립트 실행
main "$@"
