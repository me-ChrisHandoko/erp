package main

import (
	"time"

	"backend/cmd/api/bootstrap"
	"backend/internal/config"
)

// @title ERP Distribusi Sembako API
// @version 1.0
// @description Multi-Tenant ERP System API for Indonesian Food Distribution
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@erp.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 🔧 FIX: Set timezone to Asia/Jakarta (WIB) for Indonesian business
	// This ensures all timestamps are consistent between database and application
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if Asia/Jakarta not available
		loc = time.UTC
	}
	time.Local = loc

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := bootstrap.InitLogger(cfg.Logging)
	defer bootstrap.SyncLogger(log)

	log.Infof("Starting %s (Environment: %s)", cfg.Server.AppName, cfg.Server.Environment)

	// Initialize database
	db := bootstrap.InitDatabase(cfg.Database, cfg.Server.Debug)
	defer bootstrap.CloseDatabase(db)

	// Initialize HTTP server
	router := bootstrap.InitServer(cfg.Server)

	// Setup CORS
	bootstrap.SetupCORS(router, cfg.CORS)

	// Setup health check
	bootstrap.SetupHealthCheck(router)

	// Setup routes
	setupRoutes(router, cfg, db)

	// Print routes in development
	if cfg.IsDevelopment() {
		bootstrap.PrintRoutes(router)
	}

	// Start server with graceful shutdown
	bootstrap.StartServer(router, cfg.Server)
}
