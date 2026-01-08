package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/jobs"
	"backend/internal/middleware"
	"backend/internal/service/audit"
	"backend/internal/service/auth"
	"backend/internal/service/company"
	"backend/internal/service/customer"
	"backend/internal/service/permission"
	"backend/internal/service/product"
	"backend/internal/service/supplier"
	"backend/internal/service/tenant"
	"backend/internal/service/warehouse"
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
	router.Use(middleware.SecurityHeadersMiddleware(cfg)) // OWASP Security Headers
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware(cfg))

	// Health check endpoints (no authentication required)
	healthHandler := handler.NewHealthHandler(db, redisClient, scheduler)
	router.GET("/health", healthHandler.Liveness)
	router.GET("/ready", healthHandler.Readiness)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// CSP violation reporting endpoint (no authentication required)
	// Browsers send CSP violation reports here for monitoring
	v1.POST("/csp-report", middleware.CSPReportHandler())
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
			authGroup.POST("/switch-company", authHandler.SwitchCompany) // PHASE 3: Company switching
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

		// ============================================================================
		// MULTI-COMPANY SERVICES (PHASE 5)
		// Created early so it can be used by multiple handlers
		// ============================================================================
		multiCompanyService := company.NewMultiCompanyService(db)
		permissionService := permission.NewPermissionService(db)

		// Company profile routes
		// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Day 1-4 Tasks
		// PHASE 5: Updated to use CompanyContextMiddleware for multi-company support
		companyService := company.NewCompanyService(db)
		companyHandler := handler.NewCompanyHandler(companyService, multiCompanyService)

		companyGroup := businessProtected.Group("/company")
		companyGroup.Use(middleware.CompanyContextMiddleware(db)) // PHASE 5: Add company context
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

			// ============================================================================
			// COMPANY USER MANAGEMENT ROUTES (Company-scoped)
			// Returns users with UserCompanyRole for the active company only
			// Respects X-Company-ID header via CompanyContextMiddleware
			// ============================================================================
			companyUserHandler := handler.NewCompanyUserHandler(tenantService, cfg)

			// GET /api/v1/company/users - List users in active company (all authenticated users)
			companyGroup.GET("/users", companyUserHandler.ListCompanyUsers)

			// Company user management routes (OWNER/ADMIN only)
			companyUserAdminGroup := companyGroup.Group("/users")
			companyUserAdminGroup.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
			{
				// POST /api/v1/company/users/invite - Invite user to company
				inviteGroup := companyUserAdminGroup.Group("/invite")
				inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5)) // 5 invites per minute
				{
					inviteGroup.POST("", companyUserHandler.InviteCompanyUser)
				}

				// PUT /api/v1/company/users/:userTenantId/role - Update user role in company
				companyUserAdminGroup.PUT("/:userTenantId/role", companyUserHandler.UpdateCompanyUserRole)

				// DELETE /api/v1/company/users/:userTenantId - Remove user from company
				companyUserAdminGroup.DELETE("/:userTenantId", companyUserHandler.RemoveCompanyUser)
			}
		}

		// ============================================================================
		// MULTI-COMPANY MANAGEMENT ROUTES (PHASE 3)
		// Reference: multi-company-architecture-analysis.md - PHASE 3
		// ============================================================================
		multiCompanyHandler := handler.NewMultiCompanyHandler(multiCompanyService, permissionService)

		companiesGroup := businessProtected.Group("/companies")
		{
			// List accessible companies (all authenticated users)
			// Uses optional company context middleware - no specific company required
			companiesGroup.GET("",
				middleware.OptionalCompanyContextMiddleware(db),
				multiCompanyHandler.ListCompanies,
			)

			// Create company (OWNER only)
			companiesGroup.POST("",
				middleware.RequireTier1Access(),
				multiCompanyHandler.CreateCompany,
			)

			// Get company details (requires access to that company)
			companiesGroup.GET("/:id",
				middleware.CompanyContextMiddleware(db),
				multiCompanyHandler.GetCompany,
			)

			// Update company (requires ADMIN or Tier 1)
			companiesGroup.PATCH("/:id",
				middleware.CompanyContextMiddleware(db),
				middleware.RequireCompanyAdmin(),
				multiCompanyHandler.UpdateCompany,
			)

			// Deactivate company (OWNER only)
			companiesGroup.DELETE("/:id",
				middleware.RequireTier1Access(),
				multiCompanyHandler.DeactivateCompany,
			)
		}

		// ============================================================================
		// PRODUCT MANAGEMENT ROUTES (PHASE 2 - Master Data Management)
		// Reference: 02-MASTER-DATA-MANAGEMENT.md Module 1: Product Management
		// ============================================================================
		auditService := audit.NewAuditService(db)
		productService := product.NewProductService(db, auditService)
		productHandler := handler.NewProductHandler(productService)

		productGroup := businessProtected.Group("/products")
		productGroup.Use(middleware.CompanyContextMiddleware(db))
		{
			// GET endpoints - all authenticated users can view
			productGroup.GET("", productHandler.ListProducts)
			productGroup.GET("/:id", productHandler.GetProduct)

			// POST/PUT/DELETE endpoints - OWNER/ADMIN only
			productGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.CreateProduct)
			productGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.UpdateProduct)
			productGroup.DELETE("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.DeleteProduct)

			// Product units management - OWNER/ADMIN only
			productGroup.POST("/:id/units", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.AddProductUnit)
			productGroup.PUT("/:id/units/:unitId", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.UpdateProductUnit)
			productGroup.DELETE("/:id/units/:unitId", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.DeleteProductUnit)

			// Product supplier management - OWNER/ADMIN only
			productGroup.POST("/:id/suppliers", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.AddProductSupplier)
			productGroup.PUT("/:id/suppliers/:supplierId", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.UpdateProductSupplier)
			productGroup.DELETE("/:id/suppliers/:supplierId", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), productHandler.DeleteProductSupplier)
		}

		// ============================================================================
		// CUSTOMER MANAGEMENT ROUTES (PHASE 2 - Master Data Management)
		// Reference: 02-MASTER-DATA-MANAGEMENT.md Module 2: Customer Management
		// ============================================================================
		customerService := customer.NewCustomerService(db, auditService)
		customerHandler := handler.NewCustomerHandler(customerService)

		customerGroup := businessProtected.Group("/customers")
		customerGroup.Use(middleware.CompanyContextMiddleware(db))
		{
			// GET endpoints - all authenticated users can view
			customerGroup.GET("", customerHandler.ListCustomers)
			customerGroup.GET("/:id", customerHandler.GetCustomer)

			// POST/PUT/DELETE endpoints - OWNER/ADMIN only
			customerGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), customerHandler.CreateCustomer)
			customerGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), customerHandler.UpdateCustomer)
			customerGroup.DELETE("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), customerHandler.DeleteCustomer)
		}

		// ============================================================================
		// SUPPLIER MANAGEMENT ROUTES (PHASE 2 - Master Data Management)
		// Reference: 02-MASTER-DATA-MANAGEMENT.md Module 3: Supplier Management
		// ============================================================================
		supplierService := supplier.NewSupplierService(db, auditService)
		supplierHandler := handler.NewSupplierHandler(supplierService)

		supplierGroup := businessProtected.Group("/suppliers")
		supplierGroup.Use(middleware.CompanyContextMiddleware(db))
		{
			// GET endpoints - all authenticated users can view
			supplierGroup.GET("", supplierHandler.ListSuppliers)
			supplierGroup.GET("/:id", supplierHandler.GetSupplier)

			// POST/PUT/DELETE endpoints - OWNER/ADMIN only
			supplierGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), supplierHandler.CreateSupplier)
			supplierGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), supplierHandler.UpdateSupplier)
			supplierGroup.DELETE("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), supplierHandler.DeleteSupplier)
		}

		// ============================================================================
		// WAREHOUSE MANAGEMENT ROUTES (PHASE 2 - Master Data Management)
		// Reference: 02-MASTER-DATA-MANAGEMENT.md Module 4: Warehouse Management
		// ============================================================================
		warehouseService := warehouse.NewWarehouseService(db, auditService)
		warehouseHandler := handler.NewWarehouseHandler(warehouseService)

		warehouseGroup := businessProtected.Group("/warehouses")
		warehouseGroup.Use(middleware.CompanyContextMiddleware(db))
		{
			// GET endpoints - all authenticated users can view
			warehouseGroup.GET("", warehouseHandler.ListWarehouses)
			warehouseGroup.GET("/:id", warehouseHandler.GetWarehouse)

			// POST/PUT/DELETE endpoints - OWNER/ADMIN only
			warehouseGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), warehouseHandler.CreateWarehouse)
			warehouseGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), warehouseHandler.UpdateWarehouse)
			warehouseGroup.DELETE("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), warehouseHandler.DeleteWarehouse)
		}

		// Warehouse Stock routes
		warehouseStockGroup := businessProtected.Group("/warehouse-stocks")
		warehouseStockGroup.Use(middleware.CompanyContextMiddleware(db))
		{
			// GET endpoints - all authenticated users can view
			warehouseStockGroup.GET("", warehouseHandler.ListWarehouseStocks)

			// PUT endpoint - OWNER/ADMIN only (update stock settings, not quantity)
			warehouseStockGroup.PUT("/:id", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), warehouseHandler.UpdateWarehouseStock)
		}

		// Example of role-based routes
		// adminOnly := businessProtected.Group("/admin")
		// adminOnly.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
		// {
		// 	adminOnly.GET("/users", adminHandler.ListUsers)
		// }
	}
}
