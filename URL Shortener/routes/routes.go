package routes

import (
	"url-shortener/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes cấu hình tất cả routes cho ứng dụng
func SetupRoutes(router *gin.Engine, urlHandler *handlers.URLHandler) {
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// Health check
	router.GET("/health", urlHandler.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Tạo short URL
		api.POST("/shorten", urlHandler.CreateShortURL)

		// Lấy thống kê
		api.GET("/stats/:shortCode", urlHandler.GetURLStats)

		// Xóa URL
		api.DELETE("/urls/:shortCode", urlHandler.DeleteURL)
	}

	// Redirect route (phải đặt cuối cùng vì là catch-all)
	router.GET("/:shortCode", urlHandler.RedirectToOriginal)

	// Serve static files (frontend)
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")
}

// CORSMiddleware xử lý Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
