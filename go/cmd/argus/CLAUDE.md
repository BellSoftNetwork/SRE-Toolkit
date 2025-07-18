# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with Argus.

## Argus 개요
ArgoCD로 관리되지 않는 Kubernetes 리소스를 탐지하는 도구입니다. kubectl 대비 65% 빠른 성능을 제공합니다.

## 아키텍처

### 핵심 컴포넌트
- **ScannerService**: 리소스 스캔 및 수집
- **Analyzer**: 리소스 분석 및 ArgoCD 관리 여부 판단
- **Reporter**: 다양한 형식의 보고서 생성
- **Config**: 규칙 및 설정 관리

### 디렉토리 구조
```
argus/
├── main.go           # 진입점, 의존성 주입
├── build.sh          # 멀티 플랫폼 빌드 스크립트
├── run.sh            # 플랫폼 감지 실행 스크립트
├── Makefile          # 빌드 자동화
└── rules.yaml        # 제외 규칙 및 리소스 설정
```

## 빌드 및 실행

### 빌드 방법
```bash
# 전체 플랫폼 빌드
./build.sh
```

```bash
# Makefile 사용
make all              # 전체 플랫폼
make local            # 현재 플랫폼만
make windows          # Windows만
```

### 실행 방법
```bash
# 기본 실행
./run.sh
```

```bash
# 빠른 스캔 (중요 리소스만)
./run.sh --fast -y
```

```bash
# 특정 네임스페이스
./run.sh -n default,kube-system
```

## 주요 기능

### 성능 최적화
- 병렬 처리: 네임스페이스와 리소스 타입 동시 처리
- 배치 처리: API 호출 최소화
- 캐싱: 리소스 타입 정보 메모리 캐싱
- Go 네이티브 클라이언트 사용

### 필터링 옵션
- 네임스페이스 필터: `-n` 또는 `--namespaces`
- 정규식 필터: `-r` 또는 `--regex`
- 빠른 스캔: `--fast` (rules.yaml의 important 리소스만)
- 자동 확인: `-y` 또는 `--yes`

### 보고서 형식
- Console: 기본 터미널 출력 (색상 지원)
- Markdown: 문서용 마크다운 형식
- HTML: 웹 보고서
- Image: 시각적 요약 (--image 옵션)

## 설정 파일 (rules.yaml)

### 구조
```yaml
config:
  argocd:
    labels:            # ArgoCD 관리 라벨
    annotations:       # ArgoCD 관리 어노테이션
  excluded:
    namespaces:        # 제외할 네임스페이스
    resources:         # 제외할 리소스 타입
    names:             # 제외할 리소스 이름 패턴
  resource_types:
    important:         # 빠른 스캔 시 확인할 리소스
    all:              # 전체 스캔 시 확인할 리소스
```

### 수정 시 주의사항
- YAML 들여쓰기는 2 spaces 사용
- 리스트 항목은 `-`로 시작
- 정규식 패턴 사용 가능

## 개발 가이드

### 코드 수정 시
1. `internal/argus/service/scanner.go`: 스캔 로직 수정
2. `internal/argus/analyzer/`: 분석 규칙 변경
3. `internal/argus/reporter/`: 보고서 형식 추가
4. `internal/argus/config/`: 설정 구조 변경

### 테스트
```bash
# 단위 테스트
go test ./internal/argus/...
```

```bash
# 통합 테스트
go test -tags=integration ./...
```

### 새로운 리소스 타입 추가
1. `rules.yaml`에 리소스 타입 추가
2. 필요시 `analyzer`에 특별 처리 로직 추가
3. 테스트 케이스 작성

## 주의사항
- Kubernetes 클러스터 접근 권한 필요 (kubeconfig)
- 대규모 클러스터에서는 `--fast` 옵션 권장
- 민감한 네임스페이스는 rules.yaml에서 제외 설정
