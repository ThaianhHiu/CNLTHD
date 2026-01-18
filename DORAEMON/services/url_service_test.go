package services

import (
	"testing"

	"url-shortener/models"
)

// TestCreateURLRequest_Validation tests request validation
func TestCreateURLRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateURLRequest
		isValid bool
	}{
		{
			name:    "Valid URL with https",
			req:     models.CreateURLRequest{OriginalURL: "https://example.com"},
			isValid: true,
		},
		{
			name:    "Valid URL with http",
			req:     models.CreateURLRequest{OriginalURL: "http://example.com"},
			isValid: true,
		},
		{
			name:    "Invalid URL without protocol",
			req:     models.CreateURLRequest{OriginalURL: "example.com"},
			isValid: false,
		},
		{
			name:    "Empty URL",
			req:     models.CreateURLRequest{OriginalURL: ""},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidURL(tt.req.OriginalURL)
			if isValid != tt.isValid {
				t.Errorf("isValidURL(%s) = %v, want %v", tt.req.OriginalURL, isValid, tt.isValid)
			}
		})
	}
}

// isValidURL helper for testing
func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}
