# Argus
## 개요
Kubernetes 클러스터에서 ArgoCD로 관리되지 않는 리소스를 빠르게 찾아내는 도구
- **성능**: Kubernetes Go Client 사용으로 kubectl 대비 65% 빠른 속도
- **효율**: 병렬 처리와 스마트 필터링으로 대규모 클러스터도 빠르게 검사



## 전제 조건
- Go 1.24 이상 (설치 가이드는 [Go 개발 가이드](../../README.md) 참조)
- kubectl 설치 및 클러스터 접근 권한
- kubeconfig 설정 완료



## 빠른 시작 가이드
### 1. Go 환경 확인
```bash
go version
```

### 2. 의존성 설치
```bash
cd ../.. && go mod download
```

### 3. 빌드
```bash
./build.sh
```

### 4. kubeconfig 확인
```bash
kubectl config current-context
```

### 5. 첫 실행 (도움말)
```bash
./run.sh -h
```

### 6. 기본 실행
```bash
./run.sh
```

### 7. 특정 네임스페이스 스캔
```bash
./run.sh -n default,kube-system
```

### 8. 빠른 스캔 모드 (확인 없이)
```bash
./run.sh --fast -y
```

### 9. 리포트 생성 (확인 없이)
```bash
./run.sh --image -y
```

### 10. 정규식으로 네임스페이스 필터링
```bash
./run.sh -r "^prod-.*" -y
```



## 사용법
### 기본 실행
```bash
./run.sh
```

### 주요 옵션
#### 네임스페이스 지정
```bash
# 단일 네임스페이스
./run.sh monitoring

# 여러 네임스페이스
./run.sh -n app1,app2,app3

# 정규식 패턴
./run.sh -r "^prod-.*"
```

#### 빠른 스캔 모드
```bash
# 핵심 리소스만 검사 (5초 이내)
./run.sh --fast
```

#### 리포트 생성
```bash
# HTML/이미지 리포트 생성
./run.sh --image
```

### 옵션 목록
| 옵션 | 설명 | 예시 |
|------|------|------|
| `-n` | 네임스페이스 목록 | `-n ns1,ns2` |
| `-r` | 정규식 필터 | `-r "^prod-"` |
| `--exclude` | 제외 패턴 | `--exclude "-test$"` |
| `-y` | 확인 없이 실행 | `-y` |
| `--fast` | 빠른 스캔 모드 | `--fast` |
| `--image` | 이미지 리포트 생성 | `--image` |
| `-f` | 제외 규칙 파일 | `-f rules.txt` |
| `-P` | 동시 처리 수 | `-P 25` |



## 설정 파일 (rules.yaml)
모든 설정은 `rules.yaml` 파일에서 관리

### ArgoCD 리소스 식별
```yaml
argocd:
  managed_labels:
    - "argocd.argoproj.io/instance"
    - "app.kubernetes.io/instance"
  sync_annotations:
    - "argocd.argoproj.io/sync-wave"
```

### 제외 규칙
```yaml
exclusions:
  system_namespaces:
    - "kube-system/*/*"
    - "kube-public/*/*"
  auto_generated:
    - "*/ConfigMap/istio-ca-root-cert"
    - "*/ServiceAccount/default"
```

### 리소스 타입 설정
```yaml
resource_types:
  skip:  # 검사하지 않을 리소스
    - "events"
    - "pods"
  important:  # 빠른 스캔에서 검사할 리소스
    - "deployments.apps"
    - "services"
```

### 성능 설정
```yaml
performance:
  default_max_concurrent: 10
  fast_scan_concurrent: 15
```



## 문제 해결

### Go 모듈 오류 발생 시
```bash
cd ../.. && go clean -modcache && go mod tidy
```

### 빌드 오류 발생 시
```bash
go clean -cache && ./build.sh
```

### kubeconfig 오류 시
```bash
export KUBECONFIG=~/.kube/config && kubectl config view
```

### 빌드된 바이너리가 없을 경우
```bash
./build.sh
```

### 권한 오류 발생 시
```bash
chmod +x build.sh run.sh
```

## 개발자 가이드
### 개발 환경 설정
```bash
# Go 모듈 초기화 확인
cd ../.. && go mod tidy && cd -
```

### 테스트 실행
#### Argus 전체 테스트 실행
```bash
cd ../.. && go test ./internal/argus/... -v
```

#### 테스트 커버리지 확인
```bash
cd ../.. && go test ./internal/argus/... -cover
```

#### 특정 패키지 테스트
```bash
# analyzer 패키지만 테스트
cd ../.. && go test ./internal/argus/analyzer -v
```

```bash
# config 패키지만 테스트
cd ../.. && go test ./internal/argus/config -v
```

```bash
# service 패키지만 테스트
cd ../.. && go test ./internal/argus/service -v
```

#### 빠른 테스트 (캐시 사용)
```bash
cd ../.. && go test ./internal/argus/...
```

#### 테스트 결과 요약
```bash
cd ../.. && go test ./internal/argus/... | grep -E "(PASS|FAIL|ok)"
```

### 코드 검증
```bash
cd ../.. && go vet ./... && cd -
```

### 빌드 옵션
```bash
# 전체 플랫폼 빌드 (기본)
./build.sh

# Makefile 사용
make build

# 빌드 후 현재 플랫폼에서 실행
./run.sh
```

### 개발 중 빠른 테스트
```bash
# 빌드 없이 직접 실행 (개발용)
go run . -h
```

### 프로젝트 구조
- `main.go`: 애플리케이션 진입점
- `internal/argus/`: 핵심 비즈니스 로직
- `pkg/`: 재사용 가능한 공용 패키지

### 성능 최적화
1. **Kubernetes Go Client**: 직접 API 호출로 높은 성능
2. **병렬 처리**: 기본 20개 동시 처리
3. **연결 재사용**: HTTP/2 연결 풀링
4. **캐싱**: 리소스 타입 및 네임스페이스 정보 캐싱
