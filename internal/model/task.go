package model

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus 任务状态类型
type TaskStatus string

const (
	TaskStatusQueued     TaskStatus = "queued"     // 排队中
	TaskStatusProcessing TaskStatus = "processing" // 处理中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
	TaskStatusCanceled   TaskStatus = "canceled"   // 已取消
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeFileUpload TaskType = "file_upload" // 文件上传
	TaskTypeURLConvert TaskType = "url_convert" // URL转换
)

// ConversionTask 转换任务模型
type ConversionTask struct {
	ID          string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID      string     `json:"user_id" gorm:"type:varchar(36);index"`
	Type        TaskType   `json:"type" gorm:"type:varchar(20);not null"`
	Status      TaskStatus `json:"status" gorm:"type:varchar(20);default:'queued';index"`
	Title       string     `json:"title" gorm:"type:varchar(255)"`
	Description string     `json:"description" gorm:"type:text"`

	// 文件路径信息
	InputPath    string `json:"input_path" gorm:"type:varchar(500)"`
	OutputPath   string `json:"output_path" gorm:"type:varchar(500)"`
	ThumbnailURL string `json:"thumbnail_url" gorm:"type:varchar(500)"`

	// 原始信息
	OriginalURL  string `json:"original_url" gorm:"type:varchar(1000)"` // 原始视频URL
	OriginalName string `json:"original_name" gorm:"type:varchar(255)"` // 原始文件名

	// 文件信息
	FileSize   int64   `json:"file_size"`                 // 输入文件大小(bytes)
	OutputSize int64   `json:"output_size"`               // 输出文件大小(bytes)
	Duration   float64 `json:"duration"`                  // 视频时长(秒)
	Progress   float64 `json:"progress" gorm:"default:0"` // 进度(0-100)

	// 转换参数
	AudioCodec   string `json:"audio_codec" gorm:"type:varchar(50);default:'libmp3lame'"`
	AudioBitrate string `json:"audio_bitrate" gorm:"type:varchar(20);default:'192k'"`
	SampleRate   string `json:"sample_rate" gorm:"type:varchar(20);default:'44100'"`

	// 错误信息
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// User 用户模型
type User struct {
	ID       string `json:"id" gorm:"type:varchar(36);primaryKey"`
	Username string `json:"username" gorm:"type:varchar(50);uniqueIndex"`
	Email    string `json:"email" gorm:"type:varchar(100);uniqueIndex"`
	Password string `json:"-" gorm:"type:varchar(255)"` // 不在JSON中显示
	Avatar   string `json:"avatar" gorm:"type:varchar(500)"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	// 统计信息
	TaskCount      int `json:"task_count" gorm:"default:0"`
	CompletedCount int `json:"completed_count" gorm:"default:0"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Tasks []ConversionTask `json:"tasks,omitempty" gorm:"foreignKey:UserID"`
}

// TaskSummary 任务摘要（用于列表显示）
type TaskSummary struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Type         TaskType   `json:"type"`
	Status       TaskStatus `json:"status"`
	Progress     float64    `json:"progress"`
	Duration     float64    `json:"duration"`
	FileSize     int64      `json:"file_size"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// ToSummary 转换为任务摘要
func (t *ConversionTask) ToSummary() TaskSummary {
	return TaskSummary{
		ID:           t.ID,
		Title:        t.Title,
		Type:         t.Type,
		Status:       t.Status,
		Progress:     t.Progress,
		Duration:     t.Duration,
		FileSize:     t.FileSize,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		ErrorMessage: t.ErrorMessage,
	}
}

// IsCompleted 检查任务是否已完成
func (t *ConversionTask) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsFailed 检查任务是否失败
func (t *ConversionTask) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// IsProcessing 检查任务是否正在处理
func (t *ConversionTask) IsProcessing() bool {
	return t.Status == TaskStatusProcessing
}

// CanCancel 检查任务是否可以取消
func (t *ConversionTask) CanCancel() bool {
	return t.Status == TaskStatusQueued || t.Status == TaskStatusProcessing
}

// BeforeCreate 创建前钩子
func (t *ConversionTask) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		// 如果没有设置ID，这里可以生成一个UUID
		// 但在我们的实现中，会在创建时手动设置ID
	}
	return nil
}

// TableName 指定表名
func (ConversionTask) TableName() string {
	return "conversion_tasks"
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
