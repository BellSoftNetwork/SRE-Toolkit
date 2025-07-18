# 스크립트 개발 가이드
SRE Toolkit의 스크립트 도구 개발 가이드



## 프로젝트 구조
```
scripts/
└── <tool-name>/           # 각 도구별 디렉토리
    ├── README.md          # 도구 문서
    ├── install.sh         # 설치 스크립트 (선택)
    └── <main-script>.sh   # 메인 스크립트
```



## 새 스크립트 도구 추가
### 1. 디렉토리 생성
```bash
mkdir -p scripts/my-tool
```

### 2. 메인 스크립트 작성
```bash
#!/bin/bash
set -euo pipefail

# 스크립트 설명
# 작성자: 이름
# 작성일: YYYY-MM-DD

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 메인 로직
main() {
    echo -e "${GREEN}도구 실행중...${NC}"
    # 실제 로직 구현
}

# 스크립트 실행
main "$@"
```

### 3. README.md 작성
각 도구는 다음 구조의 README.md 파일 작성 필요
- 도구 설명
- 설치 방법
- 사용법
- 옵션 설명
- 예시

### 4. 실행 권한 설정
```bash
chmod +x scripts/my-tool/*.sh
```



## 스크립팅 규칙
### Shell 스크립트 기본 설정
```bash
#!/bin/bash
set -euo pipefail  # 엄격한 오류 처리
```

### 색상 출력
```bash
# 표준 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 사용 예시
echo -e "${GREEN}✓ 성공${NC}"
echo -e "${RED}✗ 실패${NC}"
echo -e "${YELLOW}⚠ 경고${NC}"
```

### 오류 처리
```bash
# 오류 핸들러
error_exit() {
    echo -e "${RED}오류: $1${NC}" >&2
    exit 1
}

# 사용 예시
command || error_exit "명령 실행 실패"
```

### 도움말 함수
```bash
show_help() {
    cat << EOF
사용법: $(basename "$0") [옵션]

옵션:
  -h, --help     도움말 표시
  -v, --version  버전 표시

예시:
  $(basename "$0") -h
EOF
}
```



## 크로스 플랫폼 호환성
### OS 감지
```bash
detect_os() {
    case "$(uname -s)" in
        Linux*)   echo "linux" ;;
        Darwin*)  echo "macos" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)        echo "unknown" ;;
    esac
}
```

### 조건부 실행
```bash
OS=$(detect_os)
if [[ "$OS" == "macos" ]]; then
    # macOS 전용 코드
elif [[ "$OS" == "linux" ]]; then
    # Linux 전용 코드
fi
```



## 설치 스크립트 작성
### install.sh 템플릿
```bash
#!/bin/bash
set -euo pipefail

INSTALL_DIR="$HOME/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 설치 디렉토리 생성
mkdir -p "$INSTALL_DIR"

# 스크립트 복사 또는 심볼릭 링크
ln -sf "$SCRIPT_DIR/my-tool.sh" "$INSTALL_DIR/my-tool"

# PATH 설정 안내
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "PATH에 $INSTALL_DIR 추가 필요:"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo "설치 완료!"
```



## 테스트
### 기본 테스트
```bash
# Shellcheck로 문법 검사
shellcheck scripts/my-tool/*.sh

# 실행 테스트
bash -n scripts/my-tool/main.sh  # 문법 검사만
```

### 다양한 환경 테스트
- Bash 버전 호환성 확인
- macOS, Linux에서 테스트
- 필요한 명령어 존재 여부 확인



## 문서화
1. **인라인 주석**: 복잡한 로직에 설명 추가
2. **함수 주석**: 각 함수의 목적과 매개변수 설명
3. **README 업데이트**: 새 도구 추가 시 최상위 README에도 추가
