package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"url-shortener/config"

	"github.com/go-redis/redis/v8"
)

// RedisClient là wrapper cho Redis client
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewRedisClient tạo kết nối mới đến Redis
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Kiểm tra kết nối
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Connected to Redis successfully")

	return &RedisClient{
		Client: client,
		Ctx:    ctx,
	}, nil
}

// Set lưu giá trị vào Redis với TTL
func (r *RedisClient) Set(key, value string, expiration time.Duration) error {
	return r.Client.Set(r.Ctx, key, value, expiration).Err()
}

// Get lấy giá trị từ Redis
func (r *RedisClient) Get(key string) (string, error) {
	return r.Client.Get(r.Ctx, key).Result()
}

// Delete xóa key từ Redis
func (r *RedisClient) Delete(key string) error {
	return r.Client.Del(r.Ctx, key).Err()
}

// Exists kiểm tra key có tồn tại không
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.Client.Exists(r.Ctx, key).Result()
	return result > 0, err
}

// Incr tăng giá trị của key
func (r *RedisClient) Incr(key string) (int64, error) {
	return r.Client.Incr(r.Ctx, key).Result()
}

// Close đóng kết nối Redis
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
