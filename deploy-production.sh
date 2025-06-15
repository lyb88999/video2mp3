#!/bin/bash

# 生产环境部署脚本
# 使用方法: ./deploy-production.sh your-domain.com

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

# 检查参数
if [ $# -eq 0 ]; then
    log_error "请提供域名参数"
    echo "使用方法: $0 your-domain.com"
    exit 1
fi

DOMAIN=$1
log_info "开始部署到域名: $DOMAIN"

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    log_error "Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 更新Nginx配置中的域名
log_info "更新Nginx配置..."
sed -i.bak "s/your-domain\.com/$DOMAIN/g" docker/nginx/conf.d/default.conf
log_info "域名配置已更新为: $DOMAIN"

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
docker-compose up --build -d

# 等待服务启动
log_info "等待服务启动..."
sleep 30

# 检查服务状态
log_info "检查服务状态..."
if docker-compose ps | grep -q "Up"; then
    log_info "服务启动成功！"
else
    log_error "服务启动失败，请检查日志"
    docker-compose logs
    exit 1
fi

# 健康检查
log_info "执行健康检查..."
for i in {1..10}; do
    if curl -f http://localhost/health > /dev/null 2>&1; then
        log_info "健康检查通过！"
        break
    else
        log_warn "健康检查失败，重试中... ($i/10)"
        sleep 5
    fi
    
    if [ $i -eq 10 ]; then
        log_error "健康检查失败，请检查服务状态"
        docker-compose logs app
        exit 1
    fi
done

# 显示部署信息
log_info "部署完成！"
echo
echo "==================================="
echo "部署信息:"
echo "域名: http://$DOMAIN"
echo "本地访问: http://localhost"
echo "API地址: http://$DOMAIN/api/v1"
echo "==================================="
echo
echo "服务状态:"
docker-compose ps
echo
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo "重启服务: docker-compose restart"

# SSL证书提示
echo
log_warn "重要提示:"
echo "1. 请确保域名 $DOMAIN 已正确解析到此服务器"
echo "2. 建议配置SSL证书以启用HTTPS"
echo "3. 可以使用 Let's Encrypt 免费证书"
echo "4. 生产环境建议配置防火墙规则" 