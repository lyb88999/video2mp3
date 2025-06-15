package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	File     FileConfig     `mapstructure:"file"`
	FFmpeg   FFmpegConfig   `mapstructure:"ffmpeg"`
	Download DownloadConfig `mapstructure:"download"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// FileConfig 文件配置
type FileConfig struct {
	UploadDir    string   `mapstructure:"upload_dir"`
	OutputDir    string   `mapstructure:"output_dir"`
	TempDir      string   `mapstructure:"temp_dir"`
	MaxFileSize  int64    `mapstructure:"max_file_size"` // bytes
	AllowedTypes []string `mapstructure:"allowed_types"`
}

// FFmpegConfig FFmpeg配置
type FFmpegConfig struct {
	BinaryPath   string `mapstructure:"binary_path"`
	AudioCodec   string `mapstructure:"audio_codec"`
	AudioBitrate string `mapstructure:"audio_bitrate"`
	SampleRate   string `mapstructure:"sample_rate"`
}

// DownloadConfig 视频下载服务配置
type DownloadConfig struct {
	ServiceURL       string `mapstructure:"service_url"`
	Timeout          int    `mapstructure:"timeout"`
	EnablePrefix     bool   `mapstructure:"enable_prefix"`
	DisableWatermark bool   `mapstructure:"disable_watermark"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// 设置配置文件路径和名称
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// 默认配置文件位置
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("VIDEO_CONVERTER")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("配置文件未找到，使用默认配置")
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 将配置映射到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("配置解析失败: %v", err)
	}

	return config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "Lyb1217..")
	viper.SetDefault("database.dbname", "video_converter")

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// 文件默认配置
	viper.SetDefault("file.upload_dir", "./uploads")
	viper.SetDefault("file.output_dir", "./output")
	viper.SetDefault("file.temp_dir", os.TempDir())
	viper.SetDefault("file.max_file_size", 500*1024*1024) // 500MB
	viper.SetDefault("file.allowed_types", []string{
		"video/mp4", "video/avi", "video/mov", "video/wmv", "video/flv",
		"video/webm", "video/mkv", "video/m4v",
	})

	// FFmpeg默认配置
	viper.SetDefault("ffmpeg.binary_path", "ffmpeg")
	viper.SetDefault("ffmpeg.audio_codec", "libmp3lame")
	viper.SetDefault("ffmpeg.audio_bitrate", "192k")
	viper.SetDefault("ffmpeg.sample_rate", "44100")

	// 下载服务默认配置
	viper.SetDefault("download.service_url", "http://localhost:9080/api/download")
	viper.SetDefault("download.timeout", 300)
	viper.SetDefault("download.enable_prefix", true)
	viper.SetDefault("download.disable_watermark", true)
}

// GetDatabaseDSN 获取数据库连接字符串 (MySQL格式)
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
	)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	// 检查必要的目录是否存在，不存在则创建
	dirs := []string{c.File.UploadDir, c.File.OutputDir, c.File.TempDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}

	// 检查FFmpeg是否可用
	// 这里可以添加FFmpeg可执行文件的检查逻辑

	return nil
}
