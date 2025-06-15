package handlers

import (
	"net/http"
	"strconv"

	"video-converter/internal/model"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	*BaseHandler
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(deps *Dependencies) *TaskHandler {
	return &TaskHandler{
		BaseHandler: NewBaseHandler(deps),
	}
}

// ListTasks 任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.Query("status") // 可选的状态过滤

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// 构建查询
	query := h.db.Model(&model.ConversionTask{})

	// 如果指定了状态，添加状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 查询任务列表
	var tasks []model.ConversionTask
	var total int64

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		h.InternalError(c, err)
		return
	}

	// 获取任务列表
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&tasks).Error; err != nil {
		h.InternalError(c, err)
		return
	}

	// 转换为摘要格式
	summaries := make([]model.TaskSummary, len(tasks))
	for i, task := range tasks {
		summaries[i] = task.ToSummary()
	}

	// 构造分页响应
	pagination := NewPaginationResponse(summaries, total, page, perPage)

	h.SuccessResponse(c, http.StatusOK, "获取任务列表成功", pagination)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.ValidationError(c, "任务ID不能为空")
		return
	}

	var task model.ConversionTask
	if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
		h.NotFoundError(c, "任务不存在")
		return
	}

	h.SuccessResponse(c, http.StatusOK, "获取任务详情成功", task)
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.ValidationError(c, "任务ID不能为空")
		return
	}

	var task model.ConversionTask
	if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
		h.NotFoundError(c, "任务不存在")
		return
	}

	// 检查任务是否可以取消
	if !task.CanCancel() {
		h.ErrorResponse(c, http.StatusBadRequest, "任务当前状态不允许取消", nil)
		return
	}

	// 更新任务状态为已取消
	task.Status = model.TaskStatusCanceled
	if err := h.db.Save(&task).Error; err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessResponse(c, http.StatusOK, "任务已取消", gin.H{
		"task_id": taskID,
		"status":  string(task.Status),
	})
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.ValidationError(c, "任务ID不能为空")
		return
	}

	var task model.ConversionTask
	if err := h.db.First(&task, "id = ?", taskID).Error; err != nil {
		h.NotFoundError(c, "任务不存在")
		return
	}

	h.SuccessResponse(c, http.StatusOK, "获取任务状态成功", gin.H{
		"id":            task.ID,
		"status":        string(task.Status),
		"progress":      task.Progress,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
		"error_message": task.ErrorMessage,
	})
}
