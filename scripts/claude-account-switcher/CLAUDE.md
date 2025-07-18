# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with Claude Account Switcher.

## 프로젝트 개요
Claude Code CLI의 여러 계정을 UUID 기반으로 안전하게 관리하고 전환하는 Bash 스크립트입니다.

## 개발 환경 설정

### 사전 요구사항
- Bash 4.0 이상
- jq (JSON 처리용)
- Claude Code CLI 설치

### 프로젝트 초기화
```bash
# 저장소 클론
git clone https://gitlab.bellsoft.net/devops/sre-toolkit.git
cd sre-toolkit/scripts/claude-account-switcher

# 실행 권한 부여
chmod +x *.sh

# 테스트 실행
./claude-switcher.sh --help
```

## 프로젝트 구조

### 파일 구조
```
claude-account-switcher/
├── claude-switcher.sh        # 메인 스크립트
├── claude-aliases.sh         # Shell 별칭 정의
├── claude-completion.bash    # Bash 자동완성
├── claude-completion.zsh     # Zsh 자동완성
├── install.sh               # 설치 스크립트
├── README.md                # 사용자 가이드
└── CLAUDE.md                # 개발자 가이드 (이 파일)
```

### 데이터 저장 구조
```
~/.claude/
├── accounts/                 # 계정 백업 디렉토리
│   └── <uuid>/              # 계정별 고유 디렉토리
│       ├── metadata.json    # 계정 메타데이터
│       ├── .credentials.json # OAuth 토큰
│       ├── claude.json      # 사용자 설정
│       ├── settings.json    # Claude 설정
│       └── statsig/         # 세션 데이터 디렉토리
└── scripts/                 # 설치된 스크립트
    └── claude-account-switcher/
```

## 핵심 컴포넌트

### claude-switcher.sh
메인 스크립트로 모든 계정 관리 로직을 포함합니다.

#### 주요 함수
```bash
# 계정 저장
save_account() {
    local alias="$1"
    local uuid=$(uuidgen)
    local account_dir="$ACCOUNTS_DIR/$uuid"
    
    # 디렉토리 생성
    mkdir -p "$account_dir"
    
    # 파일 백업
    backup_claude_files "$account_dir"
    
    # 메타데이터 저장
    create_metadata "$account_dir" "$alias" "$uuid"
}

# 계정 전환
switch_account() {
    local alias="$1"
    local account_dir=$(find_account_by_alias "$alias")
    
    # 파일 복원
    restore_claude_files "$account_dir"
}

# 계정 목록
list_accounts() {
    # 모든 계정의 메타데이터 읽기
    for dir in "$ACCOUNTS_DIR"/*; do
        [[ -f "$dir/metadata.json" ]] && \
            jq -r '"\(.alias) - \(.email)"' "$dir/metadata.json"
    done
}
```

### 메타데이터 구조
```json
{
  "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "alias": "work",
  "email": "user@example.com",
  "created": "2025-01-18T10:30:45Z",
  "last_used": "2025-01-18T15:45:30Z"
}
```

## 코드 작성 가이드

### 새로운 Claude 파일 백업 추가
Claude가 새로운 설정 파일을 추가한 경우:

1. `CLAUDE_FILES` 배열에 추가:
```bash
readonly CLAUDE_FILES=(
    "$HOME/.claude/.credentials.json"
    "$HOME/.claude.json"
    "$HOME/.config/claude/settings.json"
    "$HOME/.claude/new-file.json"  # 새 파일 추가
)
```

2. 디렉토리인 경우 `CLAUDE_DIRS` 배열에 추가:
```bash
readonly CLAUDE_DIRS=(
    "$HOME/.claude/statsig"
    "$HOME/.claude/new-dir"  # 새 디렉토리 추가
)
```

### 에러 처리 개선
```bash
# 에러 처리 함수
error() {
    echo -e "${RED}오류: $1${NC}" >&2
    [[ -n "$2" ]] && echo -e "${YELLOW}해결방법: $2${NC}" >&2
    exit 1
}

# 사용 예시
error "계정을 찾을 수 없습니다: $alias" \
      "claude-list 명령으로 저장된 계정을 확인하세요"
```

### 새로운 명령 추가
1. 함수 작성:
```bash
delete_account() {
    local alias="$1"
    local account_dir=$(find_account_by_alias "$alias")
    
    confirm "정말로 '$alias' 계정을 삭제하시겠습니까?" || return 1
    
    rm -rf "$account_dir"
    success "'$alias' 계정이 삭제되었습니다"
}
```

2. 메인 케이스문에 추가:
```bash
case "$1" in
    # ... 기존 케이스들 ...
    delete)
        delete_account "$2"
        ;;
esac
```

3. 별칭 추가 (`claude-aliases.sh`):
```bash
claude-delete() {
    "$SCRIPT_DIR/claude-switcher.sh" delete "$@"
}
```

## 테스트

### 단위 테스트 예시
```bash
#!/bin/bash
# test_claude_switcher.sh

# 테스트 환경 설정
export CLAUDE_HOME="/tmp/test-claude"
export ACCOUNTS_DIR="/tmp/test-claude/accounts"

# 테스트 실행
test_save_account() {
    ./claude-switcher.sh save test-account
    
    # 검증
    [[ -d "$ACCOUNTS_DIR" ]] || error "accounts 디렉토리 생성 실패"
    
    local uuid_dir=$(ls -1 "$ACCOUNTS_DIR" | head -1)
    [[ -f "$ACCOUNTS_DIR/$uuid_dir/metadata.json" ]] || \
        error "메타데이터 생성 실패"
}

# 테스트 정리
cleanup() {
    rm -rf "$CLAUDE_HOME"
}

trap cleanup EXIT
```

### 디버깅
```bash
# 디버그 모드 활성화
DEBUG=true ./claude-switcher.sh save test

# 특정 함수 디버깅
bash -x claude-switcher.sh save test 2>&1 | grep save_account
```

## 설치 스크립트 (install.sh)

### 설치 프로세스
1. 대상 디렉토리 생성
2. 스크립트 파일 복사
3. 실행 권한 설정
4. Shell RC 파일 수정
5. 자동완성 설정
6. uninstall.sh 생성

### 안전한 RC 파일 수정
```bash
add_to_rc() {
    local rc_file="$1"
    local content="$2"
    
    # 중복 확인
    if ! grep -q "claude-account-switcher" "$rc_file"; then
        echo "$content" >> "$rc_file"
    fi
}
```

## 자동완성 구현

### Bash 자동완성
```bash
# claude-completion.bash
_claude_switch() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local accounts=$(claude-list | awk '{print $1}')
    
    COMPREPLY=($(compgen -W "$accounts" -- "$cur"))
}

complete -F _claude_switch claude-switch
```

### Zsh 자동완성
```zsh
# claude-completion.zsh
#compdef claude-switch

_claude_switch() {
    local accounts=($(claude-list | awk '{print $1}'))
    _describe 'accounts' accounts
}
```

## 보안 고려사항

### 파일 권한
```bash
# 계정 디렉토리 권한 설정
chmod 700 "$ACCOUNTS_DIR"
chmod 600 "$account_dir"/*.json
```

### 민감한 정보 보호
- `.credentials.json`은 OAuth 토큰 포함
- 백업 시 권한 유지
- 공유 시스템에서 주의

## 호환성

### 플랫폼별 처리
```bash
# UUID 생성 호환성
if command -v uuidgen &> /dev/null; then
    uuid=$(uuidgen)
elif [[ -f /proc/sys/kernel/random/uuid ]]; then
    uuid=$(cat /proc/sys/kernel/random/uuid)
else
    # 대체 방법
    uuid=$(date +%s%N | sha256sum | cut -c1-32)
fi
```

### Bash 버전 확인
```bash
if [[ "${BASH_VERSION%%.*}" -lt 4 ]]; then
    warning "Bash 4.0 이상을 권장합니다"
fi
```

## 트러블슈팅

### 일반적인 문제
1. **권한 오류**: 스크립트 실행 권한 확인
2. **경로 문제**: Claude 설치 경로 확인
3. **JSON 파싱 오류**: jq 설치 확인

### 디버그 정보 수집
```bash
debug_info() {
    echo "=== 디버그 정보 ==="
    echo "Bash 버전: $BASH_VERSION"
    echo "Claude 홈: $CLAUDE_HOME"
    echo "계정 디렉토리: $ACCOUNTS_DIR"
    echo "현재 사용자: $(whoami)"
    echo "운영체제: $(uname -s)"
}
```