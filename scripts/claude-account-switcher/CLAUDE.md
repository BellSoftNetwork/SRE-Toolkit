# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with Claude Account Switcher.

## Claude Account Switcher 개요
Claude Code CLI의 여러 계정을 쉽게 전환할 수 있는 도구입니다. UUID 기반으로 계정 정보를 안전하게 저장하고 관리합니다.

## 아키텍처

### 핵심 구성 요소
- **claude-switcher.sh**: 메인 스크립트 (계정 전환 로직)
- **claude-aliases.sh**: Shell 별칭 정의
- **claude-completion.bash/zsh**: 자동완성 스크립트
- **install.sh**: 자동 설치 스크립트
- **uninstall.sh**: 제거 스크립트 (설치 시 생성됨)

### 저장 구조
```
~/.claude/
├── accounts/
│   └── <uuid>/
│       ├── .credentials.json    # OAuth 토큰
│       ├── claude.json          # 사용자 설정
│       ├── settings.json        # Claude 설정
│       ├── statsig/            # 세션 데이터
│       └── metadata.json       # 계정 메타데이터
└── scripts/
    └── claude-account-switcher/ # 설치된 스크립트
```

## 주요 기능

### 계정 관리
- **저장**: 현재 Claude 세션을 별칭으로 저장
- **전환**: 저장된 계정으로 즉시 전환
- **목록**: 모든 저장된 계정 표시
- **현재**: 현재 사용 중인 계정 확인

### 안전성
- UUID 기반 저장으로 충돌 방지
- 계정별 완전 격리
- 메타데이터로 계정 추적

## 사용 명령어

```bash
# 현재 계정을 'work' 별칭으로 저장
claude-save work
```

```bash
# 'work' 계정으로 전환
claude-switch work
```

```bash
# 저장된 계정 목록 보기
claude-list
```

```bash
# 현재 계정 확인
claude-current
```

```bash
# 이메일을 별칭으로 빠른 저장
claude-save-default
```

## 코드 수정 가이드

### 새로운 파일 백업 추가
`claude-switcher.sh`의 `save_account()` 함수에서:
```bash
# 새 파일 추가
local new_file="$HOME/.config/claude/newfile.json"
if [[ -f "$new_file" ]]; then
    cp "$new_file" "$account_dir/"
fi
```

### 새로운 명령어 추가
1. `claude-aliases.sh`에 별칭 추가
2. `claude-switcher.sh`에 함수 구현
3. `claude-completion.bash/zsh`에 자동완성 추가

### 에러 메시지 개선
모든 에러 메시지는 한국어로 작성:
```bash
error "계정을 찾을 수 없습니다: $alias"
```

## 설치 및 제거

### 설치 과정
1. 스크립트를 `~/.claude/scripts/`에 복사
2. `~/.local/bin`에 심볼릭 링크 생성
3. Shell RC 파일에 자동완성 설정 추가
4. `uninstall.sh` 자동 생성

### 제거 과정
```bash
~/.claude/scripts/claude-account-switcher/uninstall.sh
```

## 테스트 방법

### 수동 테스트
```bash
# 1. 테스트 계정 저장
claude-save test

# 2. 다른 계정으로 전환
claude-switch test

# 3. 목록 확인
claude-list | grep test

# 4. 정리
rm -rf ~/.claude/accounts/*/test
```

### 디버그 모드
```bash
DEBUG=true claude-switch work
```

## 주의사항
- Claude Code가 설치되어 있어야 함
- 계정 전환 시 현재 세션이 덮어씌워짐
- 민감한 토큰 정보 포함 - 공유 금지
- macOS에서는 최신 bash 설치 권장

## 호환성
- Linux: 완전 지원
- macOS: 완전 지원 (bash 4.0+ 권장)
- Windows: WSL 환경에서만 지원
