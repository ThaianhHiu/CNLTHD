package handlers

import (
	"net/http"
	"strings"

	"url-shortener/models"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
)

// URLHandler xử lý các HTTP requests
type URLHandler struct {
	urlService *services.URLServiceImpl
}

// NewURLHandler tạo instance mới của URLHandler
func NewURLHandler(urlService *services.URLServiceImpl) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// CreateShortURL tạo short URL mới
// POST /api/shorten
func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req models.CreateURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate URL
	if !isValidURL(req.OriginalURL) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_url",
			Message: "URL must start with http:// or https://",
		})
		return
	}

	response, err := h.urlService.CreateShortURL(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectToOriginal redirect từ short URL sang original URL
// GET /:shortCode
func (h *URLHandler) RedirectToOriginal(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "short_code_required",
		})
		return
	}

	originalURL, err := h.urlService.GetOriginalURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	// Ghi nhận click bất đồng bộ (không block response)
	h.urlService.RecordClick(
		shortCode,
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.Referer(),
	)

	// Redirect với status 301 (Permanent) hoặc 302 (Temporary)
	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// GetURLStats lấy thống kê của URL
// GET /api/stats/:shortCode
func (h *URLHandler) GetURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "short_code_required",
		})
		return
	}

	stats, err := h.urlService.GetStats(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DeleteURL xóa short URL
// DELETE /api/urls/:shortCode
func (h *URLHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "short_code_required",
		})
		return
	}

	if err := h.urlService.DeleteURL(shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "URL deleted successfully",
	})
}

// HealthCheck kiểm tra trạng thái server
// GET /health
func (h *URLHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "url-shortener",
	})
}

// isValidURL kiểm tra URL có hợp lệ không
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
