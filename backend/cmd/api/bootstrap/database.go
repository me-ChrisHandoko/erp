package bootstrap

import (
	"backend/internal/config"
	"backend/pkg/logger"

	"gorm.io/gorm"
)

// InitDatabase initializes the database connection
func InitDatabase(cfg config.DatabaseConfig, debug bool) *gorm.DB {
	log := logger.GetDefault()

	log.Info("Initializing database connection...")

	db, err := config.NewDatabase(cfg, debug)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Infof("Database connected successfully (MaxOpenConns: %d, MaxIdleConns: %d)",
		cfg.MaxOpenConns, cfg.MaxIdleConns)

	return db
}

// CloseDatabase closes the database connection gracefully
func CloseDatabase(db *gorm.DB) {
	log := logger.GetDefault()

	log.Info("Closing database connection...")

	if err := config.CloseDatabase(db); err != nil {
		log.Errorf("Error closing database: %v", err)
	} else {
		log.Info("Database connection closed")
	}
}
