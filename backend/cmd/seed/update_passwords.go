// Package main - Update user passwords to Argon2id
// This script updates existing user passwords from bcrypt to Argon2id
// Run with: go run cmd/seed/update_passwords.go
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/internal/config"
	"backend/models"
	"backend/pkg/security"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection
	dbDriver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	var db *gorm.DB
	var err error

	if dbDriver == "postgres" || dsn[:10] == "postgresql" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		log.Fatal("This script requires PostgreSQL")
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database")
	log.Println("🔄 Updating user passwords to Argon2id...")
	log.Println("=" + string(make([]byte, 60)))

	// Use Argon2id password hasher (same as auth service)
	argon2Config := config.Argon2Config{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
	passwordHasher := security.NewPasswordHasher(argon2Config)

	// Hash the password
	hashedPassword, err := passwordHasher.HashPassword("password123")
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	log.Printf("✓ Generated Argon2id hash (length: %d)", len(hashedPassword))

	// List of user emails to update
	userEmails := []string{
		"budi@example.com",
		"siti@example.com",
		"ahmad@example.com",
		"joko@example.com",
		"dewi@example.com",
	}

	// Update each user
	for _, email := range userEmails {
		var user models.User

		// Find user
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			log.Printf("⚠️  User not found: %s (skipping)", email)
			continue
		}

		// Update password hash
		user.PasswordHash = hashedPassword
		if err := db.Save(&user).Error; err != nil {
			log.Printf("❌ Failed to update %s: %v", email, err)
			continue
		}

		log.Printf("  ✓ Updated password for: %s (%s)", user.FullName, email)
	}

	log.Println("\n🎉 Password update completed!")
	log.Println("All users can now login with password: password123")
}
