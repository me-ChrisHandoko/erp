package auth

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TokenCleanupService handles periodic cleanup of expired and orphaned tokens
type TokenCleanupService struct {
	db   *gorm.DB
	done chan bool
}

// NewTokenCleanupService creates a new token cleanup service
func NewTokenCleanupService(db *gorm.DB) *TokenCleanupService {
	return &TokenCleanupService{
		db:   db,
		done: make(chan bool),
	}
}

// Start begins the periodic token cleanup process
// This runs in the background and cleans up:
// 1. Expired refresh tokens (past expires_at date)
// 2. Old revoked tokens (revoked > 30 days ago)
func (s *TokenCleanupService) Start() {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	go func() {
		// Run immediately on startup
		s.cleanupTokens()

		for {
			select {
			case <-ticker.C:
				s.cleanupTokens()
			case <-s.done:
				ticker.Stop()
				return
			}
		}
	}()
	fmt.Println("🧹 Token cleanup service started (runs every 1 hour)")
}

// Stop gracefully stops the cleanup service
func (s *TokenCleanupService) Stop() {
	fmt.Println("🛑 Stopping token cleanup service...")
	s.done <- true
}

// cleanupTokens performs the actual cleanup operation
func (s *TokenCleanupService) cleanupTokens() {
	fmt.Println("🧹 DEBUG [TokenCleanup]: Starting scheduled token cleanup...")

	// STEP 1: Revoke expired but still active tokens
	now := time.Now()
	result1 := s.db.Model(&RefreshToken{}).
		Where("expires_at < ? AND is_revoked = ?", now, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
			"updated_at": now,
		})

	if result1.Error != nil {
		fmt.Printf("⚠️  WARNING [TokenCleanup]: Failed to revoke expired tokens: %v\n", result1.Error)
	} else if result1.RowsAffected > 0 {
		fmt.Printf("✅ DEBUG [TokenCleanup]: Revoked %d expired tokens\n", result1.RowsAffected)
	}

	// STEP 2: Delete old revoked tokens (older than 30 days)
	// This helps reduce database size while maintaining recent history for auditing
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	result2 := s.db.Where("is_revoked = ? AND revoked_at < ?", true, thirtyDaysAgo).
		Delete(&RefreshToken{})

	if result2.Error != nil {
		fmt.Printf("⚠️  WARNING [TokenCleanup]: Failed to delete old revoked tokens: %v\n", result2.Error)
	} else if result2.RowsAffected > 0 {
		fmt.Printf("✅ DEBUG [TokenCleanup]: Deleted %d old revoked tokens (>30 days)\n", result2.RowsAffected)
	}

	// STEP 3: Count remaining active tokens (for monitoring)
	var activeCount int64
	if err := s.db.Model(&RefreshToken{}).
		Where("is_revoked = ? AND expires_at > ?", false, now).
		Count(&activeCount).Error; err != nil {
		fmt.Printf("⚠️  WARNING [TokenCleanup]: Failed to count active tokens: %v\n", err)
	} else {
		fmt.Printf("📊 DEBUG [TokenCleanup]: Current active tokens across all users: %d\n", activeCount)
	}

	fmt.Println("✅ DEBUG [TokenCleanup]: Scheduled cleanup completed")
}

// CleanupUserTokens manually cleans up tokens for a specific user
// Useful for administrative actions like forcing user logout
func (s *TokenCleanupService) CleanupUserTokens(userID string) error {
	fmt.Printf("🧹 DEBUG [TokenCleanup]: Manual cleanup for user %s\n", userID)

	result := s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("✅ DEBUG [TokenCleanup]: Revoked %d tokens for user %s\n", result.RowsAffected, userID)
	return nil
}
