package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/database"
	"backend/pkg/jwt"
)

func setupAuthTest(t *testing.T) (*jwt.TokenService, *gorm.DB) {
	// Create JWT token service
	cfg := config.JWTConfig{
		Secret:    "test-secret-key-for-auth-middleware-testing",
		Algorithm: "HS256",
		Expiry:    15 * time.Minute,
	}
	tokenService, err := jwt.NewTokenService(cfg)
	require.NoError(t, err)

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Register tenant isolation callbacks
	tenantCfg := &config.TenantIsolationConfig{
		StrictMode:  true,
		LogWarnings: false,
		AllowBypass: false,
	}
	database.RegisterTenantCallbacks(db, tenantCfg)

	return tokenService, db
}

func createTestToken(t *testing.T, tokenService *jwt.TokenService, userID, email, tenantID, role string) string {
	token, err := tokenService.GenerateAccessToken(userID, email, tenantID, role)
	require.NoError(t, err)
	return token
}

// Test 1: JWTAuthMiddleware - Valid Token
func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	// Create valid token
	token := createTestToken(t, tokenService, "user-1", "test@example.com", "tenant-1", "ADMIN")

	// Setup Gin
	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		tenantID, _ := c.Get("tenant_id")
		role, _ := c.Get("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"email":     email,
			"tenant_id": tenantID,
			"role":      role,
		})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-1")
	assert.Contains(t, w.Body.String(), "test@example.com")
	assert.Contains(t, w.Body.String(), "tenant-1")
	assert.Contains(t, w.Body.String(), "ADMIN")
}

// Test 2: JWTAuthMiddleware - Missing Authorization Header
func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

// Test 3: JWTAuthMiddleware - Invalid Authorization Format
func TestJWTAuthMiddleware_InvalidFormat(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "invalid-token"},
		{"Wrong scheme", "Basic invalid-token"},
		{"Only Bearer", "Bearer"},
		{"Empty token", "Bearer "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tc.header)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// Test 4: JWTAuthMiddleware - Invalid Token
func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// Test 5: JWTAuthMiddleware - Expired Token
func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create token service with very short expiry
	cfg := config.JWTConfig{
		Secret:    "test-secret",
		Algorithm: "HS256",
		Expiry:    1 * time.Millisecond, // 1ms expiry
	}
	tokenService, err := jwt.NewTokenService(cfg)
	require.NoError(t, err)

	// Create token
	token, err := tokenService.GenerateAccessToken("user-1", "test@example.com", "tenant-1", "USER")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// Test 6: TenantContextMiddleware - Valid Tenant Context
func TestTenantContextMiddleware_ValidContext(t *testing.T) {
	_, db := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Simulate JWTAuthMiddleware setting tenant_id
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", "tenant-1")
		c.Next()
	})
	router.Use(TenantContextMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		// Verify tenant DB is set
		tenantDB, exists := c.Get("db")
		assert.True(t, exists)
		assert.NotNil(t, tenantDB)

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 7: TenantContextMiddleware - Missing Tenant Context
func TestTenantContextMiddleware_MissingContext(t *testing.T) {
	_, db := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// No tenant_id set (simulating missing JWT middleware)
	router.Use(TenantContextMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Tenant context not found")
}

// Test 8: OptionalAuthMiddleware - With Valid Token
func TestOptionalAuthMiddleware_WithToken(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	token := createTestToken(t, tokenService, "user-1", "test@example.com", "tenant-1", "USER")

	router := gin.New()
	router.Use(OptionalAuthMiddleware(tokenService))
	router.GET("/public", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"authenticated": exists,
			"user_id":       userID,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":true`)
	assert.Contains(t, w.Body.String(), "user-1")
}

// Test 9: OptionalAuthMiddleware - Without Token
func TestOptionalAuthMiddleware_WithoutToken(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(OptionalAuthMiddleware(tokenService))
	router.GET("/public", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	// No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":false`)
}

// Test 10: OptionalAuthMiddleware - Invalid Token (Should Continue)
func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	tokenService, _ := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(OptionalAuthMiddleware(tokenService))
	router.GET("/public", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	router.ServeHTTP(w, req)

	// Should continue without authentication
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":false`)
}

// Test 11: RequireRoleMiddleware - Allowed Role
func TestRequireRoleMiddleware_AllowedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Simulate JWT middleware setting role
	router.Use(func(c *gin.Context) {
		c.Set("role", "ADMIN")
		c.Next()
	})
	router.Use(RequireRoleMiddleware("ADMIN", "OWNER"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 12: RequireRoleMiddleware - Forbidden Role
func TestRequireRoleMiddleware_ForbiddenRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "USER") // User role
		c.Next()
	})
	router.Use(RequireRoleMiddleware("ADMIN", "OWNER")) // Requires ADMIN or OWNER
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

// Test 13: RequireRoleMiddleware - Missing Role Context
func TestRequireRoleMiddleware_MissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// No role set (simulating missing JWT middleware)
	router.Use(RequireRoleMiddleware("ADMIN"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "User role not found in context")
}

// Test 14: Full Auth Chain - JWT + Tenant + Role
func TestFullAuthChain_Success(t *testing.T) {
	tokenService, db := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	token := createTestToken(t, tokenService, "user-1", "admin@example.com", "tenant-1", "ADMIN")

	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.Use(TenantContextMiddleware(db))
	router.Use(RequireRoleMiddleware("ADMIN", "OWNER"))
	router.GET("/admin/resource", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		tenantDB, _ := c.Get("db")

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"user_id":   userID,
			"has_db":    tenantDB != nil,
			"tenant_id": "tenant-1",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), "user-1")
}

// Test 15: Full Auth Chain - Insufficient Role
func TestFullAuthChain_InsufficientRole(t *testing.T) {
	tokenService, db := setupAuthTest(t)
	gin.SetMode(gin.TestMode)

	// Create token with USER role
	token := createTestToken(t, tokenService, "user-1", "user@example.com", "tenant-1", "USER")

	router := gin.New()
	router.Use(JWTAuthMiddleware(tokenService))
	router.Use(TenantContextMiddleware(db))
	router.Use(RequireRoleMiddleware("ADMIN")) // Requires ADMIN
	router.GET("/admin/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}
