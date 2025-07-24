# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with K8s-Diff.

## 프로젝트 개요

K8s-Diff는 두 Kubernetes 클러스터 간의 리소스 차이를 비교하는 Go 기반 CLI 도구입니다. Argus의 코드베이스를 기반으로 개발되었으며, 클러스터 마이그레이션이나 동기화 상태 확인에 활용됩니다.

## 개발 환경 설정

### 사전 요구사항
- Go 1.24.4 이상
- GitLab 저장소 접근 권한
- 두 개 이상의 Kubernetes 클러스터 접근 권한
- kubeconfig 설정

### 프로젝트 초기화
```bash
# 저장소 클론
git clone https://gitlab.bellsoft.net/devops/sre-workbench.git
cd sre-workbench/go

# 의존성 설치
go mod download

# K8s-Diff 디렉토리로 이동
cd cmd/k8s-diff
```

## 프로젝트 구조

### 디렉토리 구조
```
k8s-diff/
├── main.go              # 애플리케이션 진입점, CLI 처리
├── build.sh             # 멀티 플랫폼 빌드 스크립트
├── run.sh               # 플랫폼 감지 실행 스크립트
├── Makefile             # 빌드 자동화
├── rules.yaml           # 제외 규칙 설정 파일
├── README.md            # 사용자 가이드
├── CLAUDE.md            # 개발자 가이드
└── bin/                 # 빌드된 바이너리 (gitignore)
    ├── k8s-diff-linux-amd64
    ├── k8s-diff-darwin-amd64
    └── k8s-diff-windows-amd64.exe
```

### 내부 패키지 구조
```
internal/k8s-diff/
├── domain/              # 도메인 모델
│   └── resource.go      # 리소스 구조체, 비교 결과
├── service/             # 비즈니스 로직
│   └── scanner.go       # 클러스터 스캔 및 비교 서비스
├── analyzer/            # 분석 로직
│   ├── analyzer.go      # 리소스 비교 및 차이점 분석
│   └── converter.go     # 리소스 변환 유틸리티
├── reporter/            # 리포트 생성
│   ├── interface.go     # Reporter 인터페이스
│   ├── console.go       # 콘솔 출력
│   ├── html.go          # HTML 리포트
│   ├── markdown.go      # Markdown 리포트
│   └── image.go         # 이미지 리포트 생성
├── k8sclient/           # Kubernetes 클라이언트 래퍼
│   ├── factory.go       # 클라이언트 팩토리
│   └── wrapper.go       # K8s 클라이언트 인터페이스
├── utils/               # 유틸리티 함수
│   └── cluster.go       # 클러스터 관련 헬퍼
└── config/              # 설정 관리
    └── config.go        # 설정 구조체 및 로더
```

## 빌드 및 테스트

### 빌드
```bash
# 전체 플랫폼 빌드
./build.sh

# Makefile 사용
make all              # 전체 플랫폼
make local            # 현재 플랫폼만
make clean            # 빌드 정리
```

### 테스트
```bash
# 전체 테스트 (go 디렉토리에서)
cd ../../ && go test ./internal/k8s-diff/... -v

# 단위 테스트만
cd ../../ && go test ./internal/k8s-diff/... -short

# 커버리지 확인
cd ../../ && go test ./internal/k8s-diff/... -cover
```

### 개발 중 실행
```bash
# go run 사용
go run main.go -n default

# 또는 make 사용
make run ARGS="-n default"
```

## 코드 작성 가이드

### 핵심 컴포넌트

#### Domain 모델
```go
// KubernetesResource - 리소스 정보
type KubernetesResource struct {
    Namespace    string
    Kind         string
    APIVersion   string
    Name         string
    Labels       map[string]string
    Annotations  map[string]string
    CreationTime time.Time
    UID          string
    ResourceHash string // 내용 비교용 해시
}

// ComparisonResult - 비교 결과
type ComparisonResult struct {
    OnlyInSource      []KubernetesResource
    OnlyInTarget      []KubernetesResource
    ModifiedResources []ResourceDiff
}
```

#### Scanner Service
두 클러스터의 리소스를 수집하고 비교합니다.
```go
type ScannerService struct {
    sourceClient k8sinterface.K8sClient
    targetClient k8sinterface.K8sClient
    analyzer     *analyzer.Analyzer
    config       *config.Config
}
```

#### Analyzer
리소스 차이점을 분석합니다.
```go
func (a *Analyzer) CompareResources(source, target []KubernetesResource) ComparisonResult {
    // 리소스 맵 생성 및 비교 로직
}
```

### 성능 최적화 전략

#### 병렬 처리
- 네임스페이스별 병렬 스캔
- 리소스 타입 배치 처리
- 두 클러스터 동시 조회

#### 메모리 최적화
- 스트리밍 방식의 리소스 처리
- 필요한 필드만 저장
- 대용량 클러스터 고려

### 설정 파일 구조

```yaml
# 제외 규칙
exclusion_rules:
  - namespace: kube-system
    kind: "*"
    name: "*"
  - namespace: "*"
    kind: Event
    name: "*"

# 스캔에서 제외할 리소스 타입 (하위 리소스 제외)
skip_resource_types:
  - pods                    # Deployment에 의해 생성
  - replicasets.apps       # Deployment에 의해 생성
  - endpoints              # Service에 의해 생성
  - podmetrics.metrics.k8s.io  # 메트릭 데이터
  - nodemetrics.metrics.k8s.io # 메트릭 데이터

# 빠른 스캔 모드에서 확인할 리소스
important_resource_types:
  - deployments.apps
  - services
  - configmaps
  - secrets

# 비교 옵션
strict_api_version: false  # false: Kind만 비교 (기본값), true: API 버전도 비교

# 성능 설정
max_concurrent: 20
batch_size: 10
```

## 새 기능 추가 가이드

### 새 리포터 추가
1. `reporter/` 디렉토리에 새 파일 생성
2. `Reporter` 인터페이스 구현
3. `main.go`에서 리포터 등록 로직 추가

### 새 비교 로직 추가
1. `domain/resource.go`에 필요한 구조체 추가
2. `analyzer/analyzer.go`에 비교 로직 구현
3. 리포터에서 새 정보 표시

### CLI 옵션 추가
1. `main.go`의 `CLIFlags` 구조체에 추가
2. `parseCommandLineFlags()`에서 플래그 정의
3. 해당 로직 구현

## 디버깅 가이드

### 상세 로그 출력
```bash
# 현재는 컨텍스트를 통해 상세 정보 확인
kubectl config get-contexts
```

### 특정 네임스페이스 디버깅
```bash
# 단일 네임스페이스만 테스트
./run.sh -n default
```

### 성능 기능
```bash
# 동시 처리 수 조절로 성능 최적화
./run.sh -P 5  # 동시에 5개의 리소스만 처리

# 빠른 스캔 모드 사용
./run.sh -fast  # 중요 리소스만 비교
```

## 알려진 제한사항

1. **리소스 내용 비교**: 현재는 리소스 존재 여부만 비교 (해시 기반 비교는 옵션)
2. **CRD 지원**: Custom Resource는 기본 리소스 타입에 포함되지 않음
3. **대용량 클러스터**: 수천 개 이상의 리소스가 있는 경우 메모리 사용량 증가

## 향후 개선 사항

### 구현 예정 기능
1. **리소스 내용 차이점 상세 표시** - 현재는 존재 여부만 비교
2. **필터링 옵션 강화** - 라벨, 어노테이션 기반 필터
3. **증분 비교** - 이전 스캔 결과와 비교
4. **Slack/이메일 알림 지원**
5. **Web UI 대시보드**
6. **디버그 모드** - DEBUG 환경 변수 지원
7. **성능 프로파일링** - CPU/메모리 프로파일링 옵션

### 이미 구현된 기능
- **Markdown 리포터** - 이미 reporter/markdown.go에 구현됨
- **이미지 리포터** - reporter/image.go에 구현됨

## 트러블슈팅

### 컨텍스트 전환 오류
```bash
# kubeconfig 확인
kubectl config get-contexts

# 특정 kubeconfig 사용
export KUBECONFIG=/path/to/kubeconfig
```

### 권한 오류
```bash
# 필요한 권한 확인
kubectl auth can-i list deployments --all-namespaces --context=source-context
kubectl auth can-i list deployments --all-namespaces --context=target-context
```

### 메모리 부족
```bash
# 동시 처리 수 줄이기
./run.sh -P 5

# 빠른 스캔 모드 사용
./run.sh -fast
```
