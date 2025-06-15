#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}[INFO]${NC} 开始清理MySQL..."

# 停止所有服务
echo -e "${GREEN}[INFO]${NC} 停止所有服务..."
docker compose down

# 删除MySQL容器
echo -e "${GREEN}[INFO]${NC} 删除MySQL容器..."
docker rm -f video-converter-mysql-fresh 2>/dev/null || true

# 删除MySQL数据卷
echo -e "${GREEN}[INFO]${NC} 删除MySQL数据卷..."
docker volume rm video-converter_mysql_data_fresh 2>/dev/null || true

# 重新启动服务
echo -e "${GREEN}[INFO]${NC} 重新启动服务..."
docker compose up -d

echo -e "${GREEN}[INFO]${NC} 清理完成！" 
