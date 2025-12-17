package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/service/auth"
	"backend/pkg/errors"
	customValidator "backend/pkg/validator"
)

// AuthHandler handles HTTP requests for authentication
// Reference: BACKEND-IMPLEMENTATION.md lines 921-1160
type AuthHandler struct {
	authService *auth.AuthService
	cfg         *config.Config
	validator   *customValidator.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
		validator:   customValidator.New(),
	}
}

// Login handles user authentication
// POST /api/v1/auth/login
// Reference: BACKEND-IMPLEMENTATION.md lines 979-1013
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info from request
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	// Call auth service
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, response.RefreshToken)

	// Set CSRF token for protecting future requests
	// This must be done after successful login to establish session security
	h.setCSRFToken(c)

	// Don't send refresh token in response body
	response.RefreshToken = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
// Reference: BACKEND-IMPLEMENTATION.md lines 1014-1053
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("Refresh token not found"))
		return
	}

	req := &dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Call auth service
	response, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set new refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, response.RefreshToken)

	// Don't send refresh token in response body
	response.RefreshToken = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
// Reference: BACKEND-IMPLEMENTATION.md lines 1054-1071
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Even if cookie not found, clear it anyway
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}

	req := &dto.LogoutRequest{
		RefreshToken: refreshToken,
	}

	// Call auth service
	if err := h.authService.Logout(c.Request.Context(), req); err != nil {
		h.handleError(c, err)
		return
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ForgotPassword handles password reset request
// POST /api/v1/auth/forgot-password
// Reference: PHASE2-MVP-ANALYSIS.md lines 180-220
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call auth service
	if err := h.authService.ForgotPassword(c.Request.Context(), &req, ipAddress, userAgent); err != nil {
		h.handleError(c, err)
		return
	}

	// Always return success (even if email doesn't exist)
	// This prevents email enumeration attacks
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset confirmation
// POST /api/v1/auth/reset-password
// Reference: PHASE2-MVP-ANALYSIS.md lines 180-220
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call auth service
	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		h.handleError(c, err)
		return
	}

	// Clear CSRF cookie on password reset (force re-login)
	h.clearCSRFCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password has been reset successfully. Please login with your new password.",
	})
}

// VerifyEmail handles email verification
// POST /api/v1/auth/verify-email
// Reference: PHASE3-MVP-ANALYSIS.md
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call auth service
	user, err := h.authService.VerifyEmail(req.Token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := &dto.VerifyEmailResponse{
		Message: "Email verified successfully. You can now log in.",
		Email:   user.Email,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// SwitchTenant handles tenant switching for multi-tenant users
// POST /api/v1/auth/switch-tenant
// Requires authentication
// Reference: PHASE3-MVP-ANALYSIS.md
func (h *AuthHandler) SwitchTenant(c *gin.Context) {
	// Get user ID from context (set by JWTAuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	var req dto.SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call auth service
	accessToken, tenant, err := h.authService.SwitchTenant(userID.(string), req.TenantID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Build tenant info
	tenantInfo := &dto.TenantInfo{
		ID:     tenant.ID,
		Name:   tenant.Name,
		Status: string(tenant.Status),
		// Role will be set by the service
	}

	// Get user's role in this tenant
	var userTenant auth.UserTenant
	if err := h.authService.DB().Where("user_id = ? AND tenant_id = ?", userID.(string), tenant.ID).
		First(&userTenant).Error; err == nil {
		tenantInfo.Role = string(userTenant.Role)
	}

	response := &dto.SwitchTenantResponse{
		AccessToken:  accessToken,
		ExpiresIn:    int(h.cfg.JWT.Expiry.Seconds()),
		TokenType:    "Bearer",
		ActiveTenant: tenantInfo,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetUserTenants returns all tenants accessible to the current user
// GET /api/v1/auth/tenants
// Requires authentication
// Reference: PHASE3-MVP-ANALYSIS.md
func (h *AuthHandler) GetUserTenants(c *gin.Context) {
	// Get user ID from context (set by JWTAuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	// Call auth service
	userTenants, tenants, err := h.authService.GetUserTenants(userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Build tenant info list
	tenantInfos := make([]dto.TenantInfo, 0, len(tenants))

	// Create a map for quick tenant lookup
	tenantMap := make(map[string]*auth.Tenant)
	for i := range tenants {
		tenantMap[tenants[i].ID] = &tenants[i]
	}

	// Build response with role information
	for _, ut := range userTenants {
		if tenant, ok := tenantMap[ut.TenantID]; ok {
			tenantInfos = append(tenantInfos, dto.TenantInfo{
				ID:     tenant.ID,
				Name:   tenant.Name,
				Status: string(tenant.Status),
				Role:   string(ut.Role),
			})
		}
	}

	response := &dto.GetUserTenantsResponse{
		Tenants: tenantInfos,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetCurrentUser returns current authenticated user's information
// GET /api/v1/auth/me
// Requires authentication
// Reference: PHASE3-MVP-ANALYSIS.md
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID and tenant ID from context (set by JWTAuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	tenantID, _ := c.Get("tenant_id")

	// Get user details
	var user auth.User
	if err := h.authService.DB().Where("id = ?", userID.(string)).First(&user).Error; err != nil {
		h.handleError(c, errors.NewInternalError(err))
		return
	}

	// Build user info
	userInfo := &dto.UserInfo{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Phone:    user.Phone,
		IsActive: user.IsActive,
	}

	// Get all user's tenants
	userTenants, tenants, err := h.authService.GetUserTenants(userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Build tenant info list
	tenantInfos := make([]dto.TenantInfo, 0, len(tenants))
	var activeTenant *dto.TenantInfo

	// Create tenant map for lookup
	tenantMap := make(map[string]*auth.Tenant)
	for i := range tenants {
		tenantMap[tenants[i].ID] = &tenants[i]
	}

	// Build tenant list with roles
	for _, ut := range userTenants {
		if tenant, ok := tenantMap[ut.TenantID]; ok {
			info := dto.TenantInfo{
				ID:     tenant.ID,
				Name:   tenant.Name,
				Status: string(tenant.Status),
				Role:   string(ut.Role),
			}
			tenantInfos = append(tenantInfos, info)

			// Set active tenant if matches current tenant ID
			if tenantID != nil && tenant.ID == tenantID.(string) {
				activeTenant = &info
			}
		}
	}

	response := &dto.CurrentUserResponse{
		User:         userInfo,
		ActiveTenant: activeTenant,
		Tenants:      tenantInfos,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ChangePassword handles password change for authenticated users
// POST /api/v1/auth/change-password
// Requires authentication
// Reference: PHASE3-MVP-ANALYSIS.md
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by JWTAuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call auth service
	if err := h.authService.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword); err != nil {
		h.handleError(c, err)
		return
	}

	// Clear refresh token cookie (force re-login)
	h.clearRefreshTokenCookie(c)

	// Clear CSRF cookie
	h.clearCSRFCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully. Please log in again with your new password.",
	})
}

// Helper functions

// setRefreshTokenCookie sets the refresh token as an httpOnly cookie
// Reference: BACKEND-IMPLEMENTATION.md lines 826-850 (Cookie Security)
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := int(h.cfg.JWT.RefreshExpiry.Seconds())

	c.SetCookie(
		"refresh_token",          // name
		token,                    // value
		maxAge,                   // maxAge in seconds
		"/",                      // path
		h.cfg.Cookie.Domain,      // domain
		h.cfg.Cookie.Secure,      // secure (HTTPS only in production)
		true,                     // httpOnly (prevent XSS)
	)

	// Set SameSite attribute for CSRF protection
	c.SetSameSite(http.SameSiteStrictMode)
}

// clearRefreshTokenCookie clears the refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1, // maxAge -1 means delete
		"/",
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		true,
	)
}

// clearCSRFCookie clears the CSRF token cookie
func (h *AuthHandler) clearCSRFCookie(c *gin.Context) {
	c.SetCookie(
		"csrf_token",
		"",
		-1, // maxAge -1 means delete
		"/",
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		false,
	)
}

// setCSRFToken generates and sets a CSRF token cookie
// This is called after successful login to protect subsequent requests
func (h *AuthHandler) setCSRFToken(c *gin.Context) {
	// Note: We need to import the middleware package for this
	// For now, we'll generate the token inline
	// TODO: Move this to a shared utility if needed
	token, err := generateCSRFToken()
	if err != nil {
		// Log error but don't fail the request
		// CSRF token will be missing but auth still works
		return
	}

	c.SetCookie(
		"csrf_token",
		token,
		24*60*60, // 24 hours
		"/",
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		false, // NOT httpOnly - frontend needs to read it
	)

	// Set SameSite=Strict
	c.SetSameSite(http.SameSiteStrictMode)
}

// generateCSRFToken creates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// handleValidationError formats and returns validation errors
func (h *AuthHandler) handleValidationError(c *gin.Context, err error) {
	// Check if it's a validator error
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		// Format each validation error
		formattedErrors := make([]errors.ValidationError, 0, len(validationErrs))

		for _, fieldErr := range validationErrs {
			formattedErrors = append(formattedErrors, errors.ValidationError{
				Field:   getJSONFieldName(fieldErr),
				Message: formatValidationMessage(fieldErr),
			})
		}

		appErr := errors.NewValidationError(formattedErrors)
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

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

// handleError handles errors and returns appropriate HTTP responses
func (h *AuthHandler) handleError(c *gin.Context, err error) {
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

// getJSONFieldName extracts JSON field name from validation error
func getJSONFieldName(fe validator.FieldError) string {
	// Convert field name to camelCase (JSON format)
	field := fe.Field()
	if len(field) > 0 {
		// Simple camelCase conversion
		return string(field[0]-'A'+'a') + field[1:]
	}
	return field
}

// formatValidationMessage creates user-friendly validation messages
func formatValidationMessage(fe validator.FieldError) string {
	field := getJSONFieldName(fe)

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		if fe.Type().String() == "string" {
			return field + " must be at least " + fe.Param() + " characters long"
		}
		return field + " must be at least " + fe.Param()
	case "max":
		if fe.Type().String() == "string" {
			return field + " must not exceed " + fe.Param() + " characters"
		}
		return field + " must not exceed " + fe.Param()
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "password_strength":
		return field + " must contain at least one uppercase letter, one lowercase letter, and one digit"
	case "phone_number":
		return field + " must be a valid Indonesian phone number (08xxxxxxxxxx or +628xxxxxxxxxx)"
	case "nefield":
		return field + " must be different from " + fe.Param()
	default:
		return field + " is invalid"
	}
}
