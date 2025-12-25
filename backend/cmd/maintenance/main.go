package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/internal/config"
	"backend/internal/service/auth"
)

// Database maintenance script
// Purpose: Clean up old refresh tokens and maintain database health
// Usage: go run cmd/maintenance/main.go
// Or schedule as cron job: 0 2 * * * /path/to/maintenance

func main() {
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  ERP Database Maintenance Script")
	fmt.Println("  Refresh Token Cleanup")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	log.Println("📡 Connecting to database...")
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress GORM logs
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get SQL DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	log.Println("✅ Connected to database")
	fmt.Println()

	// Run cleanup operations
	if err := cleanupRefreshTokens(db); err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  ✅ Maintenance Complete!")
	fmt.Println("═══════════════════════════════════════════════")
}

func cleanupRefreshTokens(db *gorm.DB) error {
	log.Println("🧹 Starting refresh token cleanup...")
	fmt.Println()

	// Step 1: Show statistics BEFORE cleanup
	type TokenStats struct {
		Total   int64
		Active  int64
		Revoked int64
		Users   int64
	}

	var beforeStats TokenStats
	db.Model(&auth.RefreshToken{}).Count(&beforeStats.Total)
	db.Model(&auth.RefreshToken{}).Where("is_revoked = ?", false).Count(&beforeStats.Active)
	db.Model(&auth.RefreshToken{}).Where("is_revoked = ?", true).Count(&beforeStats.Revoked)
	db.Model(&auth.RefreshToken{}).Distinct("user_id").Count(&beforeStats.Users)

	fmt.Println("📊 Current Statistics:")
	fmt.Printf("   Total tokens:    %d\n", beforeStats.Total)
	fmt.Printf("   Active tokens:   %d\n", beforeStats.Active)
	fmt.Printf("   Revoked tokens:  %d\n", beforeStats.Revoked)
	fmt.Printf("   Unique users:    %d\n", beforeStats.Users)
	fmt.Println()

	// Step 2: Show users with multiple active tokens
	type UserTokenCount struct {
		UserID     string
		TokenCount int64
	}
	var multiTokenUsers []UserTokenCount
	db.Model(&auth.RefreshToken{}).
		Select("user_id, COUNT(*) as token_count").
		Where("is_revoked = ?", false).
		Group("user_id").
		Having("COUNT(*) > 3").
		Find(&multiTokenUsers)

	if len(multiTokenUsers) > 0 {
		fmt.Printf("⚠️  Found %d users with >3 active tokens:\n", len(multiTokenUsers))
		for _, user := range multiTokenUsers {
			fmt.Printf("   • User %s: %d tokens\n", user.UserID[:8]+"...", user.TokenCount)
		}
		fmt.Println()
	}

	// Step 3: Revoke expired tokens
	log.Println("🔍 Revoking expired tokens...")
	result := db.Model(&auth.RefreshToken{}).
		Where("is_revoked = ? AND expires_at < ?", false, time.Now()).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke expired tokens: %w", result.Error)
	}
	fmt.Printf("   ✅ Revoked %d expired tokens\n", result.RowsAffected)
	fmt.Println()

	// Step 4: Keep only 3 newest tokens per user
	log.Println("🔍 Enforcing token limit (3 per user)...")

	// Get all users with more than 3 active tokens
	var usersToCleanup []string
	db.Model(&auth.RefreshToken{}).
		Select("user_id").
		Where("is_revoked = ?", false).
		Group("user_id").
		Having("COUNT(*) > 3").
		Pluck("user_id", &usersToCleanup)

	totalRevoked := 0
	for _, userID := range usersToCleanup {
		// Get token count for this user
		var tokenCount int64
		db.Model(&auth.RefreshToken{}).
			Where("user_id = ? AND is_revoked = ?", userID, false).
			Count(&tokenCount)

		if tokenCount > 3 {
			// Get oldest tokens (keep newest 3)
			var oldTokens []auth.RefreshToken
			db.Where("user_id = ? AND is_revoked = ?", userID, false).
				Order("created_at ASC").
				Limit(int(tokenCount - 3)).
				Find(&oldTokens)

			// Revoke each old token
			for _, token := range oldTokens {
				db.Model(&auth.RefreshToken{}).
					Where("id = ?", token.ID).
					Updates(map[string]interface{}{
						"is_revoked": true,
						"revoked_at": time.Now(),
						"updated_at": time.Now(),
					})
				totalRevoked++
			}
		}
	}
	fmt.Printf("   ✅ Revoked %d old tokens (keeping 3 newest per user)\n", totalRevoked)
	fmt.Println()

	// Step 5: Delete very old revoked tokens (>30 days)
	log.Println("🔍 Deleting old revoked tokens (>30 days)...")
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	result = db.Where("is_revoked = ? AND revoked_at < ?", true, thirtyDaysAgo).
		Delete(&auth.RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete old tokens: %w", result.Error)
	}
	fmt.Printf("   ✅ Deleted %d old revoked tokens\n", result.RowsAffected)
	fmt.Println()

	// Step 6: Show statistics AFTER cleanup
	var afterStats TokenStats
	db.Model(&auth.RefreshToken{}).Count(&afterStats.Total)
	db.Model(&auth.RefreshToken{}).Where("is_revoked = ?", false).Count(&afterStats.Active)
	db.Model(&auth.RefreshToken{}).Where("is_revoked = ?", true).Count(&afterStats.Revoked)
	db.Model(&auth.RefreshToken{}).Distinct("user_id").Count(&afterStats.Users)

	fmt.Println("📊 Final Statistics:")
	fmt.Printf("   Total tokens:    %d (was %d)\n", afterStats.Total, beforeStats.Total)
	fmt.Printf("   Active tokens:   %d (was %d)\n", afterStats.Active, beforeStats.Active)
	fmt.Printf("   Revoked tokens:  %d (was %d)\n", afterStats.Revoked, beforeStats.Revoked)
	fmt.Printf("   Unique users:    %d (was %d)\n", afterStats.Users, beforeStats.Users)
	fmt.Println()

	// Calculate improvements
	tokensRemoved := beforeStats.Total - afterStats.Total
	activeReduced := beforeStats.Active - afterStats.Active

	fmt.Println("💡 Cleanup Summary:")
	fmt.Printf("   • Removed %d tokens from database\n", tokensRemoved)
	fmt.Printf("   • Reduced active tokens by %d\n", activeReduced)
	fmt.Printf("   • Database health: %s\n", getDatabaseHealth(afterStats))

	return nil
}

func getDatabaseHealth(stats TokenStats) string {
	if stats.Total == 0 {
		return "⚠️  WARNING: No tokens in database"
	}

	avgTokensPerUser := float64(stats.Active) / float64(stats.Users)
	if avgTokensPerUser <= 2.0 {
		return "✅ EXCELLENT (avg " + fmt.Sprintf("%.1f", avgTokensPerUser) + " tokens/user)"
	} else if avgTokensPerUser <= 3.0 {
		return "👍 GOOD (avg " + fmt.Sprintf("%.1f", avgTokensPerUser) + " tokens/user)"
	} else {
		return "⚠️  NEEDS ATTENTION (avg " + fmt.Sprintf("%.1f", avgTokensPerUser) + " tokens/user)"
	}
}
