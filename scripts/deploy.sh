#!/bin/bash

# 视频转换系统 Docker 部署脚本
# 作者: AI助手
# 版本: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    
    mkdir -p uploads output temp logs/nginx
    chmod 755 uploads output temp logs
    
    log_success "目录创建完成"
}

# 检查环境变量文件
check_env_file() {
    if [ ! -f .env ]; then
        log_warning ".env 文件不存在，创建默认配置..."
        
        cat > .env << EOF
MYSQL_ROOT_PASSWORD=video_converter_2024
MYSQL_DATABASE=video_converter
MYSQL_USER=video_user
MYSQL_PASSWORD=video_pass_2024
GIN_MODE=release
TZ=Asia/Shanghai
EOF
        
        log_success "默认 .env 文件已创建"
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    docker-compose build --no-cache
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    docker-compose up -d
    
    log_success "服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -f http://localhost/health &> /dev/null; then
            log_success "服务已就绪"
            return 0
        fi
        
        attempt=$((attempt + 1))
        echo -n "."
        sleep 5
    done
    
    log_error "服务启动超时"
    return 1
}

# 显示状态
show_status() {
    log_info "服务状态:"
    docker-compose ps
    
    echo ""
    log_info "访问地址:"
    echo "  🌐 前端页面: http://localhost"
    echo "  🔍 健康检查: http://localhost/health"
    echo "  📊 API文档: http://localhost/api/v1"
    
    echo ""
    log_info "管理命令:"
    echo "  查看日志: docker-compose logs -f"
    echo "  停止服务: docker-compose down"
    echo "  重启服务: docker-compose restart"
    echo "  查看状态: docker-compose ps"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose down
    log_success "服务已停止"
}

# 清理数据
cleanup() {
    log_warning "清理所有数据..."
    read -p "确认要删除所有数据吗？(y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v
        docker system prune -f
        log_success "数据清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 查看日志
view_logs() {
    local service=${1:-}
    
    if [ -z "$service" ]; then
        log_info "查看所有服务日志..."
        docker-compose logs -f
    else
        log_info "查看 $service 服务日志..."
        docker-compose logs -f "$service"
    fi
}

# 帮助信息
show_help() {
    echo "视频转换系统 Docker 部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  install     - 完整安装部署"
    echo "  start       - 启动服务"
    echo "  stop        - 停止服务"
    echo "  restart     - 重启服务"
    echo "  status      - 查看状态"
    echo "  logs [服务] - 查看日志"
    echo "  cleanup     - 清理数据"
    echo "  help        - 显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 install          # 完整部署"
    echo "  $0 logs app         # 查看应用日志"
    echo "  $0 logs mysql       # 查看MySQL日志"
}

# 完整安装
full_install() {
    log_info "开始完整安装部署..."
    
    check_dependencies
    create_directories
    check_env_file
    build_images
    start_services
    
    if wait_for_services; then
        show_status
        log_success "🎉 部署完成！"
    else
        log_error "部署失败，请检查日志"
        docker-compose logs
    fi
}

# 主函数
main() {
    case "${1:-install}" in
        "install")
            full_install
            ;;
        "start")
            check_dependencies
            create_directories
            check_env_file
            start_services
            wait_for_services && show_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            start_services
            wait_for_services && show_status
            ;;
        "status")
            show_status
            ;;
        "logs")
            view_logs "$2"
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 