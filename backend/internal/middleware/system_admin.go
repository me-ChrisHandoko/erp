package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/service/auth"
	"backend/pkg/errors"
)

// RequireSystemAdminMiddleware checks if the authenticated user is a system administrator
// Must be used after JWTAuthMiddleware
// System admins have cross-tenant capabilities and can perform administrative operations
func RequireSystemAdminMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTAuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("User context not found"),
			})
			c.Abort()
			return
		}

		// Query user to check isSystemAdmin flag
		var user auth.User
		result := db.Where("id = ? AND is_active = ?", userID, true).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   errors.NewAuthorizationError("User not found or inactive"),
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errors.NewInternalError(result.Error),
				})
			}
			c.Abort()
			return
		}

		// Check if user is system admin
		if !user.IsSystemAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("System administrator privileges required"),
			})
			c.Abort()
			return
		}

		// Store system admin flag in context for audit logging
		c.Set("is_system_admin", true)

		c.Next()
	}
}
