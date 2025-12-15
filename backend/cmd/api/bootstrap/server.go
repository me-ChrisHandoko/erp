package bootstrap

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// InitServer initializes and configures the HTTP server
func InitServer(cfg config.ServerConfig) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	return router
}

// StartServer starts the HTTP server with graceful shutdown
func StartServer(router *gin.Engine, cfg config.ServerConfig) {
	log := logger.GetDefault()

	// Create HTTP server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Starting %s server on port %s (Environment: %s)",
			cfg.AppName, cfg.Port, cfg.Environment)
		log.Infof("Server running at http://localhost:%s", cfg.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped")
}

// SetupHealthCheck adds a health check endpoint
func SetupHealthCheck(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Service is running",
		})
	})
}

// SetupCORS adds CORS middleware
func SetupCORS(router *gin.Engine, cfg config.CORSConfig) {
	router.Use(func(c *gin.Context) {
		// Set CORS headers
		origin := c.Request.Header.Get("Origin")
		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// Helper functions

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func joinStrings(strings []string, separator string) string {
	result := ""
	for i, s := range strings {
		if i > 0 {
			result += separator
		}
		result += s
	}
	return result
}

// PrintRoutes prints all registered routes (development only)
func PrintRoutes(router *gin.Engine) {
	log := logger.GetDefault()

	routes := router.Routes()
	log.Infof("Registered %d routes:", len(routes))
	for _, route := range routes {
		log.Infof("  %-6s %s", route.Method, route.Path)
	}
}
