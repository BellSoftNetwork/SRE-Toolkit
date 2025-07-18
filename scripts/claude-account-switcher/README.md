# Claude Account Switcher



## 소개
Claude Code CLI의 여러 계정을 쉽게 전환할 수 있는 도구입니다.

### 주요 특징
- 🔐 **안전한 계정 관리**: UUID 기반으로 계정 정보 격리
- 🏷️ **별칭 지원**: 기억하기 쉬운 이름으로 계정 관리
- 🔄 **빠른 전환**: 한 번의 명령으로 계정 전환
- 🎯 **자동완성**: Tab 키로 계정 이름 자동완성



## 설치
### 자동 설치 (권장)
1. 저장소 클론
```shell
git clone https://gitlab.bellsoft.net/devops/sre-toolkit.git
```

2. 설치 디렉토리로 이동
```shell
cd sre-toolkit/scripts/claude-account-switcher
```

3. 설치 스크립트 실행
```shell
./install.sh
```

설치 후 터미널을 재시작하거나 다음 명령 실행:

- Bash 사용자
```shell
source ~/.bashrc
```

- Zsh 사용자
```shell
source ~/.zshrc
```

### 제거
```shell
~/.claude/scripts/claude-account-switcher/uninstall.sh
```



## 사용법
### 계정 저장
현재 로그인된 계정을 저장합니다.

- 별칭으로 저장 (권장)
```shell
claude-save work
```

- 별칭 없이 저장 (이메일이 별칭이 됨)
```shell
claude-save
```

### 계정 전환
저장된 계정으로 전환합니다.

- 별칭으로 전환
```shell
claude-switch work
```

- Tab 자동완성 사용
```shell
claude-switch <Tab>
```

### 계정 확인
- 현재 계정 확인
```shell
claude-current
```

- 저장된 모든 계정 목록
```shell
claude-list
```



## 실제 사용 예시
### 업무/개인 계정 설정
1. 업무 계정으로 로그인
```shell
claude code /login
```

2. 업무 계정 저장
```shell
claude-save work
```

3. 로그아웃 후 개인 계정으로 로그인
```shell
claude code /logout
claude code /login
```

4. 개인 계정 저장
```shell
claude-save personal
```

### 계정 간 빠른 전환
- 업무 시작
```shell
claude-switch work
```

- 개인 프로젝트 작업
```shell
claude-switch personal
```

### 팀 계정 관리
- 프로젝트 A 계정 저장
```shell
claude-save project-a
```

- 프로젝트 B 계정 저장
```shell
claude-save project-b
```

- 개발팀 공용 계정 저장
```shell
claude-save dev-team
```

- 필요에 따라 전환
```shell
claude-switch project-a
```



## 명령어 정리
| 명령어 | 설명 | 예시 |
|--------|------|------|
| `claude-save [별칭]` | 현재 계정 저장 | `claude-save work` |
| `claude-switch <별칭>` | 계정 전환 | `claude-switch personal` |
| `claude-list` | 저장된 계정 목록 | `claude-list` |
| `claude-current` | 현재 계정 확인 | `claude-current` |



## 문제 해결
### 계정이 제대로 전환되지 않는 경우
1. Claude Code 완전히 종료
```shell
pkill -f "claude code"
```

2. 계정 전환
```shell
claude-switch work
```

3. Claude Code 재시작
```shell
claude code
```

### 자동완성이 작동하지 않는 경우
- Bash 사용자
```shell
source ~/.claude/scripts/claude-account-switcher/claude-completion.bash
```

- Zsh 사용자
```shell
source ~/.claude/scripts/claude-account-switcher/claude-completion.zsh
```

### 저장된 계정 삭제
1. 계정 디렉토리 확인
```shell
ls ~/.claude/accounts/
```

2. 특정 계정 삭제 (UUID 확인 후)
```shell
rm -rf ~/.claude/accounts/<uuid>
```



## 주의사항
- ⚠️ 계정 전환 후 Claude Code를 재시작하는 것을 권장합니다
- ⚠️ 계정 정보에는 인증 토큰이 포함되어 있으니 공유하지 마세요
- ⚠️ 활성 세션이 있는 경우 종료 후 전환하세요

## 추가 정보
- 개발 가이드: [CLAUDE.md](./CLAUDE.md)
- 상위 프로젝트: [SRE Toolkit](../../README.md)
- 문제 신고: [GitLab Issues](https://gitlab.bellsoft.net/devops/sre-toolkit/issues)
