package config

import (
	"fmt"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server           ServerConfig
	Database         DatabaseConfig
	JWT              JWTConfig
	Argon2           Argon2Config
	Security         SecurityConfig
	Email            EmailConfig
	Cookie           CookieConfig
	Tenant           TenantConfig
	TenantIsolation  TenantIsolationConfig
	Tax              TaxConfig
	Logging          LoggingConfig
	CORS             CORSConfig
	Worker           WorkerConfig
	Upload           UploadConfig
	RateLimit        RateLimitConfig
	Cache            CacheConfig
	Job              JobConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	AppName     string
	Environment string
	Port        string
	Debug       bool
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret           string
	Algorithm        string // HS256 or RS256
	Expiry           time.Duration
	RefreshExpiry    time.Duration
	PrivateKeyPath   string // For RS256
	PublicKeyPath    string // For RS256
}

// Argon2Config holds password hashing configuration
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MaxLoginAttempts     int
	LoginLockoutDuration time.Duration
	// Exponential backoff tiers
	LockoutTier1Attempts  int
	LockoutTier1Duration  time.Duration
	LockoutTier2Attempts  int
	LockoutTier2Duration  time.Duration
	LockoutTier3Attempts  int
	LockoutTier3Duration  time.Duration
	LockoutTier4Attempts  int
	LockoutTier4Duration  time.Duration
}

// EmailConfig holds email/SMTP configuration
type EmailConfig struct {
	SMTPHost             string
	SMTPPort             int
	SMTPUser             string
	SMTPPassword         string
	SMTPFromName         string
	SMTPFromEmail        string
	SMTPTLS              bool
	VerificationExpiry   time.Duration
	PasswordResetExpiry  time.Duration
}

// CookieConfig holds cookie configuration
type CookieConfig struct {
	Secure bool
	Domain string
}

// TenantConfig holds tenant-related configuration
type TenantConfig struct {
	DefaultSubscriptionPrice int64
	TrialPeriodDays          int
	GracePeriodDays          int
}

// TenantIsolationConfig holds tenant isolation enforcement configuration
// This configures the behavior of GORM callback-based tenant filtering
type TenantIsolationConfig struct {
	StrictMode   bool // If true, ERROR on missing tenant context (recommended: true in production)
	LogWarnings  bool // Log all queries without tenant context for debugging
	AllowBypass  bool // Allow db.Set("bypass_tenant", true) for system operations
}

// TaxConfig holds Indonesian tax configuration
type TaxConfig struct {
	DefaultPPNRate      float64
	DefaultInvoicePrefix string
	DefaultSOPrefix      string
	DefaultPOPrefix      string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// WorkerConfig holds background worker configuration
type WorkerConfig struct {
	Enabled                    bool
	SubscriptionBillingCron    string
	ExpiryMonitoringCron       string
	OutstandingCalculationCron string
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	MaxUploadSize int64
	UploadPath    string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled  bool
	RedisURL string
	TTL      time.Duration
}

// JobConfig holds background job configuration
type JobConfig struct {
	EnableCleanup          bool
	RefreshTokenCleanup    string
	EmailCleanup           string
	PasswordCleanup        string
	LoginCleanup           string
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Database validation
	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	// JWT validation
	if c.JWT.Secret == "" && c.JWT.Algorithm != "RS256" {
		return fmt.Errorf("JWT secret is required for HS256 algorithm")
	}
	if c.JWT.Algorithm == "HS256" && len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters for HS256")
	}
	if c.JWT.Algorithm == "RS256" {
		if c.JWT.PrivateKeyPath == "" {
			return fmt.Errorf("JWT private key path is required for RS256")
		}
		if c.JWT.PublicKeyPath == "" {
			return fmt.Errorf("JWT public key path is required for RS256")
		}
	}
	if c.JWT.Expiry <= 0 {
		return fmt.Errorf("JWT expiry must be positive")
	}
	if c.JWT.RefreshExpiry <= 0 {
		return fmt.Errorf("JWT refresh expiry must be positive")
	}

	// Argon2 validation
	if c.Argon2.Memory < 8192 {
		return fmt.Errorf("Argon2 memory must be at least 8192 (8 MB)")
	}
	if c.Argon2.Iterations < 1 {
		return fmt.Errorf("Argon2 iterations must be at least 1")
	}
	if c.Argon2.Parallelism < 1 {
		return fmt.Errorf("Argon2 parallelism must be at least 1")
	}

	// Security validation
	if c.Security.MaxLoginAttempts < 3 {
		return fmt.Errorf("max login attempts must be at least 3")
	}

	// Email validation (if SMTP configured)
	if c.Email.SMTPHost != "" {
		if c.Email.SMTPPort == 0 {
			return fmt.Errorf("SMTP port required when SMTP host is set")
		}
		if c.Email.SMTPUser == "" {
			return fmt.Errorf("SMTP user required when SMTP host is set")
		}
		if c.Email.SMTPFromEmail == "" {
			return fmt.Errorf("SMTP from email required when SMTP host is set")
		}
	}

	// Production-specific validation
	if c.IsProduction() {
		if !c.Cookie.Secure {
			return fmt.Errorf("cookie secure must be true in production")
		}
		if len(c.CORS.AllowedOrigins) == 0 || (len(c.CORS.AllowedOrigins) == 1 && c.CORS.AllowedOrigins[0] == "*") {
			return fmt.Errorf("CORS allowed origins must be explicitly set in production (not *)")
		}
		if c.Cache.Enabled && c.Cache.RedisURL == "" {
			return fmt.Errorf("Redis URL required when cache is enabled in production")
		}
		if c.RateLimit.Enabled && c.Cache.RedisURL == "" {
			return fmt.Errorf("Redis URL required for rate limiting in production")
		}
	}

	// Tenant validation
	if c.Tenant.DefaultSubscriptionPrice <= 0 {
		return fmt.Errorf("default subscription price must be positive")
	}
	if c.Tenant.TrialPeriodDays < 0 {
		return fmt.Errorf("trial period days cannot be negative")
	}
	if c.Tenant.GracePeriodDays < 0 {
		return fmt.Errorf("grace period days cannot be negative")
	}

	// Tax validation
	if c.Tax.DefaultPPNRate < 0 || c.Tax.DefaultPPNRate > 100 {
		return fmt.Errorf("PPN rate must be between 0 and 100")
	}

	// Tenant Isolation validation
	if c.IsProduction() && !c.TenantIsolation.StrictMode {
		return fmt.Errorf("tenant isolation strict mode must be enabled in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf(":%s", c.Server.Port)
}
