package auth

import "time"

// User represents a user in the system
// Maps to users table in database
type User struct {
	ID              string     `gorm:"primaryKey;size:255"`
	Email           string     `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash    string     `gorm:"column:password;size:255;not null"` // Maps to DB column 'password'
	FullName        string     `gorm:"column:name;size:255;not null"` // Maps to DB column 'name'
	Phone           string     `gorm:"size:50"`
	IsActive        bool       `gorm:"default:true;not null"`
	IsSystemAdmin   bool       `gorm:"default:false;not null"`
	EmailVerified   bool       `gorm:"default:false;not null"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// Tenant represents a tenant (company/organization)
// Maps to tenants table in database
type Tenant struct {
	ID                   string     `gorm:"primaryKey;size:255"`
	Name                 string     `gorm:"size:255;not null"`
	Status               string     `gorm:"size:50;not null"` // ACTIVE, TRIAL, SUSPENDED, EXPIRED
	TrialEndsAt          *time.Time
	SubscriptionEndsAt   *time.Time
	CurrentPeriodStart   *time.Time
	CurrentPeriodEnd     *time.Time
	CreatedAt            time.Time `gorm:"not null"`
	UpdatedAt            time.Time `gorm:"not null"`
}

// TableName specifies the table name for Tenant model
func (Tenant) TableName() string {
	return "tenants"
}

// UserTenant represents the relationship between users and tenants
// Maps to user_tenants table in database
type UserTenant struct {
	ID        string    `gorm:"primaryKey;size:255"`
	UserID    string    `gorm:"size:255;not null;index:idx_user_tenant"`
	TenantID  string    `gorm:"size:255;not null;index:idx_user_tenant"`
	Role      string    `gorm:"size:50;not null"` // OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF
	IsActive  bool      `gorm:"default:true;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for UserTenant model
func (UserTenant) TableName() string {
	return "user_tenants"
}

// Company represents company profile
// Maps to companies table in database
type Company struct {
	ID        string    `gorm:"primaryKey;size:255"`
	TenantID  string    `gorm:"size:255;not null;index"`
	Code      string    `gorm:"size:50;not null"`
	Name      string    `gorm:"size:255;not null"`
	LegalName string    `gorm:"size:255;not null"`
	Type      string    `gorm:"size:50"` // PT, CV, UD, etc.
	PpnRate   float64   `gorm:"default:11.0"`
	IsActive  bool      `gorm:"default:true;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for Company model
func (Company) TableName() string {
	return "companies"
}

// RefreshToken represents a refresh token for JWT authentication
// Maps to refresh_tokens table in database
type RefreshToken struct {
	ID                string     `gorm:"primaryKey;size:255"`
	UserID            string     `gorm:"size:255;not null;index"`
	TokenHash         string     `gorm:"size:255;not null;uniqueIndex"`
	DeviceFingerprint string     `gorm:"size:64;index:idx_user_device,priority:2"` // SHA256 hash of device info for per-device token management
	DeviceInfo        string     `gorm:"size:500"`
	IPAddress         string     `gorm:"size:45"`
	UserAgent         string     `gorm:"size:500"`
	IsRevoked         bool       `gorm:"default:false;not null;index:idx_user_device,priority:3"`
	RevokedAt         *time.Time
	ExpiresAt         time.Time `gorm:"not null;index"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// EmailVerification represents an email verification token
// Maps to email_verifications table in database
type EmailVerification struct {
	ID        string     `gorm:"primaryKey;size:255"`
	UserID    string     `gorm:"size:255;not null;index"`
	Email     string     `gorm:"size:255;not null"`
	Token     string     `gorm:"size:255;not null;uniqueIndex"`
	IsUsed    bool       `gorm:"default:false"`
	UsedAt    *time.Time `gorm:"type:timestamp"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for EmailVerification model
func (EmailVerification) TableName() string {
	return "email_verifications"
}

// PasswordReset represents a password reset token
// Maps to password_resets table in database
type PasswordReset struct {
	ID         string     `gorm:"primaryKey;size:255"`
	UserID     string     `gorm:"size:255;not null;index"`
	Token      string     `gorm:"size:255;not null;uniqueIndex"`
	IPAddress  string     `gorm:"size:45"`
	UserAgent  string     `gorm:"size:500"`
	ExpiresAt  time.Time  `gorm:"not null"`
	UsedAt     *time.Time
	CreatedAt  time.Time `gorm:"not null"`
}

// TableName specifies the table name for PasswordReset model
func (PasswordReset) TableName() string {
	return "password_resets"
}

// LoginAttempt represents a login attempt for brute force protection
// Maps to login_attempts table in database
type LoginAttempt struct {
	ID            string    `gorm:"primaryKey;size:255"`
	Email         string    `gorm:"size:255;not null;index"`
	IPAddress     string    `gorm:"size:45;index"`
	UserAgent     string    `gorm:"size:500"`
	Success       bool      `gorm:"column:is_success;not null"`
	FailureReason *string   `gorm:"size:255"` // Nullable field for failure reason
	AttemptedAt   time.Time `gorm:"column:created_at;not null;index"`

	// Unlock metadata for soft delete and audit trail
	UnlockedAt   *time.Time `gorm:"index"`                    // When admin unlocked this attempt
	UnlockedBy   *string    `gorm:"size:255"`                 // Email of admin who unlocked
	UnlockReason *string    `gorm:"size:500"`                 // Admin reason for unlocking
}

// TableName specifies the table name for LoginAttempt model
func (LoginAttempt) TableName() string {
	return "login_attempts"
}
