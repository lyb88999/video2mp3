#!/bin/bash

# 视频转换功能测试脚本
BASE_URL="http://localhost:8080"

echo "🎵 开始测试视频转MP3功能..."
echo "================================"

# 测试1: 先上传一个视频文件
echo "📤 步骤1: 上传测试视频..."
if [ ! -f "test_video.mp4" ]; then
    if command -v ffmpeg > /dev/null; then
        echo "📹 创建测试视频文件..."
        ffmpeg -f lavfi -i testsrc=duration=10:size=320x240:rate=1 -f lavfi -i sine=frequency=1000:duration=10 -c:v libx264 -c:a aac -t 10 test_video.mp4 -y 2>/dev/null
        echo "✅ 测试视频文件创建完成"
    else
        echo "❌ 需要FFmpeg来创建测试文件"
        exit 1
    fi
fi

# 上传文件
upload_response=$(curl -s -X POST "$BASE_URL/api/v1/upload" \
  -F "file=@test_video.mp4" \
  -F "title=转换测试视频" \
  -F "description=用于测试转换功能的视频")

echo "上传响应:"
echo "$upload_response" | jq '.'

# 提取任务ID
task_id=$(echo "$upload_response" | jq -r '.data.task_id // empty')

if [ -z "$task_id" ] || [ "$task_id" = "null" ]; then
    echo "❌ 文件上传失败，无法继续测试"
    exit 1
fi

echo "✅ 文件上传成功，任务ID: $task_id"

# 测试2: 启动转换任务
echo ""
echo "🔄 步骤2: 启动视频转MP3转换..."
convert_response=$(curl -s -X POST "$BASE_URL/api/v1/convert/file" \
  -H "Content-Type: application/json" \
  -d "{
    \"task_id\": \"$task_id\",
    \"audio_codec\": \"libmp3lame\",
    \"audio_bitrate\": \"192k\",
    \"sample_rate\": \"44100\"
  }")

echo "转换响应:"
echo "$convert_response" | jq '.'

# 检查转换是否成功启动
convert_success=$(echo "$convert_response" | jq -r '.success // false')
if [ "$convert_success" != "true" ]; then
    echo "❌ 转换任务启动失败"
    exit 1
fi

echo "✅ 转换任务已启动"

# 测试3: 监控转换进度
echo ""
echo "📊 步骤3: 监控转换进度..."
max_attempts=30  # 最多等待30次 (约5分钟)
attempt=0

while [ $attempt -lt $max_attempts ]; do
    # 获取任务状态
    status_response=$(curl -s "$BASE_URL/api/v1/tasks/$task_id/status")
    status=$(echo "$status_response" | jq -r '.data.status // "unknown"')
    progress=$(echo "$status_response" | jq -r '.data.progress // 0')
    
    echo "进度更新: 状态=$status, 进度=${progress}%"
    
    # 检查是否完成
    if [ "$status" = "completed" ]; then
        echo "✅ 转换完成！"
        break
    elif [ "$status" = "failed" ]; then
        echo "❌ 转换失败"
        echo "$status_response" | jq '.data.error_message'
        exit 1
    elif [ "$status" = "canceled" ]; then
        echo "⚠️ 转换被取消"
        exit 1
    fi
    
    # 等待10秒后重试
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "⏰ 转换超时，可能需要更长时间"
    echo "   当前状态: $status"
    echo "   您可以稍后检查任务状态: curl -s '$BASE_URL/api/v1/tasks/$task_id/status'"
fi

# 测试4: 检查输出文件
echo ""
echo "📁 步骤4: 检查输出文件..."
file_check_response=$(curl -s -I "$BASE_URL/api/v1/download/$task_id")
file_exists=$(echo "$file_check_response" | grep -i "x-file-exists: true")

if [ -n "$file_exists" ]; then
    echo "✅ 输出文件已生成"
    
    # 测试5: 下载文件
    echo ""
    echo "⬇️ 步骤5: 测试文件下载..."
    echo "下载链接: $BASE_URL/api/v1/download/$task_id"
    echo "您可以在浏览器中访问上述链接下载MP3文件"
    
    # 可选：实际下载文件到本地
    echo "正在下载文件到本地..."
    curl -s -L -o "converted_audio.mp3" "$BASE_URL/api/v1/download/$task_id"
    
    if [ -f "converted_audio.mp3" ]; then
        file_size=$(stat -f%z "converted_audio.mp3" 2>/dev/null || stat -c%s "converted_audio.mp3" 2>/dev/null)
        echo "✅ 文件下载成功: converted_audio.mp3 (${file_size} bytes)"
    else
        echo "❌ 文件下载失败"
    fi
else
    echo "❌ 输出文件不存在或未准备好"
fi

# 测试6: 获取任务详情
echo ""
echo "📋 步骤6: 获取任务详细信息..."
task_detail=$(curl -s "$BASE_URL/api/v1/tasks/$task_id")
echo "$task_detail" | jq '.'

# 测试7: 列出所有任务
echo ""
echo "📝 步骤7: 列出所有任务..."
all_tasks=$(curl -s "$BASE_URL/api/v1/tasks")
echo "$all_tasks" | jq '.data.items[] | {id, title, status, progress}'

echo ""
echo "🎉 转换功能测试完成！"
echo ""
echo "📋 测试总结："
echo "   - 任务ID: $task_id"
echo "   - 下载链接: $BASE_URL/api/v1/download/$task_id"
if [ -f "converted_audio.mp3" ]; then
    echo "   - 本地文件: converted_audio.mp3"
fi
echo ""
echo "🧹 清理命令："
echo "   rm test_video.mp4 converted_audio.mp3" 