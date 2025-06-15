#!/bin/bash

# 文件上传功能测试脚本
BASE_URL="http://localhost:8080"

echo "🧪 开始测试文件上传功能..."
echo "================================"

# 创建测试视频文件（如果不存在）
create_test_video() {
    if command -v ffmpeg > /dev/null; then
        echo "📹 创建测试视频文件..."
        ffmpeg -f lavfi -i testsrc=duration=10:size=320x240:rate=1 -f lavfi -i sine=frequency=1000:duration=10 -c:v libx264 -c:a aac -t 10 test_video.mp4 -y 2>/dev/null
        echo "✅ 测试视频文件创建完成: test_video.mp4"
    else
        echo "⚠️ FFmpeg未安装，请手动创建一个视频文件命名为 test_video.mp4"
        echo "   或者下载一个小的MP4文件用于测试"
        return 1
    fi
}

# 测试1: 检查服务器状态
echo "🔍 测试1: 检查服务器状态..."
response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/health")
if [ "$response" = "200" ]; then
    echo "✅ 服务器运行正常"
else
    echo "❌ 服务器无响应 (HTTP $response)"
    echo "请确保服务器已启动: make run"
    exit 1
fi

# 测试2: 上传列表（空列表）
echo ""
echo "📋 测试2: 获取上传列表（初始状态）..."
curl -s "$BASE_URL/api/v1/upload/list" | jq '.'

# 测试3: 创建测试文件并上传
echo ""
echo "📤 测试3: 文件上传..."

# 检查是否有测试文件
if [ ! -f "test_video.mp4" ]; then
    echo "🎬 未找到测试文件，尝试创建..."
    if ! create_test_video; then
        echo "❌ 无法创建测试文件，请手动提供一个名为 test_video.mp4 的视频文件"
        exit 1
    fi
fi

# 执行文件上传
echo "正在上传 test_video.mp4..."
upload_response=$(curl -s -X POST "$BASE_URL/api/v1/upload" \
  -F "file=@test_video.mp4" \
  -F "title=测试视频" \
  -F "description=这是一个测试上传的视频文件")

echo "上传响应:"
echo "$upload_response" | jq '.'

# 提取任务ID
task_id=$(echo "$upload_response" | jq -r '.data.task_id // empty')

if [ -n "$task_id" ] && [ "$task_id" != "null" ]; then
    echo "✅ 文件上传成功，任务ID: $task_id"

    # 测试4: 查询上传进度
    echo ""
    echo "📊 测试4: 查询上传进度..."
    curl -s "$BASE_URL/api/v1/upload/progress/$task_id" | jq '.'

    # 测试5: 更新后的上传列表
    echo ""
    echo "📋 测试5: 获取上传列表（包含新上传）..."
    curl -s "$BASE_URL/api/v1/upload/list" | jq '.'

    # 测试6: 查询任务状态
    echo ""
    echo "📈 测试6: 查询任务状态..."
    curl -s "$BASE_URL/api/v1/tasks/$task_id/status" | jq '.'

    echo ""
    echo "💡 测试完成！任务ID: $task_id"
    echo "   可以使用这个ID进行后续的转换测试"

else
    echo "❌ 文件上传失败"
    echo "错误响应: $upload_response"
fi

# 测试7: 错误情况测试
echo ""
echo "🚫 测试7: 错误情况测试..."

# 测试上传不存在的文件
echo "测试无效文件上传:"
curl -s -X POST "$BASE_URL/api/v1/upload" \
  -F "file=@nonexistent.mp4" 2>/dev/null | jq '.' || echo "请求失败（预期行为）"

# 测试无文件上传
echo ""
echo "测试空文件上传:"
curl -s -X POST "$BASE_URL/api/v1/upload" | jq '.'

# 测试获取不存在的任务进度
echo ""
echo "测试不存在的任务进度:"
curl -s "$BASE_URL/api/v1/upload/progress/nonexistent-id" | jq '.'

echo ""
echo "🎉 所有测试完成！"
echo ""
echo "📁 检查上传的文件:"
echo "   - 上传目录: ./uploads/"
echo "   - 输出目录: ./output/"
echo "   - 临时目录: ./temp/"
echo ""
echo "🧹 清理测试文件:"
echo "   rm test_video.mp4  # 删除测试视频文件"