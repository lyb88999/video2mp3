package converter

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"video-converter/internal/config"
)

// FFmpegConverter FFmpeg转换器
type FFmpegConverter struct {
	BinaryPath   string
	AudioCodec   string
	AudioBitrate string
	SampleRate   string
}

// ConversionOptions 转换选项
type ConversionOptions struct {
	InputPath    string
	OutputPath   string
	AudioCodec   string
	AudioBitrate string
	SampleRate   string
	OnProgress   func(progress float64) // 进度回调函数
}

// VideoInfo 视频信息
type VideoInfo struct {
	Duration float64 `json:"duration"` // 时长（秒）
	Size     int64   `json:"size"`     // 文件大小（字节）
	Format   string  `json:"format"`   // 格式
}

// NewFFmpegConverter 创建FFmpeg转换器
func NewFFmpegConverter(cfg *config.FFmpegConfig) *FFmpegConverter {
	return &FFmpegConverter{
		BinaryPath:   cfg.BinaryPath,
		AudioCodec:   cfg.AudioCodec,
		AudioBitrate: cfg.AudioBitrate,
		SampleRate:   cfg.SampleRate,
	}
}

// GetVideoInfo 获取视频信息
func (fc *FFmpegConverter) GetVideoInfo(inputPath string) (*VideoInfo, error) {
	// 使用ffprobe获取视频信息
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		inputPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe执行失败: %v", err)
	}

	// 解析duration (这里简化处理，使用正则提取)
	durationRegex := regexp.MustCompile(`"duration":"([^"]+)"`)
	matches := durationRegex.FindStringSubmatch(string(output))

	var duration float64
	if len(matches) > 1 {
		if d, err := strconv.ParseFloat(matches[1], 64); err == nil {
			duration = d
		}
	}

	// 获取文件大小
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return &VideoInfo{
		Duration: duration,
		Size:     fileInfo.Size(),
		Format:   filepath.Ext(inputPath),
	}, nil
}

// ConvertToMP3 转换视频为MP3
func (fc *FFmpegConverter) ConvertToMP3(ctx context.Context, options *ConversionOptions) error {
	// 确保输出目录存在
	outputDir := filepath.Dir(options.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 设置默认转换参数
	audioCodec := options.AudioCodec
	if audioCodec == "" {
		audioCodec = fc.AudioCodec
	}

	audioBitrate := options.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = fc.AudioBitrate
	}

	sampleRate := options.SampleRate
	if sampleRate == "" {
		sampleRate = fc.SampleRate
	}

	// 构建FFmpeg命令
	args := []string{
		"-i", options.InputPath, // 输入文件
		"-vn",                 // 不包含视频
		"-acodec", audioCodec, // 音频编码器
		"-ab", audioBitrate, // 音频比特率
		"-ar", sampleRate, // 音频采样率
		"-y",                  // 覆盖输出文件
		"-progress", "pipe:1", // 输出进度到标准输出
		options.OutputPath, // 输出文件
	}

	cmd := exec.CommandContext(ctx, fc.BinaryPath, args...)

	// 创建管道来读取进度信息
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动FFmpeg失败: %v", err)
	}

	// 获取视频总时长用于计算进度
	videoInfo, _ := fc.GetVideoInfo(options.InputPath)
	totalDuration := videoInfo.Duration

	// 读取进度信息
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// 解析进度信息
			if strings.HasPrefix(line, "out_time_us=") {
				timeStr := strings.TrimPrefix(line, "out_time_us=")
				if timeUs, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
					currentTime := float64(timeUs) / 1000000.0 // 转换为秒

					var progress float64
					if totalDuration > 0 {
						progress = (currentTime / totalDuration) * 100
						if progress > 100 {
							progress = 100
						}
					}

					// 调用进度回调
					if options.OnProgress != nil {
						options.OnProgress(progress)
					}
				}
			}
		}
	}()

	// 读取错误信息
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// 这里可以记录错误日志
			fmt.Printf("FFmpeg stderr: %s\n", scanner.Text())
		}
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg转换失败: %v", err)
	}

	// 转换完成，设置进度为100%
	if options.OnProgress != nil {
		options.OnProgress(100)
	}

	return nil
}

// ValidateInput 验证输入文件
func (fc *FFmpegConverter) ValidateInput(inputPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在: %s", inputPath)
	}

	// 检查是否为支持的格式
	ext := strings.ToLower(filepath.Ext(inputPath))
	supportedFormats := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v"}

	for _, format := range supportedFormats {
		if ext == format {
			return nil
		}
	}

	return fmt.Errorf("不支持的文件格式: %s", ext)
}

// GenerateOutputPath 生成输出文件路径
func (fc *FFmpegConverter) GenerateOutputPath(inputPath, outputDir string) string {
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s.mp3", baseName, timestamp)
	return filepath.Join(outputDir, fileName)
}
