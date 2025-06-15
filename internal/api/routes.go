package api

import (
	"net/http"
	"video-converter/internal/api/handlers"
	"video-converter/internal/api/middleware"
	"video-converter/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// SetupRoutes 设置API路由
func SetupRoutes(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 添加全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	router.Use(middleware.RateLimit(redisClient))

	// 创建handlers依赖
	deps := &handlers.Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
	}

	// 健康检查端点
	router.GET("/health", handlers.HealthCheck)

	// 静态文件服务（前端资源）
	router.Static("/static", "./web/static")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// 前端首页
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 文件上传相关
		upload := v1.Group("/upload")
		{
			upload.POST("", handlers.NewUploadHandler(deps).UploadFile)
			upload.GET("/progress/:id", handlers.NewUploadHandler(deps).GetUploadProgress)
			upload.GET("/list", handlers.NewUploadHandler(deps).ListUploads)
			upload.DELETE("/:id", handlers.NewUploadHandler(deps).DeleteUpload)
		}

		// 转换相关
		convert := v1.Group("/convert")
		{
			convert.POST("/file", handlers.NewConvertHandler(deps).ConvertFile)
			convert.POST("/url", handlers.NewConvertHandler(deps).ConvertURL)
		}

		// 任务管理
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", handlers.NewTaskHandler(deps).ListTasks)
			tasks.GET("/:id", handlers.NewTaskHandler(deps).GetTask)
			tasks.POST("/:id/cancel", handlers.NewTaskHandler(deps).CancelTask)
			tasks.GET("/:id/status", handlers.NewTaskHandler(deps).GetTaskStatus)
		}

		// 文件下载
		download := v1.Group("/download")
		{
			download.GET("/:id", handlers.NewDownloadHandler(deps).DownloadFile)
			download.HEAD("/:id", handlers.NewDownloadHandler(deps).CheckFile)
		}

		// WebSocket连接（进度推送）
		// v1.GET("/ws/progress", handlers.NewWebSocketHandler(deps).HandleProgress)
	}

	// SPA路由处理 - 处理前端路由
	router.NoRoute(func(c *gin.Context) {
		// 如果请求路径以/api开头，返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "API端点不存在",
			})
			return
		}

		// 否则服务前端SPA
		c.File("./web/index.html")
	})

	return router
}

// APIResponse 标准API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string, err error) APIResponse {
	response := APIResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}
