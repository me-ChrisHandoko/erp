package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/internal/middleware"
	"backend/models"
)

func setupInvitationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Tenant{}, &models.UserTenant{}, &models.Company{})
	require.NoError(t, err)

	return db
}

func setupInvitationTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	// Create miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return redisClient, mr
}

func createInvitationTestUser(t *testing.T, db *gorm.DB, role models.UserRole) (string, string) {
	// Create company
	company := &models.Company{
		ID:       "company-1",
		Name:     "Test Company",
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)

	// Create tenant
	tenant := &models.Tenant{
		ID:        "tenant-1",
		CompanyID: company.ID,
		Status:    models.TenantStatusActive,
	}
	err = db.Create(tenant).Error
	require.NoError(t, err)

	// Create user
	user := &models.User{
		ID:           "user-1",
		Email:        "admin@test.com",
		Username:     "admin",
		PasswordHash: "hashed",
		FullName:     "Admin User",
		IsActive:     true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create user-tenant link with specified role
	userTenant := &models.UserTenant{
		ID:       user.ID + "-" + tenant.ID,
		UserID:   user.ID,
		TenantID: tenant.ID,
		Role:     role,
		IsActive: true,
	}
	err = db.Create(userTenant).Error
	require.NoError(t, err)

	return user.ID, tenant.ID
}

// Test Issue #8 Fix: Rate limiting applied to invitation endpoint (5 per minute)
func TestInvitationEndpoint_RateLimitEnforced(t *testing.T) {
	db := setupInvitationTestDB(t)
	redisClient, mr := setupInvitationTestRedis(t)
	defer mr.Close()

	_, _ = createInvitationTestUser(t, db, models.UserRoleOwner)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Simulate protected group with invitation route
	inviteGroup := router.Group("/api/v1/tenant/users/invite")
	inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5)) // 5 requests per minute
	{
		inviteGroup.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// Make 5 requests - should all succeed
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1") // Simulate same IP
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 6th request - should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "6th request should be rate limited")
}

// Test Issue #8 Fix: Rate limit resets after 1 minute
func TestInvitationEndpoint_RateLimitResetsAfterMinute(t *testing.T) {
	db := setupInvitationTestDB(t)
	redisClient, mr := setupInvitationTestRedis(t)
	defer mr.Close()

	_, _ = createInvitationTestUser(t, db, models.UserRoleOwner)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	inviteGroup := router.Group("/api/v1/tenant/users/invite")
	inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5))
	{
		inviteGroup.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// Make 5 requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Fast-forward time by 61 seconds in miniredis
	mr.FastForward(61 * time.Second)

	// Next request should succeed (rate limit reset)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Request after TTL expiry should succeed")
}

// Test Issue #8 Fix: Different IPs have separate rate limits
func TestInvitationEndpoint_SeparateRateLimitPerIP(t *testing.T) {
	db := setupInvitationTestDB(t)
	redisClient, mr := setupInvitationTestRedis(t)
	defer mr.Close()

	_, _ = createInvitationTestUser(t, db, models.UserRoleOwner)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	inviteGroup := router.Group("/api/v1/tenant/users/invite")
	inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5))
	{
		inviteGroup.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// IP 1: Make 5 requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 2: Should still be able to make requests
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Different IP should have separate rate limit")
}

// Test Issue #8 Fix: Rate limiting gracefully handles Redis unavailability
func TestInvitationEndpoint_GracefulRedisFailure(t *testing.T) {
	db := setupInvitationTestDB(t)

	_, _ = createInvitationTestUser(t, db, models.UserRoleOwner)

	// Setup Gin router with nil Redis client
	gin.SetMode(gin.TestMode)
	router := gin.New()

	inviteGroup := router.Group("/api/v1/tenant/users/invite")
	inviteGroup.Use(middleware.RateLimitMiddleware(nil, 5)) // nil Redis client
	{
		inviteGroup.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// Request should succeed even without Redis
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow request when Redis unavailable")
}

// Test Issue #8 Fix: RBAC middleware blocks non-OWNER/ADMIN from invitation endpoint
func TestInvitationEndpoint_RBACProtection(t *testing.T) {
	db := setupInvitationTestDB(t)
	redisClient, mr := setupInvitationTestRedis(t)
	defer mr.Close()

	createInvitationTestUser(t, db, models.UserRoleStaff) // STAFF role

	// Setup Gin router with RBAC middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set role in context (simulating JWT middleware)
	router.Use(func(c *gin.Context) {
		c.Set("role", "STAFF") // STAFF role
		c.Next()
	})

	tenantGroup := router.Group("/api/v1/tenant")
	tenantGroup.Use(middleware.RequireRoleMiddleware("OWNER", "ADMIN"))
	{
		inviteGroup := tenantGroup.Group("/users/invite")
		inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5))
		{
			inviteGroup.POST("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})
		}
	}

	// Simulate request from STAFF user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/users/invite", nil)

	router.ServeHTTP(w, req)

	// STAFF should be blocked by RBAC middleware before reaching rate limiter
	assert.Equal(t, http.StatusForbidden, w.Code, "STAFF role should be blocked from invitation endpoint")
}
