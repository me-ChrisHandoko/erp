package main

import (
	"backend/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("")
		{
			// Welcome endpoint
			public.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Welcome to ERP Distribusi Sembako API",
					"version": "1.0.0",
					"docs":    "/api/v1/docs",
				})
			})

			// TODO: Auth routes
			// auth := public.Group("/auth")
			// {
			// 	auth.POST("/login", authHandler.Login)
			// 	auth.POST("/register", authHandler.Register)
			// 	auth.POST("/refresh", authHandler.RefreshToken)
			// }
		}

		// Protected routes (authentication required)
		// protected := v1.Group("")
		// protected.Use(middleware.Auth(cfg.JWT))
		// protected.Use(middleware.TenantContext())
		// {
		// 	// TODO: Add protected routes here
		// 	// users := protected.Group("/users")
		// 	// products := protected.Group("/products")
		// 	// sales := protected.Group("/sales")
		// 	// etc.
		// }
	}

	// Swagger/OpenAPI documentation endpoint (development only)
	if cfg.IsDevelopment() {
		router.GET("/api/v1/docs/*any", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "API documentation endpoint",
				"note":    "Swagger UI will be integrated here",
			})
		})
	}
}
