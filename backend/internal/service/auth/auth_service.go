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
	"gorm.io/gorm/clause"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/models"
	"backend/pkg/errors"
	"backend/pkg/jwt"
	"backend/pkg/logger"
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

	// Log authentication attempt (without exposing sensitive data)
	logger.Debugf("[Auth.Login] Validating credentials for user: %s", user.Email)

	// Check if user is active
	if !user.IsActive {
		logger.Debugf("[Auth.Login] User is inactive: %s", user.Email)
		// Record failed attempt - account inactive
		reason := FailureReasonAccountInactive
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewAuthenticationError("Account is inactive")
	}
	logger.Debugf("[Auth.Login] User is active: %s", user.Email)

	// Verify password
	logger.Debugf("[Auth.Login] Verifying password for: %s", user.Email)
	valid, err := s.passwordHasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		logger.Errorf("[Auth.Login] Password verification error for %s: %v", user.Email, err)
		return nil, errors.NewInternalError(err)
	}
	if !valid {
		logger.Debugf("[Auth.Login] Invalid password for: %s", user.Email)
		// Record failed attempt - invalid password
		reason := FailureReasonInvalidPassword
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewAuthenticationError("Invalid email or password")
	}
	logger.Debugf("[Auth.Login] Password valid for: %s", user.Email)

	// NOTE: Email verification check disabled for internal ERP system
	// In production with self-registration, you may want to enable this:
	// if !user.EmailVerified {
	//     s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false)
	//     return nil, errors.NewAuthenticationError("Email not verified")
	// }

	// Get user's tenant access - Support both Tier 1 (user_tenants) and Tier 2 (user_company_roles)
	// IMPORTANT: Bypass tenant isolation for authentication - we don't have tenant context yet!
	logger.Debugf("[Auth.Login] Getting tenants for user: %s", user.Email)
	var userTenants []UserTenant
	if err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND is_active = ?", user.ID, true).Find(&userTenants).Error; err != nil {
		logger.Errorf("[Auth.Login] Error getting tenants: %v", err)
		return nil, errors.NewInternalError(err)
	}
	logger.Debugf("[Auth.Login] Found %d user_tenants records (Tier 1)", len(userTenants))

	var tenantID string
	var userRole string

	// Tier 1: Check user_tenants (OWNER, TENANT_ADMIN, etc.)
	if len(userTenants) > 0 {
		// User has Tier 1 access - use first tenant as default
		defaultUserTenant := userTenants[0]
		tenantID = defaultUserTenant.TenantID
		userRole = defaultUserTenant.Role
		logger.Debugf("[Auth.Login] Using Tier 1 access - Tenant: %s, Role: %s", tenantID, userRole)
	} else {
		// Tier 2: Check user_company_roles (SALES, FINANCE, WAREHOUSE, etc.)
		logger.Debugf("[Auth.Login] No Tier 1 access, checking user_company_roles (Tier 2)")
		var userCompanyRoles []models.UserCompanyRole
		if err := s.db.Set("bypass_tenant", true).
			Where("user_id = ? AND is_active = ?", user.ID, true).
			Find(&userCompanyRoles).Error; err != nil {
			logger.Errorf("[Auth.Login] Error getting user_company_roles: %v", err)
			return nil, errors.NewInternalError(err)
		}
		logger.Debugf("[Auth.Login] Found %d user_company_roles records (Tier 2)", len(userCompanyRoles))

		if len(userCompanyRoles) == 0 {
			// No access at all - neither Tier 1 nor Tier 2
			logger.Debugf("[Auth.Login] No Tier 1 or Tier 2 access for user: %s", user.Email)
			reason := FailureReasonNoTenantAccess
			s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
			return nil, errors.NewAuthorizationError("User has no active tenant access")
		}

		// User has Tier 2 access - get tenant_id from first company role
		tenantID = userCompanyRoles[0].TenantID
		userRole = string(userCompanyRoles[0].Role)
		logger.Debugf("[Auth.Login] Using Tier 2 access - Tenant: %s, Role: %s", tenantID, userRole)
	}

	// Get tenant details
	// IMPORTANT: Bypass tenant isolation - we're validating tenant before setting context
	logger.Debugf("[Auth.Login] Getting tenant details for: %s", tenantID)
	var tenant Tenant
	if err := s.db.Set("bypass_tenant", true).Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		logger.Errorf("[Auth.Login] Error getting tenant details: %v", err)
		return nil, errors.NewInternalError(err)
	}
	logger.Debugf("[Auth.Login] Tenant status: %s", tenant.Status)

	// Check tenant status and subscription
	if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
		logger.Debugf("[Auth.Login] Tenant not active or trial: %s", tenant.Status)
		// Record failed attempt - tenant inactive
		reason := FailureReasonTenantInactive
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewSubscriptionError("Tenant subscription is not active")
	}

	// Check trial expiry
	if tenant.Status == "TRIAL" && tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
		logger.Debugf("[Auth.Login] Trial expired for tenant: %s", tenantID)
		// Record failed attempt - trial expired
		reason := FailureReasonTrialExpired
		s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, false, &reason)
		return nil, errors.NewSubscriptionError("Trial period has expired")
	}

	// Build company access list for JWT (PHASE 3)
	companyAccess, err := s.buildCompanyAccess(user.ID, tenantID)
	if err != nil {
		// Log but don't fail - fallback to no company access
		companyAccess = []jwt.CompanyAccess{}
	}

	// Generate JWT tokens
	logger.Debugf("[Auth.Login] Generating tokens for user: %s", user.Email)
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		tenantID,
		userRole,
		"", // No active company on login - user must select
		companyAccess,
	)
	if err != nil {
		logger.Errorf("[Auth.Login] Error generating access token: %v", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		logger.Errorf("[Auth.Login] Error generating refresh token: %v", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}

	// Store refresh token
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken, req.DeviceInfo, req.IPAddress, req.UserAgent); err != nil {
		logger.Errorf("[Auth.Login] Error storing refresh token: %v", err)
		return nil, err
	}

	// Record successful login
	s.recordLoginAttempt(ctx, req.Email, req.IPAddress, req.UserAgent, true, nil) // nil = no failure reason for success
	logger.Infof("[Auth.Login] Login successful for user: %s", user.Email)

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
				ID:     tenantID,
				Name:   tenant.Name,
				Status: tenant.Status,
				Role:   userRole,
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

	// Begin database transaction with row-level locking to prevent race conditions
	// This ensures that concurrent refresh attempts will serialize at the database level
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to begin transaction: %w", tx.Error))
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Check if token exists and is not revoked
	// CRITICAL: Use SELECT FOR UPDATE to lock the row for the duration of the transaction
	// This prevents concurrent transactions from reading the same token simultaneously
	var refreshTokenRecord RefreshToken
	tokenHash := hashToken(req.RefreshToken)

	// Log token hash being searched
	logger.Debugf("[RefreshToken] Searching for token hash with row lock: %s...", tokenHash[:16])

	// Use Clauses(clause.Locking{Strength: "UPDATE"}) for row-level locking
	// Second concurrent request will WAIT here until first transaction commits
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ? AND is_revoked = ?", tokenHash, false).
		First(&refreshTokenRecord).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			// Check if token exists but is revoked
			var revokedToken RefreshToken
			if err2 := s.db.Where("token_hash = ?", tokenHash).First(&revokedToken).Error; err2 == nil {
				logger.Warnf("[RefreshToken] Token found but REVOKED at %v", revokedToken.RevokedAt)
				logger.Debugf("[RefreshToken] Token was created at %v, revoked after %v",
					revokedToken.CreatedAt, revokedToken.RevokedAt.Sub(revokedToken.CreatedAt))
			} else {
				logger.Warnf("[RefreshToken] Token hash not found in database at all")
			}
			return nil, errors.NewAuthenticationError("Refresh token not found or revoked")
		}
		return nil, errors.NewInternalError(err)
	}

	logger.Debugf("[RefreshToken] Token found and locked - Created: %v, Expires: %v",
		refreshTokenRecord.CreatedAt, refreshTokenRecord.ExpiresAt)

	// Check expiry
	if time.Now().After(refreshTokenRecord.ExpiresAt) {
		tx.Rollback()
		return nil, errors.NewAuthenticationError("Refresh token expired")
	}

	// Get user
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthenticationError("User not found")
		}
		return nil, errors.NewInternalError(err)
	}

	// Check if user is active
	if !user.IsActive {
		tx.Rollback()
		logger.Warnf("[RefreshToken] User %s is inactive", user.ID)
		return nil, errors.NewAuthenticationError("Account is inactive")
	}
	logger.Debugf("[RefreshToken] User %s is active", user.ID)

	// Get user's active tenants
	// IMPORTANT: Bypass tenant isolation for refresh - we don't have tenant context yet!
	logger.Debugf("[RefreshToken] Getting user tenants for user %s", user.ID)
	var userTenants []UserTenant
	if err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND is_active = ?", user.ID, true).Find(&userTenants).Error; err != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error getting tenants: %v", err)
		return nil, errors.NewInternalError(err)
	}
	logger.Debugf("[RefreshToken] Found %d tenants for user %s", len(userTenants), user.ID)

	if len(userTenants) == 0 {
		tx.Rollback()
		logger.Warnf("[RefreshToken] No active tenants for user %s", user.ID)
		return nil, errors.NewAuthorizationError("User has no active tenant access")
	}

	// Use first tenant as default
	defaultUserTenant := userTenants[0]
	logger.Debugf("[RefreshToken] Using tenant: %s (role: %s)", defaultUserTenant.TenantID, defaultUserTenant.Role)

	// Get tenant and check subscription status
	// IMPORTANT: Bypass tenant isolation - we're validating tenant before setting context
	logger.Debugf("[RefreshToken] Getting tenant details for %s", defaultUserTenant.TenantID)
	var tenant Tenant
	if err := s.db.Set("bypass_tenant", true).Where("id = ?", defaultUserTenant.TenantID).First(&tenant).Error; err != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error getting tenant details: %v", err)
		return nil, errors.NewInternalError(err)
	}
	logger.Debugf("[RefreshToken] Tenant %s status: %s", tenant.ID, tenant.Status)

	// CRITICAL: Check subscription status before issuing new tokens
	// Reference: BACKEND-IMPLEMENTATION-ANALYSIS.md - Recommendation #2
	if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
		tx.Rollback()
		return nil, errors.NewSubscriptionError("Tenant subscription is not active")
	}

	// Check trial expiry
	if tenant.Status == "TRIAL" && tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
		tx.Rollback()
		logger.Warnf("[RefreshToken] Trial expired for tenant %s", tenant.ID)
		return nil, errors.NewSubscriptionError("Trial period has expired")
	}

	// Build company access list for JWT (PHASE 3)
	companyAccess, err := s.buildCompanyAccess(user.ID, defaultUserTenant.TenantID)
	if err != nil {
		// Log but don't fail - fallback to no company access
		companyAccess = []jwt.CompanyAccess{}
	}

	// Generate new access token (token generation is stateless, no rollback needed)
	logger.Debugf("[RefreshToken] Generating new access token for user %s", user.ID)
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		defaultUserTenant.TenantID,
		defaultUserTenant.Role,
		"", // No active company on refresh - maintain session continuity
		companyAccess,
	)
	if err != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error generating access token: %v", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}
	logger.Debugf("[RefreshToken] Access token generated for user %s", user.ID)

	// Optionally rotate refresh token (best practice)
	logger.Debugf("[RefreshToken] Generating new refresh token for user %s", user.ID)
	newRefreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error generating refresh token: %v", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}
	logger.Debugf("[RefreshToken] New refresh token generated for user %s", user.ID)

	// Revoke old refresh token within the transaction
	// CRITICAL: Use tx instead of s.db to ensure atomicity
	logger.Debugf("[RefreshToken] Revoking old token in transaction (hash: %s...)", tokenHash[:16])
	result := tx.Model(&RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error revoking old token: %v", result.Error)
		return nil, errors.NewInternalError(result.Error)
	}

	logger.Debugf("[RefreshToken] Old token revoked in transaction (%d rows affected)", result.RowsAffected)
	if result.RowsAffected == 0 {
		tx.Rollback()
		logger.Warnf("[RefreshToken] No rows affected during revocation - token may have been already revoked")
		return nil, errors.NewAuthenticationError("Failed to revoke old token")
	}

	// Store new refresh token within the transaction
	// CRITICAL: Use tx for storeRefreshToken to ensure atomicity
	logger.Debugf("[RefreshToken] Storing new refresh token in transaction for user %s", user.ID)
	if err := s.storeRefreshTokenTx(ctx, tx, user.ID, newRefreshToken, refreshTokenRecord.DeviceInfo, refreshTokenRecord.IPAddress, refreshTokenRecord.UserAgent); err != nil {
		tx.Rollback()
		logger.Errorf("[RefreshToken] Error storing new token: %v", err)
		return nil, err
	}
	logger.Debugf("[RefreshToken] New refresh token stored in transaction for user %s", user.ID)

	// Commit the transaction
	// All operations succeeded - commit atomically
	if err := tx.Commit().Error; err != nil {
		logger.Errorf("[RefreshToken] Error committing transaction: %v", err)
		return nil, errors.NewInternalError(fmt.Errorf("failed to commit transaction: %w", err))
	}
	logger.Infof("[RefreshToken] Token refresh completed successfully for user %s", user.ID)

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

// storeRefreshTokenTx stores refresh token in database within a transaction
// Also enforces a limit of 3 active tokens per user to prevent token accumulation
// CRITICAL: This method accepts a transaction to maintain atomicity with other operations
func (s *AuthService) storeRefreshTokenTx(ctx context.Context, tx *gorm.DB, userID, token, deviceInfo, ipAddress, userAgent string) error {
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(s.cfg.JWT.RefreshExpiry)

	// STEP 1: Cleanup expired tokens first (regardless of revocation status)
	// This prevents accumulation of expired but unrevoked tokens
	expiredCount := int64(0)
	if result := tx.Model(&RefreshToken{}).
		Where("user_id = ? AND expires_at < ? AND is_revoked = ?", userID, time.Now(), false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}); result.Error != nil {
		logger.Warnf("[StoreTokenTx] Failed to revoke expired tokens: %v", result.Error)
	} else {
		expiredCount = result.RowsAffected
		if expiredCount > 0 {
			logger.Debugf("[StoreTokenTx] Revoked %d expired tokens", expiredCount)
		}
	}

	// STEP 2: Check how many active (non-expired) tokens this user currently has
	var activeTokenCount int64
	if err := tx.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Count(&activeTokenCount).Error; err != nil {
		logger.Warnf("[StoreTokenTx] Failed to count active tokens: %v", err)
	} else {
		logger.Debugf("[StoreTokenTx] User has %d active tokens (after removing %d expired)", activeTokenCount, expiredCount)

		// STEP 3: Enforce maximum active token limit
		// UPDATED: Changed from >= 3 to >= 2 to keep only 1 newest + new one = 2 total
		// This prevents accumulation and ensures proper cleanup
		const maxActiveTokens = 2 // Keep only 1 old token + 1 new token
		if activeTokenCount >= maxActiveTokens {
			tokensToRevoke := activeTokenCount - (maxActiveTokens - 1) // Keep maxActiveTokens-1, revoke the rest
			logger.Warnf("[StoreTokenTx] User has %d active tokens (limit: %d), revoking %d oldest tokens",
				activeTokenCount, maxActiveTokens, tokensToRevoke)

			// Get oldest tokens to revoke
			var oldTokens []RefreshToken
			if err := tx.Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
				Order("created_at ASC").
				Limit(int(tokensToRevoke)).
				Find(&oldTokens).Error; err != nil {
				logger.Warnf("[StoreTokenTx] Failed to fetch old tokens: %v", err)
			} else {
				// Revoke each old token
				revokedCount := 0
				for _, oldToken := range oldTokens {
					if err := tx.Model(&RefreshToken{}).
						Where("id = ?", oldToken.ID).
						Updates(map[string]interface{}{
							"is_revoked": true,
							"revoked_at": time.Now(),
							"updated_at": time.Now(),
						}).Error; err != nil {
						logger.Warnf("[StoreTokenTx] Failed to revoke old token %s: %v", oldToken.ID, err)
					} else {
						revokedCount++
						logger.Debugf("[StoreTokenTx] Revoked old token ID %s from %v (age: %v)",
							oldToken.ID[:8], oldToken.CreatedAt, time.Since(oldToken.CreatedAt))
					}
				}
				logger.Debugf("[StoreTokenTx] Successfully revoked %d/%d old tokens", revokedCount, len(oldTokens))
			}
		}
	}

	// Create new refresh token
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

	if err := tx.Create(&refreshToken).Error; err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to store refresh token: %w", err))
	}

	logger.Debugf("[StoreTokenTx] New token stored (expires: %v)", expiresAt)
	return nil
}

// storeRefreshToken stores refresh token in database (non-transactional version)
// Also enforces a limit of 2 active tokens per user to prevent token accumulation
// UPDATED: Matches storeRefreshTokenTx logic with expired token cleanup
func (s *AuthService) storeRefreshToken(ctx context.Context, userID, token, deviceInfo, ipAddress, userAgent string) error {
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(s.cfg.JWT.RefreshExpiry)

	// STEP 1: Cleanup expired tokens first (regardless of revocation status)
	expiredCount := int64(0)
	if result := s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND expires_at < ? AND is_revoked = ?", userID, time.Now(), false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}); result.Error != nil {
		logger.Warnf("[StoreToken] Failed to revoke expired tokens: %v", result.Error)
	} else {
		expiredCount = result.RowsAffected
		if expiredCount > 0 {
			logger.Debugf("[StoreToken] Revoked %d expired tokens", expiredCount)
		}
	}

	// STEP 2: Check how many active (non-expired) tokens this user currently has
	var activeTokenCount int64
	if err := s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Count(&activeTokenCount).Error; err != nil {
		logger.Warnf("[StoreToken] Failed to count active tokens: %v", err)
	} else {
		logger.Debugf("[StoreToken] User has %d active tokens (after removing %d expired)", activeTokenCount, expiredCount)

		// STEP 3: Enforce maximum active token limit
		const maxActiveTokens = 2 // Keep only 1 old token + 1 new token
		if activeTokenCount >= maxActiveTokens {
			tokensToRevoke := activeTokenCount - (maxActiveTokens - 1)
			logger.Warnf("[StoreToken] User has %d active tokens (limit: %d), revoking %d oldest tokens",
				activeTokenCount, maxActiveTokens, tokensToRevoke)

			// Get oldest tokens to revoke
			var oldTokens []RefreshToken
			if err := s.db.Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
				Order("created_at ASC").
				Limit(int(tokensToRevoke)).
				Find(&oldTokens).Error; err != nil {
				logger.Warnf("[StoreToken] Failed to fetch old tokens: %v", err)
			} else {
				// Revoke each old token
				revokedCount := 0
				for _, oldToken := range oldTokens {
					if err := s.db.Model(&RefreshToken{}).
						Where("id = ?", oldToken.ID).
						Updates(map[string]interface{}{
							"is_revoked": true,
							"revoked_at": time.Now(),
							"updated_at": time.Now(),
						}).Error; err != nil {
						logger.Warnf("[StoreToken] Failed to revoke old token %s: %v", oldToken.ID, err)
					} else {
						revokedCount++
						logger.Debugf("[StoreToken] Revoked old token ID %s from %v (age: %v)",
							oldToken.ID[:8], oldToken.CreatedAt, time.Since(oldToken.CreatedAt))
					}
				}
				logger.Debugf("[StoreToken] Successfully revoked %d/%d old tokens", revokedCount, len(oldTokens))
			}
		}
	}

	// Create new refresh token
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

	logger.Debugf("[StoreToken] New token stored (expires: %v)", expiresAt)
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
		logger.Warnf("[RecordLoginAttempt] Failed to record login attempt: %v", err)
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

	// 6. Get user's accessible companies for JWT (PHASE 3)
	companyAccess, err := s.buildCompanyAccess(userID, newTenantID)
	if err != nil {
		// Log but don't fail - fallback to no company access
		companyAccess = []jwt.CompanyAccess{}
	}

	// 7. Generate new access token with new tenantID and company access
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		newTenantID,
		userTenant.Role,
		"", // No active company on tenant switch
		companyAccess,
	)
	if err != nil {
		return "", nil, errors.NewInternalError(err)
	}

	return accessToken, &tenant, nil
}

// SwitchCompany switches user's active company within current tenant (PHASE 3)
// Returns new access token with updated activeCompanyID claim
func (s *AuthService) SwitchCompany(userID, tenantID, newCompanyID string) (string, *models.Company, error) {
	logger.Infof("[SwitchCompany] START - UserID: %s, TenantID: %s, CompanyID: %s", userID, tenantID, newCompanyID)

	// 1. Validate company exists and belongs to tenant
	var company models.Company
	// IMPORTANT: Bypass tenant isolation because we're manually filtering by tenant_id
	// and the session doesn't have tenant context set yet (chicken-and-egg problem)
	err := s.db.Set("bypass_tenant", true).Where("id = ? AND tenant_id = ? AND is_active = ?",
		newCompanyID, tenantID, true).First(&company).Error
	if err != nil {
		logger.Warnf("[SwitchCompany] Company not found: %v", err)
		if err == gorm.ErrRecordNotFound {
			return "", nil, errors.NewNotFoundError("Company")
		}
		return "", nil, errors.NewInternalError(err)
	}
	logger.Debugf("[SwitchCompany] Company found: %s", company.Name)

	// 2. Validate user has access to this company
	// Check Tier 1 access (OWNER/TENANT_ADMIN)
	var userTenant UserTenant
	logger.Debugf("[SwitchCompany] Checking Tier 1 access...")
	// Bypass tenant isolation - we're checking access during switch operation
	err = s.db.Set("bypass_tenant", true).Where("user_id = ? AND tenant_id = ? AND is_active = ? AND role IN ?",
		userID, tenantID, true, []string{"OWNER", "TENANT_ADMIN"}).First(&userTenant).Error

	if err != nil {
		logger.Debugf("[SwitchCompany] Tier 1 check error: %v", err)
	} else {
		logger.Debugf("[SwitchCompany] Tier 1 access granted - Role: %s", userTenant.Role)
	}

	hasTier1Access := err == nil

	// If no Tier 1 access, check Tier 2 access (UserCompanyRole)
	if !hasTier1Access {
		logger.Debugf("[SwitchCompany] Checking Tier 2 access...")
		var userCompanyRole models.UserCompanyRole
		// Bypass tenant isolation - we're checking access during switch operation
		err = s.db.Set("bypass_tenant", true).Where("user_id = ? AND company_id = ? AND is_active = ?",
			userID, newCompanyID, true).First(&userCompanyRole).Error
		if err != nil {
			logger.Warnf("[SwitchCompany] Tier 2 check error: %v", err)
			if err == gorm.ErrRecordNotFound {
				return "", nil, errors.NewAuthorizationError("You don't have access to this company")
			}
			return "", nil, errors.NewInternalError(err)
		}
		logger.Debugf("[SwitchCompany] Tier 2 access granted")
	}

	// 3. Get user details for token generation
	logger.Debugf("[SwitchCompany] Getting user details...")
	var user User
	// Bypass tenant isolation - users table doesn't have tenant_id but middleware might still block
	err = s.db.Set("bypass_tenant", true).Where("id = ?", userID).First(&user).Error
	if err != nil {
		logger.Errorf("[SwitchCompany] User not found: %v", err)
		return "", nil, errors.NewInternalError(err)
	}
	logger.Debugf("[SwitchCompany] User found: %s", user.Email)

	// 4. Determine user's role based on access tier
	var userRole string
	if hasTier1Access {
		// Tier 1: Use role from user_tenants
		userRole = userTenant.Role
		logger.Debugf("[SwitchCompany] Using Tier 1 role: %s", userRole)
	} else {
		// Tier 2: Get role from user_company_roles for the specific company
		logger.Debugf("[SwitchCompany] Getting Tier 2 role for company...")
		var userCompanyRole models.UserCompanyRole
		err = s.db.Set("bypass_tenant", true).Where("user_id = ? AND company_id = ? AND is_active = ?",
			userID, newCompanyID, true).First(&userCompanyRole).Error
		if err != nil {
			logger.Errorf("[SwitchCompany] Failed to get company role: %v", err)
			return "", nil, errors.NewInternalError(err)
		}
		userRole = string(userCompanyRole.Role)
		logger.Debugf("[SwitchCompany] Using Tier 2 role: %s", userRole)
	}

	// 5. Build company access list for JWT
	companyAccess, err := s.buildCompanyAccess(userID, tenantID)
	if err != nil {
		// Log but don't fail - fallback to current company only
		companyAccess = []jwt.CompanyAccess{
			{CompanyID: newCompanyID, Role: ""},
		}
	}

	// 6. Generate new access token with active company
	accessToken, err := s.tokenService.GenerateAccessToken(
		user.ID,
		user.Email,
		tenantID,
		userRole,
		newCompanyID, // Set active company
		companyAccess,
	)
	if err != nil {
		return "", nil, errors.NewInternalError(err)
	}

	return accessToken, &company, nil
}

// buildCompanyAccess builds company access list for JWT claims (PHASE 3)
// Returns all companies user has access to with their respective roles
func (s *AuthService) buildCompanyAccess(userID, tenantID string) ([]jwt.CompanyAccess, error) {
	var companyAccess []jwt.CompanyAccess

	// 1. Check if user has Tier 1 access (OWNER/TENANT_ADMIN)
	var userTenant UserTenant
	// Bypass tenant isolation - we're building access list during token generation
	err := s.db.Set("bypass_tenant", true).Where("user_id = ? AND tenant_id = ? AND is_active = ? AND role IN ?",
		userID, tenantID, true, []string{"OWNER", "TENANT_ADMIN"}).First(&userTenant).Error

	if err == nil {
		// User has Tier 1 access - get ALL companies in tenant
		var companies []models.Company
		// Bypass tenant isolation - we're fetching all companies for this tenant
		err = s.db.Set("bypass_tenant", true).Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&companies).Error
		if err != nil {
			return nil, err
		}

		for _, company := range companies {
			companyAccess = append(companyAccess, jwt.CompanyAccess{
				CompanyID: company.ID,
				Role:      userTenant.Role, // Use tenant-level role
			})
		}
		return companyAccess, nil
	}

	// 2. No Tier 1 access - get companies via Tier 2 (UserCompanyRole)
	var userCompanyRoles []models.UserCompanyRole
	// Bypass tenant isolation - we're fetching company roles for access list
	err = s.db.Set("bypass_tenant", true).Preload("Company").
		Where("user_company_roles.user_id = ? AND user_company_roles.is_active = ?", userID, true).
		Joins("JOIN companies ON companies.id = user_company_roles.company_id").
		Where("companies.tenant_id = ? AND companies.is_active = ?", tenantID, true).
		Find(&userCompanyRoles).Error

	if err != nil {
		return nil, err
	}

	for _, ucr := range userCompanyRoles {
		companyAccess = append(companyAccess, jwt.CompanyAccess{
			CompanyID: ucr.CompanyID,
			Role:      string(ucr.Role),
		})
	}

	return companyAccess, nil
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
	if verification.IsUsed || verification.UsedAt != nil {
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
	verification.IsUsed = true
	verification.UsedAt = &now
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
