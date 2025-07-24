# SRE Workbench
SRE 작업을 위한 도구 모음



## 목차
- [도구 목록](#도구-목록)
  - [Go 유틸리티](#go-유틸리티)
  - [Python 유틸리티](#python-유틸리티)
  - [스크립트 도구](#스크립트-도구)
- [프로젝트 구조](#프로젝트-구조)
- [개발 가이드](#개발-가이드)



## 도구 목록
### Go 유틸리티
| 도구 | 설명 | 문서 |
|------|------|------|
| **Argus** | ArgoCD 미관리 K8s 리소스 탐지 | [README](go/cmd/argus/README.md) |
| **K8s-Diff** | 두 K8s 클러스터 간 리소스 비교 | [README](go/cmd/k8s-diff/README.md) |


### Python 유틸리티
| 도구 | 설명 | 문서 |
|------|------|------|
| **sysinfo** | 시스템 정보 출력 CLI 도구 | [README](python/sysinfo/README.md) |


### 스크립트 도구
| 도구 | 설명 | 문서 |
|------|------|------|
| **Claude Account Switcher** | Claude CLI 계정 전환 도구 | [README](scripts/claude-account-switcher/README.md) |



## 프로젝트 구조
```
sre-workbench/
├── go/                    # Go 유틸리티
│   ├── cmd/              # 실행 파일
│   ├── internal/         # 내부 패키지
│   └── pkg/              # 공용 패키지
├── python/               # Python 유틸리티
│   ├── sysinfo/          # 시스템 정보 도구
│   ├── common/           # 공통 모듈
│   └── tests/            # 테스트 코드
└── scripts/              # 스크립트 도구
```



## 개발 가이드
- **Go 개발**: [Go 개발 가이드](go/README.md) 참조
- **Python 개발**: [Python 개발 가이드](python/README.md) 참조
- **스크립트 개발**: [스크립트 가이드](scripts/README.md) 참조
