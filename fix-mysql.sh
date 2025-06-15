#!/bin/bash

echo "🔧 开始修复MySQL问题..."

# 1. 停止所有服务
echo "停止所有服务..."
docker-compose down --volumes --remove-orphans

# 2. 清理所有MySQL相关的卷
echo "清理MySQL数据卷..."
docker volume ls | grep mysql | awk '{print $2}' | xargs docker volume rm 2>/dev/null || true
docker volume ls | grep video-converter | awk '{print $2}' | xargs docker volume rm 2>/dev/null || true

# 3. 清理所有相关容器
echo "清理容器..."
docker container prune -f

# 4. 清理网络
echo "清理网络..."
docker network prune -f

# 5. 重新构建应用镜像
echo "重新构建应用镜像..."
docker-compose build --no-cache app

# 6. 先启动MySQL
echo "启动MySQL..."
docker-compose up -d mysql

# 7. 等待MySQL就绪
echo "等待MySQL启动..."
sleep 30

# 8. 检查MySQL状态
echo "检查MySQL状态..."
docker-compose logs mysql --tail=10

# 9. 启动其他服务
echo "启动Redis..."
docker-compose up -d redis

sleep 10

echo "启动应用..."
docker-compose up -d app

sleep 15

echo "启动Nginx..."
docker-compose up -d nginx

# 10. 显示最终状态
echo "📊 服务状态:"
docker-compose ps

echo "🎉 修复完成！访问 http://localhost 测试应用" 
