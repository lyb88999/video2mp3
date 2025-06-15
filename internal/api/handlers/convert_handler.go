package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"video-converter/internal/model"
	"video-converter/internal/storage"
	"video-converter/pkg/converter"

	"video-converter/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConvertHandler 转换处理器
type ConvertHandler struct {
	*BaseHandler
	ffmpegConverter *converter.FFmpegConverter
	redisManager    *storage.RedisManager
}

// ConvertFileRequest 文件转换请求
type ConvertFileRequest struct {
	TaskID       string `json:"task_id" binding:"required"` // 上传任务ID
	AudioCodec   string `json:"audio_codec,omitempty"`      // 音频编码器
	AudioBitrate string `json:"audio_bitrate,omitempty"`    // 音频比特率
	SampleRate   string `json:"sample_rate,omitempty"`      // 采样率
}

// ConvertURLRequest URL转换请求
type ConvertURLRequest struct {
	URL          string `json:"url" binding:"required"`  // 视频URL
	Title        string `json:"title,omitempty"`         // 任务标题
	Description  string `json:"description,omitempty"`   // 任务描述
	AudioCodec   string `json:"audio_codec,omitempty"`   // 音频编码器
	AudioBitrate string `json:"audio_bitrate,omitempty"` // 音频比特率
	SampleRate   string `json:"sample_rate,omitempty"`   // 采样率
}

// ConvertResponse 转换响应
type ConvertResponse struct {
	TaskID            string    `json:"task_id"`
	Status            string    `json:"status"`
	Message           string    `json:"message"`
	CreatedAt         time.Time `json:"created_at"`
	EstimatedDuration string    `json:"estimated_duration,omitempty"`
}

// NewConvertHandler 创建转换处理器
func NewConvertHandler(deps *Dependencies) *ConvertHandler {
	// 创建FFmpeg转换器
	ffmpegConverter := converter.NewFFmpegConverter(&deps.Config.FFmpeg)

	// 创建Redis管理器
	redisManager := storage.NewRedisManager(deps.Redis)

	return &ConvertHandler{
		BaseHandler:     NewBaseHandler(deps),
		ffmpegConverter: ffmpegConverter,
		redisManager:    redisManager,
	}
}

// ConvertFile 文件转换
func (h *ConvertHandler) ConvertFile(c *gin.Context) {
	var req ConvertFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ValidationError(c, "请求参数格式错误: "+err.Error())
		return
	}

	// 查找上传任务
	var uploadTask model.ConversionTask
	if err := h.db.First(&uploadTask, "id = ? AND type = ?", req.TaskID, model.TaskTypeFileUpload).Error; err != nil {
		h.NotFoundError(c, "找不到对应的上传任务")
		return
	}

	// 检查上传任务状态
	if uploadTask.Status != model.TaskStatusQueued {
		h.ErrorResponse(c, http.StatusBadRequest, "只能转换处于排队状态的上传任务", nil)
		return
	}

	// 验证输入文件
	if err := h.ffmpegConverter.ValidateInput(uploadTask.InputPath); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "输入文件验证失败", err)
		return
	}

	// 设置转换参数（使用请求参数或默认值）
	audioCodec := req.AudioCodec
	if audioCodec == "" {
		audioCodec = h.cfg.FFmpeg.AudioCodec
	}

	audioBitrate := req.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = h.cfg.FFmpeg.AudioBitrate
	}

	sampleRate := req.SampleRate
	if sampleRate == "" {
		sampleRate = h.cfg.FFmpeg.SampleRate
	}

	// 更新任务转换参数
	uploadTask.AudioCodec = audioCodec
	uploadTask.AudioBitrate = audioBitrate
	uploadTask.SampleRate = sampleRate
	uploadTask.UpdatedAt = time.Now()

	if err := h.db.Save(&uploadTask).Error; err != nil {
		h.InternalError(c, err)
		return
	}

	// 获取视频信息估算时长
	estimatedDuration := "未知"
	if videoInfo, err := h.ffmpegConverter.GetVideoInfo(uploadTask.InputPath); err == nil {
		// 简单估算：通常转换时间约为视频时长的10-50%
		estimated := int(videoInfo.Duration * 0.3) // 假设30%的时间
		if estimated < 10 {
			estimated = 10 // 最少10秒
		}
		estimatedDuration = time.Duration(estimated * int(time.Second)).String()

		// 更新数据库中的视频时长
		uploadTask.Duration = videoInfo.Duration
		h.db.Model(&uploadTask).Update("duration", videoInfo.Duration)
	}

	response := ConvertResponse{
		TaskID:            uploadTask.ID,
		Status:            string(uploadTask.Status),
		Message:           "转换任务已创建，正在排队处理",
		CreatedAt:         uploadTask.CreatedAt,
		EstimatedDuration: estimatedDuration,
	}

	h.SuccessResponse(c, http.StatusOK, "转换任务创建成功", response)
}

// ConvertURL URL转换
func (h *ConvertHandler) ConvertURL(c *gin.Context) {
	var req ConvertURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 提取真实的视频URL
	extractedURL := utils.ExtractVideoURL(req.URL)
	if extractedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到有效的视频链接"})
		return
	}

	// 如果提取的URL与原始URL不同，记录日志
	if extractedURL != req.URL {
		fmt.Printf("URL提取: 原始='%s' -> 提取='%s'\n", req.URL, extractedURL)
	}

	// 使用提取的URL
	req.URL = extractedURL

	// 设置默认标题
	if req.Title == "" {
		req.Title = "URL视频转换"
	}

	// 设置转换参数
	audioCodec := req.AudioCodec
	if audioCodec == "" {
		audioCodec = h.cfg.FFmpeg.AudioCodec
	}

	audioBitrate := req.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = h.cfg.FFmpeg.AudioBitrate
	}

	sampleRate := req.SampleRate
	if sampleRate == "" {
		sampleRate = h.cfg.FFmpeg.SampleRate
	}

	// 创建URL转换任务
	task := &model.ConversionTask{
		ID:           uuid.New().String(),
		UserID:       "5ba6d3f8-4478-11f0-8574-74df036e1f50", // 使用默认用户ID
		Type:         model.TaskTypeURLConvert,
		Status:       model.TaskStatusQueued,
		Title:        req.Title,
		Description:  req.Description,
		OriginalURL:  req.URL,
		AudioCodec:   audioCodec,
		AudioBitrate: audioBitrate,
		SampleRate:   sampleRate,
		Progress:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := h.db.Create(task).Error; err != nil {
		h.InternalError(c, err)
		return
	}

	// 设置初始状态到Redis
	ctx := c.Request.Context()
	h.redisManager.SetTaskStatus(ctx, task.ID, string(task.Status))
	h.redisManager.SetTaskProgress(ctx, task.ID, task.Progress)

	// 启动异步下载任务
	go h.downloadAndProcessTask(task.ID, req.URL)

	response := ConvertResponse{
		TaskID:            task.ID,
		Status:            string(task.Status),
		Message:           "URL转换任务已创建，正在下载视频",
		CreatedAt:         task.CreatedAt,
		EstimatedDuration: "根据视频长度而定",
	}

	h.SuccessResponse(c, http.StatusCreated, "URL转换任务创建成功", response)
}

// downloadAndProcessTask 下载视频并更新任务状态
func (h *ConvertHandler) downloadAndProcessTask(taskID, videoURL string) {
	ctx := context.Background()

	// 更新任务状态为处理中
	task := &model.ConversionTask{}
	if err := h.db.First(task, "id = ?", taskID).Error; err != nil {
		return
	}

	task.Status = model.TaskStatusProcessing
	h.db.Save(task)
	h.redisManager.SetTaskStatus(ctx, taskID, string(model.TaskStatusProcessing))

	// 直接从下载服务下载视频文件
	inputPath, originalName, err := h.downloadVideoFromService(taskID, videoURL)
	if err != nil {
		// 更新任务为失败状态
		task.Status = model.TaskStatusFailed
		task.ErrorMessage = fmt.Sprintf("下载视频失败: %v", err)
		h.db.Save(task)
		h.redisManager.SetTaskStatus(ctx, taskID, string(model.TaskStatusFailed))
		return
	}

	// 更新任务的输入路径和原始文件名
	task.InputPath = inputPath
	task.OriginalName = originalName
	task.Status = model.TaskStatusQueued // 重新放入队列等待转换
	h.db.Save(task)
	h.redisManager.SetTaskStatus(ctx, taskID, string(model.TaskStatusQueued))
}

// downloadVideoFromService 从下载服务下载视频文件
func (h *ConvertHandler) downloadVideoFromService(taskID, videoURL string) (string, string, error) {
	// 构建请求URL
	params := url.Values{}
	params.Add("url", videoURL)
	params.Add("prefix", fmt.Sprintf("%t", h.cfg.Download.EnablePrefix))
	params.Add("with_watermark", fmt.Sprintf("%t", !h.cfg.Download.DisableWatermark))

	requestURL := h.cfg.Download.ServiceURL + "?" + params.Encode()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(h.cfg.Download.Timeout) * time.Second,
	}

	// 下载视频
	resp, err := client.Get(requestURL)
	if err != nil {
		return "", "", fmt.Errorf("调用下载服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("下载服务返回错误状态: %d", resp.StatusCode)
	}

	// 检查Content-Type，如果是JSON则说明出错了
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 读取错误响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("读取错误响应失败: %v", err)
		}

		// 尝试解析错误信息
		var errorResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", "", fmt.Errorf("下载服务错误[%d]: %s", errorResp.Code, errorResp.Message)
		}

		return "", "", fmt.Errorf("下载失败，可能是链接已过期或无效: %s", string(body))
	}

	// 从Content-Disposition头获取文件名
	contentDisposition := resp.Header.Get("Content-Disposition")
	var originalName string
	if contentDisposition != "" {
		// 解析 Content-Disposition 头
		if matches := contentDispositionRegex.FindStringSubmatch(contentDisposition); len(matches) > 1 {
			originalName = matches[1]
		}
	}

	// 如果没有获取到文件名，使用默认名称
	if originalName == "" {
		originalName = fmt.Sprintf("%s.mp4", taskID)
	}

	// 生成本地文件路径
	inputPath := filepath.Join(h.cfg.File.TempDir, originalName)

	// 创建本地文件
	file, err := os.Create(inputPath)
	if err != nil {
		return "", "", fmt.Errorf("创建本地文件失败: %v", err)
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("保存视频文件失败: %v", err)
	}

	return inputPath, originalName, nil
}

// 正则表达式用于解析Content-Disposition头
var contentDispositionRegex = regexp.MustCompile(`filename="([^"]+)"`)

// isValidURL 简单的URL验证
func isValidURL(url string) bool {
	// 简化验证，检查是否以http或https开头
	return len(url) > 0 && (len(url) > 7 && url[:7] == "http://" ||
		len(url) > 8 && url[:8] == "https://")
}
