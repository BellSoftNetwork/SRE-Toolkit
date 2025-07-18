# sysinfo
시스템의 다양한 정보를 수집하여 표시하는 CLI 도구입니다.



## 🚀 빠른 시작
```bash
# 기본 시스템 정보 표시
sysinfo
```

```bash
# 모든 정보 표시
sysinfo --all
```



## 📋 사용법
### 기본 정보 표시
```bash
sysinfo
```
운영체제, 호스트명, 가동시간 등 기본 정보만 표시합니다.

### 특정 정보 선택
```bash
# CPU 정보만
sysinfo --cpu
```

```bash
# 메모리 정보만
sysinfo --memory
```

```bash
# 디스크 정보만
sysinfo --disk
```

```bash
# 네트워크 정보만
sysinfo --network
```

### 여러 정보 조합
```bash
# CPU와 메모리 정보
sysinfo --cpu --memory
```

```bash
# CPU, 메모리, 디스크 정보
sysinfo --cpu --memory --disk
```

### 출력 형식 변경
```bash
# JSON 형식으로 출력
sysinfo --all --json
```

```bash
# JSON 파일로 저장
sysinfo --all --json > system_info.json
```

### 디버그 모드
```bash
# 상세 로그와 함께 실행
sysinfo --debug --all
```



## 🔧 옵션
| 옵션 | 설명 |
|------|------|
| `--all` | 모든 정보 표시 |
| `--cpu` | CPU 정보 표시 |
| `--memory` | 메모리 정보 표시 |
| `--disk` | 디스크 정보 표시 |
| `--network` | 네트워크 정보 표시 |
| `--json` | JSON 형식으로 출력 |
| `--config PATH` | 설정 파일 경로 지정 |
| `--debug` | 디버그 모드 활성화 |
| `--help` | 도움말 표시 |



## 💻 출력 예시
### 기본 출력
```
╭─ 시스템 정보 ────────────────────────────────────────╮
│ 운영체제    │ Linux                                  │
│ OS 버전     │ 6.6.87.2-microsoft-standard-WSL2      │
│ 아키텍처    │ x86_64                                 │
│ 호스트명    │ myhost                                 │
│ Python 버전 │ 3.12.0                                 │
│ 부팅 시간   │ 2025-01-13 10:00:00 UTC              │
│ 가동 시간   │ 5일 3시간 24분                         │
╰──────────────────────────────────────────────────────╯
```

### CPU 정보
```
╭─ CPU 정보 ───────────────────────────────────────────╮
│ 모델명      │ Intel(R) Core(TM) i7-9750H           │
│ 코어 수     │ 물리: 6, 논리: 12                     │
│ 현재 주파수 │ 2.60 GHz                              │
│ 사용률      │ 23.5%                                 │
╰──────────────────────────────────────────────────────╯
```

### JSON 출력
```json
{
  "system": {
    "운영체제": "Linux",
    "OS 버전": "6.6.87.2-microsoft-standard-WSL2",
    "아키텍처": "x86_64",
    "호스트명": "myhost",
    "Python 버전": "3.12.0",
    "부팅 시간": "2025-01-13 10:00:00 UTC",
    "가동 시간": "5일 3시간 24분"
  },
  "cpu": {
    "모델명": "Intel(R) Core(TM) i7-9750H",
    "물리 코어": 6,
    "논리 코어": 12,
    "현재 주파수": "2.60 GHz",
    "사용률": 23.5
  }
}
```



## 🔍 문제 해결
### Python 디렉토리 찾기
```bash
# sre-workbench 루트에서
cd python
```

### 가상환경 활성화 확인
```bash
# 프롬프트에 (.venv)가 표시되어야 함
which python
# 출력: /path/to/python/.venv/bin/python
```

### 의존성 재설치
```bash
# python 디렉토리에서
uv pip install -e .
```



## 📚 추가 정보
- 이 도구는 `psutil` 라이브러리를 사용하여 시스템 정보를 수집합니다
- 일부 정보는 운영체제나 권한에 따라 표시되지 않을 수 있습니다
- 개발 관련 정보는 [python/CLAUDE.md](../CLAUDE.md)를 참조하세요
