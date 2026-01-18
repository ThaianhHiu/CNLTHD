package database

import (
	"fmt"
	"log"

	"url-shortener/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDB là wrapper cho GORM database connection
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresDB tạo kết nối mới đến PostgreSQL
func NewPostgresDB(cfg config.DatabaseConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully")

	return &PostgresDB{DB: db}, nil
}

// AutoMigrate thực hiện auto migration cho các models
func (p *PostgresDB) AutoMigrate(models ...interface{}) error {
	return p.DB.AutoMigrate(models...)
}

// Close đóng kết nối database
func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
