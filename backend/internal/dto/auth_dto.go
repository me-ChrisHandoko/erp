package dto

// LoginRequest represents user login request
// Reference: BACKEND-IMPLEMENTATION.md lines 979-1013
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255" validate:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=1,max=72" validate:"required,min=1,max=72"`

	// Optional device tracking
	DeviceInfo string `json:"deviceInfo,omitempty" validate:"omitempty,max=500"`
	IPAddress  string `json:"-"` // Set from request context
	UserAgent  string `json:"-"` // Set from request context
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required,min=32" validate:"required,min=32"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required,min=32" validate:"required,min=32"`
}

// AuthResponse represents authentication response
// Returns access token, refresh token, and user info
type AuthResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"` // Access token expiry in seconds
	TokenType    string    `json:"tokenType"` // "Bearer"
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information in auth response
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone,omitempty"`
	IsActive bool   `json:"isActive"`

	// Current tenant info (if user has access to multiple tenants)
	CurrentTenant *TenantInfo `json:"currentTenant,omitempty"`
}

// TenantInfo represents tenant information
type TenantInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"` // ACTIVE, TRIAL, SUSPENDED, etc.
	Role   string `json:"role"`   // User's role in this tenant
}

// EmailVerificationRequest represents email verification request
type EmailVerificationRequest struct {
	Token string `json:"token" binding:"required,min=32,max=255" validate:"required,min=32,max=255"`
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email,max=255" validate:"required,email,max=255"`
}

// PasswordResetConfirmRequest represents password reset confirmation
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required,min=32,max=255" validate:"required,min=32,max=255"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=72" validate:"required,min=8,max=72,password_strength"`
}

// ChangePasswordRequest represents password change request (authenticated user)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=1,max=72" validate:"required,min=1,max=72"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72" validate:"required,min=8,max=72,password_strength,nefield=CurrentPassword"`
}

// SwitchTenantRequest represents tenant switching request
type SwitchTenantRequest struct {
	TenantID string `json:"tenantId" binding:"required" validate:"required"`
}

// SwitchTenantResponse represents tenant switching response
type SwitchTenantResponse struct {
	AccessToken  string      `json:"accessToken"`
	ExpiresIn    int         `json:"expiresIn"` // Access token expiry in seconds
	TokenType    string      `json:"tokenType"` // "Bearer"
	ActiveTenant *TenantInfo `json:"activeTenant"`
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" validate:"required"`
}

// VerifyEmailResponse represents email verification response
type VerifyEmailResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// GetUserTenantsResponse represents user's accessible tenants response
type GetUserTenantsResponse struct {
	Tenants []TenantInfo `json:"tenants"`
}

// CurrentUserResponse represents current authenticated user response
type CurrentUserResponse struct {
	User         *UserInfo    `json:"user"`
	ActiveTenant *TenantInfo  `json:"activeTenant"`
	Tenants      []TenantInfo `json:"tenants"` // All accessible tenants
}
