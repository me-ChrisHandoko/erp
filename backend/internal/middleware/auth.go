package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/database"
	"backend/pkg/errors"
	"backend/pkg/jwt"
)

// JWTAuthMiddleware validates JWT access token and sets user context
// Reference: BACKEND-IMPLEMENTATION.md lines 1161-1203
func JWTAuthMiddleware(tokenService *jwt.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("Authorization header required"),
			})
			c.Abort()
			return
		}

		// Check Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("Invalid authorization header format"),
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("Invalid or expired token"),
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// TenantContextMiddleware sets tenant context for database operations
// CRITICAL: This enables GORM callbacks for automatic tenant isolation
// Reference: BACKEND-IMPLEMENTATION.md lines 395-415, RLS-REMOVAL-IMPLEMENTATION-SUMMARY.md
func TenantContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant_id from context (set by JWTAuthMiddleware)
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Tenant context not found"),
			})
			c.Abort()
			return
		}

		// Set tenant context in GORM session
		// This activates the dual-layer tenant isolation system:
		// 1. GORM Callbacks (automatic WHERE tenant_id = ?) - PRIMARY DEFENSE
		// 2. GORM Scopes (manual filtering when needed) - SECONDARY DEFENSE
		tenantDB := database.SetTenantSession(db, tenantID.(string))

		// Store tenant-scoped DB in context
		c.Set("db", tenantDB)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT token if present, but doesn't require it
// Useful for endpoints that have different behavior for authenticated vs anonymous users
func OptionalAuthMiddleware(tokenService *jwt.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRoleMiddleware checks if user has required role
// Must be used after JWTAuthMiddleware
// Reference: BACKEND-IMPLEMENTATION.md lines 1204-1234 (RBAC)
func RequireRoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("User role not found in context"),
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		roleStr := userRole.(string)
		allowed := false
		for _, role := range allowedRoles {
			if roleStr == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Insufficient permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
