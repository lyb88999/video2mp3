#!/bin/bash

# 快速构建脚本
# 包含多种优化选项

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_step "开始快速构建..."

# 检查Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker 未安装，请先安装 Docker"
    exit 1
fi

# 设置Docker构建参数
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

log_info "启用 Docker BuildKit 加速构建"

# 清理构建缓存（可选）
read -p "是否清理Docker构建缓存？这会让首次构建更慢，但确保使用最新代码 (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_step "清理构建缓存..."
    docker builder prune -f
fi

# 预拉取基础镜像
log_step "预拉取基础镜像..."
docker pull golang:1.22-alpine &
docker pull alpine:3.18 &
docker pull nginx:alpine &
docker pull mysql:8.0 &
docker pull redis:7-alpine &
wait

log_info "基础镜像拉取完成"

# 构建应用
log_step "构建应用镜像..."
start_time=$(date +%s)

# 使用并行构建
docker-compose build --parallel --progress=plain

end_time=$(date +%s)
build_time=$((end_time - start_time))

log_info "构建完成！耗时: ${build_time}秒"

# 显示镜像大小
log_step "镜像信息:"
docker images | grep video-converter

# 启动服务
read -p "是否立即启动服务？(Y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    log_step "启动服务..."
    docker-compose up -d
    
    log_info "等待服务启动..."
    sleep 15
    
    # 显示服务状态
    log_step "服务状态:"
    docker-compose ps
    
    log_info "构建和启动完成！"
    echo "前端访问: http://localhost:9001"
    echo "后端API: http://localhost:9002"
else
    log_info "构建完成，使用 'docker-compose up -d' 启动服务"
fi

# 优化建议
echo
log_warn "构建优化建议:"
echo "1. 如果经常构建，建议保留Docker缓存"
echo "2. 可以使用 'docker system df' 查看磁盘使用情况"
echo "3. 定期清理无用镜像: 'docker image prune'"
echo "4. 使用 '.dockerignore' 文件排除不必要的文件" 
