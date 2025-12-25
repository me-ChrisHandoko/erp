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
	"backend/internal/service/company"
	"backend/internal/service/tenant"
	"backend/pkg/email"
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

	// ============================================================================
	// AUTH-ONLY PROTECTED ROUTES
	// Authentication required, but NO subscription validation
	// Users must ALWAYS be able to logout, view profile, and switch tenants
	// regardless of subscription status (security best practice)
	// ============================================================================
	authOnlyProtected := rg.Group("")
	authOnlyProtected.Use(middleware.RateLimitMiddleware(redisClient, 60)) // 60 requests per minute
	authOnlyProtected.Use(middleware.JWTAuthMiddleware(tokenService))
	authOnlyProtected.Use(middleware.TenantContextMiddleware(db))
	authOnlyProtected.Use(middleware.CSRFMiddleware()) // CSRF protection for state-changing operations
	{
		authGroup := authOnlyProtected.Group("/auth")
		{
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/change-password", authHandler.ChangePassword)
			authGroup.POST("/switch-tenant", authHandler.SwitchTenant)
			authGroup.GET("/me", authHandler.GetCurrentUser)
			authGroup.GET("/tenants", authHandler.GetUserTenants)
		}
	}

	// ============================================================================
	// BUSINESS PROTECTED ROUTES
	// Authentication + Active Subscription required
	// Business operations require active subscription to prevent billing bypass
	// ============================================================================
	businessProtected := rg.Group("")
	businessProtected.Use(middleware.RateLimitMiddleware(redisClient, 60)) // 60 requests per minute
	businessProtected.Use(middleware.JWTAuthMiddleware(tokenService))
	businessProtected.Use(middleware.TenantContextMiddleware(db))
	businessProtected.Use(middleware.ValidateSubscriptionMiddleware(db)) // ✅ Issue #2: Subscription validation
	businessProtected.Use(middleware.CSRFMiddleware())                   // ✨ CSRF protection for all state-changing operations
	businessProtected.Use(middleware.IdempotencyMiddleware(redisClient)) // ✅ Issue #9: Idempotency support for POST/PUT/PATCH
	{
		// Tenant management routes (OWNER/ADMIN only)
		// Reference: 01-TENANT-COMPANY-SETUP.md lines 732-929
		// Issue #8 Fix: Rate limiting applied to invitation endpoint
		emailService := email.NewEmailService(cfg)
		tenantService := tenant.NewTenantService(db, cfg, passwordHasher, emailService)
		tenantHandler := handler.NewTenantHandler(tenantService, cfg)

		tenantGroup := businessProtected.Group("/tenant")
		{
			// GET /api/v1/tenant - Get tenant details (all authenticated users)
			tenantGroup.GET("", tenantHandler.GetTenantDetails)

			// Tenant user management routes (OWNER/ADMIN only)
			adminGroup := tenantGroup.Group("")
			adminGroup.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
			{
				// GET /api/v1/tenant/users - List tenant users
				adminGroup.GET("/users", tenantHandler.ListTenantUsers)

				// User invitation endpoint with stricter rate limiting (5 invites per minute)
				// Issue #8: Prevent spam/abuse on invitation endpoint
				inviteGroup := adminGroup.Group("/users/invite")
				inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5)) // 5 invites per minute
				{
					// POST /api/v1/tenant/users/invite - Invite user
					inviteGroup.POST("", tenantHandler.InviteUser)
				}

				// PUT /api/v1/tenant/users/:userTenantId/role - Update user role
				adminGroup.PUT("/users/:userTenantId/role", tenantHandler.UpdateUserRole)

				// DELETE /api/v1/tenant/users/:userTenantId - Remove user from tenant
				adminGroup.DELETE("/users/:userTenantId", tenantHandler.RemoveUser)
			}
		}

		// Company profile routes
		// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Day 1-4 Tasks
		companyService := company.NewCompanyService(db)
		companyHandler := handler.NewCompanyHandler(companyService)

		companyGroup := businessProtected.Group("/company")
		{
			// Company profile endpoints (OWNER/ADMIN only for updates)
			companyGroup.GET("", companyHandler.GetCompanyProfile)
			companyGroup.PUT("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), companyHandler.UpdateCompanyProfile)

			// Bank account endpoints
			bankGroup := companyGroup.Group("/banks")
			{
				// GET endpoints - accessible to all authenticated users
				bankGroup.GET("", companyHandler.GetBankAccounts)
				bankGroup.GET("/:id", companyHandler.GetBankAccount)

				// POST/PUT/DELETE endpoints - OWNER/ADMIN only
				bankGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), companyHandler.AddBankAccount)
				bankGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), companyHandler.UpdateBankAccount)
				bankGroup.DELETE("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), companyHandler.DeleteBankAccount)
			}
		}

		// Example of role-based routes
		// adminOnly := businessProtected.Group("/admin")
		// adminOnly.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
		// {
		// 	adminOnly.GET("/users", adminHandler.ListUsers)
		// }
	}
}
