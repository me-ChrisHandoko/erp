package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/jobs"
	"backend/internal/middleware"
	"backend/internal/service/auth"
	"backend/pkg/jwt"
	"backend/pkg/security"
)

// SetupRouter configures all routes and middleware
// Reference: BACKEND-IMPLEMENTATION.md lines 1525-1606 (Route Organization)
func SetupRouter(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	passwordHasher *security.PasswordHasher,
	tokenService *jwt.TokenService,
	scheduler *jobs.Scheduler,
) *gin.Engine {
	// Create Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware(cfg))

	// Health check endpoints (no authentication required)
	healthHandler := handler.NewHealthHandler(db, redisClient, scheduler)
	router.GET("/health", healthHandler.Liveness)
	router.GET("/ready", healthHandler.Readiness)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no authentication required)
		setupAuthRoutes(v1, cfg, db, redisClient, passwordHasher, tokenService)

		// Protected routes (authentication required)
		setupProtectedRoutes(v1, cfg, db, redisClient, tokenService)
	}

	return router
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(
	rg *gin.RouterGroup,
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	passwordHasher *security.PasswordHasher,
	tokenService *jwt.TokenService,
) {
	// Create auth service
	authService := auth.NewAuthService(db, cfg, passwordHasher, tokenService)

	// Create auth handler
	authHandler := handler.NewAuthHandler(authService, cfg)

	// Auth routes with stricter rate limiting
	authGroup := rg.Group("/auth")
	authGroup.Use(middleware.AuthRateLimitMiddleware(redisClient, 10)) // 10 requests per minute
	{
		// Public auth endpoints (no CSRF required)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)

		// Password reset endpoints (Phase 2)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)

		// Email verification endpoint (Phase 3)
		authGroup.POST("/verify-email", authHandler.VerifyEmail)
	}
}

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(
	rg *gin.RouterGroup,
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	tokenService *jwt.TokenService,
) {
	// Create auth service and handler for protected auth routes
	passwordHasher := security.NewPasswordHasher(cfg.Argon2)
	authService := auth.NewAuthService(db, cfg, passwordHasher, tokenService)
	authHandler := handler.NewAuthHandler(authService, cfg)

	// Protected routes group with JWT authentication, tenant context, and CSRF protection
	protected := rg.Group("")
	protected.Use(middleware.RateLimitMiddleware(redisClient, 60)) // 60 requests per minute
	protected.Use(middleware.JWTAuthMiddleware(tokenService))
	protected.Use(middleware.TenantContextMiddleware(db))
	protected.Use(middleware.CSRFMiddleware()) // ✨ CSRF protection for all state-changing operations
	{
		// Protected auth endpoints (require authentication + CSRF)
		authProtected := protected.Group("/auth")
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.POST("/change-password", authHandler.ChangePassword)
			authProtected.POST("/switch-tenant", authHandler.SwitchTenant)
			authProtected.GET("/me", authHandler.GetCurrentUser)
			authProtected.GET("/tenants", authHandler.GetUserTenants)
		}

		// Tenant management routes (OWNER/ADMIN only)
		// TODO: Implement in next phase
		// tenantGroup := protected.Group("/tenants")
		// tenantGroup.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
		// {
		// 	tenantGroup.GET("", tenantHandler.List)
		// 	tenantGroup.GET("/:id", tenantHandler.Get)
		// 	tenantGroup.PUT("/:id", tenantHandler.Update)
		// }

		// Company profile routes
		// TODO: Implement in next phase
		// companyGroup := protected.Group("/company")
		// {
		// 	companyGroup.GET("", companyHandler.Get)
		// 	companyGroup.PUT("", companyHandler.Update)
		// }

		// Example of role-based routes
		// adminOnly := protected.Group("/admin")
		// adminOnly.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
		// {
		// 	adminOnly.GET("/users", adminHandler.ListUsers)
		// }
	}
}
