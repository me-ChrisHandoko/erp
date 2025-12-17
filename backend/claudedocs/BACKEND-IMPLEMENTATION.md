# Backend Authentication Implementation Guide

**Target:** Go Backend Developers
**Version:** 2.0
**Last Updated:** 2025-12-17

---

## 🎯 Quick Start

Dokumen ini adalah **panduan implementasi backend** untuk authentication system multi-tenant ERP. Untuk detail lengkap, lihat `authentication-mvp-design.md`.

> **⚠️ PENTING:** Sebelum mulai implementasi, WAJIB baca section **[🚨 CRITICAL SECURITY REQUIREMENTS](#-critical-security-requirements)** untuk mencegah vulnerability kritis seperti cross-tenant data leakage dan XSS attacks!

### Technology Stack

```
Language:     Go 1.25+
Framework:    Gin HTTP Router
ORM:          GORM
Database:     PostgreSQL / SQLite
Auth:         JWT (golang-jwt/jwt)
Password:     Argon2id (golang.org/x/crypto/argon2)
Validation:   go-playground/validator
Logging:      go.uber.org/zap
```

---

## 📦 Required Database Models

### 4 New Tables Dibutuhkan:

1. **`refresh_tokens`** - JWT refresh token storage dengan revocation
2. **`email_verifications`** - Email verification tokens
3. **`password_resets`** - Password reset tokens
4. **`login_attempts`** - Brute force protection tracking

### Migration Files:

```bash
db/migrations/
├── 004_create_refresh_tokens.up.sql
├── 005_create_email_verifications.up.sql
├── 006_create_password_resets.up.sql
└── 007_create_login_attempts.up.sql
```

**Detail lengkap models:** Lihat section "Database Models" di `authentication-mvp-design.md` lines 88-352

---

## 🔐 Security Implementation Checklist

### 1. Password Hashing (Argon2id)

```go
import "golang.org/x/crypto/argon2"

// Default parameters
Memory:      64 * 1024  // 64 MB
Iterations:  3
Parallelism: 4
SaltLength:  16
KeyLength:   32

// Hash format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
```

**Implementation:** Lines 1303-1425 di `authentication-mvp-design.md`

### 2. JWT Strategy

**Access Token:**

- Lifetime: 30 minutes
- Contains: userID, email, tenantID, role
- Storage: Frontend memory (Redux)

**Refresh Token:**

- Lifetime: 30 days
- Stored in: Database (revocable)
- Rotation: New token on each refresh

**Implementation:** Lines 1098-1297 di `authentication-mvp-design.md`

### 3. Brute Force Protection

```go
// Account lockout rules
Max Attempts:   5 failed logins
Time Window:    15 minutes
Lockout:        Exponential backoff
                - 5 attempts: 5 min
                - 10 attempts: 15 min
                - 15 attempts: 1 hour
                - 20+ attempts: 24 hours
```

**Implementation:** Lines 1374-1461 di `authentication-mvp-design.md`

### 4. Rate Limiting

```yaml
Login: 5 requests/minute per IP
Registration: 3 requests/hour per IP
Password Reset: 3 requests/hour per email
General API: 100 requests/minute per user
```

**Implementation:** Lines 1462-1500 di `authentication-mvp-design.md`

---

## 🛣️ API Endpoints

### Public Endpoints (No Auth)

| Method | Endpoint                       | Purpose                |
| ------ | ------------------------------ | ---------------------- |
| POST   | `/api/v1/auth/register`        | User registration      |
| POST   | `/api/v1/auth/verify-email`    | Email verification     |
| POST   | `/api/v1/auth/login`           | User login             |
| POST   | `/api/v1/auth/refresh`         | Refresh access token   |
| POST   | `/api/v1/auth/forgot-password` | Request password reset |
| POST   | `/api/v1/auth/reset-password`  | Reset password         |

### Protected Endpoints (Auth Required)

| Method | Endpoint                       | Purpose               |
| ------ | ------------------------------ | --------------------- |
| POST   | `/api/v1/auth/logout`          | Logout & revoke token |
| GET    | `/api/v1/auth/me`              | Get current user      |
| POST   | `/api/v1/auth/change-password` | Change password       |
| POST   | `/api/v1/auth/switch-tenant`   | Switch active tenant  |
| GET    | `/api/v1/auth/tenants`         | Get user's tenants    |

**Detail API specs:** Lines 727-1097 di `authentication-mvp-design.md`

---

## 🏗️ Architecture Layers

### 1. Controller Layer (HTTP Handlers)

```go
// cmd/api/handlers/auth.go
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Handle validation error
    }

    resp, err := authService.Login(req)
    // ... handle response
}
```

### 2. Service Layer (Business Logic)

```go
// internal/services/auth_service.go
type AuthService struct {
    userRepo    UserRepository
    tokenRepo   RefreshTokenRepository
    securitySvc SecurityService
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // 1. Validate credentials
    // 2. Check brute force
    // 3. Generate tokens
    // 4. Create audit log
}
```

### 3. Repository Layer (Data Access)

```go
// internal/repositories/user_repository.go
type UserRepository interface {
    Create(user *User) error
    FindByEmail(email string) (*User, error)
    Update(user *User) error
}
```

**Architecture detail:** Lines 46-86 di `authentication-mvp-design.md`

---

## 🔧 Middleware Stack

```
Request Flow:
CORS → Rate Limiter → Logger → JWT Validator → Tenant Context → Role Checker → Handler
```

### Key Middleware:

1. **JWTAuthMiddleware** - Validates JWT, extracts claims
2. **TenantContextMiddleware** - Validates tenant access, checks subscription
3. **RequireRole** - Enforces role-based access
4. **RateLimitMiddleware** - Prevents abuse

**Implementation:** Lines 2197-2386 di `authentication-mvp-design.md`

---

## 🎪 Multi-Tenant Implementation

### Critical Rule:

```go
// ❌ WRONG - Cross-tenant data leakage
db.Where("is_active = ?", true).Find(&products)

// ✅ CORRECT - Always filter by tenantID
db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&products)
```

### Tenant Context Flow:

1. JWT contains `tenantID` claim
2. Middleware validates user has access to tenant
3. Middleware checks tenant subscription status
4. Middleware injects tenant context into request
5. All queries MUST include `tenantID` filter

**Implementation:** Lines 1842-2196 di `authentication-mvp-design.md`

---

## 🚨 CRITICAL SECURITY REQUIREMENTS

### ⚠️ MUST IMPLEMENT BEFORE PRODUCTION

Bagian ini berisi **requirements kritis** yang WAJIB diimplementasikan untuk mencegah security vulnerabilities. Jangan skip section ini!

---

### 1. Tenant Query Enforcement (🔴 HIGHEST PRIORITY)

**Problem:** Manual tenant filtering rentan human error → cross-tenant data leakage

**Solution:** Implementasi GORM Scope untuk auto-inject tenant filter

```go
// internal/database/scopes.go
package database

import (
    "gorm.io/gorm"
)

// TenantScope automatically filters queries by tenant_id
// Usage: db.Scopes(database.TenantScope(tenantID)).Find(&products)
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}

// TenantScopedModel - Base model for all tenant-scoped tables
type TenantScopedModel struct {
    TenantID string `gorm:"type:varchar(255);not null;index"`
}

// Example usage in repository
func (r *ProductRepository) FindAll(tenantID string) ([]*Product, error) {
    var products []*Product

    // ✅ CORRECT - Using scope
    err := r.db.Scopes(TenantScope(tenantID)).
        Where("is_active = ?", true).
        Find(&products).Error

    return products, err
}

// ❌ WRONG - Manual filtering (easy to forget)
func (r *ProductRepository) FindAllWrong(tenantID string) ([]*Product, error) {
    var products []*Product

    // Dangerous - what if you forget tenant_id in other queries?
    err := r.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
        Find(&products).Error

    return products, err
}
```

**Testing Tenant Isolation:**

```go
// Test cross-tenant access prevention
func TestCrossTenantDataLeakage(t *testing.T) {
    // Create products for Tenant A
    tenantA := "tenant-a-id"
    productA := &Product{TenantID: tenantA, Name: "Product A"}
    db.Create(productA)

    // Create products for Tenant B
    tenantB := "tenant-b-id"
    productB := &Product{TenantID: tenantB, Name: "Product B"}
    db.Create(productB)

    // Query with Tenant A scope - should only see Product A
    var productsA []*Product
    db.Scopes(TenantScope(tenantA)).Find(&productsA)

    assert.Equal(t, 1, len(productsA))
    assert.Equal(t, "Product A", productsA[0].Name)

    // Verify Tenant B products are NOT visible
    for _, p := range productsA {
        assert.NotEqual(t, tenantB, p.TenantID)
    }
}
```

**Audit Script:**

```bash
# Find all db.Where() calls that might be missing tenant_id
grep -rn "db\.Where" internal/ cmd/ | grep -v "tenant_id" | grep -v "Scopes(TenantScope"

# This will show queries that don't use tenant filtering - audit each one!
```

---

### 2. Refresh Token Client Storage (🔴 HIGH PRIORITY)

**Problem:** Refresh token storage di localStorage = vulnerable to XSS attacks

**Solution:** WAJIB gunakan httpOnly cookies (immune to XSS)

```go
// internal/services/auth_service.go

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // ... validate credentials, generate tokens ...

    accessToken, _ := s.generateAccessToken(user)
    refreshToken, _ := s.generateRefreshToken(user)

    // Store refresh token in database
    s.tokenRepo.Create(&RefreshToken{
        UserID:    user.ID,
        Token:     hashToken(refreshToken), // Store hashed version
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    })

    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken, // Will be set as httpOnly cookie
        User:         user,
    }, nil
}

// cmd/api/handlers/auth.go

func LoginHandler(c *gin.Context) {
    // ... bind request ...

    resp, err := authService.Login(req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // ✅ CORRECT - Set refresh token as httpOnly cookie
    c.SetCookie(
        "refresh_token",           // name
        resp.RefreshToken,         // value
        30*24*60*60,               // maxAge (30 days in seconds)
        "/api/v1/auth/refresh",    // path (only sent to refresh endpoint)
        "",                        // domain (empty = current domain)
        true,                      // secure (HTTPS only in production)
        true,                      // httpOnly (NOT accessible via JavaScript)
    )

    // Add SameSite attribute manually (Gin doesn't have built-in support)
    c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie")+"; SameSite=Strict")

    // Return access token in JSON (stored in memory by frontend)
    c.JSON(http.StatusOK, gin.H{
        "access_token": resp.AccessToken,
        "user":         resp.User,
    })
}

func RefreshTokenHandler(c *gin.Context) {
    // ✅ CORRECT - Read refresh token from cookie
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token"})
        return
    }

    // Validate and rotate token
    newAccessToken, newRefreshToken, err := authService.RefreshToken(refreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
        return
    }

    // Set new refresh token cookie (rotation)
    c.SetCookie("refresh_token", newRefreshToken, 30*24*60*60, "/api/v1/auth/refresh", "", true, true)
    c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie")+"; SameSite=Strict")

    c.JSON(http.StatusOK, gin.H{
        "access_token": newAccessToken,
    })
}
```

**Cookie Security Checklist:**

- ✅ `httpOnly=true` - Prevents JavaScript access (XSS protection)
- ✅ `secure=true` - HTTPS only (production)
- ✅ `SameSite=Strict` - Prevents CSRF attacks
- ✅ `path=/api/v1/auth/refresh` - Only sent to refresh endpoint (minimize exposure)
- ✅ Token rotation on each refresh (prevents replay attacks)

**Environment Variable:**

```env
# Add to deployment config
COOKIE_SECURE=true  # Set false for local development (HTTP), true for production (HTTPS)
```

---

### 3. CSRF Protection Implementation (🔴 HIGH PRIORITY)

**Problem:** State-changing operations vulnerable to CSRF without protection

**Solution:** Implement CSRF token middleware (Double Submit Cookie Pattern)

```go
// cmd/api/middleware/csrf.go
package middleware

import (
    "crypto/rand"
    "encoding/base64"
    "net/http"

    "github.com/gin-gonic/gin"
)

// CSRFMiddleware implements CSRF protection using double-submit cookie pattern
func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip CSRF for GET, HEAD, OPTIONS (safe methods)
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }

        // Get CSRF token from cookie
        cookieToken, err := c.Cookie("csrf_token")
        if err != nil {
            c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
            c.Abort()
            return
        }

        // Get CSRF token from header
        headerToken := c.GetHeader("X-CSRF-Token")
        if headerToken == "" {
            c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required in header"})
            c.Abort()
            return
        }

        // Validate tokens match
        if cookieToken != headerToken {
            c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// GenerateCSRFToken creates a new CSRF token
func GenerateCSRFToken() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// SetCSRFCookie sets CSRF token cookie
func SetCSRFCookie(c *gin.Context) {
    token, _ := GenerateCSRFToken()

    // Set cookie (NOT httpOnly - frontend needs to read it)
    c.SetCookie(
        "csrf_token",
        token,
        24*60*60,      // 24 hours
        "/",
        "",
        true,          // secure
        false,         // NOT httpOnly (frontend needs access)
    )
    c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie")+"; SameSite=Strict")
}
```

**Usage in Routes:**

```go
// cmd/api/router.go

func SetupRoutes(r *gin.Engine, authService *services.AuthService) {
    api := r.Group("/api/v1")

    // Public routes - CSRF required for state-changing operations
    auth := api.Group("/auth")
    {
        // Login sets CSRF token for future requests
        auth.POST("/login", func(c *gin.Context) {
            // ... handle login ...

            // Set CSRF token after successful login
            middleware.SetCSRFCookie(c)

            // ... return response ...
        })

        // GET endpoints don't need CSRF
        auth.POST("/register", middleware.CSRFMiddleware(), handlers.RegisterHandler)
        auth.POST("/forgot-password", middleware.CSRFMiddleware(), handlers.ForgotPasswordHandler)
        auth.POST("/reset-password", middleware.CSRFMiddleware(), handlers.ResetPasswordHandler)
    }

    // Protected routes - always require CSRF
    protected := api.Group("")
    protected.Use(middleware.JWTAuthMiddleware())
    protected.Use(middleware.CSRFMiddleware()) // Apply to all protected routes
    {
        protected.POST("/auth/logout", handlers.LogoutHandler)
        protected.POST("/auth/change-password", handlers.ChangePasswordHandler)
        protected.POST("/auth/switch-tenant", handlers.SwitchTenantHandler)
    }
}
```

**Frontend Integration:**

```javascript
// Frontend reads CSRF token from cookie and sends in header
function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(";").shift();
}

// Add to all state-changing requests
fetch("/api/v1/auth/change-password", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-CSRF-Token": getCookie("csrf_token"), // Read from cookie
    Authorization: `Bearer ${accessToken}`,
  },
  body: JSON.stringify({ old_password: "...", new_password: "..." }),
});
```

---

### 4. Complete Environment Variables (🔴 DEPLOYMENT BLOCKER)

**Missing critical variables - tambahkan ke config:**

```env
# ============================================
# DATABASE
# ============================================
DATABASE_URL=postgresql://user:password@localhost:5432/erp_db?sslmode=disable
# For SQLite (development): DATABASE_URL=file:./erp.db

# ============================================
# SERVER
# ============================================
SERVER_PORT=8080
ENVIRONMENT=production  # development|staging|production
LOG_LEVEL=info         # debug|info|warn|error

# ============================================
# CORS
# ============================================
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Authorization,X-CSRF-Token
CORS_ALLOW_CREDENTIALS=true

# ============================================
# JWT
# ============================================
JWT_SECRET=<min-32-byte-random-string>
JWT_ALGORITHM=HS256                    # HS256 (symmetric) or RS256 (asymmetric)
JWT_ACCESS_TOKEN_EXPIRY=30m
JWT_REFRESH_TOKEN_EXPIRY=30d

# For RS256 (recommended for distributed systems):
# JWT_ALGORITHM=RS256
# JWT_PRIVATE_KEY_PATH=/path/to/private.pem
# JWT_PUBLIC_KEY_PATH=/path/to/public.pem

# ============================================
# ARGON2ID
# ============================================
ARGON2_MEMORY=65536      # 64 MB
ARGON2_ITERATIONS=3
ARGON2_PARALLELISM=4
ARGON2_SALT_LENGTH=16
ARGON2_KEY_LENGTH=32

# Development (faster hashing):
# ARGON2_MEMORY=32768      # 32 MB
# ARGON2_ITERATIONS=2

# ============================================
# SECURITY
# ============================================
MAX_LOGIN_ATTEMPTS=5
LOGIN_LOCKOUT_DURATION=15m
RATE_LIMIT_PER_MINUTE=100

# Exponential backoff thresholds
LOCKOUT_TIER1_ATTEMPTS=5
LOCKOUT_TIER1_DURATION=5m
LOCKOUT_TIER2_ATTEMPTS=10
LOCKOUT_TIER2_DURATION=15m
LOCKOUT_TIER3_ATTEMPTS=15
LOCKOUT_TIER3_DURATION=1h
LOCKOUT_TIER4_ATTEMPTS=20
LOCKOUT_TIER4_DURATION=24h

# Rate limiting per endpoint
RATE_LIMIT_LOGIN=5          # per minute per IP
RATE_LIMIT_REGISTER=3       # per hour per IP
RATE_LIMIT_PASSWORD_RESET=3 # per hour per email

# ============================================
# RATE LIMITING STORAGE
# ============================================
RATE_LIMIT_STORE=redis      # redis|memory (use redis in production)
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0

# ============================================
# EMAIL / SMTP
# ============================================
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=<smtp-password>
SMTP_FROM_NAME=ERP System
SMTP_FROM_EMAIL=noreply@example.com
SMTP_TLS=true              # Use TLS encryption

# Email templates
EMAIL_VERIFICATION_EXPIRY=24h
PASSWORD_RESET_EXPIRY=1h

# ============================================
# FRONTEND
# ============================================
FRONTEND_URL=https://app.example.com

# ============================================
# COOKIES
# ============================================
COOKIE_SECURE=true         # Set false for development (HTTP)
COOKIE_DOMAIN=             # Empty = current domain

# ============================================
# BACKGROUND JOBS
# ============================================
JOB_CLEANUP_INTERVAL=1h    # Token cleanup interval
JOB_LOGIN_CLEANUP=24h      # Old login attempts cleanup
LOGIN_ATTEMPTS_RETENTION=7d # Keep login attempts for 7 days
```

**Configuration Loading:**

```go
// internal/config/config.go
package config

import (
    "os"
    "time"
)

type Config struct {
    // Database
    DatabaseURL string

    // Server
    ServerPort  string
    Environment string
    LogLevel    string

    // CORS
    CORSAllowedOrigins  []string
    CORSAllowedMethods  []string
    CORSAllowedHeaders  []string
    CORSAllowCredentials bool

    // JWT
    JWTSecret          string
    JWTAlgorithm       string
    JWTAccessExpiry    time.Duration
    JWTRefreshExpiry   time.Duration
    JWTPrivateKeyPath  string  // For RS256
    JWTPublicKeyPath   string  // For RS256

    // Argon2
    Argon2Memory      uint32
    Argon2Iterations  uint32
    Argon2Parallelism uint8
    Argon2SaltLength  uint32
    Argon2KeyLength   uint32

    // Security
    MaxLoginAttempts      int
    LoginLockoutDuration  time.Duration
    RateLimitPerMinute    int

    // Rate Limiting
    RateLimitStore    string
    RedisURL          string
    RedisPassword     string
    RedisDB           int

    // Email
    SMTPHost       string
    SMTPPort       int
    SMTPUser       string
    SMTPPassword   string
    SMTPFromName   string
    SMTPFromEmail  string
    SMTPTLS        bool

    // Frontend
    FrontendURL string

    // Cookies
    CookieSecure bool
    CookieDomain string
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", ""),
        ServerPort:  getEnv("SERVER_PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),

        JWTSecret:        getEnv("JWT_SECRET", ""),
        JWTAlgorithm:     getEnv("JWT_ALGORITHM", "HS256"),
        JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_TOKEN_EXPIRY", "30m")),
        JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_TOKEN_EXPIRY", "30d")),

        // ... load all other config ...
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func parseDuration(s string) time.Duration {
    d, _ := time.ParseDuration(s)
    return d
}
```

---

### 5. JWT Algorithm Specification (🟡 MEDIUM PRIORITY)

**Current Issue:** Document tidak specify JWT algorithm

**Solution:** Pilih algorithm sesuai deployment architecture

#### **Option 1: HS256 (HMAC-SHA256) - Symmetric**

**Use Case:** Single server atau monolith application

**Pros:**

- ✅ Simple implementation
- ✅ Fast performance
- ✅ Single secret shared across instances

**Cons:**

- ❌ Shared secret = higher risk if compromised
- ❌ Can't verify tokens without secret (no public verification)

```go
// Using HS256
import "github.com/golang-jwt/jwt/v5"

func generateToken(claims jwt.MapClaims, secret string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func validateToken(tokenString, secret string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate algorithm
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })
}
```

#### **Option 2: RS256 (RSA-SHA256) - Asymmetric** ⭐ RECOMMENDED

**Use Case:** Microservices, distributed systems, mobile apps

**Pros:**

- ✅ Private key only on auth service (signing)
- ✅ Public key can be distributed (verification)
- ✅ Better security (private key compromise doesn't expose verification)
- ✅ Supports key rotation

**Cons:**

- ❌ More complex setup (need key pairs)
- ❌ Slightly slower than HS256

```go
// Using RS256
import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "os"

    "github.com/golang-jwt/jwt/v5"
)

// Load RSA private key
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
    keyData, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyData)
    return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// Load RSA public key
func loadPublicKey(path string) (*rsa.PublicKey, error) {
    keyData, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyData)
    pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    return pubKey.(*rsa.PublicKey), nil
}

// Generate token with RS256
func generateTokenRS256(claims jwt.MapClaims, privateKey *rsa.PrivateKey) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(privateKey)
}

// Validate token with RS256
func validateTokenRS256(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate algorithm
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return publicKey, nil
    })
}
```

**Generate RSA Key Pair:**

```bash
# Generate private key
openssl genrsa -out private.pem 2048

# Extract public key
openssl rsa -in private.pem -pubout -out public.pem

# Store securely (don't commit to git!)
# Add to .gitignore:
echo "*.pem" >> .gitignore
```

**Recommendation:**

- **Development:** HS256 (simple, fast iteration)
- **Production:** RS256 (better security, scalability)

---

## 🟡 MEDIUM PRIORITY ENHANCEMENTS

### 📌 Implementation Improvements untuk Production Readiness

Section ini berisi **enhancements** yang strongly recommended untuk production deployment. Tidak se-kritis section sebelumnya, tapi sangat meningkatkan code quality dan maintainability.

---

### 1. Background Job System (🟡 PRODUCTION REQUIREMENT)

**Current Issue:** Document hanya mention cleanup jobs tanpa implementation details

**Solution:** Implementasi job scheduler dengan robfig/cron

#### **Installation:**

```bash
go get github.com/robfig/cron/v3
```

#### **Job Scheduler Implementation:**

```go
// internal/jobs/scheduler.go
package jobs

import (
    "context"
    "log"
    "time"

    "github.com/robfig/cron/v3"
    "gorm.io/gorm"
)

type Scheduler struct {
    cron *cron.Cron
    db   *gorm.DB
}

func NewScheduler(db *gorm.DB) *Scheduler {
    return &Scheduler{
        cron: cron.New(cron.WithSeconds()),
        db:   db,
    }
}

func (s *Scheduler) Start() {
    // Hourly: Cleanup expired refresh tokens
    s.cron.AddFunc("0 0 * * * *", func() {
        s.cleanupExpiredTokens()
    })

    // Hourly: Cleanup expired email verifications
    s.cron.AddFunc("0 5 * * * *", func() {
        s.cleanupExpiredEmailVerifications()
    })

    // Hourly: Cleanup expired password resets
    s.cron.AddFunc("0 10 * * * *", func() {
        s.cleanupExpiredPasswordResets()
    })

    // Daily at 2 AM: Cleanup old login attempts (7 days retention)
    s.cron.AddFunc("0 0 2 * * *", func() {
        s.cleanupOldLoginAttempts()
    })

    // Daily at 3 AM: Send security summary reports
    s.cron.AddFunc("0 0 3 * * *", func() {
        s.sendSecuritySummary()
    })

    s.cron.Start()
    log.Println("Background job scheduler started")
}

func (s *Scheduler) Stop() context.Context {
    return s.cron.Stop()
}

// Cleanup expired refresh tokens
func (s *Scheduler) cleanupExpiredTokens() {
    result := s.db.Where("expires_at < ?", time.Now()).
        Delete(&RefreshToken{})

    log.Printf("Cleaned up %d expired refresh tokens", result.RowsAffected)
}

// Cleanup expired email verifications
func (s *Scheduler) cleanupExpiredEmailVerifications() {
    result := s.db.Where("expires_at < ? OR is_used = ?", time.Now(), true).
        Delete(&EmailVerification{})

    log.Printf("Cleaned up %d email verifications", result.RowsAffected)
}

// Cleanup expired password resets
func (s *Scheduler) cleanupExpiredPasswordResets() {
    result := s.db.Where("expires_at < ? OR is_used = ?", time.Now(), true).
        Delete(&PasswordReset{})

    log.Printf("Cleaned up %d password resets", result.RowsAffected)
}

// Cleanup old login attempts (keep last 7 days)
func (s *Scheduler) cleanupOldLoginAttempts() {
    retentionDate := time.Now().AddDate(0, 0, -7)

    result := s.db.Where("created_at < ?", retentionDate).
        Delete(&LoginAttempt{})

    log.Printf("Cleaned up %d old login attempts", result.RowsAffected)
}

// Send daily security summary
func (s *Scheduler) sendSecuritySummary() {
    // Count failed login attempts in last 24 hours
    var failedLogins int64
    s.db.Model(&LoginAttempt{}).
        Where("is_success = ? AND created_at > ?", false, time.Now().AddDate(0, 0, -1)).
        Count(&failedLogins)

    // Count locked accounts
    var lockedAccounts int64
    // Implementation depends on how you track locked accounts

    log.Printf("Security Summary - Failed Logins: %d, Locked Accounts: %d",
        failedLogins, lockedAccounts)

    // TODO: Send email to admins
}
```

#### **Integration in main.go:**

```go
// cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "your-project/internal/jobs"
    // ... other imports
)

func main() {
    // ... database setup ...

    // Start background job scheduler
    scheduler := jobs.NewScheduler(db)
    scheduler.Start()

    // ... server setup ...

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // Stop background jobs
    ctx := scheduler.Stop()
    <-ctx.Done()

    log.Println("Background jobs stopped")

    // ... shutdown server ...
}
```

#### **Cron Expression Format:**

```
# ┌───────────── second (0 - 59)
# │ ┌───────────── minute (0 - 59)
# │ │ ┌───────────── hour (0 - 23)
# │ │ │ ┌───────────── day of month (1 - 31)
# │ │ │ │ ┌───────────── month (1 - 12)
# │ │ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
# │ │ │ │ │ │
# * * * * * *

Examples:
"0 0 * * * *"     - Every hour
"0 0 2 * * *"     - Daily at 2 AM
"0 */15 * * * *"  - Every 15 minutes
"0 0 0 * * 0"     - Every Sunday at midnight
```

---

### 2. API Request/Response Examples (🟡 DOCUMENTATION)

**Current Issue:** Document hanya list endpoints tanpa payload examples

**Solution:** Tambahkan complete request/response examples untuk setiap endpoint

#### **POST /api/v1/auth/register**

**Request:**

```json
{
  "email": "john.doe@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "name": "John Doe",
  "tenant_name": "PT Example Indonesia",
  "tenant_legal_name": "PT Example Indonesia Sejahtera"
}
```

**Response (201 Created):**

```json
{
  "message": "Registration successful. Please check your email to verify your account.",
  "user": {
    "id": "cm1abc123xyz",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "name": "John Doe",
    "is_active": false
  },
  "tenant": {
    "id": "cm1tenant123",
    "name": "PT Example Indonesia",
    "status": "TRIAL",
    "trial_ends_at": "2025-12-31T23:59:59Z"
  }
}
```

**Error Response (400 Bad Request):**

```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "email",
      "message": "Email already registered"
    },
    {
      "field": "password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

#### **POST /api/v1/auth/login**

**Request:**

```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "cm1abc123xyz",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "name": "John Doe",
    "is_system_admin": false
  },
  "tenants": [
    {
      "id": "cm1tenant123",
      "name": "PT Example Indonesia",
      "role": "OWNER",
      "is_active": true
    }
  ],
  "active_tenant": {
    "id": "cm1tenant123",
    "name": "PT Example Indonesia",
    "role": "OWNER"
  }
}
```

**Note:** Refresh token set as httpOnly cookie (not in JSON response)

**Error Responses:**

```json
// 401 Unauthorized - Invalid credentials
{
  "error": "Invalid email or password"
}

// 403 Forbidden - Account locked
{
  "error": "Account temporarily locked due to multiple failed login attempts",
  "locked_until": "2025-12-17T15:30:00Z",
  "retry_after": 300
}

// 403 Forbidden - Email not verified
{
  "error": "Email not verified. Please check your email.",
  "verification_sent": true
}
```

#### **POST /api/v1/auth/refresh**

**Request:**

```
Cookie: refresh_token=<token-value>
```

**Response (200 OK):**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Note:** New refresh token set as httpOnly cookie (rotation)

**Error Response (401 Unauthorized):**

```json
{
  "error": "Invalid or expired refresh token"
}
```

#### **POST /api/v1/auth/logout**

**Request:**

```
Authorization: Bearer <access-token>
Cookie: refresh_token=<token-value>
```

**Response (200 OK):**

```json
{
  "message": "Logged out successfully"
}
```

#### **GET /api/v1/auth/me**

**Request:**

```
Authorization: Bearer <access-token>
X-CSRF-Token: <csrf-token>
```

**Response (200 OK):**

```json
{
  "user": {
    "id": "cm1abc123xyz",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "name": "John Doe",
    "is_system_admin": false,
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z"
  },
  "active_tenant": {
    "id": "cm1tenant123",
    "name": "PT Example Indonesia",
    "role": "OWNER",
    "subscription_status": "ACTIVE"
  },
  "tenants": [
    {
      "id": "cm1tenant123",
      "name": "PT Example Indonesia",
      "role": "OWNER",
      "is_active": true
    }
  ]
}
```

#### **POST /api/v1/auth/switch-tenant**

**Request:**

```json
{
  "tenant_id": "cm1tenant456"
}
```

**Response (200 OK):**

```json
{
  "message": "Tenant switched successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "active_tenant": {
    "id": "cm1tenant456",
    "name": "PT Another Company",
    "role": "ADMIN"
  }
}
```

**Error Response (403 Forbidden):**

```json
{
  "error": "You don't have access to this tenant"
}
```

---

### 3. Error Handling Standards (🟡 CODE QUALITY)

**Current Issue:** No standardized error response format

**Solution:** Implementasi unified error handling system

#### **Custom Error Types:**

```go
// pkg/errors/errors.go
package errors

import "net/http"

type AppError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    StatusCode int                    `json:"-"`
    Details    []ValidationError      `json:"details,omitempty"`
    Internal   error                  `json:"-"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Predefined error constructors
func NewValidationError(details []ValidationError) *AppError {
    return &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        StatusCode: http.StatusBadRequest,
        Details:    details,
    }
}

func NewAuthenticationError(message string) *AppError {
    return &AppError{
        Code:       "AUTHENTICATION_ERROR",
        Message:    message,
        StatusCode: http.StatusUnauthorized,
    }
}

func NewAuthorizationError(message string) *AppError {
    return &AppError{
        Code:       "AUTHORIZATION_ERROR",
        Message:    message,
        StatusCode: http.StatusForbidden,
    }
}

func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Code:       "NOT_FOUND",
        Message:    resource + " not found",
        StatusCode: http.StatusNotFound,
    }
}

func NewInternalError(err error) *AppError {
    return &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "Internal server error",
        StatusCode: http.StatusInternalServerError,
        Internal:   err,
    }
}

func NewRateLimitError(retryAfter int) *AppError {
    return &AppError{
        Code:       "RATE_LIMIT_EXCEEDED",
        Message:    "Too many requests",
        StatusCode: http.StatusTooManyRequests,
    }
}

func NewAccountLockedError(lockedUntil string, retryAfter int) *AppError {
    return &AppError{
        Code:       "ACCOUNT_LOCKED",
        Message:    "Account temporarily locked due to multiple failed login attempts",
        StatusCode: http.StatusForbidden,
    }
}
```

#### **Error Middleware:**

```go
// cmd/api/middleware/error_handler.go
package middleware

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "your-project/pkg/errors"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Check if there are any errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            // Check if it's our custom AppError
            if appErr, ok := err.(*errors.AppError); ok {
                // Log internal error if exists
                if appErr.Internal != nil {
                    log.Printf("Internal error: %v", appErr.Internal)
                }

                c.JSON(appErr.StatusCode, gin.H{
                    "error":   appErr.Message,
                    "code":    appErr.Code,
                    "details": appErr.Details,
                })
                return
            }

            // Unknown error - return 500
            log.Printf("Unknown error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
                "code":  "INTERNAL_ERROR",
            })
        }
    }
}
```

#### **Usage in Handlers:**

```go
// cmd/api/handlers/auth.go

func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Validation error
        c.Error(errors.NewValidationError([]errors.ValidationError{
            {Field: "email", Message: "Valid email is required"},
        }))
        return
    }

    user, err := authService.Login(req)
    if err != nil {
        // Pass error to middleware
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"user": user})
}

// In service layer
func (s *AuthService) Login(req LoginRequest) (*User, error) {
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, errors.NewAuthenticationError("Invalid email or password")
    }

    if !verifyPassword(user.Password, req.Password) {
        return nil, errors.NewAuthenticationError("Invalid email or password")
    }

    return user, nil
}
```

---

### 4. Enhanced Troubleshooting Guide (🟡 DOCUMENTATION)

**Expand dari 4 → 10+ common issues dengan diagnostic steps**

#### **Issue 1: Password Hashing Too Slow**

**Symptoms:**

- Login/registration takes >2 seconds
- High CPU usage during authentication

**Diagnosis:**

```bash
# Test password hashing performance
go test -bench=BenchmarkHashPassword -benchtime=10s
```

**Solutions:**

- **Development:** Reduce `ARGON2_MEMORY=32768` (32 MB) dan `ARGON2_ITERATIONS=2`
- **Production:** Keep recommended values, consider horizontal scaling
- **⚠️ WARNING:** Never use development settings in production!

**How to Measure:**

```go
start := time.Now()
hashedPassword, _ := hashPassword(password)
duration := time.Since(start)
log.Printf("Password hashing took: %v", duration)
// Should be < 500ms for production settings
```

---

#### **Issue 2: Token Validation Failing**

**Symptoms:**

- "Invalid token" errors on valid requests
- Intermittent authentication failures

**Diagnosis:**

```bash
# Check JWT secret consistency
echo $JWT_SECRET | wc -c  # Should be >= 32 bytes

# Decode JWT to inspect claims (use jwt.io or jwt-cli)
jwt decode <your-token>
```

**Common Causes:**

1. **Secret Mismatch:** Different `JWT_SECRET` across services
2. **Clock Skew:** Server time differences causing premature expiry
3. **Token Expiry:** Access token expired (check `exp` claim)
4. **Algorithm Mismatch:** Signed with HS256 but validating with RS256

**Solutions:**

```bash
# Sync server time
sudo ntpdate -s time.nist.gov

# Verify JWT_SECRET is consistent
grep JWT_SECRET /path/to/.env

# Check token expiry
jwt decode $TOKEN | jq '.exp'
date -r $(jwt decode $TOKEN | jq '.exp')
```

---

#### **Issue 3: Cross-Tenant Data Leakage**

**Symptoms:**

- Users seeing data from other tenants
- API returning unauthorized data

**Diagnosis:**

```bash
# Audit all database queries for missing tenant_id
grep -rn "db\.Where" internal/ cmd/ | grep -v "tenant_id" | grep -v "Scopes(TenantScope"

# Check specific file
grep "db\.Where\|db\.Find\|db\.First" internal/repositories/product_repository.go
```

**Prevention:**

```go
// Create test case
func TestTenantIsolation(t *testing.T) {
    // Create data for tenant A
    productA := &Product{TenantID: "tenant-a", Name: "Product A"}
    db.Create(productA)

    // Create data for tenant B
    productB := &Product{TenantID: "tenant-b", Name: "Product B"}
    db.Create(productB)

    // Query with tenant A scope
    var products []*Product
    db.Scopes(TenantScope("tenant-a")).Find(&products)

    // Should only see Product A
    assert.Equal(t, 1, len(products))
    assert.Equal(t, "Product A", products[0].Name)
}
```

---

#### **Issue 4: Rate Limiting Too Aggressive**

**Symptoms:**

- Legitimate users getting "Too many requests" errors
- Normal usage patterns blocked

**Diagnosis:**

```bash
# Check current rate limits
curl -I http://localhost:8080/api/v1/auth/login
# Look for X-RateLimit-Limit, X-RateLimit-Remaining headers

# Check Redis for rate limit keys (if using Redis)
redis-cli KEYS "rate_limit:*"
redis-cli GET "rate_limit:192.168.1.100:login"
```

**Solutions:**

```env
# Adjust per-endpoint limits
RATE_LIMIT_LOGIN=10          # Increase from 5 to 10
RATE_LIMIT_REGISTER=5        # Increase from 3 to 5
RATE_LIMIT_PASSWORD_RESET=5  # Increase from 3 to 5
RATE_LIMIT_PER_MINUTE=200    # Increase general API limit
```

---

#### **Issue 5: Email Not Sending**

**Symptoms:**

- Email verification not received
- Password reset emails missing

**Diagnosis:**

```bash
# Test SMTP connection
telnet smtp.example.com 587

# Test with swaks (SMTP test tool)
swaks --to test@example.com --from noreply@example.com --server smtp.example.com:587 --auth-user noreply@example.com --auth-password 'password' --tls

# Check application logs
grep "SMTP\|email" /var/log/app.log
```

**Common Causes:**

1. **Wrong SMTP credentials**
2. **Firewall blocking port 587**
3. **SMTP server requires TLS**
4. **FROM email not authorized**

**Solutions:**

```env
# Verify SMTP settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_TLS=true
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=<app-specific-password>  # Not your regular password!

# For Gmail: Generate app-specific password
# https://myaccount.google.com/apppasswords
```

---

#### **Issue 6: Migration Failures**

**Symptoms:**

- `make migrate-up` fails
- "relation already exists" errors

**Diagnosis:**

```bash
# Check migration status
migrate -path db/migrations -database $DATABASE_URL version

# Check database connection
psql $DATABASE_URL -c "SELECT version();"

# List existing tables
psql $DATABASE_URL -c "\dt"
```

**Solutions:**

```bash
# Rollback and retry
migrate -path db/migrations -database $DATABASE_URL down 1
migrate -path db/migrations -database $DATABASE_URL up

# Force version (if migration table corrupted)
migrate -path db/migrations -database $DATABASE_URL force <version>

# Fresh start (⚠️ DESTROYS ALL DATA)
migrate -path db/migrations -database $DATABASE_URL drop
migrate -path db/migrations -database $DATABASE_URL up
```

---

#### **Issue 7: Session/Logout Problems**

**Symptoms:**

- Logout doesn't work (still authenticated)
- Refresh token rotation failing

**Diagnosis:**

```bash
# Check refresh token in database
psql $DATABASE_URL -c "SELECT * FROM refresh_tokens WHERE user_id = '<user-id>';"

# Check cookie in browser DevTools
# Application → Cookies → refresh_token
```

**Common Causes:**

1. **Token not revoked in database**
2. **Cookie not cleared**
3. **Frontend still using old access token**

**Solutions:**

```go
// Ensure token revocation
func (s *AuthService) Logout(refreshToken string) error {
    // Find and revoke token
    token, err := s.tokenRepo.FindByToken(hashToken(refreshToken))
    if err != nil {
        return err
    }

    token.IsRevoked = true
    token.RevokedAt = time.Now()
    return s.tokenRepo.Update(token)
}

// In handler, clear cookie
func LogoutHandler(c *gin.Context) {
    refreshToken, _ := c.Cookie("refresh_token")
    authService.Logout(refreshToken)

    // Clear cookie
    c.SetCookie("refresh_token", "", -1, "/", "", true, true)

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
```

---

#### **Issue 8: Performance Problems**

**Symptoms:**

- Slow API responses (>1 second)
- High database CPU usage
- Memory leaks

**Diagnosis:**

```bash
# Enable query logging
# Set LOG_LEVEL=debug

# Profile API endpoints
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/auth/me

# Check database connection pool
psql $DATABASE_URL -c "SELECT count(*) FROM pg_stat_activity;"

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Solutions:**

```go
// Add database indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_login_attempts_email_created ON login_attempts(email, created_at);

// Configure connection pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)

// Add query timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
db.WithContext(ctx).Find(&users)
```

---

#### **Issue 9: Subscription Validation Failing**

**Symptoms:**

- Valid subscriptions showing as expired
- Trial period not working
- Grace period ignored

**Diagnosis:**

```sql
-- Check tenant subscription status
SELECT t.id, t.name, t.status, t.trial_ends_at,
       s.status as sub_status, s.current_period_end, s.grace_period_ends
FROM tenants t
LEFT JOIN subscriptions s ON t.id = s.tenant_id
WHERE t.id = '<tenant-id>';
```

**Solutions:**

```go
// Proper subscription validation
func (m *TenantContextMiddleware) validateSubscription(tenant *Tenant) error {
    now := time.Now()

    // Trial period check
    if tenant.Status == "TRIAL" {
        if now.After(tenant.TrialEndsAt) {
            return errors.NewAuthorizationError("Trial period expired")
        }
        return nil // Trial is valid
    }

    // Active subscription check
    if tenant.Status != "ACTIVE" {
        return errors.NewAuthorizationError("Subscription inactive")
    }

    // Get latest subscription
    sub, err := m.subRepo.GetByTenantID(tenant.ID)
    if err != nil {
        return err
    }

    // Check subscription status
    switch sub.Status {
    case "ACTIVE":
        return nil
    case "PAST_DUE":
        // Check grace period
        if now.Before(sub.GracePeriodEnds) {
            return nil // Still in grace period
        }
        return errors.NewAuthorizationError("Subscription payment overdue")
    case "EXPIRED", "CANCELLED":
        return errors.NewAuthorizationError("Subscription expired")
    default:
        return errors.NewInternalError(fmt.Errorf("unknown subscription status: %s", sub.Status))
    }
}
```

---

#### **Issue 10: CORS Errors**

**Symptoms:**

- "CORS policy blocked" errors in browser console
- Preflight requests failing
- Credentials not sent

**Diagnosis:**

```bash
# Test CORS with curl
curl -H "Origin: https://app.example.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     --verbose \
     http://localhost:8080/api/v1/auth/login
```

**Solutions:**

```go
// Proper CORS configuration
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()

    // CORS middleware
    config := cors.Config{
        AllowOrigins:     []string{"https://app.example.com", "https://admin.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
        ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    r.Use(cors.New(config))

    // ... routes ...
}
```

```env
# Environment variables
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Authorization,X-CSRF-Token
CORS_ALLOW_CREDENTIALS=true
```

---

## 📊 Testing Requirements

### Unit Tests (40% of timeline)

```go
// Password hashing
TestHashPassword()
TestVerifyPassword()

// JWT generation
TestGenerateAccessToken()
TestGenerateRefreshToken()
TestTokenExpiry()

// Security
TestBruteForceProtection()
TestRateLimiting()
```

### Integration Tests (30% of timeline)

```go
// Authentication flows
TestRegistrationFlow()
TestLoginFlow()
TestPasswordResetFlow()
TestTenantSwitching()
```

### Security Tests (20% of timeline)

```go
// Attack prevention
TestSQLInjection()
TestXSS()
TestTokenTampering()
TestCrosstenantAccess()
```

**Target:** 80%+ code coverage

**Testing guide:** Lines 2742-3025 di `authentication-mvp-design.md`

---

## 🚀 Deployment Checklist

### Environment Variables

```env
# JWT
JWT_SECRET=<min-32-byte-random>
JWT_ACCESS_TOKEN_EXPIRY=30m
JWT_REFRESH_TOKEN_EXPIRY=30d

# Argon2id
ARGON2_MEMORY=65536
ARGON2_ITERATIONS=3
ARGON2_PARALLELISM=4

# Security
MAX_LOGIN_ATTEMPTS=5
LOGIN_LOCKOUT_DURATION=15m
RATE_LIMIT_PER_MINUTE=100

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=<password>
FRONTEND_URL=https://app.example.com
```

### Database Setup

```bash
# Run migrations
make migrate-up

# Verify tables created
psql -d erp_db -c "\dt"
# Should show: refresh_tokens, email_verifications, password_resets, login_attempts
```

### Background Jobs

```go
// Hourly: Cleanup expired tokens
// Daily: Remove old login attempts
// Daily: Send summary reports
```

**Deployment guide:** Lines 3026-3293 di `authentication-mvp-design.md`

---

## 📅 Backend Implementation Timeline

> **⚠️ Note:** Checkmarks (✅) indicate planned tasks, not completion status. Remove checkmarks when using this as implementation tracker.

### Phase 1: Core Authentication (Week 1-2) ✅ COMPLETED

**Status:** ✅ Fully Implemented (2025-12-17)

#### Database & Migrations ✅

- [x] Migration 000002: Auth tables (`db/migrations/000002_create_auth_tables.up.sql`)
  - [x] refresh_tokens - JWT refresh token storage dengan revocation
  - [x] email_verifications - Email verification tokens
  - [x] password_resets - Password reset tokens
  - [x] login_attempts - Brute force protection tracking
- [x] Database initialization (`internal/database/database.go`)
  - [x] Connection pool configuration
  - [x] GORM logger setup
  - [x] SQLite (development) & PostgreSQL (production) support
- [x] Enhanced Dual-layer tenant isolation (`internal/database/tenant.go`)
  - [x] Layer 1: GORM Callbacks with strict mode (automatic, enforced)
  - [x] Layer 2: Manual GORM Scopes (explicit, optional)
  - [x] SetTenantSession - Sets tenant context in GORM session
  - [x] RegisterTenantCallbacks - Auto-inject tenant filters dengan configurable strict mode
  - [x] TenantScope - Manual GORM scope untuk explicit filtering
  - [x] Bypass mechanism untuk system/admin operations
  - [x] Tenant ID immutability protection

#### Core Security Services ✅

- [x] Password hashing (`pkg/security/password.go`)
  - [x] Argon2id implementation (64MB, 3 iterations, parallelism 4)
  - [x] Constant-time comparison
  - [x] Standard hash format support
- [x] JWT service (`pkg/jwt/jwt.go`)
  - [x] Dual algorithm (HS256 & RS256)
  - [x] Access token generation (30 min)
  - [x] Refresh token generation (30 days)
  - [x] Token validation & claims extraction
  - [x] RSA key loading for RS256
- [x] Error handling (`pkg/errors/errors.go`)
  - [x] Unified AppError type
  - [x] Predefined constructors (Authentication, Authorization, Validation, etc.)

#### Authentication Business Logic ✅

- [x] Auth service (`internal/service/auth/auth_service.go`)
  - [x] RegisterUser - User registration dengan tenant creation
  - [x] Login - Authentication dengan brute force protection
  - [x] RefreshToken - Token refresh dengan subscription validation
  - [x] Logout - Token revocation
  - [x] Login attempt tracking
  - [x] Email verification token generation
- [x] Domain models (`internal/service/auth/models.go`)
  - [x] User, Tenant, UserTenant, Company
  - [x] RefreshToken, EmailVerification, PasswordReset, LoginAttempt

#### HTTP Layer ✅

- [x] Auth handler (`internal/handler/auth_handler.go`)
  - [x] POST /api/v1/auth/register
  - [x] POST /api/v1/auth/login
  - [x] POST /api/v1/auth/refresh
  - [x] POST /api/v1/auth/logout
  - [x] httpOnly cookie management
  - [x] Error handling
- [x] Health handler (`internal/handler/health.go`)
  - [x] GET /health - Liveness probe
  - [x] GET /ready - Readiness probe (DB + Redis check)
- [x] DTOs (`internal/dto/auth_dto.go`)
  - [x] Request DTOs (Register, Login, Refresh, Logout)
  - [x] Response DTOs (AuthResponse, UserInfo, TenantInfo)
  - [x] Validation tags

#### Middleware Stack ✅

- [x] Authentication (`internal/middleware/auth.go`)
  - [x] JWTAuthMiddleware - JWT validation
  - [x] TenantContextMiddleware - Tenant context setup
  - [x] OptionalAuthMiddleware - Optional auth
  - [x] RequireRoleMiddleware - RBAC
- [x] Error handling (`internal/middleware/error.go`)
  - [x] ErrorHandlerMiddleware - Panic recovery
- [x] CORS (`internal/middleware/cors.go`)
  - [x] CORSMiddleware - Cross-origin handling
- [x] Rate limiting (`internal/middleware/ratelimit.go`)
  - [x] RateLimitMiddleware - General (60/min)
  - [x] AuthRateLimitMiddleware - Auth endpoints (10/min)

#### Router & Server ✅

- [x] Router (`internal/router/router.go`)
  - [x] Route organization
  - [x] Middleware chain
  - [x] Auth routes with rate limiting
  - [x] Protected routes with JWT + tenant context
- [x] Main server (`cmd/server/main.go`)
  - [x] Config loading & validation
  - [x] DB initialization
  - [x] Redis connection (optional)
  - [x] Service initialization
  - [x] Graceful shutdown

#### Configuration ✅

- [x] Config structures (`internal/config/config.go`)
  - [x] ServerConfig, DatabaseConfig, JWTConfig
  - [x] Argon2Config, SecurityConfig, EmailConfig
  - [x] CookieConfig, CORSConfig, CacheConfig
  - [x] Production validation
- [x] Environment loading (`internal/config/env.go`)
  - [x] Auth environment variables
  - [x] Default values
  - [x] Type conversion helpers
- [x] .env.example - Complete template

#### Security Features ✅

- [x] **Enhanced Dual-layer tenant isolation** (CRITICAL)
  - [x] Layer 1: GORM Callbacks with Strict Mode (automatic, enforced)
    - [x] Auto-inject WHERE tenant_id filter pada semua queries
    - [x] Configurable strict mode (ERROR vs WARNING)
    - [x] Bypass mechanism untuk system operations
    - [x] Tenant ID immutability protection
  - [x] Layer 2: Manual GORM Scopes (explicit, optional)
    - [x] TenantScope untuk explicit filtering
    - [x] Developer-level defense layer
- [x] **Brute force protection**
  - [x] Login attempt tracking
  - [x] 4-tier exponential backoff
- [x] **Token security**
  - [x] httpOnly cookies (XSS protection)
  - [x] SameSite attribute (CSRF protection)
  - [x] Token rotation

**Deliverable:** ✅ Fully working registration, login, refresh, logout system dengan production-grade security

### Phase 2: Security Hardening (Week 3) ✅ COMPLETED

**Status:** ✅ Fully Implemented (2025-12-17)

#### Rate Limiting ✅

- [x] Rate limiting middleware implementation (`internal/middleware/ratelimit.go`)
  - [x] RateLimitMiddleware - General API (60 requests/min)
  - [x] AuthRateLimitMiddleware - Auth endpoints (10 requests/min)
  - [x] Redis integration for distributed rate limiting
  - [x] Configurable limits via environment variables
- [x] Per-endpoint rate limit configuration
  - [x] Applied to /api/v1/auth routes (`router.go:72`)
  - [x] Applied to protected routes (`router.go:109`)

#### Brute Force Protection ✅

- [x] Login attempt tracking (`internal/service/auth/auth_service.go:510-614`)
  - [x] 4-tier exponential backoff system
    - [x] Tier 1: 3-4 attempts → 5 minute lockout
    - [x] Tier 2: 5-9 attempts → 15 minute lockout
    - [x] Tier 3: 10-14 attempts → 1 hour lockout
    - [x] Tier 4: 15+ attempts → 24 hour lockout
  - [x] checkLoginAttempts() with sliding window calculation
  - [x] recordLoginAttempt() for audit trail
  - [x] IP-based and email-based tracking
- [x] login_attempts table with indexes (migration 000002)

#### Password Reset Flow ✅

- [x] Password reset service methods (`internal/service/auth/auth_service.go`)
  - [x] ForgotPassword() - Generates reset token (lines 309-379)
  - [x] ResetPassword() - Validates token and updates password (lines 382-478)
  - [x] Rate limiting (max 3 requests per hour per email)
  - [x] Secure token generation (32-byte random)
  - [x] Token expiry (1 hour from .env)
  - [x] Email enumeration prevention
  - [x] Auto-revoke all refresh tokens on password change
- [x] HTTP handlers (`internal/handler/auth_handler.go`)
  - [x] ForgotPassword handler (lines 144-167)
  - [x] ResetPassword handler (lines 172-192)
  - [x] CSRF cookie clearing on reset
- [x] Routes (`internal/router/router.go`)
  - [x] POST /api/v1/auth/forgot-password (line 79)
  - [x] POST /api/v1/auth/reset-password (line 80)
- [x] password_resets table with indexes (migration 000002)

#### Input Validation & Sanitization ✅

- [x] go-playground/validator integration (`go.mod:7`)
- [x] Custom validator package (`pkg/validator/validator.go`)
  - [x] Password strength validator
  - [x] Phone number validator (Indonesian format)
  - [x] Custom validation rules
- [x] DTO validation in handlers
  - [x] LoginRequest validation
  - [x] RefreshTokenRequest validation
  - [x] PasswordResetRequest validation
  - [x] PasswordResetConfirmRequest validation
- [x] Validation error formatting (`auth_handler.go:278-311`)
  - [x] User-friendly error messages
  - [x] Field-level error details
  - [x] JSON field name conversion

#### CSRF Protection ✅

- [x] CSRF middleware (`internal/middleware/csrf.go`)
  - [x] Double Submit Cookie Pattern
  - [x] Token generation (32-byte random)
  - [x] Cookie and header validation
  - [x] Safe methods exemption (GET, HEAD, OPTIONS)
  - [x] Token rotation on login
- [x] Applied to protected routes (`router.go:112`)
- [x] CSRF token generation in login handler (`auth_handler.go:61, 241-266`)
- [x] SameSite=Strict cookie attribute

**Deliverable:** ✅ Production-grade security layer fully operational

### Phase 3: Multi-Tenant Integration (Week 4) ✅ COMPLETED

**Status:** ✅ Fully Implemented (2025-12-17)

#### Tenant Context Middleware ✅

- [x] Tenant context middleware (`internal/middleware/auth.go:191-284`)
  - [x] TenantContextMiddleware - Sets tenant session
  - [x] Subscription validation (ACTIVE/TRIAL status check)
  - [x] Trial expiry date validation
  - [x] Dual-layer tenant isolation enforcement
  - [x] GORM scopes auto-applied to all queries
- [x] Already implemented in Phase 1

#### Email Verification System ✅

- [x] Email verification service (`internal/service/auth/auth_service.go`)
  - [x] VerifyEmail() method - Token validation (24hr expiry)
  - [x] Updates user.email_verified flag
  - [x] Marks verification token as used
  - [x] Returns verified user details
- [x] Email verification handler (`internal/handler/auth_handler.go:194-220`)
  - [x] VerifyEmail handler - POST /api/v1/auth/verify-email
  - [x] Token validation
  - [x] Success response with verified email
- [x] Login enforcement (`internal/service/auth/auth_service.go`)
  - [x] Check email_verified before allowing login
  - [x] Clear error message for unverified emails
- [x] Routes (`internal/router/router.go:83`)
  - [x] POST /api/v1/auth/verify-email (public route)

#### Tenant Switching ✅

- [x] Tenant switching service (`internal/service/auth/auth_service.go`)
  - [x] SwitchTenant(userID, tenantID) method
  - [x] User-tenant relationship validation
  - [x] Tenant status validation (ACTIVE/TRIAL)
  - [x] Trial expiry check
  - [x] New access token generation with updated tenantId
  - [x] Security: Full access control validation
- [x] Tenant switching handler (`internal/handler/auth_handler.go:222-273`)
  - [x] SwitchTenant handler - POST /api/v1/auth/switch-tenant
  - [x] User ID extraction from JWT context
  - [x] Tenant info response builder
  - [x] Role information included
- [x] DTOs (`internal/dto/auth_dto.go:77-88`)
  - [x] SwitchTenantRequest
  - [x] SwitchTenantResponse
- [x] Routes (`internal/router/router.go:118`)
  - [x] POST /api/v1/auth/switch-tenant (protected)

#### User Tenant Management ✅

- [x] Get user tenants service (`internal/service/auth/auth_service.go`)
  - [x] GetUserTenants(userID) method
  - [x] Returns user_tenants relationships
  - [x] Returns full tenant details
  - [x] Includes role information per tenant
- [x] Get user tenants handler (`internal/handler/auth_handler.go:275-323`)
  - [x] GetUserTenants handler - GET /api/v1/auth/tenants
  - [x] User ID extraction from JWT
  - [x] Tenant list building with roles
  - [x] Active relationship filtering
- [x] DTOs (`internal/dto/auth_dto.go:101-104`)
  - [x] GetUserTenantsResponse
  - [x] TenantInfo structure
- [x] Routes (`internal/router/router.go:120`)
  - [x] GET /api/v1/auth/tenants (protected)

#### Current User Profile ✅

- [x] Current user handler (`internal/handler/auth_handler.go:325-400`)
  - [x] GetCurrentUser handler - GET /api/v1/auth/me
  - [x] User details retrieval
  - [x] Active tenant identification from JWT
  - [x] All accessible tenants with roles
  - [x] Comprehensive profile response
- [x] DTOs (`internal/dto/auth_dto.go:106-111`)
  - [x] CurrentUserResponse
  - [x] UserInfo structure
- [x] Routes (`internal/router/router.go:119`)
  - [x] GET /api/v1/auth/me (protected)

#### Password Change ✅

- [x] Password change service (`internal/service/auth/auth_service.go`)
  - [x] ChangePassword(userID, oldPassword, newPassword) method
  - [x] Old password verification
  - [x] Argon2id hashing for new password
  - [x] Revoke all refresh tokens (force re-login)
- [x] Password change handler (`internal/handler/auth_handler.go:402-436`)
  - [x] ChangePassword handler - POST /api/v1/auth/change-password
  - [x] User ID extraction from JWT
  - [x] Cookie clearing (refresh token + CSRF)
  - [x] Re-login prompt message
- [x] DTOs (`internal/dto/auth_dto.go:71-75`)
  - [x] ChangePasswordRequest
  - [x] Password strength validation
  - [x] Current password != new password validation
- [x] Routes (`internal/router/router.go:117`)
  - [x] POST /api/v1/auth/change-password (protected)

#### GORM Scopes for Tenant Filtering ✅

- [x] Already implemented in Phase 1
  - [x] Dual-layer tenant isolation
  - [x] GORM callbacks (automatic)
  - [x] Manual scopes (explicit)
  - [x] Tested in tenant_test.go

#### Role-Based Authorization Middleware ✅

- [x] Already implemented in Phase 2
  - [x] RequireRoleMiddleware (`internal/middleware/auth.go:286-320`)
  - [x] Multi-role support
  - [x] Tenant-specific role validation
  - [x] Used in protected routes

#### Cross-Tenant Security Tests ✅

- [x] Already implemented in Phase 2
  - [x] Tenant isolation tests (`internal/database/tenant_test.go`)
  - [x] GORM callback tests
  - [x] Manual scope tests
  - [x] Cross-tenant access prevention

#### Service Layer Enhancements ✅

- [x] DB() method added (`internal/service/auth/auth_service.go:45-47`)
  - [x] Allows handler access to database
  - [x] Used for direct queries in handlers

**Deliverable:** ✅ Complete multi-tenant isolation with email verification and user management

### Phase 4: Background Jobs & Cleanup (Week 4-5) ✅ COMPLETED

**Status:** ✅ Fully Implemented (2025-12-17)

#### Job Scheduler Infrastructure ✅

- [x] robfig/cron/v3 dependency (`go.mod`)
  - [x] Cron-based scheduler with 6-field format (second minute hour day month weekday)
  - [x] UTC timezone support to avoid DST issues
  - [x] Graceful shutdown with context-based completion waiting
- [x] Scheduler implementation (`internal/jobs/scheduler.go`)
  - [x] NewScheduler - Creates scheduler with UTC timezone and WithSeconds() option
  - [x] Start() - Registers and starts all cleanup jobs
  - [x] Stop() - Graceful shutdown returning context
  - [x] IsRunning() - Health check status
  - [x] GetLastCleanupTime() - Monitoring support

#### Cleanup Jobs Implementation ✅

- [x] Cleanup jobs (`internal/jobs/cleanup.go`)
  - [x] cleanupExpiredRefreshTokens() - Hourly at :00 (30-day retention)
  - [x] cleanupExpiredEmailVerifications() - Hourly at :05 (24-hour retention)
  - [x] cleanupExpiredPasswordResets() - Hourly at :10 (1-hour retention)
  - [x] cleanupOldLoginAttempts() - Daily at 2 AM (7-day retention)
  - [x] recoverFromPanic() - Panic recovery to prevent crashes
- [x] UTC time handling in all jobs
- [x] Structured logging (INFO/ERROR levels)
- [x] Execution time tracking
- [x] Row count reporting

#### Configuration System ✅

- [x] Job configuration (`internal/config/config.go`)
  - [x] JobConfig struct with enable/disable flag
  - [x] Configurable cron schedules per job
- [x] Environment loading (`internal/config/env.go`)
  - [x] JOB_ENABLE_CLEANUP (default: true)
  - [x] JOB_REFRESH_TOKEN_CLEANUP (default: "0 0 * * * *")
  - [x] JOB_EMAIL_CLEANUP (default: "0 5 * * * *")
  - [x] JOB_PASSWORD_CLEANUP (default: "0 10 * * * *")
  - [x] JOB_LOGIN_CLEANUP (default: "0 0 2 * * *")
- [x] .env.example documentation
  - [x] Comprehensive comments on cron format
  - [x] Testing configuration examples
  - [x] Production schedules

#### Application Integration ✅

- [x] Main server integration (`cmd/server/main.go`)
  - [x] Scheduler initialization after database setup
  - [x] Graceful shutdown sequence:
    - [x] HTTP server stop (30s timeout)
    - [x] Job scheduler stop (60s timeout)
    - [x] Database connection cleanup
- [x] Router integration (`internal/router/router.go`)
  - [x] Scheduler parameter passed to health handler
  - [x] Health check integration

#### Health Monitoring ✅

- [x] Health check enhancement (`internal/handler/health.go`)
  - [x] HealthHandler accepts scheduler parameter
  - [x] Scheduler status in /ready endpoint
  - [x] Last cleanup timestamp tracking
  - [x] Stale detection (degraded if >2 hours)
  - [x] Detailed scheduler info in JSON response
- [x] Status indicators:
  - [x] "healthy" - Scheduler running with recent cleanup
  - [x] "degraded" - No cleanup in last 2 hours
  - [x] "unhealthy" - Scheduler not running
  - [x] "not_configured" - Jobs disabled

#### Testing ✅

- [x] Comprehensive unit tests (`internal/jobs/scheduler_test.go`)
  - [x] TestNewScheduler - Scheduler initialization
  - [x] TestSchedulerStartStop - Lifecycle management
  - [x] TestSchedulerDisabled - Configuration respect
  - [x] TestCleanupExpiredRefreshTokens - Refresh token cleanup
  - [x] TestCleanupExpiredEmailVerifications - Email verification cleanup
  - [x] TestCleanupExpiredPasswordResets - Password reset cleanup
  - [x] TestCleanupOldLoginAttempts - Login attempt cleanup
  - [x] TestCleanupEmptyTables - Edge case (empty tables)
  - [x] TestGetLastCleanupTime - Timestamp tracking
  - [x] TestPanicRecovery - Panic handling
- [x] Test coverage: 90%+ across all files
- [x] All tests passing
- [x] In-memory SQLite for fast execution

#### Documentation ✅

- [x] README.md updates
  - [x] Background Jobs section (110+ lines)
  - [x] Configuration guide with examples
  - [x] Monitoring and health check documentation
  - [x] Testing instructions
  - [x] Graceful shutdown explanation
  - [x] Updated implementation status
- [x] Implementation summary document
  - [x] PHASE4-IMPLEMENTATION-COMPLETE.md

#### Production Readiness ✅

- [x] Error handling - Graceful error logging without crashes
- [x] Panic recovery - Each job recovers from panics
- [x] Performance - <100ms execution time per cleanup
- [x] Monitoring - Health check integration complete
- [x] Configuration - All schedules configurable via env vars
- [x] Graceful shutdown - Jobs complete before termination

**Deliverable:** ✅ Production-ready background job system with automated database cleanup

### Phase 5: Testing & Deployment (Week 5-6)

- Unit tests (80%+ coverage target)
- Integration tests (end-to-end flows)
- Security audit & penetration testing
- Performance optimization (database indexes, connection pooling)
- Production deployment with monitoring
- **Deliverable:** Tested, audited, deployed system

**Total Timeline:** 5-6 weeks

**Week Breakdown:**

- Week 1-2: Core Auth (Phase 1)
- Week 3: Security (Phase 2)
- Week 4: Multi-Tenant (Phase 3)
- Week 4-5: Operations (Phase 4)
- Week 5-6: Testing & Deploy (Phase 5)

**Critical Path:**

1. Phase 1 → Phase 2 → Phase 3 (sequential, each builds on previous)
2. Phase 4 can overlap with Phase 3 (week 4)
3. Phase 5 requires all previous phases complete

---

## 📚 Reference Sections

| Topic                                 | Line Numbers | File                                      |
| ------------------------------------- | ------------ | ----------------------------------------- |
| **🚨 Critical Security Requirements** | **230-920**  | **BACKEND-IMPLEMENTATION.md (this file)** |
| **🟡 Medium Priority Enhancements**   | **924-1940** | **BACKEND-IMPLEMENTATION.md (this file)** |
| Database Models                       | 88-352       | authentication-mvp-design.md              |
| Authentication Flows                  | 353-726      | authentication-mvp-design.md              |
| API Endpoints                         | 727-1097     | authentication-mvp-design.md              |
| JWT Strategy                          | 1098-1297    | authentication-mvp-design.md              |
| Security Implementation               | 1298-1841    | authentication-mvp-design.md              |
| Multi-Tenant Context                  | 1842-2196    | authentication-mvp-design.md              |
| Middleware Stack                      | 2197-2386    | authentication-mvp-design.md              |
| Error Handling                        | 2387-2547    | authentication-mvp-design.md              |
| Testing Strategy                      | 2742-3025    | authentication-mvp-design.md              |
| Deployment                            | 3026-3293    | authentication-mvp-design.md              |

---

## 🆘 Common Issues

**Q: Password hashing too slow?**  
A: Tune argon2id parameters. Reduce `ARGON2_MEMORY` to 32MB or `ARGON2_ITERATIONS` to 2 untuk development.

**Q: Token validation failing?**  
A: Check `JWT_SECRET` consistency across services. Verify token expiry times.

**Q: Cross-tenant data leakage?**  
A: Audit all queries. Use `grep -r "db.Where" .` dan verify semua include `tenant_id`.

**Q: Rate limiting too aggressive?**  
A: Adjust `RATE_LIMIT_PER_MINUTE` di environment variables.

---

**For complete implementation details, refer to: `authentication-mvp-design.md`**
