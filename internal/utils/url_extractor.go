package utils

import (
	"regexp"
	"strings"
)

// ExtractVideoURL 从分享文本中提取视频URL
func ExtractVideoURL(text string) string {
	text = strings.TrimSpace(text)

	// 如果输入已经是有效的URL，直接返回
	if isValidVideoURL(text) {
		return text
	}

	// 定义支持的URL模式
	patterns := []string{
		// 抖音URL模式 - 包含下划线、短横线等特殊字符
		`https?://v\.douyin\.com/[A-Za-z0-9_-]+/?`,
		`https?://www\.douyin\.com/video/\d+`,
		`https?://(?:www\.)?douyin\.com/[^\s]+`,

		// 快手URL模式 - 同样包含特殊字符
		`https?://v\.kuaishou\.com/[A-Za-z0-9_-]+/?`,
		`https?://www\.kuaishou\.com/[^\s]+`,

		// TikTok URL模式（扩展支持）
		`https?://(?:www\.)?tiktok\.com/[^\s]+`,
		`https?://vm\.tiktok\.com/[A-Za-z0-9_-]+/?`,
	}

	// 尝试匹配每个模式
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)

		if len(matches) > 0 {
			url := matches[0]

			// 清理URL末尾的标点符号
			url = cleanURL(url)

			// 确保抖音URL以/结尾
			if strings.Contains(url, "v.douyin.com") && !strings.HasSuffix(url, "/") {
				url += "/"
			}

			return url
		}
	}

	return ""
}

// isValidVideoURL 检查是否为有效的视频URL
func isValidVideoURL(url string) bool {
	patterns := []string{
		`^https?://v\.douyin\.com/[A-Za-z0-9_-]+/?$`,
		`^https?://www\.douyin\.com/video/\d+`,
		`^https?://v\.kuaishou\.com/[A-Za-z0-9_-]+/?$`,
		`^https?://www\.kuaishou\.com/`,
		`^https?://(?:www\.)?tiktok\.com/`,
		`^https?://vm\.tiktok\.com/[A-Za-z0-9_-]+/?$`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, url)
		if matched {
			return true
		}
	}

	return false
}

// cleanURL 清理URL末尾的标点符号
func cleanURL(url string) string {
	// 移除末尾的标点符号
	punctuations := []string{".", ",", ";", "!", "?", "。", "，", "；", "！", "？"}

	for _, p := range punctuations {
		url = strings.TrimSuffix(url, p)
	}

	return url
}
