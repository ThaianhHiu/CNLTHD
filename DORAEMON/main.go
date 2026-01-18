package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"url-shortener/config"
	"url-shortener/database"
	"url-shortener/handlers"
	"url-shortener/models"
	"url-shortener/repository"
	"url-shortener/routes"
	"url-shortener/services"
	"url-shortener/workers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Starting URL Shortener Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("✅ Configuration loaded")

	// Connect to PostgreSQL
	postgresDB, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Auto migrate database schemas
	if err := postgresDB.AutoMigrate(&models.URL{}, &models.ClickEvent{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated")

	// Connect to Redis
	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	urlRepo := repository.NewURLRepository(postgresDB.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)
	analyticsRepo := repository.NewAnalyticsRepository(postgresDB.DB)

	// Initialize click analytics worker (Goroutines & Channels)
	// 4 workers, buffer size 10000 events
	clickWorker := workers.NewClickAnalyticsWorker(urlRepo, analyticsRepo, 4, 10000)
	clickWorker.Start()
	defer clickWorker.Stop()

	// Initialize services
	urlService := services.NewURLService(urlRepo, cacheRepo, analyticsRepo, cfg, clickWorker)

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(urlService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router, urlHandler)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("🛑 Shutting down server...")
		clickWorker.Stop()
		os.Exit(0)
	}()

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("🌐 Server running on http://localhost%s", addr)
	log.Printf("📝 API Endpoints:")
	log.Printf("   POST /api/shorten     - Create short URL")
	log.Printf("   GET  /:shortCode      - Redirect to original URL")
	log.Printf("   GET  /api/stats/:code - Get URL statistics")
	log.Printf("   DELETE /api/urls/:code - Delete URL")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
