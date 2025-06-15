package handlers

import (
	"net/http"
	"video-converter/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Dependencies 处理器依赖
type Dependencies struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

// BaseHandler 基础处理器
type BaseHandler struct {
	cfg   *config.Config
	db    *gorm.DB
	redis *redis.Client
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler(deps *Dependencies) *BaseHandler {
	return &BaseHandler{
		cfg:   deps.Config,
		db:    deps.DB,
		redis: deps.Redis,
	}
}

// HealthCheck 健康检查端点
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务运行正常",
		"data": gin.H{
			"status":    "healthy",
			"timestamp": gin.H{},
		},
	})
}

// Welcome 欢迎页面
func Welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "欢迎使用视频转MP3工具",
		"data": gin.H{
			"version": "1.0.0",
			"api":     "/api/v1",
			"docs":    "/docs",
		},
	})
}

// ServeSPA 服务单页应用
func ServeSPA(staticPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于SPA应用，所有未匹配的路由都返回index.html
		c.File(staticPath + "/index.html")
	}
}

// APIResponse 标准API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse 返回成功响应
func (h *BaseHandler) SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse 返回错误响应
func (h *BaseHandler) ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// ValidationError 验证错误响应
func (h *BaseHandler) ValidationError(c *gin.Context, message string) {
	h.ErrorResponse(c, http.StatusBadRequest, message, nil)
}

// InternalError 内部错误响应
func (h *BaseHandler) InternalError(c *gin.Context, err error) {
	h.ErrorResponse(c, http.StatusInternalServerError, "服务器内部错误", err)
}

// NotFoundError 未找到错误响应
func (h *BaseHandler) NotFoundError(c *gin.Context, message string) {
	h.ErrorResponse(c, http.StatusNotFound, message, nil)
}

// UnauthorizedError 未授权错误响应
func (h *BaseHandler) UnauthorizedError(c *gin.Context, message string) {
	h.ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginationResponse 创建分页响应
func NewPaginationResponse(items interface{}, total int64, page, perPage int) PaginationResponse {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	if totalPages <= 0 {
		totalPages = 1
	}

	return PaginationResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}
