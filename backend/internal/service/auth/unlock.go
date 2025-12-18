package auth

import (
	"context"
	"fmt"
	"time"

	"backend/pkg/errors"
)

// UnlockAccountRequest contains parameters for unlocking a user account
type UnlockAccountRequest struct {
	Email      string
	Reason     string // Admin reason for unlocking (for audit trail)
	AdminEmail string // Email of admin performing the unlock (for audit trail)
}

// UnlockAccount marks failed login attempts as unlocked (soft delete)
// This should only be called by system administrators
// Uses soft delete to preserve audit trail - failed attempts are marked with unlock metadata
func (s *AuthService) UnlockAccount(ctx context.Context, req *UnlockAccountRequest) error {
	if req.Email == "" {
		return errors.NewValidationError([]errors.ValidationError{
			{Field: "email", Message: "Email is required"},
		})
	}

	if req.AdminEmail == "" {
		return errors.NewValidationError([]errors.ValidationError{
			{Field: "adminEmail", Message: "Admin email is required for audit trail"},
		})
	}

	// Soft delete: Update metadata instead of hard delete
	// Only unlock attempts that haven't been unlocked yet
	now := time.Now()
	result := s.db.Model(&LoginAttempt{}).
		Where("email = ? AND is_success = ? AND unlocked_at IS NULL", req.Email, false).
		Updates(map[string]interface{}{
			"unlocked_at":   &now,
			"unlocked_by":   &req.AdminEmail,
			"unlock_reason": &req.Reason,
		})

	if result.Error != nil {
		return errors.NewInternalError(result.Error)
	}

	// Log the unlock action for audit trail
	fmt.Printf("🔓 ACCOUNT UNLOCKED (SOFT DELETE): email=%s, attempts_cleared=%d, reason=%s, unlocked_by=%s, timestamp=%s\n",
		req.Email, result.RowsAffected, req.Reason, req.AdminEmail, now.Format(time.RFC3339))

	return nil
}

// UnlockAccountByTimeWindow removes failed login attempts within a specific time window
// More conservative approach that preserves older audit data
func (s *AuthService) UnlockAccountByTimeWindow(ctx context.Context, email string, windowHours int, reason string) error {
	if email == "" {
		return errors.NewValidationError([]errors.ValidationError{
			{Field: "email", Message: "Email is required"},
		})
	}

	cutoffTime := time.Now().Add(-time.Duration(windowHours) * time.Hour)

	// Delete failed attempts within the time window
	result := s.db.Where("email = ? AND is_success = ? AND created_at > ?",
		email, false, cutoffTime).
		Delete(&LoginAttempt{})

	if result.Error != nil {
		return errors.NewInternalError(result.Error)
	}

	// Log the unlock action
	fmt.Printf("🔓 ACCOUNT UNLOCKED (time window): email=%s, window=%dh, attempts_cleared=%d, reason=%s\n",
		email, windowHours, result.RowsAffected, reason)

	return nil
}

// GetAccountLockStatus checks if an account is currently locked and returns lock details
func (s *AuthService) GetAccountLockStatus(ctx context.Context, email string) (*AccountLockStatus, error) {
	// Use the existing checkLoginAttempts logic
	locked, tier, retryAfterSeconds, attemptsCount, err := s.checkLoginAttempts(ctx, email, "")
	if err != nil {
		return nil, err
	}

	status := &AccountLockStatus{
		Email:             email,
		IsLocked:          locked,
		Tier:              tier,
		RetryAfterSeconds: retryAfterSeconds,
		FailedAttempts:    attemptsCount,
	}

	if locked {
		unlockAt := time.Now().Add(time.Duration(retryAfterSeconds) * time.Second)
		status.LockedUntil = &unlockAt
	}

	return status, nil
}

// AccountLockStatus represents the current lock status of an account
type AccountLockStatus struct {
	Email             string     `json:"email"`
	IsLocked          bool       `json:"is_locked"`
	Tier              int        `json:"tier"`
	RetryAfterSeconds int        `json:"retry_after_seconds"`
	FailedAttempts    int64      `json:"failed_attempts"`
	LockedUntil       *time.Time `json:"locked_until,omitempty"`
}
