.PHONY: all build local clean run help

BINARY_NAME := k8s-diff
OUTPUT_DIR := bin
MAIN_FILE := main.go

# 기본 타겟
all: build

# 전체 플랫폼 빌드
build:
	@echo "🔧 멀티 플랫폼 빌드 시작..."
	@./build.sh

# 현재 플랫폼만 빌드
local:
	@echo "🔨 현재 플랫폼 빌드..."
	@mkdir -p $(OUTPUT_DIR)
	@go build -ldflags="-s -w" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ 빌드 완료: $(OUTPUT_DIR)/$(BINARY_NAME)"

# 빌드 정리
clean:
	@echo "🧹 빌드 정리..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✅ 정리 완료"

# 실행
run:
	@./run.sh $(ARGS)

# 도움말
help:
	@echo "사용 가능한 명령:"
	@echo "  make all    - 모든 플랫폼 빌드 (기본)"
	@echo "  make local  - 현재 플랫폼만 빌드"
	@echo "  make clean  - 빌드 결과물 정리"
	@echo "  make run    - 프로그램 실행"
	@echo ""
	@echo "실행 예제:"
	@echo "  make run ARGS=\"-n default,kube-system\""
	@echo "  make run ARGS=\"-A -y\""
