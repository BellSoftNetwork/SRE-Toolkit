# Argus
## 소개
Argus는 Kubernetes 클러스터에서 ArgoCD로 관리되지 않는 리소스를 빠르게 찾아내는 도구입니다.

### 주요 특징
- ⚡ **빠른 성능**: kubectl 대비 65% 빠른 속도
- 🔍 **스마트 필터링**: 네임스페이스 및 리소스 타입별 필터링
- 📊 **다양한 리포트**: 콘솔, HTML, 이미지 형식 지원
- 🚀 **병렬 처리**: 대규모 클러스터도 빠르게 검사



## 설치
### 사전 요구사항
- Kubernetes 클러스터 접근 권한
- kubectl 설치 및 kubeconfig 설정

### 바이너리 다운로드
추후 예정

### 소스에서 빌드
1. 저장소 클론
```shell
git clone https://gitlab.bellsoft.net/devops/sre-toolkit.git
```

2. Argus 디렉토리로 이동
```shell
cd sre-toolkit/go/cmd/argus
```

3. 빌드 스크립트 실행 (모든 플랫폼)
```shell
./build.sh
```



## 사용법
### 기본 사용
- 전체 클러스터 스캔
```shell
./run.sh
```

- 도움말 확인
```shell
./run.sh -h
```

### 네임스페이스 필터링
- 특정 네임스페이스만 스캔
```shell
./run.sh -n default,monitoring
```

- 정규식으로 네임스페이스 필터링
```shell
./run.sh -r "^prod-.*"
```

- 특정 네임스페이스 제외
```shell
./run.sh --exclude ".*-test$"
```

### 스캔 모드
- 빠른 스캔 (중요 리소스만)
```shell
./run.sh --fast
```

- 확인 없이 자동 실행
```shell
./run.sh -y
```

- 빠른 스캔 + 자동 실행
```shell
./run.sh --fast -y
```

### 리포트 생성
- HTML 리포트 생성
```shell
./run.sh --image
```

- 특정 디렉토리에 리포트 저장
```shell
./run.sh --image --output ./reports
```



## 실행 예제
### 개발 환경 스캔
- dev- 로 시작하는 네임스페이스만 빠르게 스캔
```shell
./run.sh -r "^dev-" --fast -y
```

### 프로덕션 환경 전체 검사
- prod- 로 시작하는 네임스페이스 스캔 후 리포트 생성
```shell
./run.sh -r "^prod-" --image
```

### 특정 앱 네임스페이스 확인
- 지정한 네임스페이스만 스캔
```shell
./run.sh -n app-frontend,app-backend,app-database
```



## 설정
### 제외 규칙 설정
`rules.yaml` 파일을 통해 제외할 리소스를 설정할 수 있습니다:

```yaml
exclusions:
  # 시스템 네임스페이스 제외
  system_namespaces:
    - "kube-system/*/*"
    - "kube-public/*/*"
  
  # 자동 생성되는 리소스 제외
  auto_generated:
    - "*/ConfigMap/istio-ca-root-cert"
    - "*/ServiceAccount/default"
```

### 커스텀 설정 파일 사용
- 커스텀 규칙 파일 사용
```shell
./run.sh -f custom-rules.yaml
```



## 출력 예시
### 콘솔 출력
```
🔍 Argus - ArgoCD 미관리 리소스 탐지

네임스페이스: default
✗ Deployment/nginx-manual
✗ Service/nginx-service
✗ ConfigMap/app-config

총 3개의 미관리 리소스 발견
```

### HTML 리포트
브라우저에서 열 수 있는 대화형 HTML 리포트가 생성됩니다.



## 문제 해결
### kubeconfig 오류
1. kubeconfig 위치 확인
```shell
echo $KUBECONFIG
```

2. 기본 위치로 설정
```shell
export KUBECONFIG=~/.kube/config
```

3. 현재 컨텍스트 확인
```shell
kubectl config current-context
```

### 권한 오류
- 클러스터 접근 권한 확인
```shell
kubectl auth can-i list deployments --all-namespaces
```

### 느린 성능
- 빠른 스캔 모드 사용
```shell
./run.sh --fast
```

- 동시 처리 수 조정 (기본값: 20)
```shell
./run.sh -P 50
```

## 추가 정보
- 개발 가이드: [CLAUDE.md](./CLAUDE.md)
- 이슈 트래커: [GitLab Issues](https://gitlab.bellsoft.net/devops/sre-toolkit/issues)
- 상위 프로젝트: [SRE Toolkit](../../README.md)
