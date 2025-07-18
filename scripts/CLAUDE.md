# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with script tools in this repository.

## 스크립트 개발 공통 가이드

### 환경 요구사항
- Bash 4.0 이상 (macOS는 brew install bash 권장)
- 기본 Unix 도구들 (grep, sed, awk 등)
- jq (JSON 처리용, 선택사항)

### 디렉토리 구조
```
scripts/
└── <tool-name>/
    ├── <tool-name>.sh    # 메인 스크립트
    ├── install.sh        # 설치 스크립트
    ├── uninstall.sh      # 제거 스크립트 (선택)
    ├── README.md         # 사용자 문서
    └── lib/              # 공통 함수 (선택)
```

## 스크립트 작성 원칙

### 기본 구조
```bash
#!/bin/bash
set -euo pipefail

# 스크립트 설명
# 작성일: YYYY-MM-DD
# 용도: 간단한 설명

# 색상 정의
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# 스크립트 디렉토리
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
```

### 에러 처리
```bash
# 에러 처리 함수
error() {
    echo -e "${RED}오류: $1${NC}" >&2
    exit 1
}

# 성공 메시지
success() {
    echo -e "${GREEN}성공: $1${NC}"
}

# 경고 메시지
warning() {
    echo -e "${YELLOW}경고: $1${NC}"
}
```

### 플랫폼 감지
```bash
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "macos";;
        CYGWIN*|MINGW*|MSYS*) echo "windows";;
        *)          error "지원하지 않는 OS: $(uname -s)";;
    esac
}
```

### 함수 작성 규칙
- 함수명은 소문자와 언더스코어 사용
- 단일 책임 원칙 준수
- 복잡한 로직은 별도 함수로 분리
- 반환값 확인 필수

```bash
# 좋은 예
check_dependency() {
    local cmd="$1"
    if ! command -v "$cmd" &> /dev/null; then
        error "$cmd 명령어를 찾을 수 없습니다"
    fi
}

# 사용
check_dependency "git"
check_dependency "docker"
```

## 설치 스크립트 패턴

### 기본 구조
```bash
#!/bin/bash
set -euo pipefail

# 설치 디렉토리
readonly INSTALL_DIR="$HOME/.local/bin"
readonly CONFIG_DIR="$HOME/.config/<tool-name>"

# 설치 함수
install() {
    # 디렉토리 생성
    mkdir -p "$INSTALL_DIR" "$CONFIG_DIR"
    
    # 심볼릭 링크 생성
    ln -sf "$SCRIPT_DIR/<tool-name>.sh" "$INSTALL_DIR/<tool-name>"
    
    # Shell 설정
    setup_shell_completion
    
    success "설치가 완료되었습니다"
}

# Shell 자동완성 설정
setup_shell_completion() {
    # Bash
    if [[ -f "$HOME/.bashrc" ]]; then
        # 중복 방지
        grep -q "<tool-name>" "$HOME/.bashrc" || \
            echo "source $SCRIPT_DIR/completion.bash" >> "$HOME/.bashrc"
    fi
    
    # Zsh
    if [[ -f "$HOME/.zshrc" ]]; then
        grep -q "<tool-name>" "$HOME/.zshrc" || \
            echo "source $SCRIPT_DIR/completion.zsh" >> "$HOME/.zshrc"
    fi
}
```

## 코드 품질

### 검증 도구
```bash
# ShellCheck 사용 (설치: brew install shellcheck)
shellcheck *.sh
```

### 테스트
```bash
# 기본 테스트 구조
test_function() {
    local result
    result=$(my_function "input")
    
    if [[ "$result" != "expected" ]]; then
        error "테스트 실패: $result != expected"
    fi
    
    success "테스트 통과"
}

# dry-run 모드 지원
if [[ "${DRY_RUN:-false}" == "true" ]]; then
    echo "[DRY RUN] 실제 실행하지 않음"
fi
```

### 디버깅
```bash
# 디버그 모드
if [[ "${DEBUG:-false}" == "true" ]]; then
    set -x  # 명령어 출력
fi

# 디버그 로그
debug() {
    [[ "${DEBUG:-false}" == "true" ]] && echo "[DEBUG] $*" >&2
}
```

## 보안 고려사항
- 사용자 입력 검증 필수
- 파일 경로는 절대 경로 사용
- 임시 파일은 mktemp 사용
- 민감한 정보는 환경 변수나 설정 파일 사용

```bash
# 임시 파일 안전하게 사용
readonly TMP_FILE=$(mktemp)
trap "rm -f $TMP_FILE" EXIT

# 사용자 입력 검증
validate_input() {
    local input="$1"
    [[ "$input" =~ ^[a-zA-Z0-9_-]+$ ]] || error "잘못된 입력: $input"
}
```

## 문서화
- 각 함수 위에 용도 설명
- 복잡한 로직은 인라인 설명
- 사용 예제 포함
- 주요 변수는 readonly 선언
