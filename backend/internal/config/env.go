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
			URL:             getEnv("DATABASE_URL", "file:./erp.db"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			Expiry:        getEnvAsDuration("JWT_EXPIRY", 24*time.Hour),
			RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),
		},
		Tenant: TenantConfig{
			DefaultSubscriptionPrice: getEnvAsInt64("DEFAULT_SUBSCRIPTION_PRICE", 300000),
			TrialPeriodDays:          getEnvAsInt("TRIAL_PERIOD_DAYS", 14),
			GracePeriodDays:          getEnvAsInt("GRACE_PERIOD_DAYS", 7),
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
