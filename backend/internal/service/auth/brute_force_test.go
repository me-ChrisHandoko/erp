package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/internal/config"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate test tables
	err = db.AutoMigrate(&LoginAttempt{})
	require.NoError(t, err)

	return db
}

// setupTestAuthService creates a test auth service with test configuration
func setupTestAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	db := setupTestDB(t)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxLoginAttempts:     5,
			LoginLockoutDuration: 15 * time.Minute,
			// 4-tier configuration
			LockoutTier1Attempts: 3,
			LockoutTier1Duration: 5 * time.Minute,
			LockoutTier2Attempts: 5,
			LockoutTier2Duration: 15 * time.Minute,
			LockoutTier3Attempts: 10,
			LockoutTier3Duration: 1 * time.Hour,
			LockoutTier4Attempts: 15,
			LockoutTier4Duration: 24 * time.Hour,
		},
	}

	service := &AuthService{
		db:  db,
		cfg: cfg,
	}

	return service, db
}

// Helper function to create login attempts
func createLoginAttempts(t *testing.T, db *gorm.DB, email, ipAddress string, count int, success bool) {
	for i := 0; i < count; i++ {
		attempt := LoginAttempt{
			ID:          uuid.New().String(),
			Email:       email,
			IPAddress:   ipAddress,
			UserAgent:   "test-agent",
			Success:     success,
			AttemptedAt: time.Now(),
		}
		err := db.Create(&attempt).Error
		require.NoError(t, err)
	}
}

func TestCheckLoginAttempts_NoAttempts(t *testing.T) {
	service, _ := setupTestAuthService(t)

	locked, tier, retryAfter, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.False(t, locked)
	assert.Equal(t, 0, tier)
	assert.Equal(t, 0, retryAfter)
	assert.Equal(t, int64(0), count)
}

func TestCheckLoginAttempts_BelowTier1(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 2 failed attempts (below Tier 1 threshold of 3)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 2, false)

	locked, tier, _, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.False(t, locked, "Should not be locked with 2 attempts")
	assert.Equal(t, 0, tier)
	assert.Equal(t, int64(2), count)
}

func TestCheckLoginAttempts_Tier1(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 3 failed attempts (Tier 1: 3-4 attempts → 5 min lockout)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 3, false)

	locked, tier, retryAfter, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.True(t, locked, "Should be locked at Tier 1")
	assert.Equal(t, 1, tier, "Should be Tier 1")
	assert.Greater(t, retryAfter, 0, "Should have retry time")
	assert.LessOrEqual(t, retryAfter, int(5*time.Minute.Seconds()), "Retry time should be <= 5 minutes")
	assert.Equal(t, int64(3), count)
}

func TestCheckLoginAttempts_Tier2(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 5 failed attempts (Tier 2: 5-9 attempts → 15 min lockout)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 5, false)

	locked, tier, retryAfter, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.True(t, locked, "Should be locked at Tier 2")
	assert.Equal(t, 2, tier, "Should be Tier 2")
	assert.Greater(t, retryAfter, 0, "Should have retry time")
	assert.LessOrEqual(t, retryAfter, int(15*time.Minute.Seconds()), "Retry time should be <= 15 minutes")
	assert.Equal(t, int64(5), count)
}

func TestCheckLoginAttempts_Tier3(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 10 failed attempts (Tier 3: 10-14 attempts → 1 hour lockout)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 10, false)

	locked, tier, retryAfter, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.True(t, locked, "Should be locked at Tier 3")
	assert.Equal(t, 3, tier, "Should be Tier 3")
	assert.Greater(t, retryAfter, 0, "Should have retry time")
	assert.LessOrEqual(t, retryAfter, int(1*time.Hour.Seconds()), "Retry time should be <= 1 hour")
	assert.Equal(t, int64(10), count)
}

func TestCheckLoginAttempts_Tier4(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 15 failed attempts (Tier 4: 15+ attempts → 24 hour lockout)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 15, false)

	locked, tier, retryAfter, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.True(t, locked, "Should be locked at Tier 4")
	assert.Equal(t, 4, tier, "Should be Tier 4")
	assert.Greater(t, retryAfter, 0, "Should have retry time")
	assert.LessOrEqual(t, retryAfter, int(24*time.Hour.Seconds()), "Retry time should be <= 24 hours")
	assert.Equal(t, int64(15), count)
}

func TestCheckLoginAttempts_ExpiredLockout(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 3 failed attempts from 10 minutes ago (Tier 1 is 5 min, so should be expired)
	for i := 0; i < 3; i++ {
		attempt := LoginAttempt{
			ID:          uuid.New().String(),
			Email:       "test@example.com",
			IPAddress:   "192.168.1.1",
			UserAgent:   "test-agent",
			Success:     false,
			AttemptedAt: time.Now().Add(-10 * time.Minute),
		}
		err := db.Create(&attempt).Error
		require.NoError(t, err)
	}

	locked, tier, _, _, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.False(t, locked, "Should not be locked - lockout expired")
	assert.Equal(t, 0, tier, "Tier should be 0 when lockout expired")
}

func TestCheckLoginAttempts_IPAddressTracking(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 3 failed attempts from different IP
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 3, false)

	// Check from same IP - should be locked
	locked1, tier1, _, count1, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.True(t, locked1, "Should be locked for same IP")
	assert.Equal(t, 1, tier1)
	assert.Equal(t, int64(3), count1)

	// Check from different IP - should also be locked (email tracking)
	locked2, tier2, retryAfter2, count2, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.2",
	)

	assert.NoError(t, err)
	assert.True(t, locked2, "Should be locked for different IP (email tracked)")
	assert.Equal(t, 1, tier2)
	assert.Greater(t, retryAfter2, 0, "Should have retry time")
	assert.Equal(t, int64(3), count2)
}

func TestCheckLoginAttempts_SuccessfulAttemptsIgnored(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create 2 failed attempts and 1 successful
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 2, false)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 1, true)

	locked, tier, _, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.False(t, locked, "Should not be locked - only 2 failed attempts")
	assert.Equal(t, int64(2), count, "Should only count failed attempts")
	assert.Equal(t, 0, tier)
}

func TestCheckLoginAttempts_TierProgression(t *testing.T) {
	service, db := setupTestAuthService(t)
	email := "test@example.com"
	ip := "192.168.1.1"

	// Test progression through tiers
	testCases := []struct {
		attempts     int
		expectedTier int
		expectedLock bool
	}{
		{2, 0, false},  // Below Tier 1
		{3, 1, true},   // Tier 1
		{4, 1, true},   // Still Tier 1
		{5, 2, true},   // Tier 2
		{9, 2, true},   // Still Tier 2
		{10, 3, true},  // Tier 3
		{14, 3, true},  // Still Tier 3
		{15, 4, true},  // Tier 4
		{20, 4, true},  // Still Tier 4
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.attempts))+"_attempts", func(t *testing.T) {
			// Clear previous attempts
			db.Exec("DELETE FROM login_attempts")

			// Create attempts
			createLoginAttempts(t, db, email, ip, tc.attempts, false)

			locked, tier, _, _, err := service.checkLoginAttempts(
				context.Background(),
				email,
				ip,
			)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedLock, locked, "Lock status for %d attempts", tc.attempts)
			if tc.expectedLock {
				assert.Equal(t, tc.expectedTier, tier, "Tier for %d attempts", tc.attempts)
			}
		})
	}
}

func TestRecordLoginAttempt_Success(t *testing.T) {
	service, db := setupTestAuthService(t)

	err := service.recordLoginAttempt(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
		"test-agent",
		true,
	)

	assert.NoError(t, err)

	// Verify attempt was recorded
	var count int64
	db.Model(&LoginAttempt{}).Where("email = ? AND success = ?", "test@example.com", true).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestRecordLoginAttempt_Failure(t *testing.T) {
	service, db := setupTestAuthService(t)

	err := service.recordLoginAttempt(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
		"test-agent",
		false,
	)

	assert.NoError(t, err)

	// Verify attempt was recorded
	var count int64
	db.Model(&LoginAttempt{}).Where("email = ? AND success = ?", "test@example.com", false).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCheckLoginAttempts_LookbackWindow(t *testing.T) {
	service, db := setupTestAuthService(t)

	// Create attempts outside the 24-hour lookback window
	for i := 0; i < 10; i++ {
		attempt := LoginAttempt{
			ID:          uuid.New().String(),
			Email:       "test@example.com",
			IPAddress:   "192.168.1.1",
			UserAgent:   "test-agent",
			Success:     false,
			AttemptedAt: time.Now().Add(-25 * time.Hour), // Outside 24h window
		}
		err := db.Create(&attempt).Error
		require.NoError(t, err)
	}

	// Create 2 recent attempts (within window)
	createLoginAttempts(t, db, "test@example.com", "192.168.1.1", 2, false)

	locked, _, _, count, err := service.checkLoginAttempts(
		context.Background(),
		"test@example.com",
		"192.168.1.1",
	)

	assert.NoError(t, err)
	assert.False(t, locked, "Old attempts should not count")
	assert.Equal(t, int64(2), count, "Should only count recent attempts")
}
