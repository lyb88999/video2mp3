package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"video-converter/internal/api"
	"video-converter/internal/config"
	"video-converter/internal/storage"
	"video-converter/pkg/converter"
	"video-converter/pkg/queue"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库连接
	db, err := storage.NewMySQLDB(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer storage.CloseDB(db)

	// 初始化Redis连接
	redisClient, err := storage.NewRedisClient(cfg.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	defer redisClient.Close()

	// 创建FFmpeg转换器
	ffmpegConverter := converter.NewFFmpegConverter(&cfg.FFmpeg)

	// 创建Redis管理器
	redisManager := storage.NewRedisManager(redisClient)

	// 创建任务处理器
	taskProcessor := queue.NewTaskProcessor(db, redisManager, ffmpegConverter, cfg.File.OutputDir, 2) // 2个worker

	// 启动任务处理器
	taskProcessor.Start()
	defer taskProcessor.Stop()

	// 创建API路由
	router := api.SetupRoutes(cfg, db, redisClient)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在 http://%s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Printf("任务处理器已启动，开始监听转换任务...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 首先停止任务处理器
	log.Println("正在停止任务处理器...")
	taskProcessor.Stop()

	// 5秒超时的优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭超时:", err)
	}

	log.Println("服务器已关闭")
}
