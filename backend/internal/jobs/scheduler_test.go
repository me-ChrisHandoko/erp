package jobs

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/service/auth"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate test models
	if err := db.AutoMigrate(
		&auth.RefreshToken{},
		&auth.EmailVerification{},
		&auth.PasswordReset{},
		&auth.LoginAttempt{},
	); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupTestScheduler creates a scheduler with test configuration
func setupTestScheduler(t *testing.T) (*Scheduler, *gorm.DB) {
	db := setupTestDB(t)

	cfg := &config.Config{
		Job: config.JobConfig{
			EnableCleanup:       true,
			RefreshTokenCleanup: "0 0 * * * *",  // Hourly
			EmailCleanup:        "0 5 * * * *",  // Hourly
			PasswordCleanup:     "0 10 * * * *", // Hourly
			LoginCleanup:        "0 0 2 * * *",  // Daily at 2 AM
		},
	}

	scheduler := NewScheduler(db, cfg)
	return scheduler, db
}

// TestNewScheduler tests scheduler initialization
func TestNewScheduler(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		Job: config.JobConfig{
			EnableCleanup: true,
		},
	}

	scheduler := NewScheduler(db, cfg)

	if scheduler == nil {
		t.Fatal("Expected scheduler to be created, got nil")
	}

	if scheduler.db != db {
		t.Error("Expected scheduler.db to match provided db")
	}

	if scheduler.config != cfg {
		t.Error("Expected scheduler.config to match provided config")
	}

	if scheduler.cron == nil {
		t.Error("Expected scheduler.cron to be initialized")
	}

	if scheduler.isRunning {
		t.Error("Expected new scheduler to not be running")
	}
}

// TestSchedulerStartStop tests starting and stopping the scheduler
func TestSchedulerStartStop(t *testing.T) {
	scheduler, _ := setupTestScheduler(t)

	// Test start
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("Expected scheduler to be running after Start()")
	}

	// Test stop
	ctx := scheduler.Stop()
	if ctx == nil {
		t.Error("Expected Stop() to return a context")
	}

	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running after Stop()")
	}
}

// TestSchedulerDisabled tests that scheduler respects EnableCleanup=false
func TestSchedulerDisabled(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		Job: config.JobConfig{
			EnableCleanup: false,
		},
	}

	scheduler := NewScheduler(db, cfg)
	err := scheduler.Start()

	if err != nil {
		t.Fatalf("Expected Start() to succeed when disabled, got: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running when disabled")
	}
}

// TestCleanupExpiredRefreshTokens tests refresh token cleanup
func TestCleanupExpiredRefreshTokens(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	now := time.Now().UTC()
	expired := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Create test data: 2 expired, 1 active
	testTokens := []auth.RefreshToken{
		{
			ID:        uuid.New().String(),
			TokenHash: "expired1_hash",
			UserID:    "user1",
			ExpiresAt: expired,
		},
		{
			ID:        uuid.New().String(),
			TokenHash: "expired2_hash",
			UserID:    "user2",
			ExpiresAt: expired,
		},
		{
			ID:        uuid.New().String(),
			TokenHash: "active_hash",
			UserID:    "user3",
			ExpiresAt: future,
		},
	}

	for _, token := range testTokens {
		if err := db.Create(&token).Error; err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}
	}

	// Run cleanup
	scheduler.cleanupExpiredRefreshTokens()

	// Verify only active token remains
	var count int64
	db.Model(&auth.RefreshToken{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 token to remain, got %d", count)
	}

	// Verify it's the active token
	var remaining auth.RefreshToken
	db.First(&remaining)
	if remaining.TokenHash != "active_hash" {
		t.Errorf("Expected active token to remain, got %s", remaining.TokenHash)
	}
}

// TestCleanupExpiredEmailVerifications tests email verification cleanup
func TestCleanupExpiredEmailVerifications(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	now := time.Now().UTC()
	expired := now.Add(-25 * time.Hour) // 25 hours ago (>24hr expiry)
	future := now.Add(1 * time.Hour)

	// Create test data: 1 expired, 1 verified, 1 active
	testVerifications := []auth.EmailVerification{
		{
			ID:        uuid.New().String(),
			UserID:    "user1",
			Token:     "EXPIRED1",
			ExpiresAt: expired,
		},
		{
			ID:         uuid.New().String(),
			UserID:     "user2",
			Token:      "VERIFIED1",
			ExpiresAt:  future,
			VerifiedAt: &now, // Already verified
		},
		{
			ID:        uuid.New().String(),
			UserID:    "user3",
			Token:     "ACTIVE1",
			ExpiresAt: future,
		},
	}

	for _, verification := range testVerifications {
		if err := db.Create(&verification).Error; err != nil {
			t.Fatalf("Failed to create test verification: %v", err)
		}
	}

	// Run cleanup
	scheduler.cleanupExpiredEmailVerifications()

	// Verify only active verification remains
	var count int64
	db.Model(&auth.EmailVerification{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 verification to remain, got %d", count)
	}

	// Verify it's the active verification
	var remaining auth.EmailVerification
	db.First(&remaining)
	if remaining.Token != "ACTIVE1" {
		t.Errorf("Expected active verification to remain, got %s", remaining.Token)
	}
}

// TestCleanupExpiredPasswordResets tests password reset cleanup
func TestCleanupExpiredPasswordResets(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	now := time.Now().UTC()
	expired := now.Add(-2 * time.Hour) // 2 hours ago (>1hr expiry)
	future := now.Add(30 * time.Minute)

	// Create test data: 1 expired, 1 used, 1 active
	testResets := []auth.PasswordReset{
		{
			ID:        uuid.New().String(),
			UserID:    "user1",
			Token:     "EXPIRED1",
			ExpiresAt: expired,
		},
		{
			ID:        uuid.New().String(),
			UserID:    "user2",
			Token:     "USED1",
			ExpiresAt: future,
			UsedAt:    &now, // Already used
		},
		{
			ID:        uuid.New().String(),
			UserID:    "user3",
			Token:     "ACTIVE1",
			ExpiresAt: future,
		},
	}

	for _, reset := range testResets {
		if err := db.Create(&reset).Error; err != nil {
			t.Fatalf("Failed to create test reset: %v", err)
		}
	}

	// Run cleanup
	scheduler.cleanupExpiredPasswordResets()

	// Verify only active reset remains
	var count int64
	db.Model(&auth.PasswordReset{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 reset to remain, got %d", count)
	}

	// Verify it's the active reset
	var remaining auth.PasswordReset
	db.First(&remaining)
	if remaining.Token != "ACTIVE1" {
		t.Errorf("Expected active reset to remain, got %s", remaining.Token)
	}
}

// TestCleanupOldLoginAttempts tests login attempt cleanup
func TestCleanupOldLoginAttempts(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	now := time.Now().UTC()
	old := now.Add(-8 * 24 * time.Hour) // 8 days ago (>7 day retention)
	recent := now.Add(-3 * 24 * time.Hour) // 3 days ago

	// Create test data: 2 old, 1 recent
	testAttempts := []auth.LoginAttempt{
		{
			ID:          uuid.New().String(),
			Email:       "old1@test.com",
			IPAddress:   "1.1.1.1",
			Success:     false,
			AttemptedAt: old,
		},
		{
			ID:          uuid.New().String(),
			Email:       "old2@test.com",
			IPAddress:   "2.2.2.2",
			Success:     false,
			AttemptedAt: old,
		},
		{
			ID:          uuid.New().String(),
			Email:       "recent@test.com",
			IPAddress:   "3.3.3.3",
			Success:     true,
			AttemptedAt: recent,
		},
	}

	for _, attempt := range testAttempts {
		if err := db.Create(&attempt).Error; err != nil {
			t.Fatalf("Failed to create test attempt: %v", err)
		}
	}

	// Run cleanup
	scheduler.cleanupOldLoginAttempts()

	// Verify only recent attempt remains
	var count int64
	db.Model(&auth.LoginAttempt{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 attempt to remain, got %d", count)
	}

	// Verify it's the recent attempt
	var remaining auth.LoginAttempt
	db.First(&remaining)
	if remaining.Email != "recent@test.com" {
		t.Errorf("Expected recent attempt to remain, got %s", remaining.Email)
	}
}

// TestCleanupEmptyTables tests cleanup with empty tables
func TestCleanupEmptyTables(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Run all cleanup functions on empty tables
	scheduler.cleanupExpiredRefreshTokens()
	scheduler.cleanupExpiredEmailVerifications()
	scheduler.cleanupExpiredPasswordResets()
	scheduler.cleanupOldLoginAttempts()

	// Verify no errors occurred (check counts are still 0)
	var count int64
	db.Model(&auth.RefreshToken{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 refresh tokens, got %d", count)
	}
}

// TestGetLastCleanupTime tests lastCleanup timestamp tracking
func TestGetLastCleanupTime(t *testing.T) {
	scheduler, _ := setupTestScheduler(t)

	// Initially should be zero
	lastCleanup := scheduler.GetLastCleanupTime()
	if !lastCleanup.IsZero() {
		t.Error("Expected lastCleanup to be zero before any cleanup")
	}

	// Run a cleanup
	scheduler.cleanupExpiredRefreshTokens()

	// Now should be set
	lastCleanup = scheduler.GetLastCleanupTime()
	if lastCleanup.IsZero() {
		t.Error("Expected lastCleanup to be set after cleanup")
	}

	// Should be recent (within last second)
	timeSince := time.Since(lastCleanup)
	if timeSince > time.Second {
		t.Errorf("Expected lastCleanup to be recent, got %v ago", timeSince)
	}
}

// TestPanicRecovery tests that panics in cleanup jobs are recovered
func TestPanicRecovery(t *testing.T) {
	scheduler, _ := setupTestScheduler(t)

	// This should not panic the test
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Expected panic to be recovered, but test panicked: %v", r)
		}
	}()

	// Simulate panic by calling recoverFromPanic
	func() {
		defer scheduler.recoverFromPanic("test_job")
		panic("simulated panic for testing")
	}()

	// If we reach here, panic was recovered successfully
}
