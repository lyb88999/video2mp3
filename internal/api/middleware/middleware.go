package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		method := c.Request.Method

		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", "*") // 生产环境应该设置具体域名
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理OPTIONS预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// Security 安全头中间件
func Security() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		c.Next()
	})
}

// RateLimit 速率限制中间件
func RateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()

		// 构造Redis键
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := context.Background()

		// 检查当前请求数
		current, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Redis错误，允许请求通过
			c.Next()
			return
		}

		// 速率限制配置：每分钟100个请求
		limit := 100
		window := time.Minute

		if current >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求过于频繁，请稍后再试",
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// 增加计数
		pipe := redisClient.TxPipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err = pipe.Exec(ctx)

		if err != nil {
			// Redis错误，记录日志但允许请求通过
			fmt.Printf("Rate limit Redis error: %v\n", err)
		}

		// 设置响应头显示限制信息
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(limit-current-1))

		c.Next()
	})
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 生成或获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置到context和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	})
}

// Logger 自定义日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s \"%s %s %s\" %d %s \"%s\" \"%s\" %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.ErrorMessage,
		)
	})
}

// Auth 认证中间件（简化版本）
func Auth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")

		// 简单的token验证（生产环境应该使用JWT或其他安全方案）
		if authHeader == "" {
			// 对于演示项目，我们暂时跳过认证
			// 在生产环境中，这里应该返回401错误
			c.Next()
			return
		}

		// TODO: 实现JWT token验证
		// 这里可以添加JWT解析和验证逻辑

		c.Next()
	})
}

// FileUpload 文件上传中间件
func FileUpload(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置最大文件大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单的时间戳生成方案
	// 生产环境建议使用UUID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Recovery 自定义恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录panic信息
		fmt.Printf("Panic recovered: %v\n", recovered)

		// 返回500错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器内部错误",
			"error":   "Internal server error",
		})
	})
}
