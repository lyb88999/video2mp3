#!/bin/bash

# IP地址部署脚本
# 前端端口: 9001, 后端端口: 9002

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

SERVER_IP="47.93.190.244"
FRONTEND_PORT="9001"
BACKEND_PORT="9002"

log_info "开始部署到服务器: $SERVER_IP"
log_info "前端端口: $FRONTEND_PORT"
log_info "后端端口: $BACKEND_PORT"

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    log_error "Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 创建必要的目录
log_info "创建必要的目录..."
mkdir -p uploads output temp logs/nginx data

# 设置目录权限
log_info "设置目录权限..."
chmod 755 uploads output temp logs data
chmod -R 755 docker/

# 停止现有服务
log_info "停止现有服务..."
docker-compose down --remove-orphans || true

# 清理旧的镜像（可选）
read -p "是否清理旧的Docker镜像？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "清理旧镜像..."
    docker system prune -f
fi

# 构建并启动服务
log_info "构建并启动服务..."
docker compose up --build -d

# 等待服务启动
log_info "等待服务启动..."
sleep 30

# 检查服务状态
log_info "检查服务状态..."
if docker compose ps | grep -q "Up"; then
    log_info "服务启动成功！"
else
    log_error "服务启动失败，请检查日志"
    docker-compose logs
    exit 1
fi

# 健康检查
log_info "执行健康检查..."
for i in {1..10}; do
    if curl -f http://localhost:$FRONTEND_PORT/health > /dev/null 2>&1; then
        log_info "前端健康检查通过！"
        break
    else
        log_warn "前端健康检查失败，重试中... ($i/10)"
        sleep 5
    fi
    
    if [ $i -eq 10 ]; then
        log_error "前端健康检查失败，请检查服务状态"
        docker-compose logs nginx
        exit 1
    fi
done

# 后端健康检查
log_info "检查后端API..."
for i in {1..10}; do
    if curl -f http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
        log_info "后端API健康检查通过！"
        break
    else
        log_warn "后端API健康检查失败，重试中... ($i/10)"
        sleep 5
    fi
    
    if [ $i -eq 10 ]; then
        log_error "后端API健康检查失败，请检查服务状态"
        docker-compose logs app
        exit 1
    fi
done

# 显示部署信息
log_info "部署完成！"
echo
echo "==================================="
echo "部署信息:"
echo "服务器IP: $SERVER_IP"
echo "前端访问: http://$SERVER_IP:$FRONTEND_PORT"
echo "后端API: http://$SERVER_IP:$BACKEND_PORT/api/v1"
echo "本地前端: http://localhost:$FRONTEND_PORT"
echo "本地后端: http://localhost:$BACKEND_PORT"
echo "==================================="
echo
echo "服务状态:"
docker-compose ps
echo
echo "查看日志: docker compose logs -f"
echo "停止服务: docker compose down"
echo "重启服务: docker compose restart"

# 防火墙提示
echo
log_warn "重要提示:"
echo "1. 请确保服务器防火墙已开放端口 $FRONTEND_PORT 和 $BACKEND_PORT"
echo "2. 如果使用云服务器，请在安全组中开放这些端口"
echo "3. 前端页面: http://$SERVER_IP:$FRONTEND_PORT"
echo "4. 后端API: http://$SERVER_IP:$BACKEND_PORT/api/v1"
echo
echo "防火墙命令示例 (Ubuntu/CentOS):"
echo "sudo ufw allow $FRONTEND_PORT"
echo "sudo ufw allow $BACKEND_PORT"
echo "或者:"
echo "sudo firewall-cmd --permanent --add-port=$FRONTEND_PORT/tcp"
echo "sudo firewall-cmd --permanent --add-port=$BACKEND_PORT/tcp"
echo "sudo firewall-cmd --reload" 
