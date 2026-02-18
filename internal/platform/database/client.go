package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mairuu/mp-api/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewClient establishes a connection to the database and returns a GORM instance
func NewClient(cfg *config.DatabaseConfig, log *slog.Logger) (*gorm.DB, error) {
	logLevel := logger.Warn
	switch cfg.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		TranslateError: true,
		Logger: logger.NewSlogLogger(log, logger.Config{
			LogLevel: logLevel,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// set connection pool settings
	// todo: make configurable via cfg
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
