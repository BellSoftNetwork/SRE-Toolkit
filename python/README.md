# Python 기반 SRE 유틸리티
Python으로 작성된 SRE 도구 모음입니다. 각 도구는 독립적으로 실행 가능하며, 공용 라이브러리를 통해 일관된 기능을 제공합니다.



## 🚀 빠른 시작
### 환경 설정 (최초 1회)
```bash
# uv 설치 (Python 패키지 관리자)
curl -LsSf https://astral.sh/uv/install.sh | sh
```

```bash
# Python 디렉토리로 이동
cd python
```

```bash
# 가상환경 생성 및 패키지 설치
uv venv && source .venv/bin/activate && uv pip install -e .
```

### 도구 실행
#### sysinfo - 시스템 정보 출력
```bash
# 기본 정보 출력
sysinfo
```

```bash
# 모든 정보 출력
sysinfo --all
```

```bash
# CPU와 메모리 정보만 출력
sysinfo --cpu --memory
```

```bash
# JSON 형식으로 출력
sysinfo --all --json
```

```bash
# 도움말 보기
sysinfo --help
```



## 📦 사용 가능한 도구
| 도구명 | 설명 | 실행 명령 |
|--------|------|-----------|
| **sysinfo** | 시스템 정보 수집 및 출력 | `sysinfo` |

> 💡 **Tip**: 가상환경 활성화 후 도구 이름만 입력하면 바로 실행됩니다!  
> 각 도구의 상세 사용법은 `도구명 --help`로 확인하세요.



## 🛠️ 도구별 상세 가이드
### sysinfo
시스템의 다양한 정보를 수집하여 표시합니다.

**지원 기능:**
- 시스템 기본 정보 (OS, 호스트명, 가동시간)
- CPU 사용률 및 코어 정보
- 메모리 및 스왑 사용 현황
- 디스크 파티션 및 사용률
- 네트워크 인터페이스 정보

[상세 문서 →](sysinfo/README.md)



## 📚 추가 정보
### 시스템 요구사항
- Python 3.12 이상
- uv (Python 패키지 관리자)

### 프로젝트 구조
```
python/
├── common/           # 공용 모듈
├── sysinfo/          # 시스템 정보 도구
├── tests/            # 테스트 코드
├── pyproject.toml    # 프로젝트 설정
├── README.md         # 이 문서 (사용자 가이드)
└── CLAUDE.md         # 개발자 가이드
```

### 문제 해결
#### uv 명령을 찾을 수 없을 때
```bash
# PATH에 추가 (macOS/Linux)
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

#### 가상환경 활성화 확인
```bash
# 가상환경이 활성화되었는지 확인
which python
# 출력: /path/to/python/.venv/bin/python
```

### 개발자를 위한 정보
새로운 도구를 추가하거나 기존 도구를 개선하려면 [CLAUDE.md](./CLAUDE.md)를 참조하세요.
