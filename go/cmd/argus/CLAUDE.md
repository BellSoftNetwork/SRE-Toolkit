# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with Argus.

## 프로젝트 개요
Argus는 ArgoCD로 관리되지 않는 Kubernetes 리소스를 탐지하는 Go 기반 CLI 도구입니다.

## 개발 환경 설정

### 사전 요구사항
- Go 1.24.4 이상
- GitLab 저장소 접근 권한
- Kubernetes 클러스터 접근 권한
- golangci-lint (코드 품질 검사용)

### 프로젝트 초기화
```bash
# 저장소 클론
git clone https://gitlab.bellsoft.net/devops/sre-toolkit.git
cd sre-toolkit/go

# 의존성 설치
go mod download

# Argus 디렉토리로 이동
cd cmd/argus
```

## 프로젝트 구조

### 디렉토리 구조
```
argus/
├── main.go              # 애플리케이션 진입점, 의존성 주입
├── build.sh             # 멀티 플랫폼 빌드 스크립트
├── run.sh               # 플랫폼 감지 실행 스크립트
├── Makefile             # 빌드 자동화
├── rules.yaml           # 제외 규칙 및 리소스 설정
└── bin/                 # 빌드된 바이너리 (gitignore)
    ├── argus-linux-amd64
    ├── argus-darwin-amd64
    └── argus-windows-amd64.exe
```

### 내부 패키지 구조
```
internal/argus/
├── domain/              # 도메인 모델, 인터페이스
│   ├── resource.go      # Resource 구조체
│   └── interfaces.go    # 서비스 인터페이스
├── service/             # 비즈니스 로직
│   └── scanner.go       # 리소스 스캔 서비스
├── analyzer/            # 분석 로직
│   ├── analyzer.go      # ArgoCD 관리 여부 판단
│   └── rules.go         # 제외 규칙 처리
├── reporter/            # 리포트 생성
│   ├── console.go       # 콘솔 출력
│   ├── markdown.go      # 마크다운 리포트
│   ├── html.go          # HTML 리포트
│   └── image.go         # 이미지 리포트
└── config/              # 설정 관리
    ├── config.go        # 설정 구조체
    └── loader.go        # YAML 로더
```

## 빌드 및 테스트

### 빌드
```bash
# 전체 플랫폼 빌드
./build.sh

# Makefile 사용
make all              # 전체 플랫폼
make local            # 현재 플랫폼만
make windows          # Windows만
make clean            # 빌드 정리
```

### 테스트
```bash
# 전체 테스트 (프로젝트 루트에서)
cd ../.. && go test ./internal/argus/... -v

# 단위 테스트만
cd ../.. && go test ./internal/argus/... -short

# 커버리지 확인
cd ../.. && go test ./internal/argus/... -cover

# 특정 패키지 테스트
cd ../.. && go test ./internal/argus/analyzer -v
```

### 코드 검증
```bash
# 프로젝트 루트에서
cd ../.. && go vet ./...
cd ../.. && golangci-lint run
```

## 코드 작성 가이드

### 핵심 컴포넌트

#### Scanner Service
리소스 수집 및 병렬 처리를 담당합니다.
```go
// internal/argus/service/scanner.go
type ScannerService struct {
    client    kubernetes.Interface
    analyzer  domain.Analyzer
    config    *config.Config
}
```

#### Analyzer
ArgoCD 관리 여부를 판단합니다.
```go
// internal/argus/analyzer/analyzer.go
func (a *Analyzer) IsManaged(resource *domain.Resource) bool {
    // 라벨 및 어노테이션 확인 로직
}
```

#### Reporter
다양한 형식의 리포트를 생성합니다.
```go
// internal/argus/reporter/interface.go
type Reporter interface {
    Generate(resources []domain.Resource) error
}
```

### 성능 최적화 전략

#### 병렬 처리
```go
// 네임스페이스별 병렬 스캔
var wg sync.WaitGroup
for _, ns := range namespaces {
    wg.Add(1)
    go func(namespace string) {
        defer wg.Done()
        s.scanNamespace(namespace)
    }(ns)
}
wg.Wait()
```

#### 연결 재사용
```go
// HTTP/2 연결 풀링 활용
config.QPS = 100
config.Burst = 100
```

#### 캐싱
```go
// 리소스 타입 정보 캐싱
var resourceTypeCache = make(map[string]*metav1.APIResource)
```

## 설정 파일 (rules.yaml)

### 구조 설명
```yaml
config:
  argocd:
    labels:                    # ArgoCD 관리 리소스 식별 라벨
      - "argocd.argoproj.io/instance"
      - "app.kubernetes.io/instance"
    annotations:               # ArgoCD 관리 리소스 식별 어노테이션
      - "argocd.argoproj.io/sync-wave"
  
  excluded:
    namespaces:               # 제외할 네임스페이스 패턴
      - "kube-system"
      - "kube-public"
    resources:                # 제외할 리소스 타입
      - "events"
      - "pods"
    names:                    # 제외할 리소스 이름 패턴
      - "*/ServiceAccount/default"
      - "*/ConfigMap/istio-ca-root-cert"
  
  resource_types:
    important:                # 빠른 스캔 시 확인할 리소스
      - "deployments.apps"
      - "services"
      - "configmaps"
    all:                      # 전체 스캔 시 확인할 리소스
      - "deployments.apps"
      - "services"
      - "configmaps"
      - "secrets"
      - "ingresses.networking.k8s.io"
  
  performance:
    default_max_concurrent: 20    # 기본 동시 처리 수
    fast_scan_concurrent: 30      # 빠른 스캔 시 동시 처리 수
```

## 개발 작업 흐름

### 새 기능 추가
1. 도메인 모델 정의 (`internal/argus/domain/`)
2. 서비스 로직 구현 (`internal/argus/service/`)
3. 단위 테스트 작성
4. 통합 테스트 추가
5. 문서 업데이트

### 새 리포터 추가
1. `Reporter` 인터페이스 구현
2. `reporter/` 디렉토리에 새 파일 생성
3. `main.go`에서 리포터 등록
4. 테스트 케이스 추가

### 디버깅
```bash
# 상세 로그 출력
go run . -v

# 특정 네임스페이스만 테스트
go run . -n default --dry-run
```

## 릴리스 프로세스

### 버전 태깅
```bash
git tag v1.0.0
git push origin v1.0.0
```

### 빌드 및 배포
```bash
# 전체 플랫폼 빌드
make release

# 바이너리 업로드
# CI/CD 파이프라인 또는 수동으로 진행
```

## 주의사항

### 보안
- kubeconfig 파일 경로 노출 주의
- 민감한 네임스페이스 접근 시 권한 확인
- 빌드된 바이너리에 크레덴셜 포함 금지

### 성능
- 대규모 클러스터에서는 동시 처리 수 조정 필요
- API 서버 부하 고려하여 QPS 설정
- 메모리 사용량 모니터링

### 호환성
- Kubernetes API 버전 호환성 확인
- ArgoCD 버전별 라벨/어노테이션 차이 고려
- Go 모듈 버전 관리 철저히

## 트러블슈팅

### 빌드 실패
```bash
# 모듈 캐시 정리
go clean -modcache

# 의존성 재설치
go mod download
```

### 테스트 실패
```bash
# 특정 테스트만 실행
go test -run TestAnalyzer

# 상세 로그 확인
go test -v
```

### 런타임 오류
```bash
# kubeconfig 확인
export KUBECONFIG=~/.kube/config
kubectl config current-context

# 권한 확인
kubectl auth can-i list deployments --all-namespaces
```
