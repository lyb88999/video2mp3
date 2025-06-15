#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080"

echo "🚀 开始测试视频转MP3 API..."
echo "================================"

# 测试健康检查
echo "📋 测试健康检查..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# 测试欢迎页面
echo "👋 测试欢迎页面..."
curl -s "$BASE_URL/" | jq '.'
echo ""

# 测试上传端点
echo "📁 测试文件上传端点..."
curl -s -X POST "$BASE_URL/api/v1/upload" | jq '.'
echo ""

# 测试URL转换端点
echo "🔗 测试URL转换端点..."
curl -s -X POST "$BASE_URL/api/v1/convert/url" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4"}' | jq '.'
echo ""

# 测试任务列表
echo "📝 测试任务列表..."
curl -s "$BASE_URL/api/v1/tasks" | jq '.'
echo ""

# 测试任务状态
echo "📊 测试任务状态..."
curl -s "$BASE_URL/api/v1/tasks/test-123/status" | jq '.'
echo ""

# 测试上传进度
echo "📈 测试上传进度..."
curl -s "$BASE_URL/api/v1/upload/progress/test-123" | jq '.'
echo ""

# 测试速率限制（快速发送多个请求）
echo "⚡测试速率限制（发送5个快速请求）..."
for i in {1..5}; do
  echo "请求 $i:"
  curl -s -w "状态码: %{http_code}\n" "$BASE_URL/health" | head -1
done
echo ""

echo "✅ API测试完成！"
echo ""
echo "💡 提示："
echo "   - 如果看到健康检查返回成功，说明基础API正常工作"
echo "   - 所有端点目前返回占位符响应，这是正常的"
echo "   - 接下来我们将实现具体的业务逻辑"