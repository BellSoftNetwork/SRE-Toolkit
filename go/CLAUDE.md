# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with Go code in this repository.

## Go 개발 환경 공통 가이드

### 환경 요구사항
- Go 1.24.4 이상
- GitLab 저장소 접근 권한 (내부 모듈 다운로드용)
- golangci-lint (선택사항, 코드 품질 검사용)

### 프로젝트 구조
```
go/
├── cmd/              # 실행 가능한 애플리케이션
│   └── <app-name>/   # 각 앱별 메인 패키지
├── internal/         # 내부 패키지 (외부 접근 불가)
│   └── <app-name>/   # 앱별 비즈니스 로직
├── pkg/              # 공용 패키지 (외부 접근 가능)
│   ├── k8s/          # Kubernetes 관련 유틸리티
│   └── utils/        # 공통 유틸리티
├── go.mod            # 모듈 정의
└── go.sum            # 의존성 체크섬
```

### 코드 작성 원칙

#### 클린 코드
- 함수는 15줄 이내로 작성 (복잡한 로직은 분리)
- 변수명과 함수명은 의미가 명확하게 작성
- 주석 대신 설명적인 함수명 사용
- 예: `CheckResourceManagement()` 대신 `isResourceManagedByArgoCD()`

#### 클린 아키텍처
```
cmd/app/
  └── main.go           # 의존성 주입, 설정
internal/app/
  ├── domain/          # 도메인 모델, 인터페이스
  ├── service/         # 비즈니스 로직
  ├── repository/      # 데이터 접근 계층
  └── handler/         # HTTP/CLI 핸들러
```

#### 에러 처리
- 에러는 즉시 처리하거나 상위로 전파
- 에러 메시지는 한국어로 작성
- 구조화된 에러 사용 권장
```go
return fmt.Errorf("리소스 조회 실패: %w", err)
```

#### 테스트
- 테이블 주도 테스트 사용
- 인터페이스를 통한 의존성 모킹
- 테스트 파일은 `_test.go` 접미사 사용
- BDD 스타일 테스트는 ginkgo/gomega 사용

### 공통 개발 명령어

```bash
# 의존성 관리
go mod tidy
```

```bash
# 전체 테스트
go test ./...
```

```bash
# 테스트 커버리지
go test -cover ./...
```

```bash
# 코드 검증
go vet ./...
```

```bash
# 포맷팅
go fmt ./...
```

### 패키지 임포트 순서
1. 표준 라이브러리
2. 외부 라이브러리
3. 내부 패키지

### 주의사항
- `internal/` 패키지는 해당 앱에서만 사용
- 공용 기능은 `pkg/`에 위치
- 순환 참조 주의
- 인터페이스는 사용하는 쪽에서 정의
