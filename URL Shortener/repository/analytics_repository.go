package repository

import (
	"time"

	"url-shortener/models"

	"gorm.io/gorm"
)

// AnalyticsRepositoryImpl là implementation của AnalyticsRepository
type AnalyticsRepositoryImpl struct {
	db *gorm.DB
}

// NewAnalyticsRepository tạo instance mới của AnalyticsRepository
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepositoryImpl {
	return &AnalyticsRepositoryImpl{db: db}
}

// SaveClickEvent lưu sự kiện click
func (r *AnalyticsRepositoryImpl) SaveClickEvent(event *models.ClickEvent) error {
	return r.db.Create(event).Error
}

// GetClicksByDate lấy số lượt click theo ngày
func (r *AnalyticsRepositoryImpl) GetClicksByDate(shortCode string, days int) (map[string]int64, error) {
	result := make(map[string]int64)

	startDate := time.Now().AddDate(0, 0, -days)

	type DateCount struct {
		Date  string
		Count int64
	}

	var counts []DateCount

	err := r.db.Model(&models.ClickEvent{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("short_code = ? AND created_at >= ?", shortCode, startDate).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	for _, c := range counts {
		result[c.Date] = c.Count
	}

	return result, nil
}

// GetTopReferers lấy top referers
func (r *AnalyticsRepositoryImpl) GetTopReferers(shortCode string, limit int) ([]models.RefererStats, error) {
	var stats []models.RefererStats

	err := r.db.Model(&models.ClickEvent{}).
		Select("referer, COUNT(*) as count").
		Where("short_code = ? AND referer != ''", shortCode).
		Group("referer").
		Order("count DESC").
		Limit(limit).
		Scan(&stats).Error

	return stats, err
}

// GetTopCountries lấy top countries
func (r *AnalyticsRepositoryImpl) GetTopCountries(shortCode string, limit int) ([]models.CountryStats, error) {
	var stats []models.CountryStats

	err := r.db.Model(&models.ClickEvent{}).
		Select("country, COUNT(*) as count").
		Where("short_code = ? AND country != ''", shortCode).
		Group("country").
		Order("count DESC").
		Limit(limit).
		Scan(&stats).Error

	return stats, err
}
