package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/service/company"
	"backend/internal/util"
	"backend/pkg/errors"
)

// CompanyContextMiddleware extracts and validates company context for PHASE 2
// Supports multi-company architecture (1 Tenant → N Companies)
//
// Company ID can be provided via:
// 1. Header: X-Company-ID
// 2. Query parameter: company_id
// 3. JWT claim: active_company_id (future implementation)
//
// This middleware:
// - Extracts company ID from request
// - Validates user has access to the company
// - Sets company context in Gin context for downstream handlers
func CompanyContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	companyService := company.NewMultiCompanyService(db)
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user ID from context (set by JWTAuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.NewAuthenticationError("User context not found"),
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errors.NewInternalError(fmt.Errorf("invalid user ID format")),
			})
			c.Abort()
			return
		}

		// Extract company ID from request
		// Priority: Header > Query Parameter
		companyID := c.GetHeader("X-Company-ID")
		if companyID == "" {
			companyID = c.Query("company_id")
		}

		// If no company ID provided, return error
		// FUTURE: Could default to user's last active company from JWT
		if companyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Company ID required (use X-Company-ID header or company_id query parameter)"),
			})
			c.Abort()
			return
		}

		// Verify user has access to this company
		accessInfo, err := companyService.CheckUserCompanyAccess(c.Request.Context(), userIDStr, companyID)
		if err != nil {
			fmt.Printf("❌ ERROR [CompanyContextMiddleware]: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err,
			})
			c.Abort()
			return
		}

		fmt.Printf("✅ DEBUG [CompanyContextMiddleware]: CheckUserCompanyAccess succeeded - HasAccess: %v, Role: %v, Tier: %d\n",
			accessInfo.HasAccess, accessInfo.Role, accessInfo.AccessTier)

		// Check if user has access
		if !accessInfo.HasAccess {
			fmt.Printf("❌ ERROR [CompanyContextMiddleware]: Access denied - HasAccess is false\n")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewAuthorizationError("Access denied: user does not have access to this company"),
			})
			c.Abort()
			return
		}

		fmt.Printf("🔧 DEBUG [CompanyContextMiddleware]: Setting company context...\n")

		// Set company context
		companyCtx.SetCompanyID(c, companyID)
		fmt.Printf("  ✅ SetCompanyID done\n")

		companyCtx.SetTenantID(c, accessInfo.TenantID)
		fmt.Printf("  ✅ SetTenantID done\n")

		companyCtx.SetUserID(c, userIDStr)
		fmt.Printf("  ✅ SetUserID done\n")

		fmt.Printf("  🔍 About to convert Role: %v (type: %T)\n", accessInfo.Role, accessInfo.Role)
		roleStr := string(accessInfo.Role)
		fmt.Printf("  ✅ Role converted to string: %s\n", roleStr)

		companyCtx.SetUserRole(c, roleStr)
		fmt.Printf("  ✅ SetUserRole done\n")

		companyCtx.SetCompanyAccess(c, accessInfo)
		fmt.Printf("  ✅ SetCompanyAccess done\n")

		// Store access info for downstream use
		c.Set("access_tier", accessInfo.AccessTier)
		fmt.Printf("✅ DEBUG [CompanyContextMiddleware]: All context set successfully, calling c.Next()\n")

		c.Next()
	}
}

// OptionalCompanyContextMiddleware is similar to CompanyContextMiddleware
// but doesn't require company ID (for endpoints that list all accessible companies)
func OptionalCompanyContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	companyService := company.NewMultiCompanyService(db)
	companyCtx := util.NewCompanyContext()

	return func(c *gin.Context) {
		// Get user ID from context (set by JWTAuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// No user context, continue without company context
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		// Extract company ID from request (optional)
		companyID := c.GetHeader("X-Company-ID")
		if companyID == "" {
			companyID = c.Query("company_id")
		}

		// If no company ID provided, skip company context setup
		if companyID == "" {
			companyCtx.SetUserID(c, userIDStr)
			c.Next()
			return
		}

		// Verify user has access to this company
		accessInfo, err := companyService.CheckUserCompanyAccess(c.Request.Context(), userIDStr, companyID)
		if err != nil {
			// Error checking access, continue without company context
			c.Next()
			return
		}

		// If user has access, set company context
		if accessInfo.HasAccess {
			companyCtx.SetCompanyID(c, companyID)
			companyCtx.SetTenantID(c, accessInfo.TenantID)
			companyCtx.SetUserID(c, userIDStr)
			companyCtx.SetUserRole(c, string(accessInfo.Role))
			companyCtx.SetCompanyAccess(c, accessInfo)
			c.Set("access_tier", accessInfo.AccessTier)
		}

		c.Next()
	}
}
