#!/bin/bash

# 快速启动脚本 - 跳过数据库，只启动核心服务
# 适用于快速测试和开发

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_step "快速启动模式 - 跳过数据库服务"

# 停止现有服务
log_step "停止现有服务..."
docker-compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
docker-compose down --remove-orphans 2>/dev/null || true

# 创建必要目录
log_step "创建必要目录..."
mkdir -p uploads output temp logs/nginx data

# 使用开发配置
log_step "使用开发环境配置..."
cp config-dev.yaml config.yaml

# 启用BuildKit
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# 只构建应用镜像（跳过数据库）
log_step "构建应用镜像..."
start_time=$(date +%s)

docker-compose -f docker-compose.dev.yml build --progress=plain

end_time=$(date +%s)
build_time=$((end_time - start_time))

log_info "构建完成！耗时: ${build_time}秒"

# 启动服务
log_step "启动核心服务..."
docker-compose -f docker-compose.dev.yml up -d

# 等待服务启动
log_info "等待服务启动..."
sleep 10

# 检查服务状态
log_step "检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

# 健康检查
log_step "执行健康检查..."
for i in {1..5}; do
    if curl -f http://localhost:9002/health > /dev/null 2>&1; then
        log_info "后端服务健康检查通过！"
        break
    else
        log_warn "等待后端服务启动... ($i/5)"
        sleep 3
    fi
done

for i in {1..5}; do
    if curl -f http://localhost:9001 > /dev/null 2>&1; then
        log_info "前端服务健康检查通过！"
        break
    else
        log_warn "等待前端服务启动... ($i/5)"
        sleep 3
    fi
done

# 显示访问信息
echo
echo "=================================="
log_info "快速启动完成！"
echo "前端访问: http://localhost:9001"
echo "后端API: http://localhost:9002"
echo "服务器访问: http://47.93.190.244:9001"
echo "=================================="
echo
echo "常用命令:"
echo "查看日志: docker-compose -f docker-compose.dev.yml logs -f"
echo "停止服务: docker-compose -f docker-compose.dev.yml down"
echo "重启服务: docker-compose -f docker-compose.dev.yml restart"

log_warn "注意: 此模式使用SQLite数据库，数据存储在 ./data/ 目录"
