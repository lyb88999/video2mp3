#!/bin/bash

echo "🌐 前端界面测试"
echo "==============="

BASE_URL="http://localhost:8080"

echo "1. 测试前端页面访问..."
response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/")
if [ "$response" = "200" ]; then
    echo "✅ 前端页面访问正常 (HTTP $response)"
else
    echo "❌ 前端页面访问失败 (HTTP $response)"
fi

echo ""
echo "2. 测试静态资源..."

# 测试CSS
css_response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/static/css/style.css")
echo "CSS文件: $([ "$css_response" = "200" ] && echo "✅ 正常" || echo "❌ 失败") (HTTP $css_response)"

# 测试JS
js_response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/static/js/app.js")
echo "JS文件: $([ "$js_response" = "200" ] && echo "✅ 正常" || echo "❌ 失败") (HTTP $js_response)"

echo ""
echo "3. 测试API端点..."

# 测试健康检查
health_response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/health")
echo "健康检查: $([ "$health_response" = "200" ] && echo "✅ 正常" || echo "❌ 失败") (HTTP $health_response)"

# 测试任务列表API
tasks_response=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/api/v1/tasks")
echo "任务列表API: $([ "$tasks_response" = "200" ] && echo "✅ 正常" || echo "❌ 失败") (HTTP $tasks_response)"

echo ""
echo "🎯 前端界面测试完成！"
echo ""
echo "📱 访问地址："
echo "   🖥️  桌面端: $BASE_URL"
echo "   📱 移动端: $BASE_URL (响应式设计)"
echo ""
echo "🔗 功能测试："
echo "   • 文件拖拽上传"
echo "   • 转换参数设置"
echo "   • 实时进度显示"
echo "   • 文件下载"
echo "   • 任务历史查看"
echo ""
echo "💡 提示："
echo "   在浏览器中打开 $BASE_URL 开始使用！" 