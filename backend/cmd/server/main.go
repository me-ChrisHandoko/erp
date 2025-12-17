package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/jobs"
	"backend/internal/router"
	"backend/pkg/jwt"
	"backend/pkg/security"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Starting ERP Backend Server (Environment: %s)", cfg.Server.Environment)

	// Initialize database
	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Initialize background job scheduler
	scheduler := jobs.NewScheduler(db, cfg)
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start job scheduler: %v", err)
	}
	if scheduler.IsRunning() {
		log.Println("Background job scheduler started successfully")
	}

	// Initialize Redis (optional, for caching and rate limiting)
	var redisClient *redis.Client
	if cfg.Cache.Enabled && cfg.Cache.RedisURL != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Cache.RedisURL,
			Password: "", // Add password if needed
			DB:       0,  // Use default DB
		})

		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v", err)
			log.Println("Continuing without Redis (caching and rate limiting will be disabled)")
			redisClient = nil
		} else {
			log.Println("Redis connected successfully")
		}
	}

	// Initialize services
	passwordHasher := security.NewPasswordHasher(cfg.Argon2)
	log.Println("Password hasher initialized (Argon2id)")

	tokenService, err := jwt.NewTokenService(cfg.JWT)
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}
	log.Printf("JWT service initialized (Algorithm: %s)", cfg.JWT.Algorithm)

	// Setup router
	r := router.SetupRouter(cfg, db, redisClient, passwordHasher, tokenService, scheduler)
	log.Println("Router configured successfully")

	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Println("Server started successfully")
	log.Println("Press Ctrl+C to shutdown")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown sequence:
	// 1. Stop HTTP server first (30s timeout)
	// 2. Stop job scheduler second (60s timeout for jobs to complete)
	// 3. Close database connection last

	// Step 1: Shutdown HTTP server with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped successfully")
	}

	// Step 2: Stop job scheduler and wait for running jobs to complete
	if scheduler.IsRunning() {
		log.Println("Stopping background job scheduler...")
		jobCtx := scheduler.Stop()

		// Wait for jobs to complete with 60 second timeout
		select {
		case <-jobCtx.Done():
			log.Println("All background jobs completed successfully")
		case <-time.After(60 * time.Second):
			log.Println("Warning: Background jobs did not complete within timeout")
		}
	}

	// Step 3: Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}

	// Close Redis connection
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}

	log.Println("Server shutdown complete")
}
