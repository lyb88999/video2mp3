package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"video-converter/internal/model"

	"github.com/gin-gonic/gin"
)

// DownloadHandler 下载处理器
type DownloadHandler struct {
	*BaseHandler
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler(deps *Dependencies) *DownloadHandler {
	return &DownloadHandler{
		BaseHandler: NewBaseHandler(deps),
	}
}

// DownloadFile 文件下载
func (h *DownloadHandler) DownloadFile(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.ValidationError(c, "任务ID不能为空")
		return
	}

	// 查找任务
	var task model.ConversionTask
	if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
		h.NotFoundError(c, "任务不存在")
		return
	}

	// 检查任务状态
	if task.Status != model.TaskStatusCompleted {
		h.ErrorResponse(c, http.StatusBadRequest, "任务尚未完成，无法下载", nil)
		return
	}

	// 检查输出文件是否存在
	if task.OutputPath == "" {
		h.ErrorResponse(c, http.StatusInternalServerError, "输出文件路径为空", nil)
		return
	}

	if _, err := os.Stat(task.OutputPath); os.IsNotExist(err) {
		h.ErrorResponse(c, http.StatusNotFound, "输出文件不存在", nil)
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(task.OutputPath)
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "输出文件不存在", nil)
		return
	}

	// 设置下载文件名
	filename := task.Title + ".mp3"
	if task.OriginalName != "" {
		// 如果有原始文件名，使用原始文件名但改为mp3扩展名
		originalBase := filepath.Base(task.OriginalName)
		ext := filepath.Ext(originalBase)
		if ext != "" {
			filename = originalBase[:len(originalBase)-len(ext)] + ".mp3"
		}
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// 发送文件
	c.File(task.OutputPath)
}

// CheckFile 检查文件是否存在（HEAD请求）
func (h *DownloadHandler) CheckFile(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// 查找任务
	var task model.ConversionTask
	if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 检查任务状态和文件
	if task.Status == model.TaskStatusCompleted && task.OutputPath != "" {
		if _, err := os.Stat(task.OutputPath); err == nil {
			c.Header("X-File-Exists", "true")
			c.Header("X-File-Size", getFileSize(task.OutputPath))
			c.Status(http.StatusOK)
			return
		}
	}

	c.Header("X-File-Exists", "false")
	c.Status(http.StatusNotFound)
}

// getFileSize 获取文件大小
func getFileSize(filePath string) string {
	if info, err := os.Stat(filePath); err == nil {
		return string(rune(info.Size()))
	}
	return "0"
}
