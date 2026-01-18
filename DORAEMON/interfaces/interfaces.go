package interfaces

import (
	"url-shortener/models"
)

// URLRepository định nghĩa các phương thức làm việc với database
type URLRepository interface {
	// Create tạo mới một URL record
	Create(url *models.URL) error

	// FindByShortCode tìm URL theo short code
	FindByShortCode(shortCode string) (*models.URL, error)

	// FindByOriginalURL tìm URL theo original URL
	FindByOriginalURL(originalURL string) (*models.URL, error)

	// IncrementClickCount tăng số lượt click
	IncrementClickCount(shortCode string) error

	// Delete xóa URL
	Delete(shortCode string) error

	// ExistsShortCode kiểm tra short code đã tồn tại chưa
	ExistsShortCode(shortCode string) (bool, error)

	// GetStats lấy thống kê của URL
	GetStats(shortCode string) (*models.URLStatsResponse, error)
}

// CacheRepository định nghĩa các phương thức làm việc với cache
type CacheRepository interface {
	// Set lưu URL vào cache
	Set(shortCode string, originalURL string) error

	// Get lấy original URL từ cache
	Get(shortCode string) (string, error)

	// Delete xóa URL khỏi cache
	Delete(shortCode string) error

	// Exists kiểm tra URL có trong cache không
	Exists(shortCode string) (bool, error)
}

// AnalyticsRepository định nghĩa các phương thức cho analytics
type AnalyticsRepository interface {
	// SaveClickEvent lưu sự kiện click
	SaveClickEvent(event *models.ClickEvent) error

	// GetClicksByDate lấy số lượt click theo ngày
	GetClicksByDate(shortCode string, days int) (map[string]int64, error)

	// GetTopReferers lấy top referers
	GetTopReferers(shortCode string, limit int) ([]models.RefererStats, error)

	// GetTopCountries lấy top countries
	GetTopCountries(shortCode string, limit int) ([]models.CountryStats, error)
}

// ShortCodeGenerator định nghĩa interface cho việc sinh short code
type ShortCodeGenerator interface {
	// Generate tạo short code mới
	Generate() string

	// GenerateWithLength tạo short code với độ dài cụ thể
	GenerateWithLength(length int) string

	// IsValid kiểm tra short code có hợp lệ không
	IsValid(code string) bool
}

// URLService định nghĩa các phương thức business logic
type URLService interface {
	// CreateShortURL tạo short URL mới
	CreateShortURL(req *models.CreateURLRequest) (*models.CreateURLResponse, error)

	// GetOriginalURL lấy original URL từ short code
	GetOriginalURL(shortCode string) (string, error)

	// GetStats lấy thống kê của URL
	GetStats(shortCode string) (*models.URLStatsResponse, error)

	// DeleteURL xóa URL
	DeleteURL(shortCode string) error

	// RecordClick ghi nhận click event (bất đồng bộ)
	RecordClick(shortCode string, ipAddress, userAgent, referer string)
}
