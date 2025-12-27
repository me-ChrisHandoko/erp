package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/service/permission"
	"backend/internal/util"
	"backend/pkg/errors"
)

// RequirePermission creates a middleware that checks if user has required permission
// for the active company context
//
// Usage:
//   router.POST("/products", middleware.RequirePermission(db, permission.PermissionCreateData), handler)
//
// This middleware requires CompanyContextMiddleware to run first
func RequirePermission(db *gorm.DB, requiredPermission permission.Permission) gin.HandlerFunc {
	permissionService := permission.NewPermissionService(db)
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := companyCtx.GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("User context not found"),
			})
			c.Abort()
			return
		}

		// Get company ID from context
		companyID, exists := companyCtx.GetCompanyID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Company context not found"),
			})
			c.Abort()
			return
		}

		// Check if user has the required permission
		hasPermission, err := permissionService.CheckPermission(c.Request.Context(), userID, companyID, requiredPermission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err,
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Access denied: insufficient permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates a middleware that checks if user has ANY of the required permissions
// Useful for endpoints that can be accessed by multiple permission levels
//
// Usage:
//   router.GET("/reports", middleware.RequireAnyPermission(db,
//     permission.PermissionViewReports,
//     permission.PermissionViewData,
//   ), handler)
func RequireAnyPermission(db *gorm.DB, permissions ...permission.Permission) gin.HandlerFunc {
	permissionService := permission.NewPermissionService(db)
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := companyCtx.GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("User context not found"),
			})
			c.Abort()
			return
		}

		// Get company ID from context
		companyID, exists := companyCtx.GetCompanyID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Company context not found"),
			})
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		for _, perm := range permissions {
			hasPermission, err := permissionService.CheckPermission(c.Request.Context(), userID, companyID, perm)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err,
				})
				c.Abort()
				return
			}

			if hasPermission {
				// User has at least one required permission
				c.Next()
				return
			}
		}

		// User doesn't have any of the required permissions
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errors.NewAuthorizationError("Access denied: insufficient permissions"),
		})
		c.Abort()
	}
}

// RequireAllPermissions creates a middleware that checks if user has ALL of the required permissions
// Useful for sensitive endpoints that require multiple permission levels
//
// Usage:
//   router.DELETE("/company/:id", middleware.RequireAllPermissions(db,
//     permission.PermissionDeleteData,
//     permission.PermissionManageSettings,
//   ), handler)
func RequireAllPermissions(db *gorm.DB, permissions ...permission.Permission) gin.HandlerFunc {
	permissionService := permission.NewPermissionService(db)
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := companyCtx.GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("User context not found"),
			})
			c.Abort()
			return
		}

		// Get company ID from context
		companyID, exists := companyCtx.GetCompanyID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Company context not found"),
			})
			c.Abort()
			return
		}

		// Check if user has all required permissions
		for _, perm := range permissions {
			hasPermission, err := permissionService.CheckPermission(c.Request.Context(), userID, companyID, perm)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err,
				})
				c.Abort()
				return
			}

			if !hasPermission {
				// User doesn't have all required permissions
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   errors.NewAuthorizationError("Access denied: insufficient permissions"),
				})
				c.Abort()
				return
			}
		}

		// User has all required permissions
		c.Next()
	}
}

// RequireTier1Access creates a middleware that requires Tier 1 access (OWNER or TENANT_ADMIN)
// Useful for tenant-wide operations like creating new companies
//
// Usage:
//   router.POST("/companies", middleware.RequireTier1Access(), handler)
func RequireTier1Access() gin.HandlerFunc {
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get access tier from context (set by CompanyContextMiddleware)
		accessTier, exists := c.Get("access_tier")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Access tier not found in context"),
			})
			c.Abort()
			return
		}

		tier, ok := accessTier.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errors.NewInternalError(fmt.Errorf("invalid access tier format")),
			})
			c.Abort()
			return
		}

		// Check if user has Tier 1 access
		if tier != 1 {
			userRole, _ := companyCtx.GetUserRole(c)
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Access denied: requires OWNER or TENANT_ADMIN role (current role: " + userRole + ")"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCompanyAdmin creates a middleware that requires ADMIN role for the company
// Useful for company-specific administrative operations
//
// Usage:
//   router.POST("/companies/:id/users", middleware.RequireCompanyAdmin(), handler)
func RequireCompanyAdmin() gin.HandlerFunc {
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := companyCtx.GetUserRole(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("User role not found in context"),
			})
			c.Abort()
			return
		}

		// Check if user is ADMIN, TENANT_ADMIN, or OWNER
		if userRole != "ADMIN" && userRole != "TENANT_ADMIN" && userRole != "OWNER" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Access denied: requires ADMIN role (current role: " + userRole + ")"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
