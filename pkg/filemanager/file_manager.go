package filemanager

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileManager 文件管理器
type FileManager struct {
	UploadDir string // 上传目录
	OutputDir string // 输出目录
	TempDir   string // 临时目录
}

// FileInfo 文件信息
type FileInfo struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	SavedName    string    `json:"saved_name"`
	FilePath     string    `json:"file_path"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Extension    string    `json:"extension"`
	MD5Hash      string    `json:"md5_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewFileManager 创建文件管理器
func NewFileManager(uploadDir, outputDir, tempDir string) *FileManager {
	fm := &FileManager{
		UploadDir: uploadDir,
		OutputDir: outputDir,
		TempDir:   tempDir,
	}

	// 确保目录存在
	fm.ensureDirectories()
	return fm
}

// SaveUploadedFile 保存上传的文件
func (fm *FileManager) SaveUploadedFile(fileHeader *multipart.FileHeader) (*FileInfo, error) {
	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 生成文件信息
	fileInfo := &FileInfo{
		ID:           uuid.New().String(),
		OriginalName: fileHeader.Filename,
		Size:         fileHeader.Size,
		Extension:    strings.ToLower(filepath.Ext(fileHeader.Filename)),
		CreatedAt:    time.Now(),
	}

	// 生成保存的文件名（使用UUID避免重名）
	fileInfo.SavedName = fm.generateSavedName(fileInfo.ID, fileInfo.Extension)
	fileInfo.FilePath = filepath.Join(fm.UploadDir, fileInfo.SavedName)

	// 创建目标文件
	dst, err := os.Create(fileInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 创建MD5哈希计算器
	hash := md5.New()

	// 复制文件内容并计算哈希
	_, err = io.Copy(dst, io.TeeReader(src, hash))
	if err != nil {
		// 如果复制失败，删除已创建的文件
		os.Remove(fileInfo.FilePath)
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 设置MD5哈希
	fileInfo.MD5Hash = fmt.Sprintf("%x", hash.Sum(nil))

	return fileInfo, nil
}

// MoveToTemp 将文件移动到临时目录
func (fm *FileManager) MoveToTemp(filePath string) (string, error) {
	// 生成临时文件路径
	tempFileName := uuid.New().String() + filepath.Ext(filePath)
	tempPath := filepath.Join(fm.TempDir, tempFileName)

	// 移动文件
	err := fm.moveFile(filePath, tempPath)
	if err != nil {
		return "", fmt.Errorf("移动文件到临时目录失败: %v", err)
	}

	return tempPath, nil
}

// MoveToOutput 将文件移动到输出目录
func (fm *FileManager) MoveToOutput(filePath, newName string) (string, error) {
	outputPath := filepath.Join(fm.OutputDir, newName)

	err := fm.moveFile(filePath, outputPath)
	if err != nil {
		return "", fmt.Errorf("移动文件到输出目录失败: %v", err)
	}

	return outputPath, nil
}

// DeleteFile 删除文件
func (fm *FileManager) DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

// GetFileInfo 获取文件信息
func (fm *FileManager) GetFileInfo(filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return &FileInfo{
		OriginalName: stat.Name(),
		SavedName:    stat.Name(),
		FilePath:     filePath,
		Size:         stat.Size(),
		Extension:    filepath.Ext(filePath),
		CreatedAt:    stat.ModTime(),
	}, nil
}

// FileExists 检查文件是否存在
func (fm *FileManager) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileSize 获取文件大小
func (fm *FileManager) GetFileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// CleanupExpiredFiles 清理过期文件
func (fm *FileManager) CleanupExpiredFiles(directory string, expiry time.Duration) error {
	cutoff := time.Now().Add(-expiry)

	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 删除过期文件
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				fmt.Printf("删除过期文件失败 %s: %v\n", path, err)
			} else {
				fmt.Printf("已删除过期文件: %s\n", path)
			}
		}

		return nil
	})
}

// GetDirectorySize 获取目录大小
func (fm *FileManager) GetDirectorySize(directory string) (int64, error) {
	var size int64

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// generateSavedName 生成保存的文件名
func (fm *FileManager) generateSavedName(id, extension string) string {
	// 使用时间戳和UUID确保文件名唯一
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s%s", timestamp, id[:8], extension)
}

// moveFile 移动文件
func (fm *FileManager) moveFile(src, dst string) error {
	// 先尝试重命名（如果在同一个文件系统中）
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 如果重命名失败，则复制后删除
	return fm.copyAndDelete(src, dst)
}

// copyAndDelete 复制文件并删除原文件
func (fm *FileManager) copyAndDelete(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		os.Remove(dst) // 删除部分复制的文件
		return err
	}

	// 删除源文件
	return os.Remove(src)
}

// ensureDirectories 确保所有必要的目录存在
func (fm *FileManager) ensureDirectories() {
	directories := []string{fm.UploadDir, fm.OutputDir, fm.TempDir}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
		}
	}
}

// FormatFileSize 格式化文件大小为可读文本
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}
