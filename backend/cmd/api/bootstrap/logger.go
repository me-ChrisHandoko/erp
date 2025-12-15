package bootstrap

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"log"
)

// InitLogger initializes the application logger
func InitLogger(cfg config.LoggingConfig) *logger.Logger {
	loggerInstance, err := logger.New(logger.Config{
		Level:  cfg.Level,
		Format: cfg.Format,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Set as default logger
	if err := logger.InitDefault(logger.Config{
		Level:  cfg.Level,
		Format: cfg.Format,
	}); err != nil {
		log.Fatalf("Failed to set default logger: %v", err)
	}

	loggerInstance.Infof("Logger initialized (Level: %s, Format: %s)", cfg.Level, cfg.Format)

	return loggerInstance
}

// SyncLogger flushes any buffered log entries
func SyncLogger(log *logger.Logger) {
	if err := log.Sync(); err != nil {
		// Ignore sync errors on stdout/stderr
		_ = err
	}
}
