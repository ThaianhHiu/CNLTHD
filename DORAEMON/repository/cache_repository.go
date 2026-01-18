package repository

import (
	"fmt"
	"time"

	"url-shortener/database"
)

// CacheRepositoryImpl là implementation của CacheRepository
type CacheRepositoryImpl struct {
	redis      *database.RedisClient
	expiration time.Duration
}

// NewCacheRepository tạo instance mới của CacheRepository
func NewCacheRepository(redis *database.RedisClient) *CacheRepositoryImpl {
	return &CacheRepositoryImpl{
		redis:      redis,
		expiration: 24 * time.Hour, // Cache 24 giờ
	}
}

// Set lưu URL vào cache
func (r *CacheRepositoryImpl) Set(shortCode string, originalURL string) error {
	key := r.buildKey(shortCode)
	return r.redis.Set(key, originalURL, r.expiration)
}

// Get lấy original URL từ cache
func (r *CacheRepositoryImpl) Get(shortCode string) (string, error) {
	key := r.buildKey(shortCode)
	return r.redis.Get(key)
}

// Delete xóa URL khỏi cache
func (r *CacheRepositoryImpl) Delete(shortCode string) error {
	key := r.buildKey(shortCode)
	return r.redis.Delete(key)
}

// Exists kiểm tra URL có trong cache không
func (r *CacheRepositoryImpl) Exists(shortCode string) (bool, error) {
	key := r.buildKey(shortCode)
	return r.redis.Exists(key)
}

// buildKey tạo key cho Redis
func (r *CacheRepositoryImpl) buildKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}
