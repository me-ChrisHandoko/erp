package config

import (
	"fmt"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	Tenant       TenantConfig
	Tax          TaxConfig
	Logging      LoggingConfig
	CORS         CORSConfig
	Worker       WorkerConfig
	Upload       UploadConfig
	RateLimit    RateLimitConfig
	Cache        CacheConfig
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
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

// TenantConfig holds tenant-related configuration
type TenantConfig struct {
	DefaultSubscriptionPrice int64
	TrialPeriodDays          int
	GracePeriodDays          int
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
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
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
