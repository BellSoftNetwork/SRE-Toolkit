# Claude Code Account Switcher
Claude Code 계정을 빠르게 전환할 수 있는 스크립트


## 특징
- **별칭 기반 계정 관리**: 간단한 별칭으로 계정 관리 (권장)
- **이메일 기반 자동 저장**: 별칭 없이도 이메일로 자동 저장
- **UUID 기반 저장**: 계정별 고유 디렉토리에 안전하게 저장
- **메타데이터 관리**: 계정 정보와 저장 시간 기록
- **자동완성 지원**: Tab 키로 계정 이름 자동완성 (Bash/Zsh)



## 개요
Claude Code는 여러 파일에 세션 정보를 저장
- `~/.claude/.credentials.json` - OAuth 인증 토큰
- `~/.claude.json` - 사용자 ID, 프로젝트 기록
- `~/.claude/statsig/` - 세션 ID와 분석 데이터
- `~/.claude/settings.json` - 사용자 설정

이 스크립트는 모든 관련 파일을 백업하고 복원하여 완전한 계정 전환 지원



## 설치
### 자동 설치 (권장)
```bash
./install.sh
```
- `~/.claude/scripts/`에 필요한 파일 복사
- Shell RC 파일에 별칭 자동 추가 (중복 방지)
- 언인스톨 스크립트 생성


### 수동 설치
1. 스크립트 복사
```bash
mkdir -p ~/.claude/scripts
```

```bash
cp claude-switcher.sh ~/.claude/scripts/
```

```bash
chmod +x ~/.claude/scripts/claude-switcher.sh
```


2. ~/.bashrc 또는 ~/.zshrc에 추가
```bash
source ~/.claude/scripts/claude-aliases.sh
```



## 사용법
### 설치 후 별칭 사용 (권장)
#### 계정 저장
별칭으로 저장 (권장)
```bash
claude-save work
```

이메일로 자동 저장
```bash
claude-save
```

#### 계정 전환
자동완성으로 전환
```bash
claude-switch <Tab>
```

별칭으로 전환
```bash
claude-switch work
```

#### 계정 확인
```bash
claude-current
```

```bash
claude-list
```


### 직접 스크립트 실행
#### 계정 저장
```bash
./claude-switcher.sh save work
```

#### 계정 전환
```bash
./claude-switcher.sh switch work
```

#### 계정 목록 확인
```bash
./claude-switcher.sh list
```

#### 현재 계정 확인
```bash
./claude-switcher.sh current
```



## 예제
### 두 계정 설정하기
1. 첫 번째 계정으로 로그인
```bash
claude code /login
```

2. 계정 저장
```bash
claude-save work
```

3. 로그아웃 후 두 번째 계정으로 로그인
```bash
claude code /logout
claude code /login
```

4. 두 번째 계정 저장
```bash
claude-save personal
```


### 계정 간 전환
업무 계정으로 전환
```bash
claude-switch work
```

개인 계정으로 전환
```bash
claude-switch personal
```



## 주의사항
1. **계정 전환 후 Claude Code를 재시작**하는 것을 권장
2. 활성 세션이 있는 경우 종료 후 전환
3. 백업 파일은 `~/.claude/` 디렉토리에 저장됨



## 파일 구조
```
scripts/claude-account-switcher/
├── claude-switcher.sh        # 메인 전환 스크립트
├── claude-aliases.sh         # Shell 별칭 정의
├── claude-completion.bash    # Bash 자동완성
├── claude-completion.zsh     # Zsh 자동완성
├── install.sh               # 설치 스크립트
└── README.md                # 이 문서
```


### 계정 저장 구조
```
~/.claude/accounts/
└── <account-uuid>/
    ├── metadata.json       # 계정 메타데이터 (이메일, 별칭, UUID)
    ├── credentials.json    # OAuth 토큰
    ├── claude.json        # 사용자 설정
    ├── settings.json      # Claude 설정
    └── statsig/           # 세션 데이터
```



## 문제 해결
### 계정이 제대로 전환되지 않는 경우
1. Claude Code를 완전히 종료
2. 계정 전환
3. Claude Code 재시작

### 세션 충돌이 발생하는 경우
statsig 디렉토리 수동 삭제

```bash
rm -rf ~/.claude/statsig
```

```bash
mkdir ~/.claude/statsig
```

### 언인스톨
```bash
~/.claude/scripts/uninstall.sh
```
