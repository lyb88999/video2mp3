# Makefile for Video Converter

.PHONY: help build run clean test deps docker-up docker-down

# 默认目标
help:
	@echo "视频转MP3工具 - 可用命令："
	@echo "  make deps     - 安装Go依赖"
	@echo "  make build    - 构建应用"
	@echo "  make run      - 运行应用"
	@echo "  make test     - 运行测试"
	@echo "  make clean    - 清理构建文件"
	@echo "  make docker-up   - 启动Docker服务"
	@echo "  make docker-down - 停止Docker服务"
	@echo "  make api-test    - 测试API端点"



# 切换到MySQL
mysql-setup:
	@echo "🔄 切换到MySQL数据库..."
	go mod edit -droprequire gorm.io/driver/postgres
	go get gorm.io/driver/mysql
	go mod tidy
	@echo "✅ MySQL依赖安装完成！"

# 安装依赖
deps:
	@echo "📦 安装Go依赖..."
	go mod tidy
	go mod download

# 构建应用
build: deps
	@echo "🔨 构建应用..."
	go build -o bin/server cmd/server/main.go

# 运行应用
run: build
	@echo "🚀 启动服务器..."
	./bin/server

# 开发模式运行（热重载需要安装air: go install github.com/cosmtrek/air@latest）
dev:
	@echo "🔥 开发模式启动..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装air: go install github.com/cosmtrek/air@latest"; \
		echo "或者使用 make run"; \
	fi

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...

# 清理
clean:
	@echo "🧹 清理构建文件..."
	rm -rf bin/
	rm -rf uploads/*
	rm -rf output/*
	rm -rf temp/*

# Docker compose启动 (开发环境)
docker-up:
	@echo "🐳 启动MySQL和Redis服务..."
	docker-compose -f docker-compose-dev.yml up -d

# Docker compose停止
docker-down:
	@echo "🛑 停止Docker服务..."
	docker-compose -f docker-compose-dev.yml down

# 测试API
api-test:
	@echo "🔍 测试API端点..."
	@if [ -f test_api.sh ]; then \
		chmod +x test_api.sh; \
		./test_api.sh; \
	else \
		echo "test_api.sh 文件不存在"; \
	fi

# 测试文件上传
upload-test:
	@echo "📤 测试文件上传功能..."
	@if [ -f test_upload.sh ]; then \
		chmod +x test_upload.sh; \
		./test_upload.sh; \
	else \
		echo "test_upload.sh 文件不存在"; \
	fi

# 格式化代码
fmt:
	@echo "✨ 格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "🔍 代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "请先安装golangci-lint: https://golangci-lint.run/usage/install/"; \
	fi

# 创建目录
setup:
	@echo "📁 创建必要目录..."
	mkdir -p uploads output temp bin logs

# 检查依赖
check-deps:
	@echo "🔍 检查系统依赖..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go未安装"; exit 1; }
	@command -v ffmpeg >/dev/null 2>&1 || { echo "⚠️  FFmpeg未安装（转换功能需要）"; }
	@command -v mysql >/dev/null 2>&1 || { echo "⚠️  MySQL客户端未安装（可选，可用Docker代替）"; }
	@command -v docker >/dev/null 2>&1 || { echo "⚠️  Docker未安装（可选）"; }
	@echo "✅ 基础依赖检查完成"

# 初始化项目
init: setup deps check-deps
	@echo "🎉 项目初始化完成！"
	@echo "💡 下一步："
	@echo "   1. 复制 config.yaml 并根据需要修改MySQL配置"
	@echo "   2. 运行 'make docker-up' 启动MySQL和Redis服务"
	@echo "   3. 运行 'make run' 启动服务器"
	@echo "   4. 运行 'make api-test' 测试API"
	@echo "   5. 访问 http://localhost:8081 打开phpMyAdmin管理数据库"