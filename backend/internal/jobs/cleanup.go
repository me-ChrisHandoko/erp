package jobs

import (
	"log"
	"time"

	"backend/internal/service/auth"
)

// cleanupExpiredRefreshTokens removes expired refresh tokens from the database
// Runs hourly at :00 minutes
func (s *Scheduler) cleanupExpiredRefreshTokens() {
	defer s.recoverFromPanic("cleanupExpiredRefreshTokens")

	start := time.Now()
	now := time.Now().UTC() // CRITICAL: Use UTC to avoid timezone issues

	result := s.db.Where("expires_at < ?", now).
		Delete(&auth.RefreshToken{})

	s.lastCleanup = time.Now()

	if result.Error != nil {
		log.Printf("[ERROR][CLEANUP] Refresh tokens failed: %v", result.Error)
		return
	}

	log.Printf("[INFO][CLEANUP] Refresh tokens: deleted %d rows (duration: %v)",
		result.RowsAffected, time.Since(start))
}

// cleanupExpiredEmailVerifications removes expired or already used email verifications
// Runs hourly at :05 minutes
func (s *Scheduler) cleanupExpiredEmailVerifications() {
	defer s.recoverFromPanic("cleanupExpiredEmailVerifications")

	start := time.Now()
	now := time.Now().UTC()

	// Delete expired (24hr) OR already verified (VerifiedAt IS NOT NULL)
	result := s.db.Where("expires_at < ? OR verified_at IS NOT NULL", now).
		Delete(&auth.EmailVerification{})

	s.lastCleanup = time.Now()

	if result.Error != nil {
		log.Printf("[ERROR][CLEANUP] Email verifications failed: %v", result.Error)
		return
	}

	log.Printf("[INFO][CLEANUP] Email verifications: deleted %d rows (duration: %v)",
		result.RowsAffected, time.Since(start))
}

// cleanupExpiredPasswordResets removes expired or already used password resets
// Runs hourly at :10 minutes
func (s *Scheduler) cleanupExpiredPasswordResets() {
	defer s.recoverFromPanic("cleanupExpiredPasswordResets")

	start := time.Now()
	now := time.Now().UTC()

	// Delete expired (1hr) OR already used (UsedAt IS NOT NULL)
	result := s.db.Where("expires_at < ? OR used_at IS NOT NULL", now).
		Delete(&auth.PasswordReset{})

	s.lastCleanup = time.Now()

	if result.Error != nil {
		log.Printf("[ERROR][CLEANUP] Password resets failed: %v", result.Error)
		return
	}

	log.Printf("[INFO][CLEANUP] Password resets: deleted %d rows (duration: %v)",
		result.RowsAffected, time.Since(start))
}

// cleanupOldLoginAttempts removes login attempts older than 7 days
// Runs daily at 2 AM
func (s *Scheduler) cleanupOldLoginAttempts() {
	defer s.recoverFromPanic("cleanupOldLoginAttempts")

	start := time.Now()
	retentionDate := time.Now().UTC().AddDate(0, 0, -7) // 7 days retention for audit/GDPR compliance

	result := s.db.Where("attempted_at < ?", retentionDate).
		Delete(&auth.LoginAttempt{})

	s.lastCleanup = time.Now()

	if result.Error != nil {
		log.Printf("[ERROR][CLEANUP] Login attempts failed: %v", result.Error)
		return
	}

	log.Printf("[INFO][CLEANUP] Login attempts: deleted %d rows (duration: %v)",
		result.RowsAffected, time.Since(start))
}

// recoverFromPanic recovers from panics in cleanup jobs to prevent scheduler crashes
// This ensures that a panic in one job doesn't stop the entire scheduler
func (s *Scheduler) recoverFromPanic(jobName string) {
	if r := recover(); r != nil {
		log.Printf("[ERROR][CLEANUP] Job %s panicked: %v", jobName, r)
	}
}
