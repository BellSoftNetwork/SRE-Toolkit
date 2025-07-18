# CLAUDE.md
This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.



## 프로젝트 개요
SRE Toolkit은 다양한 SRE 도구들을 모아둔 모노레포입니다. 각 언어별로 디렉토리를 구성하여 독립적인 유틸리티들을 개발하고 관리합니다.



## 모노레포 구조
```
sre-toolkit/
├── go/                    # Go 기반 유틸리티
│   ├── cmd/              # 실행 가능한 애플리케이션
│   ├── internal/         # 내부 패키지 (도구별)
│   └── pkg/              # 공용 패키지
├── scripts/              # 스크립트 기반 도구
│   └── <tool-name>/      # 각 도구별 디렉토리
├── CLAUDE.md             # 모노레포 공통 가이드
├── TODO.md               # 프로젝트 개발 요구사항
└── README.md             # 프로젝트 소개
```



## 개발 원칙
- **한국어 우선**: 기술적 제약이 없는 한 모든 문서와 주석은 한국어로 작성
- **테스트 주도 개발**: BDD/TDD 방법론을 기반으로 개발
- **독립적 유틸리티**: 각 도구는 독립적으로 빌드, 실행, 배포 가능
- **사용자 중심 문서**: README는 사용자 관점을 우선으로 작성
- **클린 코드**: 명확한 함수명과 변수명으로 주석 없이도 이해 가능한 코드 작성
- **클린 아키텍처**: 계층 분리와 의존성 역전으로 유지보수와 확장이 용이한 구조
- **낮은 복잡도**: 함수는 단일 책임 원칙을 따르고, 복잡한 로직은 작은 함수로 분리
- **즉시 정리**: 사용하지 않는 코드는 즉시 제거하여 코드베이스를 깔끔하게 유지



## 유틸리티별 작업 방법
각 유틸리티의 상세 정보는 해당 디렉토리의 CLAUDE.md 또는 README.md를 참조하세요:
- Go 유틸리티: `go/cmd/<utility-name>/CLAUDE.md`
- 스크립트 도구: `scripts/<tool-name>/CLAUDE.md`



## 공통 개발 규칙
### 문서 작성
- 실행 명령은 단일 기능 단위로 코드 블록 분리 (IDE에서 원클릭 실행 가능)
- README 구성 순서: 용도 → 환경 설정 → 사용법 → 개발 가이드

### 디렉토리 구조
- 각 언어별로 표준 디렉토리 구조 준수
- 유틸리티별로 독립된 빌드/실행 환경 구성

### Go 모듈
- 모듈 경로: `gitlab.bellsoft.net/devops/sre-toolkit/go`
- 내부 패키지: `internal/<tool-name>/`
- 공용 패키지: `pkg/`



## 현재 유틸리티 목록
- **Argus** (Go): ArgoCD로 관리되지 않는 K8s 리소스 탐지
- **Claude Account Switcher** (Script): Claude Code CLI 계정 전환 도구

각 유틸리티의 빌드, 실행, 개발 방법은 해당 디렉토리의 문서를 참조하세요.
