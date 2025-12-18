package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/security"
)

// AuthService handles authentication operations
// Reference: BACKEND-IMPLEMENTATION.md lines 921-1160
type AuthService struct {
	db             *gorm.DB
	cfg            *config.Config
	passwordHasher *security.PasswordHasher
	tokenService   *jwt.TokenService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	db *gorm.DB,
	cfg *config.Config,
	passwordHasher *security.PasswordHasher,
	tokenService *jwt.TokenService,
) *AuthService {
	return &AuthService{
		db:             db,
		cfg:            cfg,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
	}
}

// DB returns the database connection for handler access
func (s *AuthService) DB() *gorm.DB {
	return s.db
}

// Login authenticates a user and returns JWT tokens
// Reference: BACKEND-IMPLEMENTATION.md lines 979-1013
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Check brute force protection with 4-tier exponential backoff
	locked, tier, retryAfterSeconds, attemptsCount, err := s.checkLoginAttempts(ctx, req.Email, req.IPAddress)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, errors.NewAccountLockedError(tier, retryAfterSeconds, attemptsCount)
	}

	// Find user by email
	var user User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Record failed attempt - user not found
			reason := FailureReasonInvalidCredentials
			s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
			return nil, errors.NewAuthenticationError("Invalid email or password")
		}
		return nil, errors.NewInternalError(err)
	}

	// DEBUG: Print loaded password hash
	fmt.Printf("🔍 DEBUG: Loaded hash length: %d\n", len(user.PasswordHash))
	fmt.Printf("🔍 DEBUG: Loaded hash: %s\n", user.PasswordHash)

	// Check if user is active
	if !user.IsActive {
		fmt.Println("🚨 DEBUG: User is inactive:", user.Email)
		// Record failed attempt - account inactive
		reason := FailureReasonAccountInactive
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewAuthenticationError("Account is inactive")
	}
	fmt.Println("✅ DEBUG: User is active:", user.Email)

	// Verify password
	fmt.Println("🔍 DEBUG: Verifying password for:", user.Email)
	valid, err := s.passwordHasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		fmt.Println("🚨 DEBUG: Password verification error:", err)
		return nil, errors.NewInternalError(err)
	}
	if !valid {
		fmt.Println("🚨 DEBUG: Invalid password for:", user.Email)
		// Record failed attempt - invalid password
		reason := FailureReasonInvalidPassword
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewAuthenticationError("Invalid email or password")
	}
	fmt.Println("✅ DEBUG: Password valid for:", user.Email)

	// NOTE: Email verification check disabled for internal ERP system
	// In production with self-registration, you may want to enable this:
	// if !user.EmailVerified {
	//     s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false)
	//     return nil, errors.NewAuthenticationError("Email not verified")
	// }

	// Get user's tenants
	// IMPORTANT: Bypass tenant isolation for authentication - we don't have tenant context yet!
	fmt.Println("🔍 DEBUG: Getting user tenants...")
	var userTenants []UserTenant
	if err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND is_active = ?", user.ID, true).Find(&userTenants).Error; err != nil {
		fmt.Println("🚨 DEBUG: Error getting tenants:", err)
		return nil, errors.NewInternalError(err)
	}
	fmt.Printf("✅ DEBUG: Found %d tenants\n", len(userTenants))

	if len(userTenants) == 0 {
		fmt.Println("🚨 DEBUG: No active tenants for user")
		// Record failed attempt - no tenant access
		reason := FailureReasonNoTenantAccess
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewAuthorizationError("User has no active tenant access")
	}

	// Use first tenant as default (in real app, user might select tenant)
	defaultUserTenant := userTenants[0]
	fmt.Printf("✅ DEBUG: Using tenant: %s (role: %s)\n", defaultUserTenant.TenantID, defaultUserTenant.Role)

	// Get tenant details
	// IMPORTANT: Bypass tenant isolation - we're validating tenant before setting context
	fmt.Println("🔍 DEBUG: Getting tenant details...")
	var tenant Tenant
	if err := s.db.Set("bypass_tenant", true).Where("id = ?", defaultUserTenant.TenantID).First(&tenant).Error; err != nil {
		fmt.Println("🚨 DEBUG: Error getting tenant details:", err)
		return nil, errors.NewInternalError(err)
	}
	fmt.Printf("✅ DEBUG: Tenant status: %s\n", tenant.Status)

	// Check tenant status and subscription
	if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
		fmt.Println("🚨 DEBUG: Tenant not active or trial")
		// Record failed attempt - tenant inactive
		reason := FailureReasonTenantInactive
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewSubscriptionError("Tenant subscription is not active")
	}

	// Check trial expiry
	if tenant.Status == "TRIAL" && tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
		fmt.Println("🚨 DEBUG: Trial expired")
		// Record failed attempt - trial expired
		reason := FailureReasonTrialExpired
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewSubscriptionError("Trial period has expired")
	}

	// Generate JWT tokens
	fmt.Println("🔍 DEBUG: Generating access token...")
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		defaultUserTenant.TenantID,
		defaultUserTenant.Role,
	)
	if err != nil {
		fmt.Println("🚨 DEBUG: Error generating access token:", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}
	fmt.Println("✅ DEBUG: Access token generated")

	fmt.Println("🔍 DEBUG: Generating refresh token...")
	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		fmt.Println("🚨 DEBUG: Error generating refresh token:", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}
	fmt.Println("✅ DEBUG: Refresh token generated")

	// Store refresh token
	fmt.Println("🔍 DEBUG: Storing refresh token...")
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken, req.DeviceInfo, req.IPAddress, req.UserAgent); err != nil {
		fmt.Println("🚨 DEBUG: Error storing refresh token:", err)
		return nil, err
	}
	fmt.Println("✅ DEBUG: Refresh token stored")

	// Record successful login
	fmt.Println("🔍 DEBUG: Recording successful login...")
	s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, true, nil) // nil = no failure reason for success
	fmt.Println("✅ DEBUG: Login recorded")

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.cfg.JWT.Expiry.Seconds()),
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Phone:    user.Phone,
			IsActive: user.IsActive,
			CurrentTenant: &dto.TenantInfo{
				ID:     defaultUserTenant.TenantID,
				Name:   tenant.Name,
				Status: tenant.Status,
				Role:   defaultUserTenant.Role,
			},
		},
	}, nil
}

// RefreshToken validates refresh token and generates new access token
// Reference: BACKEND-IMPLEMENTATION.md lines 1014-1053
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Validate refresh token
	userID, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.NewAuthenticationError("Invalid refresh token")
	}

	// Check if token exists and is not revoked
	var refreshTokenRecord RefreshToken
	tokenHash := hashToken(req.RefreshToken)
	if err := s.db.Where("token_hash = ? AND is_revoked = ?", tokenHash, false).First(&refreshTokenRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthenticationError("Refresh token not found or revoked")
		}
		return nil, errors.NewInternalError(err)
	}

	// Check expiry
	if time.Now().After(refreshTokenRecord.ExpiresAt) {
		return nil, errors.NewAuthenticationError("Refresh token expired")
	}

	// Get user
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthenticationError("User not found")
		}
		return nil, errors.NewInternalError(err)
	}

	// Check if user is active
	if !user.IsActive {
		fmt.Println("🚨 DEBUG [RefreshToken]: User is inactive")
		return nil, errors.NewAuthenticationError("Account is inactive")
	}
	fmt.Println("✅ DEBUG [RefreshToken]: User is active")

	// Get user's active tenants
	// IMPORTANT: Bypass tenant isolation for refresh - we don't have tenant context yet!
	fmt.Println("🔍 DEBUG [RefreshToken]: Getting user tenants...")
	var userTenants []UserTenant
	if err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND is_active = ?", user.ID, true).Find(&userTenants).Error; err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error getting tenants:", err)
		return nil, errors.NewInternalError(err)
	}
	fmt.Printf("✅ DEBUG [RefreshToken]: Found %d tenants\n", len(userTenants))

	if len(userTenants) == 0 {
		fmt.Println("🚨 DEBUG [RefreshToken]: No active tenants for user")
		return nil, errors.NewAuthorizationError("User has no active tenant access")
	}

	// Use first tenant as default
	defaultUserTenant := userTenants[0]
	fmt.Printf("✅ DEBUG [RefreshToken]: Using tenant: %s (role: %s)\n", defaultUserTenant.TenantID, defaultUserTenant.Role)

	// Get tenant and check subscription status
	// IMPORTANT: Bypass tenant isolation - we're validating tenant before setting context
	fmt.Println("🔍 DEBUG [RefreshToken]: Getting tenant details...")
	var tenant Tenant
	if err := s.db.Set("bypass_tenant", true).Where("id = ?", defaultUserTenant.TenantID).First(&tenant).Error; err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error getting tenant details:", err)
		return nil, errors.NewInternalError(err)
	}
	fmt.Printf("✅ DEBUG [RefreshToken]: Tenant status: %s\n", tenant.Status)

	// CRITICAL: Check subscription status before issuing new tokens
	// Reference: BACKEND-IMPLEMENTATION-ANALYSIS.md - Recommendation #2
	if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
		return nil, errors.NewSubscriptionError("Tenant subscription is not active")
	}

	// Check trial expiry
	if tenant.Status == "TRIAL" && tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
		fmt.Println("🚨 DEBUG [RefreshToken]: Trial expired")
		return nil, errors.NewSubscriptionError("Trial period has expired")
	}

	// Generate new access token
	fmt.Println("🔍 DEBUG [RefreshToken]: Generating new access token...")
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		defaultUserTenant.TenantID,
		defaultUserTenant.Role,
	)
	if err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error generating access token:", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}
	fmt.Println("✅ DEBUG [RefreshToken]: Access token generated")

	// Optionally rotate refresh token (best practice)
	fmt.Println("🔍 DEBUG [RefreshToken]: Generating new refresh token...")
	newRefreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error generating refresh token:", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}
	fmt.Println("✅ DEBUG [RefreshToken]: New refresh token generated")

	// Revoke old refresh token
	fmt.Println("🔍 DEBUG [RefreshToken]: Revoking old refresh token...")
	if err := s.db.Model(&RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error revoking old token:", err)
		return nil, errors.NewInternalError(err)
	}
	fmt.Println("✅ DEBUG [RefreshToken]: Old refresh token revoked")

	// Store new refresh token
	fmt.Println("🔍 DEBUG [RefreshToken]: Storing new refresh token...")
	if err := s.storeRefreshToken(ctx, user.ID, newRefreshToken, refreshTokenRecord.DeviceInfo, refreshTokenRecord.IPAddress, refreshTokenRecord.UserAgent); err != nil {
		fmt.Println("🚨 DEBUG [RefreshToken]: Error storing new token:", err)
		return nil, err
	}
	fmt.Println("✅ DEBUG [RefreshToken]: New refresh token stored")

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(s.cfg.JWT.Expiry.Seconds()),
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Phone:    user.Phone,
			IsActive: user.IsActive,
			CurrentTenant: &dto.TenantInfo{
				ID:     defaultUserTenant.TenantID,
				Name:   tenant.Name,
				Status: tenant.Status,
				Role:   defaultUserTenant.Role,
			},
		},
	}, nil
}

// Logout revokes the refresh token
// Reference: BACKEND-IMPLEMENTATION.md lines 1054-1071
func (s *AuthService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Validate refresh token
	_, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		// Token invalid or expired, but we still try to revoke it
		// This is not an error for logout
	}

	// Revoke refresh token
	tokenHash := hashToken(req.RefreshToken)
	if err := s.db.Model(&RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		// If token not found, it's not an error (might be already revoked)
		if err != gorm.ErrRecordNotFound {
			return errors.NewInternalError(err)
		}
	}

	return nil
}

// ForgotPassword initiates password reset flow by sending reset email
// Reference: PHASE2-MVP-ANALYSIS.md lines 180-220
func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.PasswordResetRequest, ipAddress, userAgent string) error {
	// Find user by email
	var user User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if email exists (security best practice)
			// Return success anyway to prevent email enumeration
			return nil
		}
		return errors.NewInternalError(err)
	}

	// Check if user is active
	if !user.IsActive {
		// Don't reveal account status, return success
		return nil
	}

	// Rate limiting: Check recent password reset requests (max 3 per hour per email)
	var recentResets int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if err := s.db.Model(&PasswordReset{}).
		Where("user_id = ? AND created_at > ?", user.ID, oneHourAgo).
		Count(&recentResets).Error; err != nil {
		return errors.NewInternalError(err)
	}

	if recentResets >= 3 {
		// Rate limit exceeded, but don't reveal this to prevent abuse
		// Log this for monitoring
		return nil
	}

	// Generate secure reset token
	resetToken, err := generateSecureToken()
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to generate reset token: %w", err))
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(s.cfg.Email.PasswordResetExpiry)

	// Store password reset token
	passwordReset := PasswordReset{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     resetToken,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&passwordReset).Error; err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to store password reset token: %w", err))
	}

	// Send password reset email
	// Note: Email sending is best-effort - don't fail if email fails
	// User still has valid token in database
	// TODO: Uncomment when email service is configured and tested
	// emailService := email.NewEmailService(s.cfg)
	// if err := emailService.SendPasswordResetEmail(user.Email, user.FullName, resetToken); err != nil {
	// 	// Log error for monitoring
	// 	fmt.Printf("Failed to send password reset email: %v\n", err)
	// }

	return nil
}

// ResetPassword completes password reset flow by validating token and updating password
// Reference: PHASE2-MVP-ANALYSIS.md lines 180-220
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.PasswordResetConfirmRequest) error {
	// Find password reset record by token
	var passwordReset PasswordReset
	if err := s.db.Where("token = ?", req.Token).First(&passwordReset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewAuthenticationError("Invalid or expired reset token")
		}
		return errors.NewInternalError(err)
	}

	// Check if token has expired
	if time.Now().After(passwordReset.ExpiresAt) {
		return errors.NewAuthenticationError("Reset token has expired")
	}

	// Check if token has already been used
	if passwordReset.UsedAt != nil {
		return errors.NewAuthenticationError("Reset token has already been used")
	}

	// Get user
	var user User
	if err := s.db.Where("id = ?", passwordReset.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewAuthenticationError("User not found")
		}
		return errors.NewInternalError(err)
	}

	// Check if user is active
	if !user.IsActive {
		return errors.NewAuthenticationError("Account is inactive")
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.NewInternalError(tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user password
	if err := tx.Model(&User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"password_hash": hashedPassword,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError(fmt.Errorf("failed to update password: %w", err))
	}

	// Mark reset token as used
	usedAt := time.Now()
	if err := tx.Model(&PasswordReset{}).
		Where("id = ?", passwordReset.ID).
		Updates(map[string]interface{}{
			"used_at": usedAt,
		}).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError(fmt.Errorf("failed to mark token as used: %w", err))
	}

	// Revoke all active refresh tokens for this user (force re-login)
	if err := tx.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", user.ID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError(fmt.Errorf("failed to revoke refresh tokens: %w", err))
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return errors.NewInternalError(err)
	}

	// TODO: Send password changed confirmation email
	// emailService := email.NewEmailService(s.cfg)
	// emailService.SendPasswordChangedEmail(user.Email, user.FullName)

	return nil
}

// Helper functions

// storeRefreshToken stores refresh token in database
func (s *AuthService) storeRefreshToken(ctx context.Context, userID, token, deviceInfo, ipAddress, userAgent string) error {
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(s.cfg.JWT.RefreshExpiry)

	refreshToken := RefreshToken{
		ID:         uuid.New().String(),
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		IsRevoked:  false,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to store refresh token: %w", err))
	}

	return nil
}

// checkLoginAttempts checks if account is locked due to failed login attempts
// Implements 4-tier exponential backoff for brute force protection
// Reference: BACKEND-IMPLEMENTATION.md lines 1098-1134, PHASE2-MVP-ANALYSIS.md
func (s *AuthService) checkLoginAttempts(ctx context.Context, email, ipAddress string) (locked bool, tier int, retryAfterSeconds int, attemptsCount int64, err error) {
	// Determine tier based on failed attempts count
	// We need to check the longest lockout period first (Tier 4)
	var maxLockoutDuration time.Duration

	// Tier 4: 15+ attempts → 24 hour lockout
	if s.cfg.Security.LockoutTier4Attempts > 0 && s.cfg.Security.LockoutTier4Duration > maxLockoutDuration {
		maxLockoutDuration = s.cfg.Security.LockoutTier4Duration
	}
	// Tier 3: 10-14 attempts → 1 hour lockout
	if s.cfg.Security.LockoutTier3Attempts > 0 && s.cfg.Security.LockoutTier3Duration > maxLockoutDuration {
		maxLockoutDuration = s.cfg.Security.LockoutTier3Duration
	}
	// Tier 2: 5-9 attempts → 15 min lockout
	if s.cfg.Security.LockoutTier2Attempts > 0 && s.cfg.Security.LockoutTier2Duration > maxLockoutDuration {
		maxLockoutDuration = s.cfg.Security.LockoutTier2Duration
	}
	// Tier 1: 3-4 attempts → 5 min lockout
	if s.cfg.Security.LockoutTier1Attempts > 0 && s.cfg.Security.LockoutTier1Duration > maxLockoutDuration {
		maxLockoutDuration = s.cfg.Security.LockoutTier1Duration
	}

	// Use the longest tier duration as cutoff to count all recent failed attempts
	cutoffTime := time.Now().Add(-maxLockoutDuration)

	// Count failed attempts within the lookback window
	// Exclude attempts that have been unlocked (unlocked_at IS NULL)
	var count int64
	if err := s.db.Model(&LoginAttempt{}).
		Where("(email = ? OR ip_address = ?) AND is_success = ? AND created_at > ? AND unlocked_at IS NULL",
			email, ipAddress, false, cutoffTime).
		Count(&count).Error; err != nil {
		return false, 0, 0, 0, errors.NewInternalError(err)
	}

	// Determine which tier the user is in based on attempts count
	var lockoutDuration time.Duration
	var tierNumber int

	if count >= int64(s.cfg.Security.LockoutTier4Attempts) {
		// Tier 4: 15+ attempts → 24 hour lockout
		lockoutDuration = s.cfg.Security.LockoutTier4Duration
		tierNumber = 4
	} else if count >= int64(s.cfg.Security.LockoutTier3Attempts) {
		// Tier 3: 10-14 attempts → 1 hour lockout
		lockoutDuration = s.cfg.Security.LockoutTier3Duration
		tierNumber = 3
	} else if count >= int64(s.cfg.Security.LockoutTier2Attempts) {
		// Tier 2: 5-9 attempts → 15 min lockout
		lockoutDuration = s.cfg.Security.LockoutTier2Duration
		tierNumber = 2
	} else if count >= int64(s.cfg.Security.LockoutTier1Attempts) {
		// Tier 1: 3-4 attempts → 5 min lockout
		lockoutDuration = s.cfg.Security.LockoutTier1Duration
		tierNumber = 1
	} else {
		// No lockout - below Tier 1 threshold
		return false, 0, 0, count, nil
	}

	// Get the most recent failed attempt to calculate remaining lockout time
	// Exclude attempts that have been unlocked (unlocked_at IS NULL)
	var lastAttempt LoginAttempt
	if err := s.db.Model(&LoginAttempt{}).
		Where("(email = ? OR ip_address = ?) AND is_success = ? AND unlocked_at IS NULL", email, ipAddress, false).
		Order("created_at DESC").
		First(&lastAttempt).Error; err != nil {
		// If no failed attempt found, no lockout
		return false, 0, 0, count, nil
	}

	// Calculate when the lockout expires
	lockoutExpiresAt := lastAttempt.AttemptedAt.Add(lockoutDuration)

	// Check if still within lockout period
	if time.Now().Before(lockoutExpiresAt) {
		// Account is locked
		remainingSeconds := int(time.Until(lockoutExpiresAt).Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
		return true, tierNumber, remainingSeconds, count, nil
	}

	// Lockout period expired, but keep tier info for logging
	return false, 0, 0, count, nil
}

// recordLoginAttempt records login attempt for brute force protection
// failureReason should be one of the FailureReason* constants, or nil for successful login
func (s *AuthService) recordLoginAttempt(ctx context.Context, email, ipAddress, userAgent string, success bool, failureReason *string) error {
	attempt := LoginAttempt{
		ID:            uuid.New().String(),
		Email:         email,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
		AttemptedAt:   time.Now(),
	}

	if err := s.db.Create(&attempt).Error; err != nil {
		// Don't return error, just log it
		// Recording attempts should not block authentication
		fmt.Printf("⚠️  Failed to record login attempt: %v\n", err)
		return nil
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SwitchTenant switches user's active tenant
// Returns new access token with updated tenantID claim
func (s *AuthService) SwitchTenant(userID string, newTenantID string) (string, *Tenant, error) {
	// 1. Validate user-tenant relationship
	var userTenant UserTenant
	err := s.db.Where("user_id = ? AND tenant_id = ? AND is_active = ?",
		userID, newTenantID, true).First(&userTenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, errors.NewAuthorizationError("You don't have access to this tenant")
		}
		return "", nil, errors.NewInternalError(err)
	}

	// 2. Get tenant details
	var tenant Tenant
	err = s.db.Where("id = ?", newTenantID).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, errors.NewNotFoundError("Tenant")
		}
		return "", nil, errors.NewInternalError(err)
	}

	// 3. Check tenant status and subscription
	if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
		return "", nil, errors.NewSubscriptionError("Tenant subscription is not active")
	}

	// 4. If TRIAL, check expiry
	if tenant.Status == "TRIAL" {
		if tenant.TrialEndsAt != nil && tenant.TrialEndsAt.Before(time.Now()) {
			return "", nil, errors.NewSubscriptionError("Trial period has expired")
		}
	}

	// 5. Get user details for token generation
	var user User
	err = s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "", nil, errors.NewInternalError(err)
	}

	// 6. Generate new access token with new tenantID
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		newTenantID,
		userTenant.Role,
	)
	if err != nil {
		return "", nil, errors.NewInternalError(err)
	}

	return accessToken, &tenant, nil
}

// GetUserTenants returns all tenants accessible to user
func (s *AuthService) GetUserTenants(userID string) ([]UserTenant, []Tenant, error) {
	// Query user_tenants for user
	// IMPORTANT: Bypass tenant isolation - this is called during session restore and /auth/me
	// when user may not have tenant context yet (chicken-and-egg problem)
	var userTenants []UserTenant
	err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND is_active = ?", userID, true).
		Find(&userTenants).Error
	if err != nil {
		return nil, nil, errors.NewInternalError(err)
	}

	if len(userTenants) == 0 {
		return []UserTenant{}, []Tenant{}, nil
	}

	// Extract tenant IDs
	tenantIDs := make([]string, len(userTenants))
	for i, ut := range userTenants {
		tenantIDs[i] = ut.TenantID
	}

	// Get tenant details
	// IMPORTANT: Bypass tenant isolation - querying tenants user belongs to
	var tenants []Tenant
	err = s.db.Set("bypass_tenant", true).Where("id IN ?", tenantIDs).Find(&tenants).Error
	if err != nil {
		return nil, nil, errors.NewInternalError(err)
	}

	return userTenants, tenants, nil
}

// VerifyEmail verifies user's email using token from verification email
func (s *AuthService) VerifyEmail(token string) (*User, error) {
	// 1. Find email verification record
	var verification EmailVerification
	err := s.db.Where("token = ?", token).First(&verification).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewValidationError([]errors.ValidationError{
				{Field: "token", Message: "Invalid or expired verification token"},
			})
		}
		return nil, errors.NewInternalError(err)
	}

	// 2. Check if already verified
	if verification.VerifiedAt != nil {
		return nil, errors.NewValidationError([]errors.ValidationError{
			{Field: "token", Message: "Email already verified"},
		})
	}

	// 3. Check expiry (24 hours)
	if verification.ExpiresAt.Before(time.Now()) {
		return nil, errors.NewValidationError([]errors.ValidationError{
			{Field: "token", Message: "Verification link expired. Please request a new one"},
		})
	}

	// 4. Get user
	var user User
	err = s.db.Where("id = ?", verification.UserID).First(&user).Error
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// 5. Update user as verified
	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now
	err = s.db.Save(&user).Error
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// 6. Mark verification as used
	verification.VerifiedAt = &now
	err = s.db.Save(&verification).Error
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	return &user, nil
}

// ChangePassword allows authenticated user to change password
func (s *AuthService) ChangePassword(userID string, oldPassword string, newPassword string) error {
	// 1. Get user
	var user User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("User")
		}
		return errors.NewInternalError(err)
	}

	// 2. Verify old password
	match, err := s.passwordHasher.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil {
		return errors.NewInternalError(err)
	}
	if !match {
		return errors.NewAuthenticationError("Current password is incorrect")
	}

	// 3. Hash new password
	hashedPassword, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return errors.NewInternalError(err)
	}

	// 4. Update password
	user.PasswordHash = hashedPassword
	err = s.db.Save(&user).Error
	if err != nil {
		return errors.NewInternalError(err)
	}

	// 5. Revoke all refresh tokens (force re-login on all devices)
	err = s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
		}).Error
	if err != nil {
		return errors.NewInternalError(err)
	}

	return nil
}

// hashToken creates a SHA-256 hash of the token for secure storage
// This prevents token compromise if database is leaked
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
