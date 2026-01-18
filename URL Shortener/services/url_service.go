package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"url-shortener/config"
	"url-shortener/generator"
	"url-shortener/models"
	"url-shortener/repository"
	"url-shortener/workers"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// URLServiceImpl là implementation của URLService
type URLServiceImpl struct {
	urlRepo       *repository.URLRepositoryImpl
	cacheRepo     *repository.CacheRepositoryImpl
	analyticsRepo *repository.AnalyticsRepositoryImpl
	generator     *generator.ShortCodeGeneratorImpl
	config        *config.Config
	clickWorker   *workers.ClickAnalyticsWorker
}

// NewURLService tạo instance mới của URLService
func NewURLService(
	urlRepo *repository.URLRepositoryImpl,
	cacheRepo *repository.CacheRepositoryImpl,
	analyticsRepo *repository.AnalyticsRepositoryImpl,
	cfg *config.Config,
	clickWorker *workers.ClickAnalyticsWorker,
) *URLServiceImpl {
	return &URLServiceImpl{
		urlRepo:       urlRepo,
		cacheRepo:     cacheRepo,
		analyticsRepo: analyticsRepo,
		generator:     generator.NewShortCodeGenerator(cfg.App.ShortCodeLength),
		config:        cfg,
		clickWorker:   clickWorker,
	}
}

// CreateShortURL tạo short URL mới
func (s *URLServiceImpl) CreateShortURL(req *models.CreateURLRequest) (*models.CreateURLResponse, error) {
	// Kiểm tra URL đã tồn tại chưa (tránh duplicate)
	existingURL, err := s.urlRepo.FindByOriginalURL(req.OriginalURL)
	if err == nil && existingURL != nil {
		// URL đã tồn tại, trả về link cũ
		return &models.CreateURLResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.config.Server.BaseURL, existingURL.ShortCode),
			ShortCode:   existingURL.ShortCode,
			OriginalURL: existingURL.OriginalURL,
		}, nil
	}

	var shortCode string

	// Sử dụng custom code nếu được cung cấp
	if req.CustomCode != "" {
		// Validate custom code
		if !s.generator.IsValid(req.CustomCode) {
			return nil, errors.New("invalid custom code format")
		}

		// Kiểm tra custom code đã tồn tại chưa
		exists, err := s.urlRepo.ExistsShortCode(req.CustomCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom code: %w", err)
		}
		if exists {
			return nil, errors.New("custom code already exists")
		}

		shortCode = req.CustomCode
	} else {
		// Generate short code unique
		shortCode, err = s.generateUniqueShortCode()
		if err != nil {
			return nil, err
		}
	}

	// Tạo URL record
	url := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		ClickCount:  0,
	}

	// Set expiration nếu được cung cấp
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		url.ExpiresAt = &expiresAt
	}

	// Lưu vào database
	if err := s.urlRepo.Create(url); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache URL để redirect nhanh
	if err := s.cacheRepo.Set(shortCode, req.OriginalURL); err != nil {
		// Log lỗi nhưng không fail request
		log.Printf("Warning: failed to cache URL: %v", err)
	}

	response := &models.CreateURLResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.config.Server.BaseURL, shortCode),
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
	}

	if url.ExpiresAt != nil {
		response.ExpiresAt = url.ExpiresAt.Format(time.RFC3339)
	}

	return response, nil
}

// generateUniqueShortCode tạo short code unique
func (s *URLServiceImpl) generateUniqueShortCode() (string, error) {
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		shortCode := s.generator.Generate()

		exists, err := s.urlRepo.ExistsShortCode(shortCode)
		if err != nil {
			return "", fmt.Errorf("failed to check short code: %w", err)
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", errors.New("failed to generate unique short code after max attempts")
}

// GetOriginalURL lấy original URL từ short code
// Ưu tiên lấy từ cache để tối ưu hiệu năng
func (s *URLServiceImpl) GetOriginalURL(shortCode string) (string, error) {
	// 1. Thử lấy từ cache trước (Redis - cực nhanh)
	originalURL, err := s.cacheRepo.Get(shortCode)
	if err == nil && originalURL != "" {
		log.Printf("Cache HIT for short code: %s", shortCode)
		return originalURL, nil
	}

	// Cache miss hoặc lỗi Redis
	if err != nil && err != redis.Nil {
		log.Printf("Cache error: %v", err)
	}

	log.Printf("Cache MISS for short code: %s", shortCode)

	// 2. Fallback: Lấy từ database
	url, err := s.urlRepo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("short URL not found")
		}
		return "", fmt.Errorf("failed to find URL: %w", err)
	}

	// 3. Kiểm tra expiration
	if url.IsExpired() {
		return "", errors.New("short URL has expired")
	}

	// 4. Cache lại để lần sau nhanh hơn
	if err := s.cacheRepo.Set(shortCode, url.OriginalURL); err != nil {
		log.Printf("Warning: failed to cache URL: %v", err)
	}

	return url.OriginalURL, nil
}

// GetStats lấy thống kê của URL
func (s *URLServiceImpl) GetStats(shortCode string) (*models.URLStatsResponse, error) {
	// Lấy thông tin cơ bản
	stats, err := s.urlRepo.GetStats(shortCode)
	if err != nil {
		return nil, err
	}

	// Lấy clicks theo ngày (7 ngày gần nhất)
	clicksByDate, err := s.analyticsRepo.GetClicksByDate(shortCode, 7)
	if err != nil {
		log.Printf("Warning: failed to get clicks by date: %v", err)
	} else {
		stats.ClicksByDate = clicksByDate
	}

	// Lấy top referers
	topReferers, err := s.analyticsRepo.GetTopReferers(shortCode, 5)
	if err != nil {
		log.Printf("Warning: failed to get top referers: %v", err)
	} else {
		stats.TopReferers = topReferers
	}

	// Lấy top countries
	topCountries, err := s.analyticsRepo.GetTopCountries(shortCode, 5)
	if err != nil {
		log.Printf("Warning: failed to get top countries: %v", err)
	} else {
		stats.TopCountries = topCountries
	}

	return stats, nil
}

// DeleteURL xóa URL
func (s *URLServiceImpl) DeleteURL(shortCode string) error {
	// Xóa từ database
	if err := s.urlRepo.Delete(shortCode); err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	// Xóa từ cache
	if err := s.cacheRepo.Delete(shortCode); err != nil {
		log.Printf("Warning: failed to delete URL from cache: %v", err)
	}

	return nil
}

// RecordClick ghi nhận click event BẤT ĐỒNG BỘ
// Sử dụng Goroutine và Channel để không block request chính
func (s *URLServiceImpl) RecordClick(shortCode string, ipAddress, userAgent, referer string) {
	// Tạo click event
	event := &models.ClickEvent{
		ShortCode: shortCode,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
		CreatedAt: time.Now(),
	}

	// Gửi event vào worker channel (non-blocking)
	// Worker sẽ xử lý async
	s.clickWorker.Enqueue(event)
}
