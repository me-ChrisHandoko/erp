package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// ValidateSubscriptionMiddleware validates tenant subscription status
// This middleware prevents billing bypass and ensures only active tenants can access the system
// Must be used after TenantContextMiddleware
//
// Validation Logic:
// - TRIAL: Check if trial period has expired
// - ACTIVE: Allow access
// - PAST_DUE: Check if grace period has expired
// - SUSPENDED: Block access
// - EXPIRED: Block access
// - CANCELLED: Block access
//
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #2 (lines 58-113)
func ValidateSubscriptionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant_id from context (set by JWTAuthMiddleware)
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   pkgerrors.NewAuthorizationError("Tenant context required"),
			})
			c.Abort()
			return
		}

		// Load tenant with subscription data
		var tenant models.Tenant
		err := db.Preload("Subscription").
			Where("id = ?", tenantID.(string)).
			First(&tenant).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   pkgerrors.NewNotFoundError("Tenant not found"),
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   pkgerrors.NewInternalError(err),
				})
			}
			c.Abort()
			return
		}

		// Validate tenant status
		now := time.Now()

		switch tenant.Status {
		case models.TenantStatusExpired:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":    "SUBSCRIPTION_EXPIRED",
					"message": "Subscription expired. Please renew to continue.",
				},
			})
			c.Abort()
			return

		case models.TenantStatusCancelled:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":    "SUBSCRIPTION_CANCELLED",
					"message": "Subscription has been cancelled.",
				},
			})
			c.Abort()
			return

		case models.TenantStatusTrial:
			// Check if trial period has expired
			if tenant.TrialEndsAt != nil && now.After(*tenant.TrialEndsAt) {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    "TRIAL_EXPIRED",
						"message": "Trial period expired. Please subscribe to continue.",
						"trialEndedAt": tenant.TrialEndsAt.Format(time.RFC3339),
					},
				})
				c.Abort()
				return
			}
			// Trial still valid, allow access

		case models.TenantStatusSuspended:
			// For suspended status, check subscription for PAST_DUE with grace period
			if tenant.Subscription != nil && tenant.Subscription.Status == models.SubscriptionStatusPastDue {
				// Check if grace period has expired
				if tenant.Subscription.GracePeriodEnds != nil && now.After(*tenant.Subscription.GracePeriodEnds) {
					c.JSON(http.StatusForbidden, gin.H{
						"success": false,
						"error": map[string]interface{}{
							"code":               "PAYMENT_OVERDUE",
							"message":            "Payment overdue. Please update your payment method.",
							"gracePeriodEnded":   tenant.Subscription.GracePeriodEnds.Format(time.RFC3339),
						},
					})
					c.Abort()
					return
				}
				// Within grace period, allow access with warning
				c.Header("X-Subscription-Warning", "Payment overdue - grace period active")
			} else {
				// Regular suspension (not payment related)
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    "ACCOUNT_SUSPENDED",
						"message": "Account suspended. Please contact support.",
					},
				})
				c.Abort()
				return
			}

		case models.TenantStatusActive:
			// All good, check subscription expiry
			if tenant.Subscription != nil &&
				tenant.Subscription.Status == models.SubscriptionStatusExpired {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    "SUBSCRIPTION_EXPIRED",
						"message": "Subscription expired. Please renew to continue.",
					},
				})
				c.Abort()
				return
			}
			// Active and valid, allow access

		default:
			// Unknown status, block for safety
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":    "INVALID_TENANT_STATUS",
					"message": "Invalid tenant status. Please contact support.",
				},
			})
			c.Abort()
			return
		}

		// Store tenant in context for potential use in handlers
		c.Set("tenant", tenant)

		c.Next()
	}
}
