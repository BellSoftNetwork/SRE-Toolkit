# Makefile for argus

APP_NAME := argus
GO_MODULE_ROOT := ../..
SOURCE_PATH := ./cmd/argus
LDFLAGS := -ldflags="-s -w"

# 기본 타겟
.PHONY: all
all: build

# 전체 플랫폼 빌드
.PHONY: build
build:
	@echo "🚀 Building for all platforms..."
	@cd $(GO_MODULE_ROOT) && ./cmd/argus/build.sh

# 현재 플랫폼용만 빌드 (bin 디렉토리에 저장)
.PHONY: local
local:
	@echo "🔨 Building for current platform..."
	@mkdir -p bin/current
	@cd $(GO_MODULE_ROOT) && go build -o cmd/argus/bin/current/$(APP_NAME) $(LDFLAGS) $(SOURCE_PATH)
	@echo "✅ Built: bin/current/$(APP_NAME)"
	@echo "💡 실행: ./run.sh"

# Windows 빌드
.PHONY: windows
windows:
	@echo "🔨 Building for Windows..."
	@mkdir -p bin/windows
	@cd $(GO_MODULE_ROOT) && GOOS=windows GOARCH=amd64 go build -o cmd/argus/bin/windows/$(APP_NAME).exe $(LDFLAGS) $(SOURCE_PATH)
	@echo "✅ Built: bin/windows/$(APP_NAME).exe"

# Linux 빌드
.PHONY: linux
linux:
	@echo "🔨 Building for Linux..."
	@mkdir -p bin/linux
	@cd $(GO_MODULE_ROOT) && GOOS=linux GOARCH=amd64 go build -o cmd/argus/bin/linux/$(APP_NAME) $(LDFLAGS) $(SOURCE_PATH)
	@echo "✅ Built: bin/linux/$(APP_NAME)"

# macOS Intel 빌드
.PHONY: darwin-amd64
darwin-amd64:
	@echo "🔨 Building for macOS (Intel)..."
	@mkdir -p bin/darwin/amd64
	@cd $(GO_MODULE_ROOT) && GOOS=darwin GOARCH=amd64 go build -o cmd/argus/bin/darwin/amd64/$(APP_NAME) $(LDFLAGS) $(SOURCE_PATH)
	@echo "✅ Built: bin/darwin/amd64/$(APP_NAME)"

# macOS Apple Silicon 빌드
.PHONY: darwin-arm64
darwin-arm64:
	@echo "🔨 Building for macOS (Apple Silicon)..."
	@mkdir -p bin/darwin/arm64
	@cd $(GO_MODULE_ROOT) && GOOS=darwin GOARCH=arm64 go build -o cmd/argus/bin/darwin/arm64/$(APP_NAME) $(LDFLAGS) $(SOURCE_PATH)
	@echo "✅ Built: bin/darwin/arm64/$(APP_NAME)"

# 실행 (현재 플랫폼)
.PHONY: run
run: local
	@./run.sh

# 정리
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f $(APP_NAME) $(APP_NAME).exe
	@echo "✅ Cleaned"

# 설치 (현재 플랫폼용을 시스템에 설치)
.PHONY: install
install: local
	@echo "📦 Installing to /usr/local/bin..."
	@sudo cp bin/current/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "✅ Installed to /usr/local/bin/$(APP_NAME)"

# 도움말
.PHONY: help
help:
	@echo "argus Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            Build for all platforms (default)"
	@echo "  build          Build for all platforms"
	@echo "  local          Build for current platform only"
	@echo "  windows        Build for Windows only"
	@echo "  linux          Build for Linux only"
	@echo "  darwin-amd64   Build for macOS Intel only"
	@echo "  darwin-arm64   Build for macOS Apple Silicon only"
	@echo "  run            Build and run locally"
	@echo "  clean          Remove all build artifacts"
	@echo "  install        Install to /usr/local/bin (requires sudo)"
	@echo "  help           Show this help message"
