# K8s-Diff

두 Kubernetes 클러스터 간의 리소스 차이를 비교하는 CLI 도구입니다.

## 주요 기능

- 🔍 두 클러스터 간 리소스 비교
- 📊 네임스페이스별 차이점 분석
- 🚀 병렬 처리로 빠른 스캔
- 📝 다양한 출력 형식 지원 (콘솔, HTML, Markdown)
- ⚙️ 유연한 설정 옵션

## 설치 방법

### 1. 저장소 클론

```shell
git clone https://gitlab.bellsoft.net/devops/sre-workbench.git
cd sre-workbench/go/cmd/k8s-diff
```

### 2. 빌드

전체 플랫폼 빌드:
```shell
make build
```

현재 플랫폼만 빌드:
```shell
make local
```

### 3. 실행 권한 설정

```shell
chmod +x ./run.sh
```

## 사용법

### 기본 사용법

두 클러스터 비교 (소스와 타겟 컨텍스트 필수):
```shell
./run.sh -source <source-context> -target <target-context>
```

실제 사용 예제:
```shell
./run.sh -source cluster1 \
         -target cluster2
```

### 네임스페이스 지정

특정 네임스페이스만 비교:
```shell
./run.sh -source <source-context> -target <target-context> -n default,kube-system,production
```

모든 네임스페이스 비교:
```shell
./run.sh -source <source-context> -target <target-context> -A
```

### 클러스터 지정 (필수)

소스와 타겟 클러스터 컨텍스트 지정:
```shell
./run.sh -source context1 -target context2
```

AWS EKS 클러스터 예제:
```shell
./run.sh -source arn:aws:eks:ap-northeast-2:{계정ID}:cluster/cluster1 \
         -target arn:aws:eks:ap-northeast-2:{계정ID}:cluster/cluster2
```

### 고급 옵션

확인 없이 바로 실행:
```shell
./run.sh -source <source-context> -target <target-context> -A -y
```

빠른 스캔 모드 (중요 리소스만):
```shell
./run.sh -source <source-context> -target <target-context> -fast -A -y
```

정밀 분석 모드 (API 버전까지 비교):
```shell
./run.sh -source <source-context> -target <target-context> -strict-api
```

콘솔 출력만:
```shell
./run.sh -source <source-context> -target <target-context> -o console
```

HTML 리포트만 생성:
```shell
./run.sh -source <source-context> -target <target-context> -o html
```

모든 형식으로 출력 (기본값):
```shell
./run.sh -source <source-context> -target <target-context>
# 또는 명시적으로
./run.sh -source <source-context> -target <target-context> -o "console,html,markdown"
```

## 실행 예제

### 예제 1: 기본 비교

```shell
./run.sh -source cluster2 \
         -target cluster2
```

출력:
```
🔍 K8s-Diff - Kubernetes 클러스터 비교 도구
소스 클러스터: cluster1 (컨텍스트: cluster1)
타겟 클러스터: cluster2 (컨텍스트: cluster2)

📋 비교할 네임스페이스 (1개):
  1. default

계속하시겠습니까? (y/N): y
```

### 예제 2: 여러 네임스페이스 비교

```shell
./run.sh -source cluster2 \
         -target cluster2 \
         -n default,production,staging -y
```

### 예제 3: 전체 클러스터 스캔

```shell
./run.sh -source cluster2 \
         -target cluster2 \
         -A -y -P 30
```

## 출력 형식

기본적으로 콘솔, HTML, Markdown 세 가지 형식으로 동시에 출력됩니다.

### 콘솔 출력
- 실시간 진행 상황 표시
- 전체 요약 통계
- 네임스페이스별 차이점 테이블
- 네임스페이스별로 그룹화된 상세 리소스 목록

### HTML 리포트
- `reports/` 디렉토리에 생성
- 시각적인 차이점 표시
- 네임스페이스별로 구분된 리소스 테이블
- 웹 브라우저에서 보기 편한 형식

### Markdown 리포트
- `reports/` 디렉토리에 생성
- 네임스페이스별로 구조화된 리포트
- 리소스 타입별 요약 테이블
- Git 저장소나 문서에 포함하기 적합

## 설정 파일

`rules.yaml` 파일로 비교 규칙 및 기본 동작 커스터마이징:

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

max_concurrent: 20
batch_size: 10
```

### 주요 옵션 설명

- **기본 동작**: Kind가 같은 리소스는 동일한 리소스로 간주 (예: apps/v1 Deployment = extensions/v1beta1 Deployment)
- **정밀 분석 (`-strict-api`)**: API 버전이 다르면 다른 리소스로 처리

## 문제 해결

### kubeconfig 관련 오류

kubeconfig 파일 경로 확인:
```shell
export KUBECONFIG=~/.kube/config
```

### 권한 오류

클러스터 접근 권한 확인:
```shell
kubectl auth can-i list deployments --all-namespaces
```

### 타임아웃 오류

동시 처리 수 줄이기:
```shell
./run.sh -P 5
```

## 주의사항

- 대규모 클러스터의 경우 스캔 시간이 오래 걸릴 수 있습니다
- 네트워크 상태에 따라 타임아웃이 발생할 수 있습니다
- 시스템 리소스 제외 규칙이 기본 적용됩니다
