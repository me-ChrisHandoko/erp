package main

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/service/auth"
	"backend/pkg/jwt"
	"backend/pkg/security"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB) {
	// Initialize dependencies for auth
	passwordHasher := security.NewPasswordHasher(cfg.Argon2)
	tokenService, err := jwt.NewTokenService(cfg.JWT)
	if err != nil {
		panic("Failed to create token service: " + err.Error())
	}
	authService := auth.NewAuthService(db, cfg, passwordHasher, tokenService)
	authHandler := handler.NewAuthHandler(authService, cfg)
	adminHandler := handler.NewAdminHandler(authService, cfg)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Welcome endpoint
		v1.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome to ERP Distribusi Sembako API",
				"version": "1.0.0",
				"docs":    "/api/v1/docs",
			})
		})

		// Auth routes (public - no authentication required)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
			authGroup.POST("/verify-email", authHandler.VerifyEmail)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(tokenService))
		protected.Use(middleware.TenantContextMiddleware(db))
		protected.Use(middleware.CSRFMiddleware())
		{
			// Protected auth endpoints
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.POST("/change-password", authHandler.ChangePassword)
				authProtected.POST("/switch-tenant", authHandler.SwitchTenant)
				authProtected.GET("/me", authHandler.GetCurrentUser)
				authProtected.GET("/tenants", authHandler.GetUserTenants)
			}

			// System admin routes (requires isSystemAdmin = true)
			adminGroup := protected.Group("/admin")
			adminGroup.Use(middleware.RequireSystemAdminMiddleware(db))
			{
				// Account management endpoints
				adminGroup.POST("/unlock-account", adminHandler.UnlockAccount)
				adminGroup.GET("/lock-status/:email", adminHandler.GetLockStatus)
			}
		}
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
