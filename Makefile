# Makefile for Weini Quantum Proxy

.PHONY: all build clean test lint run dev install-deps help

# 变量定义
APP_NAME=weini-quantum-proxy
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS=-ldflags "\
	-s -w \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GoVersion=$(GO_VERSION)"

# 默认目标
all: clean lint test build

# 帮助信息
help:
	@echo "Weini Quantum Proxy - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build           - Build all platforms"
	@echo "  make build-windows   - Build for Windows"
	@echo "  make build-linux     - Build for Linux"
	@echo "  make build-darwin    - Build for macOS"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linters"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make run             - Run in dev mode"
	@echo "  make install-deps    - Install dependencies"
	@echo ""

# 安装依赖
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing Wails..."
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "Installing linters..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Done!"

# 构建所有平台
build: build-windows build-linux build-darwin

# Windows构建
build-windows:
	@echo "Building for Windows..."
	@cd cmd/proxy-node-studio-wails && \
		wails build -clean -platform windows/amd64 $(LDFLAGS)
	@echo "Windows build complete!"

# Linux构建
build-linux:
	@echo "Building for Linux..."
	@cd cmd/proxy-node-studio-wails && \
		wails build -clean -platform linux/amd64 $(LDFLAGS)
	@echo "Linux build complete!"

# macOS构建
build-darwin:
	@echo "Building for macOS..."
	@cd cmd/proxy-node-studio-wails && \
		wails build -clean -platform darwin/universal $(LDFLAGS)
	@echo "macOS build complete!"

# 开发模式运行
dev:
	@echo "Starting development mode..."
	@cd cmd/proxy-node-studio-wails && wails dev

# 运行测试
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete!"

# 代码检查
lint:
	@echo "Running linters..."
	@golangci-lint run --timeout=5m ./...
	@echo "Lint complete!"

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist/
	@rm -rf build/
	@rm -rf cmd/proxy-node-studio-wails/build/
	@rm -f coverage.txt
	@echo "Clean complete!"

# 下载sing-box
fetch-singbox:
	@echo "Downloading sing-box..."
	@python3 scripts/fetch_sing_box.py
	@echo "sing-box downloaded!"

# 生成节点列表
update-nodes:
	@echo "Updating node list..."
	@python3 skills/research/weini-proxy/scripts/ss_crawler.py \
		--output /tmp/weini_proxy.xlsx \
		--csv-output /tmp/weini_nodes.csv \
		--json-output /tmp/weini_summary.json
	@echo "Node list updated!"

# Docker构建
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker build -t $(APP_NAME):latest .
	@echo "Docker build complete!"

# Docker运行
docker-run:
	@echo "Starting Docker container..."
	@docker-compose up -d
	@echo "Container started!"

# Docker停止
docker-stop:
	@echo "Stopping Docker container..."
	@docker-compose down
	@echo "Container stopped!"

# 发布
release: clean lint test build
	@echo "Creating release..."
	@mkdir -p dist/release
	@echo "Release $(VERSION) created!"

# 安装到本地
install: build
	@echo "Installing to local..."
	@echo "Installation path may vary by platform"

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

# 生成文档
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060
