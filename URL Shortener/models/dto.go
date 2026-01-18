package models

// CreateURLRequest là request body để tạo short URL
type CreateURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomCode  string `json:"custom_code,omitempty"` // Optional: Custom short code
	ExpiresIn   int    `json:"expires_in,omitempty"`  // Optional: Thời gian hết hạn (giờ)
}

// CreateURLResponse là response trả về khi tạo short URL thành công
type CreateURLResponse struct {
	ShortURL    string `json:"short_url"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// URLStatsResponse là response chứa thống kê của URL
type URLStatsResponse struct {
	ShortCode    string           `json:"short_code"`
	OriginalURL  string           `json:"original_url"`
	TotalClicks  int64            `json:"total_clicks"`
	CreatedAt    string           `json:"created_at"`
	ClicksByDate map[string]int64 `json:"clicks_by_date"`
	TopReferers  []RefererStats   `json:"top_referers"`
	TopCountries []CountryStats   `json:"top_countries"`
}

// RefererStats thống kê theo referer
type RefererStats struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

// CountryStats thống kê theo quốc gia
type CountryStats struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// ErrorResponse là response trả về khi có lỗi
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse là response chung cho các thao tác thành công
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
