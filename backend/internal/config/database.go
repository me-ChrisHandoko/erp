package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection using GORM
func NewDatabase(cfg DatabaseConfig, debug bool) (*gorm.DB, error) {
	// Determine log level
	logLevel := logger.Silent
	if debug {
		logLevel = logger.Info
	}

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger:      logger.Default.LogMode(logLevel),
		PrepareStmt: true, // Prepare statement caching
	}

	// Determine database driver from URL
	var db *gorm.DB
	var err error

	if isPostgresURL(cfg.URL) {
		db, err = gorm.Open(postgres.Open(cfg.URL), gormConfig)
	} else if isSQLiteURL(cfg.URL) {
		db, err = gorm.Open(sqlite.Open(cfg.URL), gormConfig)
	} else {
		return nil, fmt.Errorf("unsupported database URL format: %s", cfg.URL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")

	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Helper functions to detect database type

func isPostgresURL(url string) bool {
	return len(url) > 10 && (url[:10] == "postgres://" || url[:14] == "postgresql://")
}

func isSQLiteURL(url string) bool {
	return len(url) > 5 && url[:5] == "file:"
}
