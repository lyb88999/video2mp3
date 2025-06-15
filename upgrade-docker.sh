#!/bin/bash

echo "🚀 开始升级Docker和Docker Compose..."

# 检查是否为root用户
if [[ $EUID -ne 0 ]]; then
   echo "此脚本需要root权限运行，请使用sudo"
   exit 1
fi

# 1. 停止当前Docker服务
echo "停止Docker服务..."
systemctl stop docker

# 2. 卸载旧版本
echo "卸载旧版本Docker..."
yum remove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine docker-ce docker-ce-cli containerd.io

# 3. 安装必要工具
echo "安装必要工具..."
yum install -y yum-utils device-mapper-persistent-data lvm2

# 4. 添加Docker官方仓库
echo "添加Docker仓库..."
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 5. 安装最新版Docker
echo "安装最新版Docker..."
yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 6. 启动并启用Docker服务
echo "启动Docker服务..."
systemctl start docker
systemctl enable docker

# 7. 下载并安装最新版docker-compose
echo "安装最新版docker-compose..."
DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 8. 创建软链接（如果需要）
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# 9. 将当前用户添加到docker组
echo "将用户添加到docker组..."
usermod -aG docker $SUDO_USER

echo "✅ Docker升级完成！"
echo "📊 版本信息："
docker --version
docker-compose --version

echo ""
echo "🔄 注意：请重新登录或运行以下命令以使docker组权限生效："
echo "newgrp docker"
echo ""
echo "然后可以测试："
echo "docker run hello-world" 
