package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	dbmigration "backend/db"
	"backend/internal/config"
)

// InitDatabase initializes database connection and registers callbacks
// Reference: BACKEND-IMPLEMENTATION.md lines 177-223, 395-415
func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	// Connect to PostgreSQL database
	dialector := postgres.Open(cfg.Database.URL)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(cfg)),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database interface
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run auto-migration for all models
	// This will create tables if they don't exist
	if err := dbmigration.AutoMigrate(db); err != nil {
		log.Printf("Warning: Auto-migration failed: %v", err)
	} else {
		log.Println("Auto-migration completed successfully")
	}

	// Register tenant isolation callbacks with configuration
	// CRITICAL: This enables automatic tenant filtering (Primary defense layer)
	// Reference: BACKEND-IMPLEMENTATION.md - Enhanced Dual-Layer Isolation
	RegisterTenantCallbacks(db, &cfg.TenantIsolation)

	log.Printf("Database connected successfully (Tenant Isolation: Strict=%v, Warnings=%v, Bypass=%v)\n",
		cfg.TenantIsolation.StrictMode,
		cfg.TenantIsolation.LogWarnings,
		cfg.TenantIsolation.AllowBypass)

	return db, nil
}

// getLogLevel returns GORM logger level based on environment
func getLogLevel(cfg *config.Config) logger.LogLevel {
	if cfg.IsProduction() {
		return logger.Error
	}
	if cfg.Server.Environment == "development" {
		return logger.Info
	}
	return logger.Warn
}

