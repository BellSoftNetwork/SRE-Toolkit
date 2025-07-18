# Go 개발 가이드
SRE Toolkit의 Go 유틸리티 개발 가이드



## Go 설치 가이드
### Ubuntu/Debian
```bash
# Go 1.24 설치
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### macOS
```bash
# Homebrew 사용
brew install go
```

### Windows
```powershell
# Chocolatey 사용
choco install golang
```

### 설치 확인
```bash
go version
```

### GOPATH 설정 (선택사항)
```bash
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```



## 프로젝트 구조
```
go/
├── cmd/                    # 실행 가능한 유틸리티
│   └── <utility-name>/    # 각 유틸리티별 디렉토리
│       ├── main.go        # 진입점
│       ├── README.md      # 유틸리티 문서
│       └── ...
├── internal/              # 내부 전용 패키지
│   └── <utility-name>/   # 유틸리티별 내부 패키지
├── pkg/                   # 공용 패키지 (외부 사용 가능)
│   ├── k8s/              # Kubernetes 관련
│   └── utils/            # 범용 유틸리티
├── go.mod                 # 모듈 정의
└── go.sum                 # 의존성 체크섬
```



## 새 유틸리티 추가
### 1. 디렉토리 생성
```bash
mkdir -p go/cmd/my-tool
mkdir -p go/internal/my-tool
```

### 2. main.go 작성
```go
package main

import (
    "gitlab.bellsoft.net/devops/sre-toolkit/go/pkg/utils/color"
    // 필요한 공용 패키지 import
)

func main() {
    // 유틸리티 로직
}
```

### 3. README.md 작성
각 유틸리티는 다음 구조의 README.md 파일 포함 필요
- 간단한 설명
- 사용법
- 옵션 설명
- 예시



## 공용 패키지 사용
### 사용 가능한 패키지
#### k8s/client
Kubernetes 클라이언트 초기화 및 연결
```go
import "gitlab.bellsoft.net/devops/sre-toolkit/go/pkg/k8s/client"

k8sClient, err := client.NewClient()
```

#### utils/color
터미널 출력 색상 지원
```go
import "gitlab.bellsoft.net/devops/sre-toolkit/go/pkg/utils/color"

color.Success("작업 완료")
color.Error("오류 발생")
```



## 빌드 및 테스트
### 로컬 빌드
```bash
cd go/cmd/my-tool
go build -o my-tool
```

### 멀티 플랫폼 빌드
```bash
# build.sh 스크립트 복사 후 수정
cp go/cmd/argus/build.sh go/cmd/my-tool/
# APP_NAME 변수 수정 후 실행
./build.sh
```

### 테스트
#### 전체 프로젝트 테스트
```bash
go test ./...
```

#### 특정 앱 테스트
```bash
# Argus 앱 전체 테스트
go test ./internal/argus/... -v
```

#### 테스트 커버리지 확인
```bash
# 전체 프로젝트
go test ./... -cover
```

```bash
# 특정 앱 (예: Argus)
go test ./internal/argus/... -cover
```

#### 테스트 캐시 클리어
```bash
go clean -testcache
```

#### 벤치마크 테스트
```bash
go test -bench=. ./...
```



## 코딩 규칙
1. **패키지 구조**
   - 공용 기능은 `pkg/` 하위에 배치
   - 유틸리티 전용 코드는 `internal/<utility-name>/` 하위에 배치

2. **에러 처리**
   - 모든 에러는 적절히 처리하고 사용자 친화적 메시지 제공
   - 상세 로그는 `-v` 플래그로 제어

3. **문서화**
   - 모든 공용 함수와 타입에 GoDoc 주석 추가
   - 복잡한 로직은 인라인 주석으로 설명



## 의존성 관리
### 새 의존성 추가
```bash
cd go
go get github.com/some/package
go mod tidy
```

### 의존성 업데이트
```bash
cd go
go get -u ./...
go mod tidy
```



## CI/CD 고려사항
- 빌드 스크립트는 CI 환경에서도 동작하도록 작성
- 크로스 컴파일을 위한 GOOS/GOARCH 설정 포함
- 바이너리 크기 최적화 (`-ldflags="-s -w"`)
