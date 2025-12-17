package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			AppName:     getEnv("APP_NAME", "ERP Distribusi Sembako"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			Debug:       getEnvAsBool("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", ""),
			Algorithm:      getEnv("JWT_ALGORITHM", "HS256"),
			Expiry:         getEnvAsDuration("JWT_EXPIRY", 30*time.Minute),
			RefreshExpiry:  getEnvAsDuration("JWT_REFRESH_EXPIRY", 30*24*time.Hour),
			PrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", ""),
			PublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", ""),
		},
		Argon2: Argon2Config{
			Memory:      uint32(getEnvAsInt("ARGON2_MEMORY", 65536)),      // 64 MB
			Iterations:  uint32(getEnvAsInt("ARGON2_ITERATIONS", 3)),
			Parallelism: uint8(getEnvAsInt("ARGON2_PARALLELISM", 4)),
			SaltLength:  uint32(getEnvAsInt("ARGON2_SALT_LENGTH", 16)),
			KeyLength:   uint32(getEnvAsInt("ARGON2_KEY_LENGTH", 32)),
		},
		Security: SecurityConfig{
			MaxLoginAttempts:     getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LoginLockoutDuration: getEnvAsDuration("LOGIN_LOCKOUT_DURATION", 15*time.Minute),
			// 4-tier exponential backoff for brute force protection
			LockoutTier1Attempts: getEnvAsInt("LOCKOUT_TIER1_ATTEMPTS", 3),
			LockoutTier1Duration: getEnvAsDuration("LOCKOUT_TIER1_DURATION", 5*time.Minute),
			LockoutTier2Attempts: getEnvAsInt("LOCKOUT_TIER2_ATTEMPTS", 5),
			LockoutTier2Duration: getEnvAsDuration("LOCKOUT_TIER2_DURATION", 15*time.Minute),
			LockoutTier3Attempts: getEnvAsInt("LOCKOUT_TIER3_ATTEMPTS", 10),
			LockoutTier3Duration: getEnvAsDuration("LOCKOUT_TIER3_DURATION", 1*time.Hour),
			LockoutTier4Attempts: getEnvAsInt("LOCKOUT_TIER4_ATTEMPTS", 15),
			LockoutTier4Duration: getEnvAsDuration("LOCKOUT_TIER4_DURATION", 24*time.Hour),
		},
		Email: EmailConfig{
			SMTPHost:            getEnv("SMTP_HOST", ""),
			SMTPPort:            getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:            getEnv("SMTP_USER", ""),
			SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
			SMTPFromName:        getEnv("SMTP_FROM_NAME", "ERP System"),
			SMTPFromEmail:       getEnv("SMTP_FROM_EMAIL", ""),
			SMTPTLS:             getEnvAsBool("SMTP_TLS", true),
			VerificationExpiry:  getEnvAsDuration("EMAIL_VERIFICATION_EXPIRY", 24*time.Hour),
			PasswordResetExpiry: getEnvAsDuration("PASSWORD_RESET_EXPIRY", 1*time.Hour),
		},
		Cookie: CookieConfig{
			Secure: getEnvAsBool("COOKIE_SECURE", false),
			Domain: getEnv("COOKIE_DOMAIN", ""),
		},
		Tenant: TenantConfig{
			DefaultSubscriptionPrice: getEnvAsInt64("DEFAULT_SUBSCRIPTION_PRICE", 300000),
			TrialPeriodDays:          getEnvAsInt("TRIAL_PERIOD_DAYS", 14),
			GracePeriodDays:          getEnvAsInt("GRACE_PERIOD_DAYS", 7),
		},
		TenantIsolation: TenantIsolationConfig{
			StrictMode:  getEnvAsBool("TENANT_STRICT_MODE", true),
			LogWarnings: getEnvAsBool("TENANT_LOG_WARNINGS", true),
			AllowBypass: getEnvAsBool("TENANT_ALLOW_BYPASS", true),
		},
		Tax: TaxConfig{
			DefaultPPNRate:       getEnvAsFloat64("DEFAULT_PPN_RATE", 11.0),
			DefaultInvoicePrefix: getEnv("DEFAULT_INVOICE_PREFIX", "INV"),
			DefaultSOPrefix:      getEnv("DEFAULT_SO_PREFIX", "SO"),
			DefaultPOPrefix:      getEnv("DEFAULT_PO_PREFIX", "PO"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
		},
		Worker: WorkerConfig{
			Enabled:                    getEnvAsBool("WORKER_ENABLED", true),
			SubscriptionBillingCron:    getEnv("WORKER_SUBSCRIPTION_BILLING_CRON", "0 0 * * *"),
			ExpiryMonitoringCron:       getEnv("WORKER_EXPIRY_MONITORING_CRON", "0 6 * * *"),
			OutstandingCalculationCron: getEnv("WORKER_OUTSTANDING_CALC_CRON", "0 1 * * *"),
		},
		Upload: UploadConfig{
			MaxUploadSize: getEnvAsInt64("MAX_UPLOAD_SIZE", 10485760), // 10MB
			UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		},
		RateLimit: RateLimitConfig{
			Enabled:  getEnvAsBool("RATE_LIMIT_ENABLED", false),
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Cache: CacheConfig{
			Enabled:  getEnvAsBool("CACHE_ENABLED", false),
			RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
			TTL:      getEnvAsDuration("CACHE_TTL", 1*time.Hour),
		},
		Job: JobConfig{
			EnableCleanup:       getEnvAsBool("JOB_ENABLE_CLEANUP", true),
			RefreshTokenCleanup: getEnv("JOB_REFRESH_TOKEN_CLEANUP", "0 0 * * * *"),   // Hourly at :00
			EmailCleanup:        getEnv("JOB_EMAIL_CLEANUP", "0 5 * * * *"),            // Hourly at :05
			PasswordCleanup:     getEnv("JOB_PASSWORD_CLEANUP", "0 10 * * * *"),        // Hourly at :10
			LoginCleanup:        getEnv("JOB_LOGIN_CLEANUP", "0 0 2 * * *"),            // Daily at 2 AM
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
