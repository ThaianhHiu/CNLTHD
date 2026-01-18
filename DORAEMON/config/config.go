package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config chứa tất cả cấu hình của ứng dụng
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	App      AppConfig
}

type ServerConfig struct {
	Port    string
	BaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type AppConfig struct {
	ShortCodeLength int
}

// LoadConfig đọc cấu hình từ file .env
func LoadConfig() (*Config, error) {
	// Load .env file nếu tồn tại
	godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	shortCodeLength, _ := strconv.Atoi(getEnv("SHORT_CODE_LENGTH", "6"))

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "url_shortener"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		App: AppConfig{
			ShortCodeLength: shortCodeLength,
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
