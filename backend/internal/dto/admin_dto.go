package dto

import "time"

// UnlockAccountRequest represents admin request to unlock a user account
type UnlockAccountRequest struct {
	Email  string `json:"email" binding:"required,email,max=255" validate:"required,email,max=255"`
	Reason string `json:"reason" binding:"required,min=5,max=500" validate:"required,min=5,max=500"`
}

// UnlockAccountResponse represents the result of an account unlock operation
type UnlockAccountResponse struct {
	Email            string `json:"email"`
	AttemptsCleared  int64  `json:"attemptsCleared"`
	Reason           string `json:"reason"`
	UnlockedBy       string `json:"unlockedBy"` // Admin email
	UnlockedAt       string `json:"unlockedAt"` // ISO8601 timestamp
}

// GetLockStatusResponse represents the current lock status of an account
type GetLockStatusResponse struct {
	Email             string     `json:"email"`
	IsLocked          bool       `json:"isLocked"`
	Tier              int        `json:"tier"`
	RetryAfterSeconds int        `json:"retryAfterSeconds"`
	FailedAttempts    int64      `json:"failedAttempts"`
	LockedUntil       *time.Time `json:"lockedUntil,omitempty"`
}
