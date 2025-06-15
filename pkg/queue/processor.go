package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"video-converter/internal/model"
	"video-converter/internal/storage"
	"video-converter/pkg/converter"

	"gorm.io/gorm"
)

// TaskProcessor 任务处理器
type TaskProcessor struct {
	db              *gorm.DB
	redisManager    *storage.RedisManager
	ffmpegConverter *converter.FFmpegConverter
	outputDir       string
	workers         int
	taskChan        chan *model.ConversionTask
	stopChan        chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.RWMutex
}

// NewTaskProcessor 创建任务处理器
func NewTaskProcessor(db *gorm.DB, redisManager *storage.RedisManager, ffmpegConverter *converter.FFmpegConverter, outputDir string, workers int) *TaskProcessor {
	return &TaskProcessor{
		db:              db,
		redisManager:    redisManager,
		ffmpegConverter: ffmpegConverter,
		outputDir:       outputDir,
		workers:         workers,
		taskChan:        make(chan *model.ConversionTask, 100),
		stopChan:        make(chan struct{}),
	}
}

// Start 启动任务处理器
func (tp *TaskProcessor) Start() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.running {
		return
	}

	tp.running = true
	log.Printf("启动 %d 个任务处理器worker", tp.workers)

	// 启动worker goroutines
	for i := 0; i < tp.workers; i++ {
		tp.wg.Add(1)
		go tp.worker(i)
	}

	// 启动任务扫描器
	tp.wg.Add(1)
	go tp.taskScanner()
}

// Stop 停止任务处理器
func (tp *TaskProcessor) Stop() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if !tp.running {
		return
	}

	log.Println("正在停止任务处理器...")
	close(tp.stopChan)
	tp.wg.Wait()
	tp.running = false
	log.Println("任务处理器已停止")
}

// AddTask 添加任务到队列
func (tp *TaskProcessor) AddTask(task *model.ConversionTask) {
	select {
	case tp.taskChan <- task:
		log.Printf("任务 %s 已添加到队列", task.ID)
	default:
		log.Printf("任务队列已满，任务 %s 添加失败", task.ID)
	}
}

// worker 处理任务的worker goroutine
func (tp *TaskProcessor) worker(workerID int) {
	defer tp.wg.Done()
	log.Printf("Worker %d 启动", workerID)

	for {
		select {
		case task := <-tp.taskChan:
			tp.processTask(workerID, task)
		case <-tp.stopChan:
			log.Printf("Worker %d 收到停止信号", workerID)
			return
		}
	}
}

// taskScanner 定期扫描数据库中的排队任务
func (tp *TaskProcessor) taskScanner() {
	defer tp.wg.Done()
	ticker := time.NewTicker(10 * time.Second) // 每10秒扫描一次
	defer ticker.Stop()

	log.Println("任务扫描器启动")

	for {
		select {
		case <-ticker.C:
			tp.scanQueuedTasks()
		case <-tp.stopChan:
			log.Println("任务扫描器收到停止信号")
			return
		}
	}
}

// scanQueuedTasks 扫描排队中的任务
func (tp *TaskProcessor) scanQueuedTasks() {
	var tasks []model.ConversionTask

	// 查找状态为排队的任务
	err := tp.db.Where("status = ?", model.TaskStatusQueued).
		Order("created_at ASC").
		Limit(10). // 限制每次处理的任务数
		Find(&tasks).Error

	if err != nil {
		log.Printf("扫描排队任务失败: %v", err)
		return
	}

	for _, task := range tasks {
		// 尝试添加到处理队列
		select {
		case tp.taskChan <- &task:
			// 更新任务状态为处理中
			tp.updateTaskStatus(&task, model.TaskStatusProcessing)
		default:
			// 队列已满，跳过
			break
		}
	}
}

// processTask 处理具体的转换任务
func (tp *TaskProcessor) processTask(workerID int, task *model.ConversionTask) {
	log.Printf("Worker %d 开始处理任务 %s", workerID, task.ID)

	ctx := context.Background()

	// 更新任务状态为处理中
	tp.updateTaskStatus(task, model.TaskStatusProcessing)
	tp.updateTaskProgress(ctx, task, 0)

	// 验证输入文件
	if err := tp.ffmpegConverter.ValidateInput(task.InputPath); err != nil {
		tp.failTask(ctx, task, fmt.Sprintf("输入文件验证失败: %v", err))
		return
	}

	// 生成输出文件路径
	outputPath := tp.ffmpegConverter.GenerateOutputPath(task.InputPath, tp.outputDir)

	// 获取视频信息
	videoInfo, err := tp.ffmpegConverter.GetVideoInfo(task.InputPath)
	if err != nil {
		log.Printf("获取视频信息失败: %v", err)
	} else {
		// 更新数据库中的视频时长
		task.Duration = videoInfo.Duration
		tp.db.Model(task).Update("duration", videoInfo.Duration)
	}

	// 设置转换选项
	options := &converter.ConversionOptions{
		InputPath:    task.InputPath,
		OutputPath:   outputPath,
		AudioCodec:   task.AudioCodec,
		AudioBitrate: task.AudioBitrate,
		SampleRate:   task.SampleRate,
		OnProgress: func(progress float64) {
			// 更新进度
			tp.updateTaskProgress(ctx, task, progress)
		},
	}

	// 执行转换
	conversionCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	err = tp.ffmpegConverter.ConvertToMP3(conversionCtx, options)
	if err != nil {
		tp.failTask(ctx, task, fmt.Sprintf("视频转换失败: %v", err))
		return
	}

	// 转换成功，更新任务
	task.OutputPath = outputPath
	task.Progress = 100
	task.Status = model.TaskStatusCompleted
	task.UpdatedAt = time.Now()

	// 获取输出文件大小
	if outputFileInfo, err := os.Stat(outputPath); err == nil {
		task.OutputSize = outputFileInfo.Size()
	}

	if err := tp.db.Save(task).Error; err != nil {
		log.Printf("更新任务状态失败: %v", err)
	}

	// 更新Redis状态
	tp.redisManager.SetTaskStatus(ctx, task.ID, string(model.TaskStatusCompleted))
	tp.redisManager.SetTaskProgress(ctx, task.ID, 100)

	log.Printf("Worker %d 完成任务 %s", workerID, task.ID)
}

// updateTaskStatus 更新任务状态
func (tp *TaskProcessor) updateTaskStatus(task *model.ConversionTask, status model.TaskStatus) {
	task.Status = status
	task.UpdatedAt = time.Now()

	if err := tp.db.Model(task).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": task.UpdatedAt,
	}).Error; err != nil {
		log.Printf("更新任务状态失败: %v", err)
	}

	// 同时更新Redis
	ctx := context.Background()
	tp.redisManager.SetTaskStatus(ctx, task.ID, string(status))
}

// updateTaskProgress 更新任务进度
func (tp *TaskProcessor) updateTaskProgress(ctx context.Context, task *model.ConversionTask, progress float64) {
	// 更新Redis中的进度
	tp.redisManager.SetTaskProgress(ctx, task.ID, progress)

	// 定期更新数据库（避免过于频繁的数据库写入）
	if int(progress)%10 == 0 || progress >= 100 {
		task.Progress = progress
		task.UpdatedAt = time.Now()
		tp.db.Model(task).Updates(map[string]interface{}{
			"progress":   progress,
			"updated_at": task.UpdatedAt,
		})
	}
}

// failTask 标记任务失败
func (tp *TaskProcessor) failTask(ctx context.Context, task *model.ConversionTask, errorMsg string) {
	log.Printf("任务 %s 失败: %s", task.ID, errorMsg)

	task.Status = model.TaskStatusFailed
	task.ErrorMessage = errorMsg
	task.UpdatedAt = time.Now()

	if err := tp.db.Save(task).Error; err != nil {
		log.Printf("更新失败任务状态失败: %v", err)
	}

	// 更新Redis状态
	tp.redisManager.SetTaskStatus(ctx, task.ID, string(model.TaskStatusFailed))
}
