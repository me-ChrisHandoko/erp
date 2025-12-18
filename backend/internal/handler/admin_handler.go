package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/service/auth"
	"backend/pkg/errors"
	customValidator "backend/pkg/validator"
)

// AdminHandler handles HTTP requests for system administration
// Only accessible to users with isSystemAdmin = true
type AdminHandler struct {
	authService *auth.AuthService
	cfg         *config.Config
	validator   *customValidator.Validator
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(authService *auth.AuthService, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		authService: authService,
		cfg:         cfg,
		validator:   customValidator.New(),
	}
}

// UnlockAccount removes failed login attempts for a user, effectively unlocking their account
// POST /api/v1/admin/unlock-account
// Requires: System Admin privileges
func (h *AdminHandler) UnlockAccount(c *gin.Context) {
	var req dto.UnlockAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get admin email for audit trail
	adminEmail, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   errors.NewAuthenticationError("Admin email not found in context"),
		})
		return
	}

	// Create service request with admin email for audit trail
	serviceReq := &auth.UnlockAccountRequest{
		Email:      req.Email,
		Reason:     req.Reason,
		AdminEmail: adminEmail.(string),
	}

	// Call auth service to unlock account
	err := h.authService.UnlockAccount(c.Request.Context(), serviceReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Get lock status after unlock to confirm
	status, err := h.authService.GetAccountLockStatus(c.Request.Context(), req.Email)
	if err != nil {
		// Non-critical error, just log it
		// Still return success response since unlock succeeded
	}

	// Prepare response
	response := dto.UnlockAccountResponse{
		Email:       req.Email,
		Reason:      req.Reason,
		UnlockedBy:  adminEmail.(string),
		UnlockedAt:  time.Now().Format(time.RFC3339),
	}

	// If we successfully got the status, include attempts cleared
	if status != nil {
		response.AttemptsCleared = status.FailedAttempts
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account unlocked successfully",
		"data":    response,
	})
}

// GetLockStatus retrieves the current lock status for a user account
// GET /api/v1/admin/lock-status/:email
// Requires: System Admin privileges
func (h *AdminHandler) GetLockStatus(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewValidationError([]errors.ValidationError{
				{Field: "email", Message: "Email parameter is required"},
			}),
		})
		return
	}

	// Call auth service to get lock status
	status, err := h.authService.GetAccountLockStatus(c.Request.Context(), email)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to DTO
	response := dto.GetLockStatusResponse{
		Email:             status.Email,
		IsLocked:          status.IsLocked,
		Tier:              status.Tier,
		RetryAfterSeconds: status.RetryAfterSeconds,
		FailedAttempts:    status.FailedAttempts,
		LockedUntil:       status.LockedUntil,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// handleValidationError handles validation errors
func (h *AdminHandler) handleValidationError(c *gin.Context, err error) {
	// Not a validation error, return generic error
	validationErrors := []errors.ValidationError{
		{
			Field:   "request",
			Message: err.Error(),
		},
	}
	appErr := errors.NewValidationError(validationErrors)
	c.JSON(appErr.StatusCode, gin.H{
		"success": false,
		"error":   appErr,
	})
}

// handleError handles service errors
func (h *AdminHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	// Unknown error - return internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
