#!/bin/bash

echo "🎯 上传和下载进度测试"
echo "===================="

BASE_URL="http://localhost:8080"

# 创建测试文件（稍大一些，便于观察进度）
echo "📁 创建测试文件..."
if [ ! -f "test_video_large.mp4" ]; then
    # 复制现有测试文件多次来创建更大的文件
    cp test_video.mp4 test_video_large.mp4
    for i in {1..5}; do
        cat test_video.mp4 >> test_video_large.mp4
    done
fi

file_size=$(wc -c < test_video_large.mp4)
echo "测试文件大小: $(($file_size / 1024)) KB"

echo ""
echo "🚀 步骤1: 开始上传文件..."

# 上传文件
upload_response=$(curl -s -X POST "$BASE_URL/api/v1/upload" \
  -F "file=@test_video_large.mp4" \
  -F "title=进度测试视频" \
  -F "description=测试上传和下载进度功能")

echo "上传响应:"
echo "$upload_response" | jq '.'

# 提取任务ID
task_id=$(echo "$upload_response" | jq -r '.data.task_id // empty')

if [ -n "$task_id" ] && [ "$task_id" != "null" ]; then
    echo "✅ 文件上传成功，任务ID: $task_id"
    
    echo ""
    echo "🔄 步骤2: 开始转换..."
    
    # 开始转换
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
    
    echo ""
    echo "📊 步骤3: 监控转换进度..."
    
    # 监控进度
    max_attempts=20
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
        
        # 等待5秒后重试
        sleep 5
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -eq $max_attempts ]; then
        echo "⏰ 转换超时"
        exit 1
    fi
    
    echo ""
    echo "📥 步骤4: 测试下载..."
    
    # 获取任务详情
    task_detail=$(curl -s "$BASE_URL/api/v1/tasks/$task_id")
    echo "任务详情:"
    echo "$task_detail" | jq '.'
    
    # 检查输出文件大小
    echo ""
    echo "📏 检查文件大小信息..."
    head_response=$(curl -s -I "$BASE_URL/api/v1/download/$task_id")
    echo "HEAD响应头:"
    echo "$head_response" | grep -i content-length
    
    # 测试下载（只下载前几个字节来测试）
    echo ""
    echo "🌐 访问前端页面进行完整测试："
    echo "   打开浏览器访问: $BASE_URL"
    echo "   1. 在上传转换标签页测试文件上传进度"
    echo "   2. 观察转换进度"
    echo "   3. 测试下载进度"
    echo ""
    echo "💻 或者使用以下命令直接下载："
    echo "   curl -O -J \"$BASE_URL/api/v1/download/$task_id\""
    
else
    echo "❌ 文件上传失败"
    echo "错误响应: $upload_response"
fi

echo ""
echo "🧹 清理测试文件:"
echo "   rm test_video_large.mp4  # 删除测试文件"

echo ""
echo "🎉 进度测试完成！" 