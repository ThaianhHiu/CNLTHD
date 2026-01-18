package models

import (
	"time"

	"gorm.io/gorm"
)

// URL là model chính cho việc lưu trữ link rút gọn
type URL struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ShortCode   string         `gorm:"uniqueIndex;size:10;not null" json:"short_code"`
	OriginalURL string         `gorm:"type:text;not null" json:"original_url"`
	ClickCount  int64          `gorm:"default:0" json:"click_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

// TableName định nghĩa tên bảng trong database
func (URL) TableName() string {
	return "urls"
}

// IsExpired kiểm tra link đã hết hạn chưa
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// ClickEvent là model để lưu thông tin click analytics
type ClickEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	URLID     uint      `gorm:"index;not null" json:"url_id"`
	ShortCode string    `gorm:"index;size:10;not null" json:"short_code"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Referer   string    `gorm:"type:text" json:"referer"`
	Country   string    `gorm:"size:100" json:"country"`
	City      string    `gorm:"size:100" json:"city"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// TableName định nghĩa tên bảng trong database
func (ClickEvent) TableName() string {
	return "click_events"
}
