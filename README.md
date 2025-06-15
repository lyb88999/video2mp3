# 🎵 Video Converter - 视频转MP3工具

一个功能强大的在线视频转MP3工具，支持本地文件上传和抖音/快手链接直接转换。

## ✨ 功能特点

### 🎯 核心功能
- **本地文件转换** - 支持多种视频格式转MP3
- **在线链接转换** - 直接输入抖音/快手链接进行转换
- **智能链接解析** - 自动从分享文本中提取视频链接
- **实时进度跟踪** - 转换过程可视化显示
- **批量处理** - 支持多任务并发处理

### 🔗 支持的平台
- ✅ 抖音 (douyin.com)
- ✅ 快手 (kuaishou.com)
- ✅ 本地视频文件

### 📁 支持的视频格式
- MP4, AVI, MOV, WMV, FLV, WebM, MKV

### 🎵 音频输出选项
- **编码器**: MP3 (libmp3lame)
- **比特率**: 可选 128k, 192k, 256k, 320k
- **采样率**: 可选 44100Hz, 48000Hz

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/your-username/video-converter.git
cd video-converter

# 使用快速启动脚本
./scripts/quick-start.sh

# 或者手动启动所有服务
docker-compose up -d

# 访问应用
open http://localhost:9002
```

### 方式二：本地开发

```bash
# 安装依赖
go mod download

# 启动MySQL和Redis
docker-compose up -d mysql redis

# 运行应用
go run cmd/main.go
```

## 📋 系统要求

### Docker部署
- Docker 20.0+
- Docker Compose 2.0+
- 2GB+ 可用内存
- 10GB+ 可用磁盘空间

### 本地开发
- Go 1.21+
- FFmpeg
- MySQL 8.0+
- Redis 6.0+

## 🏗️ 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Nginx代理     │    │   主应用服务     │    │   下载服务      │
│   (端口9002)    │────│   (Go + Gin)    │────│   (Python)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                    ┌─────────┼─────────┐
                    │                   │
            ┌───────▼────────┐  ┌──────▼──────┐
            │   MySQL数据库   │  │  Redis缓存  │
            │   (用户数据)    │  │  (任务状态)  │
            └────────────────┘  └─────────────┘
```

## 🛠️ 配置说明

### 主要配置文件
- `config.yaml` - 本地开发配置
- `docker/config.yaml` - Docker部署配置
- `docker-compose.yml` - Docker服务编排

### 环境变量
```bash
# 数据库配置
DB_HOST=mysql
DB_USER=video_user
DB_PASSWORD=video_pass_2024
DB_NAME=video_converter

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# 应用配置
APP_MODE=release
APP_PORT=8080
```

## 📖 使用指南

### 1. 本地文件转换
1. 点击"选择文件"上传视频
2. 设置音频参数（可选）
3. 点击"开始转换"
4. 等待转换完成并下载

### 2. 在线链接转换
1. 复制抖音/快手分享链接
2. 粘贴到"视频链接"输入框
3. 系统自动提取真实链接
4. 点击"开始转换"

### 3. 分享链接格式示例
```
支持格式：
✅ 直接链接：https://v.douyin.com/xxxxxxx/
✅ 分享文本：7.17 xfb:/ 复制打开抖音... https://v.douyin.com/xxxxxxx/ 
```

## 🔧 开发指南

### 项目结构
```
video-converter/
├── cmd/                    # 应用入口
├── internal/              # 内部包
│   ├── api/              # API处理器
│   ├── config/           # 配置管理
│   ├── model/            # 数据模型
│   ├── service/          # 业务逻辑
│   └── utils/            # 工具函数
├── pkg/                   # 公共包
├── web/                   # 前端文件
├── docker/               # Docker配置
└── scripts/              # 部署和管理脚本
    ├── quick-start.sh    # 快速启动脚本
    ├── deploy.sh         # 部署脚本
    ├── build-fast.sh     # 快速构建脚本
    ├── clean-mysql.sh    # 清理MySQL数据
    └── fix-mysql.sh      # 修复MySQL问题
```

### 开发命令
```bash
# 代码格式化
go fmt ./...

# 运行测试
go test ./...

# 构建应用
go build -o bin/video-converter cmd/main.go

# 热重载开发
air
```

### 🛠️ 实用脚本

项目提供了多个实用脚本来简化开发和部署：

```bash
# 快速启动所有服务
./scripts/quick-start.sh

# 快速构建应用
./scripts/build-fast.sh

# 部署到生产环境
./scripts/deploy.sh

# 清理MySQL数据（重置数据库）
./scripts/clean-mysql.sh

# 修复MySQL连接问题
./scripts/fix-mysql.sh
```

## 🚀 部署指南

### 生产环境部署
```bash
# 1. 克隆代码
git clone https://github.com/your-username/video-converter.git
cd video-converter

# 2. 配置环境
cp docker/config.yaml.example docker/config.yaml
# 编辑配置文件

# 3. 启动服务
docker-compose -f docker-compose.yml up -d

# 4. 检查服务状态
docker-compose ps
```

### 服务监控
```bash
# 查看日志
docker-compose logs -f app

# 查看服务状态
docker-compose ps

# 重启服务
docker-compose restart app
```

## 🔍 故障排除

### 常见问题

**1. 下载服务连接失败**
```bash
# 检查下载服务状态
docker logs video-download-service

# 重启下载服务
docker-compose restart download-service
```

**2. FFmpeg转换失败**
```bash
# 检查FFmpeg是否安装
docker exec video-converter-app ffmpeg -version

# 检查文件权限
docker exec video-converter-app ls -la temp/
```

**3. 数据库连接问题**
```bash
# 检查MySQL状态
docker-compose logs mysql

# 重置数据库
./scripts/clean-mysql.sh
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [FFmpeg](https://ffmpeg.org/) - 视频处理核心
- [Gin](https://gin-gonic.com/) - Go Web框架
- [evil0ctal/douyin_tiktok_download_api](https://github.com/Evil0ctal/Douyin_TikTok_Download_API) - 抖音下载API

## 📞 联系方式

- 项目地址: [https://github.com/your-username/video-converter](https://github.com/your-username/video-converter)
- 问题反馈: [Issues](https://github.com/your-username/video-converter/issues)

---

⭐ 如果这个项目对你有帮助，请给个星标支持一下！ 