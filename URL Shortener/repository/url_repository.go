package repository

import (
	"fmt"
	"time"

	"url-shortener/models"

	"gorm.io/gorm"
)

// URLRepositoryImpl là implementation của URLRepository
type URLRepositoryImpl struct {
	db *gorm.DB
}

// NewURLRepository tạo instance mới của URLRepository
func NewURLRepository(db *gorm.DB) *URLRepositoryImpl {
	return &URLRepositoryImpl{db: db}
}

// Create tạo mới một URL record
func (r *URLRepositoryImpl) Create(url *models.URL) error {
	return r.db.Create(url).Error
}

// FindByShortCode tìm URL theo short code
func (r *URLRepositoryImpl) FindByShortCode(shortCode string) (*models.URL, error) {
	var url models.URL
	err := r.db.Where("short_code = ?", shortCode).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// FindByOriginalURL tìm URL theo original URL
func (r *URLRepositoryImpl) FindByOriginalURL(originalURL string) (*models.URL, error) {
	var url models.URL
	err := r.db.Where("original_url = ?", originalURL).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// IncrementClickCount tăng số lượt click
func (r *URLRepositoryImpl) IncrementClickCount(shortCode string) error {
	return r.db.Model(&models.URL{}).
		Where("short_code = ?", shortCode).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error
}

// Delete xóa URL (soft delete)
func (r *URLRepositoryImpl) Delete(shortCode string) error {
	return r.db.Where("short_code = ?", shortCode).Delete(&models.URL{}).Error
}

// ExistsShortCode kiểm tra short code đã tồn tại chưa
func (r *URLRepositoryImpl) ExistsShortCode(shortCode string) (bool, error) {
	var count int64
	err := r.db.Model(&models.URL{}).Where("short_code = ?", shortCode).Count(&count).Error
	return count > 0, err
}

// GetStats lấy thống kê của URL
func (r *URLRepositoryImpl) GetStats(shortCode string) (*models.URLStatsResponse, error) {
	url, err := r.FindByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("URL not found: %w", err)
	}

	stats := &models.URLStatsResponse{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		TotalClicks: url.ClickCount,
		CreatedAt:   url.CreatedAt.Format(time.RFC3339),
	}

	return stats, nil
}
