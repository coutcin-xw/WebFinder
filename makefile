# 定义变量
GO = go
GOFMT = gofmt
GO_BUILD_ARGS = -ldflags "-X WebFinder/cmd.buildTime=$(shell date '+%Y-%m-%dT%H:%M:%S')"
# 默认编译目标平台和架构
TARGET_OS ?= linux
TARGET_ARCH ?= amd64

# 是否需要静态链接
STATIC_BUILD ?= true

# 根据是否静态链接选择 CGO_ENABLED
ifeq ($(STATIC_BUILD), true)
  CGO_ENABLED = 0
else
  CGO_ENABLED = 1
endif

# 设置 GO_BUILD_ENV
GO_BUILD_ENV = CGO_ENABLED=$(CGO_ENABLED) GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH)

# 编译命令
GO_BUILD = $(GO) build $(GO_BUILD_ARGS)
GO_TEST = $(GO) test -v
GO_BUILD = $(GO) build $(GO_BUILD_ARGS)
GO_INSTALL = $(GO) install
GO_CLEAN = $(GO) clean
GO_MOD_TIDY = $(GO) mod tidy

# 目标文件
BINARY_NAME=WebFinder
BUILD_DIR=bin
SRC_DIR=src

# Go 版本
GO_VERSION=$(shell go version | awk '{print $$3}')

# 默认目标
all: build

# 编译 Go 项目
build: clean fmt tidy
	@echo "Building the Go project..."
	$(GO_BUILD_ENV) $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

# 运行单元测试
test:
	@echo "Running tests..."
	$(GO_TEST) ./...

# 格式化 Go 代码
fmt:
	@echo "Running gofmt..."
	$(GOFMT) -w .

# 安装 Go 包
install:
	@echo "Installing Go packages..."
	$(GO_INSTALL)

# 清理未使用的Go 包
tidy:
	@echo "Mod tidy up..."
	$(GO_MOD_TIDY)

# 清理编译文件
clean: 
	@echo "Cleaning up..."
	$(GO_CLEAN)
	rm -rf ./$(BUILD_DIR)/*


# 获取构建信息（如版本信息）
version:
	@echo "Go version: $(GO_VERSION)"
	@echo "Project version: $(shell git describe --tags --always)"

# 运行程序（开发模式）
run: build
	@echo "Running the program..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# 创建二进制文件并压缩为 tar 文件（适用于发布）
release: clean build
	@echo "Creating release package..."
	mkdir -p release
	tar -czvf release/$(BINARY_NAME)-$(GO_VERSION).tar.gz $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all build test fmt install clean run version release