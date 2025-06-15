# 视频URL转换功能指南

## 功能概述

新增了从抖音、快手等平台视频链接直接转换为MP3的功能。集成了Douyin_TikTok_Download_API服务，用户只需粘贴视频链接，即可获得转换后的音频文件。

## 使用步骤

### 前端界面使用

1. **打开网页**：访问 `http://localhost:8080`
2. **输入链接**：在"或输入视频链接"区域粘贴抖音/快手视频链接
3. **点击添加**：点击"添加链接"按钮
4. **配置设置**：选择音频质量、采样率等参数
5. **开始转换**：点击"开始转换"按钮
6. **等待完成**：查看转换进度
7. **下载文件**：转换完成后下载MP3文件

### API接口使用

#### 创建URL转换任务

```bash
curl -X POST "http://localhost:8080/api/v1/convert/url" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://v.douyin.com/wfO0eC1BTh8/",
    "title": "我的视频转换",
    "audio_codec": "libmp3lame",
    "audio_bitrate": "192k",
    "sample_rate": "44100"
  }'
```

#### 查询任务状态

```bash
curl -X GET "http://localhost:8080/api/v1/tasks/{task_id}/status"
```

#### 下载转换结果

```bash
curl -X GET "http://localhost:8080/api/v1/download/{task_id}" -o output.mp3
```

## 配置说明

### 下载服务配置

在 `config.yaml` 中配置下载服务：

```yaml
download:
  service_url: "http://localhost:9080/api/download"  # 下载服务地址
  timeout: 300                                       # 超时时间（秒）
  enable_prefix: true                               # 启用前缀
  disable_watermark: true                           # 禁用水印
```

### 音频转换参数

- **audio_codec**: 音频编码器（默认：libmp3lame）
- **audio_bitrate**: 音频比特率（选项：128k, 192k, 256k, 320k）
- **sample_rate**: 采样率（选项：22050, 44100, 48000）

## 工作流程

1. **接收URL**：前端/API接收抖音/快手视频链接
2. **创建任务**：在数据库中创建转换任务记录
3. **异步下载**：调用下载服务直接获取视频文件
4. **文件解析**：解析Content-Disposition头获取原始文件名
5. **本地保存**：将视频文件保存到临时目录
6. **队列转换**：任务重新进入转换队列
7. **FFmpeg转换**：使用FFmpeg转换为MP3格式
8. **完成下载**：提供最终的MP3下载链接

## 技术说明

### 下载服务集成

- 使用Douyin_TikTok_Download_API服务
- 服务直接返回视频文件的二进制数据
- 通过Content-Disposition头获取文件名
- 响应示例：
  ```
  Content-Type: video/mp4
  Content-Disposition: attachment; filename="douyin.wtf_douyin_7515792557595462947.mp4"
  Content-Length: 83847419
  ```

### 异步处理

- URL解析和文件下载在后台异步进行
- 不阻塞用户请求响应
- 实时状态更新通过Redis缓存

## 支持的平台

- 抖音 (douyin.com)
- 快手 (kuaishou.com)
- TikTok (国际版)
- 其他支持的视频平台（取决于下载服务）

## 注意事项

1. **下载服务依赖**：需要先启动Douyin_TikTok_Download_API服务（端口9080）
2. **网络连接**：确保服务器能够访问视频平台
3. **文件大小**：大文件可能需要较长转换时间
4. **存储空间**：确保有足够的临时存储空间
5. **文件清理**：建议定期清理临时文件

## 故障排除

### 常见错误

1. **"调用下载服务失败"**
   - 检查下载服务是否正常运行
   - 验证 `config.yaml` 中的服务地址
   - 确认端口9080可访问

2. **"下载服务返回错误状态"**
   - 检查视频链接是否有效
   - 验证链接格式是否正确
   - 查看下载服务日志

3. **"保存视频文件失败"**
   - 检查临时目录权限
   - 确保有足够磁盘空间
   - 验证文件路径配置

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 检查下载服务状态
curl http://localhost:9080/
```

## 测试

运行测试脚本：

```bash
./test_douyin_convert.sh
```

测试步骤：
1. 创建URL转换任务
2. 检查任务状态
3. 监控转换进度
4. 验证最终结果

## 更新内容

- ✅ 集成Douyin_TikTok_Download_API服务
- ✅ 支持直接文件下载（非JSON响应）
- ✅ 自动解析原始文件名
- ✅ 异步任务处理机制
- ✅ 完整的错误处理和重试逻辑
- ✅ 实时进度跟踪 