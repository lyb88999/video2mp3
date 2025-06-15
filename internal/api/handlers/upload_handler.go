package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"video-converter/internal/model"
	"video-converter/internal/storage"
	"video-converter/pkg/filemanager"
	"video-converter/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	*BaseHandler
	fileManager   *filemanager.FileManager
	fileValidator *validator.FileValidator
	redisManager  *storage.RedisManager
}

// UploadRequest 上传请求
type UploadRequest struct {
	Title       string `json:"title"`       // 可选的任务标题
	Description string `json:"description"` // 可选的任务描述
}

// UploadResponse 上传响应
type UploadResponse struct {
	TaskID       string                `json:"task_id"`
	OriginalName string                `json:"original_name"`
	FileSize     int64                 `json:"file_size"`
	Status       string                `json:"status"`
	UploadedAt   time.Time             `json:"uploaded_at"`
	FileInfo     *filemanager.FileInfo `json:"file_info"`
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(deps *Dependencies) *UploadHandler {
	// 创建文件管理器
	fm := filemanager.NewFileManager(
		deps.Config.File.UploadDir,
		deps.Config.File.OutputDir,
		deps.Config.File.TempDir,
	)

	// 创建文件验证器
	fv := validator.NewFileValidator(
		deps.Config.File.MaxFileSize,
		deps.Config.File.AllowedTypes,
	)

	// 创建Redis管理器
	rm := storage.NewRedisManager(deps.Redis)

	return &UploadHandler{
		BaseHandler:   NewBaseHandler(deps),
		fileManager:   fm,
		fileValidator: fv,
		redisManager:  rm,
	}
}

// UploadFile 处理文件上传
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 1. 解析multipart表单
	err := c.Request.ParseMultipartForm(h.cfg.File.MaxFileSize)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "解析上传表单失败", err)
		return
	}

	// 2. 获取上传的文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		h.ValidationError(c, "未找到上传文件，请确保表单字段名为'file'")
		return
	}
	defer file.Close()

	// 3. 验证文件基本信息
	validationResult := h.fileValidator.ValidateFile(fileHeader)
	if !validationResult.Valid {
		h.ErrorResponse(c, http.StatusBadRequest, "文件验证失败", fmt.Errorf("验证错误: %v", validationResult.Errors))
		return
	}

	// 4. 验证文件内容
	contentValidation := h.fileValidator.ValidateFileContent(file)
	if !contentValidation.Valid {
		h.ErrorResponse(c, http.StatusBadRequest, "文件内容验证失败", fmt.Errorf("内容验证错误: %v", contentValidation.Errors))
		return
	}

	// 5. 保存文件
	fileInfo, err := h.fileManager.SaveUploadedFile(fileHeader)
	if err != nil {
		h.InternalError(c, fmt.Errorf("保存文件失败: %v", err))
		return
	}

	// 6. 获取可选的表单参数
	title := c.PostForm("title")
	description := c.PostForm("description")

	// 如果没有提供标题，使用原文件名
	if title == "" {
		title = fileHeader.Filename
	}

	// 7. 创建转换任务记录
	task := &model.ConversionTask{
		ID:           uuid.New().String(),
		UserID:       "5ba6d3f8-4478-11f0-8574-74df036e1f50", // 使用默认用户ID
		Type:         model.TaskTypeFileUpload,
		Status:       model.TaskStatusQueued,
		Title:        title,
		Description:  description,
		InputPath:    fileInfo.FilePath,
		OriginalName: fileInfo.OriginalName,
		FileSize:     fileInfo.Size,
		Progress:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 8. 保存到数据库
	if err := h.db.Create(task).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		h.fileManager.DeleteFile(fileInfo.FilePath)
		h.InternalError(c, fmt.Errorf("创建任务记录失败: %v", err))
		return
	}

	// 9. 设置初始状态到Redis
	ctx := context.Background()
	h.redisManager.SetTaskStatus(ctx, task.ID, string(task.Status))
	h.redisManager.SetTaskProgress(ctx, task.ID, task.Progress)

	// 10. 构造响应
	response := UploadResponse{
		TaskID:       task.ID,
		OriginalName: fileInfo.OriginalName,
		FileSize:     fileInfo.Size,
		Status:       string(task.Status),
		UploadedAt:   task.CreatedAt,
		FileInfo:     fileInfo,
	}

	h.SuccessResponse(c, http.StatusCreated, "文件上传成功", response)
}

// GetUploadProgress 获取上传进度
func (h *UploadHandler) GetUploadProgress(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.ValidationError(c, "任务ID不能为空")
		return
	}

	ctx := context.Background()

	// 从Redis获取实时进度
	progress, err := h.redisManager.GetTaskProgress(ctx, taskID)
	if err != nil {
		// 如果Redis中没有，从数据库获取
		var task model.ConversionTask
		if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
			h.NotFoundError(c, "任务不存在")
			return
		}
		progress = task.Progress
	}

	// 获取任务状态
	status, err := h.redisManager.GetTaskStatus(ctx, taskID)
	if err != nil {
		var task model.ConversionTask
		if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
			h.NotFoundError(c, "任务不存在")
			return
		}
		status = string(task.Status)
	}

	h.SuccessResponse(c, http.StatusOK, "获取进度成功", gin.H{
		"task_id":  taskID,
		"progress": progress,
		"status":   status,
	})
}

// ListUploads 获取上传列表
func (h *UploadHandler) ListUploads(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// 查询任务列表
	var tasks []model.ConversionTask
	var total int64

	query := h.db.Model(&model.ConversionTask{})

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		h.InternalError(c, fmt.Errorf("查询任务总数失败: %v", err))
		return
	}

	// 获取任务列表
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&tasks).Error; err != nil {
		h.InternalError(c, fmt.Errorf("查询任务列表失败: %v", err))
		return
	}

	// 转换为摘要格式
	summaries := make([]model.TaskSummary, len(tasks))
	for i, task := range tasks {
		summaries[i] = task.ToSummary()
	}

	// 构造分页响应
	pagination := NewPaginationResponse(summaries, total, page, perPage)

	h.SuccessResponse(c, http.StatusOK, "获取上传列表成功", pagination)
}

// DeleteUpload 删除上传的文件和任务
func (h *UploadHandler) DeleteUpload(c *gin.Context) {
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

	// 检查任务是否可以删除（正在处理的任务不能删除）
	if task.Status == model.TaskStatusProcessing {
		h.ErrorResponse(c, http.StatusConflict, "正在处理的任务不能删除", nil)
		return
	}

	// 删除文件
	if task.InputPath != "" && h.fileManager.FileExists(task.InputPath) {
		if err := h.fileManager.DeleteFile(task.InputPath); err != nil {
			fmt.Printf("删除输入文件失败: %v\n", err)
		}
	}

	if task.OutputPath != "" && h.fileManager.FileExists(task.OutputPath) {
		if err := h.fileManager.DeleteFile(task.OutputPath); err != nil {
			fmt.Printf("删除输出文件失败: %v\n", err)
		}
	}

	// 删除Redis中的数据
	ctx := context.Background()
	h.redisManager.DeleteTaskData(ctx, taskID)

	// 删除数据库记录
	if err := h.db.Delete(&task).Error; err != nil {
		h.InternalError(c, fmt.Errorf("删除任务记录失败: %v", err))
		return
	}

	h.SuccessResponse(c, http.StatusOK, "删除成功", gin.H{
		"task_id": taskID,
	})
}
