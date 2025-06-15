package storage

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient 创建Redis客户端
func NewRedisClient(addr, password string, db int) (*redis.Client, error) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("Redis连接成功")
	return client, nil
}

// RedisManager Redis管理器
type RedisManager struct {
	client *redis.Client
}

// NewRedisManager 创建Redis管理器
func NewRedisManager(client *redis.Client) *RedisManager {
	return &RedisManager{
		client: client,
	}
}

// SetTaskProgress 设置任务进度
func (rm *RedisManager) SetTaskProgress(ctx context.Context, taskID string, progress float64) error {
	key := "task_progress:" + taskID
	return rm.client.Set(ctx, key, progress, time.Hour*24).Err()
}

// GetTaskProgress 获取任务进度
func (rm *RedisManager) GetTaskProgress(ctx context.Context, taskID string) (float64, error) {
	key := "task_progress:" + taskID
	return rm.client.Get(ctx, key).Float64()
}

// SetTaskStatus 设置任务状态
func (rm *RedisManager) SetTaskStatus(ctx context.Context, taskID, status string) error {
	key := "task_status:" + taskID
	return rm.client.Set(ctx, key, status, time.Hour*24).Err()
}

// GetTaskStatus 获取任务状态
func (rm *RedisManager) GetTaskStatus(ctx context.Context, taskID string) (string, error) {
	key := "task_status:" + taskID
	return rm.client.Get(ctx, key).Result()
}

// DeleteTaskData 删除任务相关的Redis数据
func (rm *RedisManager) DeleteTaskData(ctx context.Context, taskID string) error {
	keys := []string{
		"task_progress:" + taskID,
		"task_status:" + taskID,
	}

	return rm.client.Del(ctx, keys...).Err()
}

// SetCacheWithExpiry 设置缓存（带过期时间）
func (rm *RedisManager) SetCacheWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return rm.client.Set(ctx, key, value, expiry).Err()
}

// GetCache 获取缓存
func (rm *RedisManager) GetCache(ctx context.Context, key string) (string, error) {
	return rm.client.Get(ctx, key).Result()
}

// DeleteCache 删除缓存
func (rm *RedisManager) DeleteCache(ctx context.Context, key string) error {
	return rm.client.Del(ctx, key).Err()
}
