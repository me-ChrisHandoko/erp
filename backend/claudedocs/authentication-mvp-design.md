# Authentication MVP Design - Multi-Tenant ERP System

**Created:** 2025-12-16
**Version:** 1.0
**Status:** Design Phase

## Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Database Models](#database-models)
4. [Authentication Flows](#authentication-flows)
5. [API Endpoints](#api-endpoints)
6. [JWT Strategy](#jwt-strategy)
7. [Security Implementation](#security-implementation)
8. [Multi-Tenant Context](#multi-tenant-context)
9. [Frontend Integration (Redux Toolkit + RTK Query)](#frontend-integration)
10. [Middleware Stack](#middleware-stack)
11. [Error Handling](#error-handling)
12. [Testing Strategy](#testing-strategy)
13. [Deployment & Operations](#deployment--operations)
14. [Implementation Phases](#implementation-phases)

---

## Overview

This document outlines the comprehensive design for a robust, production-ready authentication system for the multi-tenant ERP backend. The system supports:

- **Multi-tenant architecture** with per-tenant role-based access control
- **JWT-based authentication** with refresh token rotation
- **Email verification** and password reset flows
- **Brute force protection** and rate limiting
- **Comprehensive audit logging** for security compliance
- **Frontend integration** with Redux Toolkit and RTK Query

### Design Goals

1. **Security First**: Multiple layers of protection against common attacks
2. **Multi-Tenant Isolation**: Complete data separation between tenants
3. **Scalability**: Support for thousands of concurrent users
4. **Developer Experience**: Clear APIs and comprehensive error handling
5. **User Experience**: Seamless authentication with automatic token refresh

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (React)                        │
│  Redux Toolkit State + RTK Query + Automatic Token Refresh  │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTPS
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (Gin)                         │
│   CORS → Rate Limiter → Logger → JWT Validator → Handler    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service Layer                              │
│  AuthService │ TokenService │ TenantService │ SecurityService│
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                            │
│   UserRepo │ RefreshTokenRepo │ TenantRepo │ AuditLogRepo   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Database (PostgreSQL/SQLite)                    │
│   Users │ UserTenants │ RefreshTokens │ AuditLogs │ etc.   │
└─────────────────────────────────────────────────────────────┘
```

### Architecture Patterns

- **Controller → Service → Repository** pattern for clear separation of concerns
- **Dependency Injection** for testability and flexibility
- **Interface-based design** for easy mocking and testing
- **Middleware chain** for cross-cutting concerns (auth, logging, rate limiting)

---

## Database Models

### Existing Models (Already Implemented)

#### User Model
```go
type User struct {
    ID            string    `gorm:"type:varchar(255);primaryKey"`
    Email         string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    Username      string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    Password      string    `gorm:"type:varchar(255);not null"` // Argon2id hashed
    Name          string    `gorm:"type:varchar(255);not null"`
    IsSystemAdmin bool      `gorm:"default:false"`
    IsActive      bool      `gorm:"default:true"`
    CreatedAt     time.Time `gorm:"autoCreateTime"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime"`

    Tenants []UserTenant `gorm:"foreignKey:UserID"`
}
```

#### UserTenant Model (Junction Table)
```go
type UserTenant struct {
    ID        string    `gorm:"type:varchar(255);primaryKey"`
    UserID    string    `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_user_tenant"`
    TenantID  string    `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_user_tenant"`
    Role      UserRole  `gorm:"type:varchar(20);default:'STAFF';index"`
    IsActive  bool      `gorm:"default:true"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`

    User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
    Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}
```

#### UserRole Enum
```go
type UserRole string

const (
    UserRoleOwner     UserRole = "OWNER"     // Full access to tenant
    UserRoleAdmin     UserRole = "ADMIN"     // Administrative access
    UserRoleFinance   UserRole = "FINANCE"   // Finance module access
    UserRoleSales     UserRole = "SALES"     // Sales module access
    UserRoleWarehouse UserRole = "WAREHOUSE" // Warehouse module access
    UserRoleStaff     UserRole = "STAFF"     // Read-only access
)
```

### New Models Required for Authentication

#### 1. RefreshToken Model
Stores refresh tokens for secure session management with revocation support.

```go
type RefreshToken struct {
    ID        string     `gorm:"type:varchar(255);primaryKey"`
    UserID    string     `gorm:"type:varchar(255);not null;index"`
    Token     string     `gorm:"type:text;not null;uniqueIndex"`
    ExpiresAt time.Time  `gorm:"type:datetime;not null;index"`
    IsRevoked bool       `gorm:"default:false;index"`
    IPAddress *string    `gorm:"type:varchar(45)"`
    UserAgent *string    `gorm:"type:varchar(500)"`
    CreatedAt time.Time  `gorm:"autoCreateTime"`
    RevokedAt *time.Time `gorm:"type:datetime"`

    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (RefreshToken) TableName() string {
    return "refresh_tokens"
}

// BeforeCreate hook to generate CUID
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
    if rt.ID == "" {
        rt.ID = cuid.New()
    }
    return nil
}
```

**Key Features:**
- Stores hashed refresh token for security
- Tracks IP address and user agent for security auditing
- Supports revocation (logout, password change)
- Automatic cleanup of expired tokens via background job

**Indexes:**
- `user_id` - Fast lookup of user's tokens
- `token` - Unique constraint + fast validation
- `expires_at` - Efficient cleanup queries
- `is_revoked` - Quick filtering of valid tokens

#### 2. EmailVerification Model
Manages email verification tokens for new user accounts.

```go
type EmailVerification struct {
    ID        string     `gorm:"type:varchar(255);primaryKey"`
    UserID    string     `gorm:"type:varchar(255);not null;index"`
    Token     string     `gorm:"type:varchar(255);not null;uniqueIndex"`
    ExpiresAt time.Time  `gorm:"type:datetime;not null;index"`
    IsUsed    bool       `gorm:"default:false"`
    UsedAt    *time.Time `gorm:"type:datetime"`
    CreatedAt time.Time  `gorm:"autoCreateTime"`

    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (EmailVerification) TableName() string {
    return "email_verifications"
}

// BeforeCreate hook to generate CUID
func (ev *EmailVerification) BeforeCreate(tx *gorm.DB) error {
    if ev.ID == "" {
        ev.ID = cuid.New()
    }
    return nil
}
```

**Key Features:**
- 24-hour expiry for verification tokens
- One-time use tokens
- Tracks when verification was completed
- Cascade deletes when user is deleted

**Indexes:**
- `user_id` - Lookup user's verification tokens
- `token` - Unique constraint + fast verification
- `expires_at` - Cleanup expired tokens

#### 3. PasswordReset Model
Manages password reset tokens for account recovery.

```go
type PasswordReset struct {
    ID        string     `gorm:"type:varchar(255);primaryKey"`
    UserID    string     `gorm:"type:varchar(255);not null;index"`
    Token     string     `gorm:"type:varchar(255);not null;uniqueIndex"`
    ExpiresAt time.Time  `gorm:"type:datetime;not null;index"`
    IsUsed    bool       `gorm:"default:false"`
    UsedAt    *time.Time `gorm:"type:datetime"`
    IPAddress *string    `gorm:"type:varchar(45)"`
    CreatedAt time.Time  `gorm:"autoCreateTime"`

    User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (PasswordReset) TableName() string {
    return "password_resets"
}

// BeforeCreate hook to generate CUID
func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
    if pr.ID == "" {
        pr.ID = cuid.New()
    }
    return nil
}
```

**Key Features:**
- 1-hour expiry for reset tokens (shorter for security)
- One-time use tokens
- Tracks IP address for security auditing
- All user tokens invalidated on password change

**Indexes:**
- `user_id` - Lookup user's reset tokens
- `token` - Unique constraint + fast validation
- `expires_at` - Cleanup expired tokens

#### 4. LoginAttempt Model
Tracks login attempts for brute force protection.

```go
type LoginAttempt struct {
    ID         string    `gorm:"type:varchar(255);primaryKey"`
    Email      string    `gorm:"type:varchar(255);not null;index"`
    IPAddress  string    `gorm:"type:varchar(45);not null;index"`
    UserAgent  *string   `gorm:"type:varchar(500)"`
    IsSuccess  bool      `gorm:"default:false;index"`
    FailReason *string   `gorm:"type:varchar(255)"`
    CreatedAt  time.Time `gorm:"autoCreateTime;index"`
}

// TableName specifies the table name
func (LoginAttempt) TableName() string {
    return "login_attempts"
}

// BeforeCreate hook to generate CUID
func (la *LoginAttempt) BeforeCreate(tx *gorm.DB) error {
    if la.ID == "" {
        la.ID = cuid.New()
    }
    return nil
}
```

**Key Features:**
- Tracks both successful and failed login attempts
- Composite tracking by email + IP address
- Stores failure reason for security analysis
- Used for account lockout mechanism

**Indexes:**
- `email` - Lookup attempts for specific email
- `ip_address` - Track attempts from specific IP
- `is_success` - Filter failed attempts
- `created_at` - Time-based queries for lockout logic

**Composite Index:**
```go
// Add to migration
tx.Exec("CREATE INDEX idx_login_attempts_email_ip_created ON login_attempts(email, ip_address, created_at)")
```

### Database Migration Strategy

**Migration File Structure:**
```
db/migrations/
  ├── 001_create_users.up.sql
  ├── 002_create_tenants.up.sql
  ├── 003_create_user_tenants.up.sql
  ├── 004_create_refresh_tokens.up.sql        # New
  ├── 005_create_email_verifications.up.sql   # New
  ├── 006_create_password_resets.up.sql       # New
  └── 007_create_login_attempts.up.sql        # New
```

**Sample Migration (RefreshTokens):**
```sql
-- 004_create_refresh_tokens.up.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
```

---

## Authentication Flows

### 1. User Registration Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/register                                   │
   │ {email, username, password, name}                     │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                       Validate Input  │
   │                                    Check Email Unique │
   │                                       Hash Password   │
   │                                         Create User   │
   │                                      (IsActive=false) │
   │                              Generate Verification    │
   │                                            Token      │
   │                                  Send Verification    │
   │                                             Email     │
   │                                                       │
   │ 201 Created                                           │
   │ {message: "Check email for verification"}            │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "message": "Registration successful. Please check your email to verify your account.",
  "email": "user@example.com"
}
```

**Validation Rules:**
- Email: Valid format (RFC 5322), lowercase normalized, max 255 chars
- Username: 3-50 chars, alphanumeric + underscore/hyphen only
- Password: Min 8 chars, 1 uppercase, 1 lowercase, 1 number
- Name: 2-255 chars

**Error Responses:**
- `400` - Validation errors (weak password, invalid email format)
- `409` - Email or username already exists
- `500` - Server error

### 2. Email Verification Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ GET /auth/verify-email?token=xyz123                  │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                        Validate Token │
   │                                         Check Expiry  │
   │                                     Check Not Used    │
   │                                  Mark User IsActive   │
   │                                      Mark Token Used  │
   │                              Create AuditLog Entry    │
   │                                                       │
   │ 200 OK                                                │
   │ {message: "Email verified successfully"}             │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/verify-email`

**Request Body:**
```json
{
  "token": "abc123xyz789..."
}
```

**Response (200 OK):**
```json
{
  "message": "Email verified successfully. You can now log in.",
  "redirectUrl": "/login"
}
```

**Error Responses:**
- `400` - Invalid or expired token
- `409` - Email already verified
- `404` - Token not found

### 3. Login Flow (Multi-Tenant)

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/login                                      │
   │ {email, password}                                     │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                        Validate Input │
   │                                     Find User by Email│
   │                                    Check IsActive=true│
   │                                      Verify Password  │
   │                                Record Login Attempt   │
   │                                   Fetch UserTenants   │
   │                         Check At Least 1 Active Tenant│
   │                                  Generate Access JWT  │
   │                                 Generate Refresh JWT  │
   │                              Store RefreshToken in DB │
   │                              Create AuditLog Entry    │
   │                                                       │
   │ 200 OK                                                │
   │ {accessToken, refreshToken, user, tenants}           │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "ckx1a2b3c4d5e6f7g8h9i0",
    "email": "user@example.com",
    "name": "John Doe",
    "username": "johndoe",
    "isSystemAdmin": false
  },
  "tenants": [
    {
      "tenantId": "ckx1tenant123456789",
      "role": "OWNER",
      "companyName": "PT Example Indonesia",
      "status": "ACTIVE"
    },
    {
      "tenantId": "ckx1tenant987654321",
      "role": "ADMIN",
      "companyName": "CV Demo Distribusi",
      "status": "TRIAL"
    }
  ],
  "activeTenant": {
    "tenantId": "ckx1tenant123456789",
    "role": "OWNER"
  }
}
```

**Error Responses:**
- `400` - Invalid credentials
- `401` - Email not verified
- `403` - Account locked (too many failed attempts)
- `403` - Account inactive
- `403` - No active tenants available
- `500` - Server error

**Security Features:**
- Password verification using argon2id
- Failed attempt tracking per email + IP
- Account lockout after 5 failed attempts in 15 minutes
- Audit log entry for all login attempts
- Refresh token stored with IP and user agent

### 4. Token Refresh Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/refresh                                    │
   │ {refreshToken}                                        │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                        Validate Token │
   │                                  Check Token in DB    │
   │                                    Check Not Revoked  │
   │                                         Check Expiry  │
   │                              Generate New Access JWT  │
   │                             Generate New Refresh JWT  │
   │                                   (Token Rotation)    │
   │                               Revoke Old Refresh Token│
   │                              Store New RefreshToken   │
   │                                                       │
   │ 200 OK                                                │
   │ {accessToken, refreshToken}                          │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
- `401` - Invalid refresh token
- `401` - Token expired
- `401` - Token revoked
- `404` - Token not found in database

**Security Features:**
- Refresh token rotation: Old token invalidated, new token issued
- Database validation: Token must exist and not be revoked
- Prevents token reuse attacks

### 5. Tenant Switching Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/switch-tenant                              │
   │ Authorization: Bearer <accessToken>                   │
   │ {tenantId}                                            │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                     Validate JWT      │
   │                                  Extract User ID      │
   │                            Validate User Has Access   │
   │                                      to Tenant        │
   │                                  Check Tenant Status  │
   │                            Check Subscription Valid   │
   │                        Generate New Access JWT with   │
   │                                 New TenantID + Role   │
   │                              Create AuditLog Entry    │
   │                                                       │
   │ 200 OK                                                │
   │ {accessToken, activeTenant}                          │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/switch-tenant`

**Request Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{
  "tenantId": "ckx1tenant987654321"
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "activeTenant": {
    "tenantId": "ckx1tenant987654321",
    "role": "ADMIN",
    "companyName": "CV Demo Distribusi",
    "status": "TRIAL"
  }
}
```

**Error Responses:**
- `401` - Unauthorized (invalid or expired access token)
- `403` - User does not have access to requested tenant
- `403` - Tenant is suspended or expired
- `404` - Tenant not found

**Security Features:**
- Validates user has active UserTenant relationship
- Checks tenant subscription status
- Regenerates access token with new tenant context
- Refresh token remains valid (no need to re-login)
- Audit log entry for tenant switches

### 6. Password Reset Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/forgot-password                            │
   │ {email}                                               │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                    Find User by Email │
   │                              Generate Reset Token     │
   │                                   Store in DB         │
   │                             Send Reset Email          │
   │                                                       │
   │ 200 OK                                                │
   │ {message: "Reset email sent"}                        │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
   │ ... User clicks email link ...                        │
   │                                                       │
   │ POST /auth/reset-password                             │
   │ {token, newPassword}                                  │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                        Validate Token │
   │                                         Check Expiry  │
   │                                     Check Not Used    │
   │                                 Validate New Password │
   │                                      Hash Password    │
   │                                      Update User      │
   │                                      Mark Token Used  │
   │                           Revoke All Refresh Tokens   │
   │                              Create AuditLog Entry    │
   │                                                       │
   │ 200 OK                                                │
   │ {message: "Password reset successfully"}             │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
```

**Endpoint 1:** `POST /api/v1/auth/forgot-password`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

**Note:** Always return success message (even if email doesn't exist) to prevent email enumeration attacks.

**Endpoint 2:** `POST /api/v1/auth/reset-password`

**Request Body:**
```json
{
  "token": "abc123xyz789...",
  "newPassword": "NewSecurePass456!"
}
```

**Response (200 OK):**
```json
{
  "message": "Password reset successfully. Please log in with your new password.",
  "redirectUrl": "/login"
}
```

**Error Responses:**
- `400` - Invalid or expired token
- `400` - Weak password (doesn't meet requirements)
- `404` - Token not found
- `409` - Token already used

**Security Features:**
- 1-hour token expiry (shorter than email verification)
- One-time use tokens
- All existing sessions invalidated (refresh tokens revoked)
- Audit log entry with IP address
- Rate limiting on forgot-password endpoint (3 requests per hour per email)

### 7. Logout Flow

```
┌──────┐                                              ┌──────────┐
│Client│                                              │  Server  │
└──┬───┘                                              └────┬─────┘
   │                                                       │
   │ POST /auth/logout                                     │
   │ Authorization: Bearer <accessToken>                   │
   │ {refreshToken}                                        │
   ├──────────────────────────────────────────────────────>│
   │                                                       │
   │                                     Validate JWT      │
   │                                Extract User ID        │
   │                              Find RefreshToken        │
   │                                  Mark as Revoked      │
   │                              Create AuditLog Entry    │
   │                                                       │
   │ 200 OK                                                │
   │ {message: "Logged out successfully"}                 │
   │<──────────────────────────────────────────────────────┤
   │                                                       │
   │ Clear tokens from Redux state                         │
   │                                                       │
```

**Endpoint:** `POST /api/v1/auth/logout`

**Request Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

**Security Features:**
- Revokes refresh token in database
- Access token becomes invalid after expiry (cannot be revoked, short-lived by design)
- Audit log entry for logout
- Frontend clears all tokens from state

---

## API Endpoints

### Authentication Endpoints Summary

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/auth/register` | POST | No | User registration |
| `/api/v1/auth/verify-email` | POST | No | Verify email with token |
| `/api/v1/auth/resend-verification` | POST | Yes | Resend verification email |
| `/api/v1/auth/login` | POST | No | User login |
| `/api/v1/auth/logout` | POST | Yes | User logout |
| `/api/v1/auth/refresh` | POST | No | Refresh access token |
| `/api/v1/auth/me` | GET | Yes | Get current user info |
| `/api/v1/auth/change-password` | POST | Yes | Change password (authenticated) |
| `/api/v1/auth/forgot-password` | POST | No | Request password reset |
| `/api/v1/auth/reset-password` | POST | No | Reset password with token |
| `/api/v1/auth/switch-tenant` | POST | Yes | Switch active tenant |
| `/api/v1/auth/tenants` | GET | Yes | Get user's tenants |

### Detailed API Specifications

#### POST /api/v1/auth/register

**Request:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "message": "Registration successful. Please check your email to verify your account.",
  "email": "user@example.com"
}
```

**Errors:**
- `400` - Validation error (invalid email, weak password)
- `409` - Email or username already exists

---

#### POST /api/v1/auth/login

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "user": {
    "id": "ckx123...",
    "email": "user@example.com",
    "name": "John Doe",
    "username": "johndoe",
    "isSystemAdmin": false
  },
  "tenants": [
    {
      "tenantId": "ckx456...",
      "role": "OWNER",
      "companyName": "PT Example",
      "status": "ACTIVE"
    }
  ],
  "activeTenant": {
    "tenantId": "ckx456...",
    "role": "OWNER"
  }
}
```

**Errors:**
- `400` - Invalid credentials
- `401` - Email not verified
- `403` - Account locked
- `403` - Account inactive
- `403` - No active tenants

---

#### POST /api/v1/auth/refresh

**Request:**
```json
{
  "refreshToken": "eyJhbGc..."
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc..."
}
```

**Errors:**
- `401` - Invalid or expired refresh token
- `401` - Token revoked

---

#### GET /api/v1/auth/me

**Request Headers:**
```
Authorization: Bearer eyJhbGc...
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "ckx123...",
    "email": "user@example.com",
    "name": "John Doe",
    "username": "johndoe",
    "isSystemAdmin": false
  },
  "activeTenant": {
    "tenantId": "ckx456...",
    "role": "OWNER",
    "companyName": "PT Example",
    "status": "ACTIVE"
  },
  "availableTenants": [
    {
      "tenantId": "ckx456...",
      "role": "OWNER",
      "companyName": "PT Example",
      "status": "ACTIVE"
    }
  ]
}
```

---

#### POST /api/v1/auth/switch-tenant

**Request Headers:**
```
Authorization: Bearer eyJhbGc...
```

**Request:**
```json
{
  "tenantId": "ckx789..."
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGc...",
  "activeTenant": {
    "tenantId": "ckx789...",
    "role": "ADMIN",
    "companyName": "CV Demo",
    "status": "TRIAL"
  }
}
```

**Errors:**
- `403` - User does not have access to tenant
- `403` - Tenant suspended or expired

---

#### POST /api/v1/auth/forgot-password

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

---

#### POST /api/v1/auth/reset-password

**Request:**
```json
{
  "token": "abc123...",
  "newPassword": "NewSecure456!"
}
```

**Response (200 OK):**
```json
{
  "message": "Password reset successfully. Please log in with your new password.",
  "redirectUrl": "/login"
}
```

**Errors:**
- `400` - Invalid or expired token
- `400` - Weak password

---

## JWT Strategy

### Token Types

#### 1. Access Token
**Purpose:** Short-lived token for API authorization
**Lifetime:** 30 minutes
**Storage:** Frontend memory (Redux state)
**Algorithm:** HS256 or RS256

**Payload Structure:**
```json
{
  "sub": "ckx1user123456789",        // User ID (subject)
  "email": "user@example.com",
  "name": "John Doe",
  "tid": "ckx1tenant123456789",      // Tenant ID
  "role": "OWNER",                    // Role in active tenant
  "sys_admin": false,                 // Is system admin
  "iat": 1702800000,                  // Issued at (Unix timestamp)
  "exp": 1702801800,                  // Expires at (30 min later)
  "type": "access"                    // Token type
}
```

**Security Features:**
- Short expiry reduces damage from token theft
- Contains tenant context for authorization
- Cannot be revoked (by design, short-lived)
- Signed with secure secret (min 32 bytes)

#### 2. Refresh Token
**Purpose:** Long-lived token for obtaining new access tokens
**Lifetime:** 30 days
**Storage:** httpOnly cookie (recommended) or localStorage
**Algorithm:** HS256 or RS256

**Payload Structure:**
```json
{
  "sub": "ckx1user123456789",        // User ID
  "jti": "ckx1refresh987654321",     // Token ID (for revocation)
  "iat": 1702800000,                  // Issued at
  "exp": 1705392000,                  // Expires at (30 days later)
  "type": "refresh"                   // Token type
}
```

**Security Features:**
- Stored in database with revocation support
- Token rotation on each refresh
- Tracks IP address and user agent
- Revoked on logout and password change
- Long expiry for better UX (stay logged in)

### Token Generation

**Access Token Generation:**
```go
func GenerateAccessToken(user *User, tenant *UserTenant) (string, error) {
    claims := jwt.MapClaims{
        "sub":       user.ID,
        "email":     user.Email,
        "name":      user.Name,
        "tid":       tenant.TenantID,
        "role":      tenant.Role,
        "sys_admin": user.IsSystemAdmin,
        "iat":       time.Now().Unix(),
        "exp":       time.Now().Add(30 * time.Minute).Unix(),
        "type":      "access",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

**Refresh Token Generation:**
```go
func GenerateRefreshToken(user *User, ipAddress, userAgent string) (string, error) {
    tokenID := cuid.New()

    claims := jwt.MapClaims{
        "sub":  user.ID,
        "jti":  tokenID,
        "iat":  time.Now().Unix(),
        "exp":  time.Now().Add(30 * 24 * time.Hour).Unix(),
        "type": "refresh",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        return "", err
    }

    // Store in database
    refreshToken := &RefreshToken{
        ID:        tokenID,
        UserID:    user.ID,
        Token:     tokenString,
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
        IPAddress: &ipAddress,
        UserAgent: &userAgent,
    }

    if err := db.Create(refreshToken).Error; err != nil {
        return "", err
    }

    return tokenString, nil
}
```

### Token Validation

**Access Token Validation Middleware:**
```go
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Parse and validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Validate signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Extract claims
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(401, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Verify token type
        if claims["type"] != "access" {
            c.JSON(401, gin.H{"error": "Invalid token type"})
            c.Abort()
            return
        }

        // Inject claims into context
        c.Set("userID", claims["sub"])
        c.Set("tenantID", claims["tid"])
        c.Set("role", claims["role"])
        c.Set("isSystemAdmin", claims["sys_admin"])

        c.Next()
    }
}
```

**Refresh Token Validation:**
```go
func ValidateRefreshToken(tokenString string) (*RefreshToken, error) {
    // Parse JWT
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(os.Getenv("JWT_SECRET")), nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid or expired token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || claims["type"] != "refresh" {
        return nil, errors.New("invalid token type")
    }

    // Check database for revocation
    var refreshToken RefreshToken
    if err := db.Where("token = ? AND is_revoked = ?", tokenString, false).
        First(&refreshToken).Error; err != nil {
        return nil, errors.New("token not found or revoked")
    }

    // Check expiry
    if refreshToken.ExpiresAt.Before(time.Now()) {
        return nil, errors.New("token expired")
    }

    return &refreshToken, nil
}
```

### Token Refresh Strategy

**Refresh Token Rotation:**
```go
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
    // Validate refresh token
    oldToken, err := ValidateRefreshToken(refreshTokenString)
    if err != nil {
        return nil, err
    }

    // Get user
    var user User
    if err := db.First(&user, "id = ?", oldToken.UserID).Error; err != nil {
        return nil, err
    }

    // Get user's tenants (use first active tenant as default)
    var userTenant UserTenant
    if err := db.Where("user_id = ? AND is_active = ?", user.ID, true).
        First(&userTenant).Error; err != nil {
        return nil, err
    }

    // Generate new access token
    accessToken, err := GenerateAccessToken(&user, &userTenant)
    if err != nil {
        return nil, err
    }

    // Generate new refresh token (rotation)
    newRefreshToken, err := GenerateRefreshToken(&user,
        *oldToken.IPAddress, *oldToken.UserAgent)
    if err != nil {
        return nil, err
    }

    // Revoke old refresh token
    oldToken.IsRevoked = true
    revokedAt := time.Now()
    oldToken.RevokedAt = &revokedAt
    db.Save(oldToken)

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
    }, nil
}
```

---

## Security Implementation

### 1. Password Security

**Hashing Algorithm: Argon2id**
- Memory cost: 64 MB (65536 KiB)
- Time cost: 3 iterations
- Parallelism: 4 threads
- Salt length: 16 bytes (automatically generated)
- Key length: 32 bytes
- Resistant to GPU/ASIC attacks and side-channel attacks

**Implementation:**
```go
import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "errors"
    "fmt"
    "strings"

    "golang.org/x/crypto/argon2"
)

// Argon2id parameters
type Argon2Params struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

// Default parameters (can be configured via environment)
var defaultParams = &Argon2Params{
    Memory:      64 * 1024, // 64 MB
    Iterations:  3,
    Parallelism: 4,
    SaltLength:  16,
    KeyLength:   32,
}

// Hash password during registration
func HashPassword(password string) (string, error) {
    // Generate random salt
    salt := make([]byte, defaultParams.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    // Generate hash using argon2id
    hash := argon2.IDKey(
        []byte(password),
        salt,
        defaultParams.Iterations,
        defaultParams.Memory,
        defaultParams.Parallelism,
        defaultParams.KeyLength,
    )

    // Encode to base64 for storage
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)

    // Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
    encodedHash := fmt.Sprintf(
        "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        defaultParams.Memory,
        defaultParams.Iterations,
        defaultParams.Parallelism,
        b64Salt,
        b64Hash,
    )

    return encodedHash, nil
}

// Verify password during login
func VerifyPassword(encodedHash, password string) error {
    // Extract parameters and hash from encoded string
    parts := strings.Split(encodedHash, "$")
    if len(parts) != 6 {
        return errors.New("invalid hash format")
    }

    // Parse parameters
    var memory, iterations uint32
    var parallelism uint8
    _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
    if err != nil {
        return err
    }

    // Decode salt and hash
    salt, err := base64.RawStdEncoding.DecodeString(parts[4])
    if err != nil {
        return err
    }

    hash, err := base64.RawStdEncoding.DecodeString(parts[5])
    if err != nil {
        return err
    }

    // Generate hash from provided password
    keyLength := uint32(len(hash))
    comparisonHash := argon2.IDKey(
        []byte(password),
        salt,
        iterations,
        memory,
        parallelism,
        keyLength,
    )

    // Constant-time comparison to prevent timing attacks
    if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
        return nil
    }

    return errors.New("invalid password")
}
```

**Password Requirements:**
```go
type PasswordValidator struct {
    MinLength      int
    RequireUpper   bool
    RequireLower   bool
    RequireNumber  bool
    RequireSpecial bool
}

func ValidatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }

    var hasUpper, hasLower, hasNumber bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        }
    }

    if !hasUpper {
        return errors.New("password must contain at least one uppercase letter")
    }
    if !hasLower {
        return errors.New("password must contain at least one lowercase letter")
    }
    if !hasNumber {
        return errors.New("password must contain at least one number")
    }

    return nil
}
```

### 2. Brute Force Protection

**Failed Login Tracking:**
```go
func RecordLoginAttempt(email, ipAddress, userAgent string, success bool, failReason *string) error {
    attempt := &LoginAttempt{
        Email:      email,
        IPAddress:  ipAddress,
        UserAgent:  &userAgent,
        IsSuccess:  success,
        FailReason: failReason,
    }

    return db.Create(attempt).Error
}

func IsAccountLocked(email, ipAddress string) (bool, error) {
    // Count failed attempts in last 15 minutes
    fifteenMinAgo := time.Now().Add(-15 * time.Minute)

    var count int64
    err := db.Model(&LoginAttempt{}).
        Where("email = ? AND ip_address = ? AND is_success = ? AND created_at > ?",
            email, ipAddress, false, fifteenMinAgo).
        Count(&count).Error

    if err != nil {
        return false, err
    }

    // Lock after 5 failed attempts
    return count >= 5, nil
}

func GetLockoutDuration(failedAttempts int) time.Duration {
    // Exponential backoff
    switch {
    case failedAttempts < 5:
        return 0
    case failedAttempts < 10:
        return 5 * time.Minute
    case failedAttempts < 15:
        return 15 * time.Minute
    case failedAttempts < 20:
        return 1 * time.Hour
    default:
        return 24 * time.Hour
    }
}
```

**Login Flow with Brute Force Protection:**
```go
func Login(email, password, ipAddress, userAgent string) (*LoginResponse, error) {
    // Check if account is locked
    locked, err := IsAccountLocked(email, ipAddress)
    if err != nil {
        return nil, err
    }

    if locked {
        failReason := "Account temporarily locked due to too many failed attempts"
        RecordLoginAttempt(email, ipAddress, userAgent, false, &failReason)
        return nil, errors.New("AUTH_ACCOUNT_LOCKED")
    }

    // Find user
    var user User
    if err := db.Where("email = ?", email).First(&user).Error; err != nil {
        failReason := "Invalid credentials"
        RecordLoginAttempt(email, ipAddress, userAgent, false, &failReason)
        return nil, errors.New("AUTH_INVALID_CREDENTIALS")
    }

    // Verify password
    if err := VerifyPassword(user.Password, password); err != nil {
        failReason := "Invalid credentials"
        RecordLoginAttempt(email, ipAddress, userAgent, false, &failReason)
        return nil, errors.New("AUTH_INVALID_CREDENTIALS")
    }

    // Check if email is verified
    if !user.IsActive {
        failReason := "Email not verified"
        RecordLoginAttempt(email, ipAddress, userAgent, false, &failReason)
        return nil, errors.New("AUTH_EMAIL_NOT_VERIFIED")
    }

    // Record successful login
    RecordLoginAttempt(email, ipAddress, userAgent, true, nil)

    // Continue with token generation...
}
```

### 3. Rate Limiting

**Implementation using go-rate-limit:**
```go
import "github.com/gin-contrib/limiter"

func RateLimitMiddleware() gin.HandlerFunc {
    // Create limiter with different rates for different endpoints
    return func(c *gin.Context) {
        path := c.Request.URL.Path

        var limit rate.Limit
        var burst int

        switch {
        case strings.Contains(path, "/auth/login"):
            // 5 requests per minute for login
            limit = rate.Every(time.Minute / 5)
            burst = 5
        case strings.Contains(path, "/auth/register"):
            // 3 requests per hour for registration
            limit = rate.Every(time.Hour / 3)
            burst = 3
        case strings.Contains(path, "/auth/forgot-password"):
            // 3 requests per hour for password reset
            limit = rate.Every(time.Hour / 3)
            burst = 3
        default:
            // 100 requests per minute for other endpoints
            limit = rate.Every(time.Minute / 100)
            burst = 100
        }

        // Get client identifier (IP address)
        clientIP := c.ClientIP()

        // Check rate limit (implementation depends on chosen library)
        // If exceeded, return 429 Too Many Requests
    }
}
```

### 4. Token Security

**Secure Token Generation:**
```go
import "crypto/rand"
import "encoding/base64"

func GenerateSecureToken(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}
```

**Token Expiry Configuration:**
```go
const (
    AccessTokenExpiry         = 30 * time.Minute
    RefreshTokenExpiry        = 30 * 24 * time.Hour  // 30 days
    EmailVerificationExpiry   = 24 * time.Hour
    PasswordResetExpiry       = 1 * time.Hour
)
```

### 5. CSRF Protection

**For Cookie-Based Refresh Tokens:**
```go
func SetRefreshTokenCookie(c *gin.Context, token string) {
    c.SetCookie(
        "refresh_token",           // name
        token,                      // value
        30*24*60*60,               // max age (30 days)
        "/",                        // path
        "",                         // domain (empty for same domain)
        true,                       // secure (HTTPS only)
        true,                       // httpOnly (no JavaScript access)
    )
}

func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verify CSRF token for state-changing operations
        if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
            csrfToken := c.GetHeader("X-CSRF-Token")
            // Validate CSRF token (implementation depends on chosen library)
        }
        c.Next()
    }
}
```

### 6. Audit Logging

**Comprehensive Audit Trail:**
```go
func CreateAuditLog(tenantID, userID *string, action, entityType, entityID string,
    oldValues, newValues interface{}, ipAddress, userAgent string) error {

    oldJSON, _ := json.Marshal(oldValues)
    newJSON, _ := json.Marshal(newValues)

    oldStr := string(oldJSON)
    newStr := string(newJSON)

    log := &AuditLog{
        TenantID:   tenantID,
        UserID:     userID,
        Action:     action,
        EntityType: &entityType,
        EntityID:   &entityID,
        OldValues:  &oldStr,
        NewValues:  &newStr,
        IPAddress:  &ipAddress,
        UserAgent:  &userAgent,
    }

    return db.Create(log).Error
}

// Usage examples
func LogUserLogin(user *User, tenantID, ipAddress, userAgent string) {
    CreateAuditLog(
        &tenantID,
        &user.ID,
        "LOGIN",
        "User",
        user.ID,
        nil,
        map[string]interface{}{
            "email": user.Email,
            "tenantID": tenantID,
        },
        ipAddress,
        userAgent,
    )
}

func LogPasswordChange(user *User, ipAddress, userAgent string) {
    CreateAuditLog(
        nil,
        &user.ID,
        "PASSWORD_CHANGE",
        "User",
        user.ID,
        nil,
        map[string]interface{}{
            "email": user.Email,
            "timestamp": time.Now(),
        },
        ipAddress,
        userAgent,
    )
}

func LogTenantSwitch(user *User, fromTenantID, toTenantID, ipAddress, userAgent string) {
    CreateAuditLog(
        &toTenantID,
        &user.ID,
        "TENANT_SWITCH",
        "User",
        user.ID,
        map[string]interface{}{"fromTenant": fromTenantID},
        map[string]interface{}{"toTenant": toTenantID},
        ipAddress,
        userAgent,
    )
}
```

---

## Multi-Tenant Context

### Tenant Isolation Principles

**CRITICAL SECURITY RULE:**
Every database query on transactional data MUST include `tenantID` filter to prevent cross-tenant data leakage.

### Tenant Context Middleware

**Middleware Implementation:**
```go
func TenantContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract tenant ID from JWT (set by JWTAuthMiddleware)
        tenantID, exists := c.Get("tenantID")
        if !exists {
            c.JSON(401, gin.H{"error": "Tenant context missing"})
            c.Abort()
            return
        }

        userID, _ := c.Get("userID")

        // Validate user has access to this tenant
        var userTenant UserTenant
        err := db.Where("user_id = ? AND tenant_id = ? AND is_active = ?",
            userID, tenantID, true).First(&userTenant).Error

        if err != nil {
            c.JSON(403, gin.H{"error": "Access denied to this tenant"})
            c.Abort()
            return
        }

        // Get tenant details
        var tenant Tenant
        err = db.Preload("Company").First(&tenant, "id = ?", tenantID).Error
        if err != nil {
            c.JSON(404, gin.H{"error": "Tenant not found"})
            c.Abort()
            return
        }

        // Check tenant status
        if tenant.Status != TenantStatusActive && tenant.Status != TenantStatusTrial {
            c.JSON(403, gin.H{"error": "Tenant is suspended or expired"})
            c.Abort()
            return
        }

        // Check trial expiry
        if tenant.Status == TenantStatusTrial &&
           tenant.TrialEndsAt != nil &&
           tenant.TrialEndsAt.Before(time.Now()) {
            c.JSON(403, gin.H{"error": "Trial period has expired"})
            c.Abort()
            return
        }

        // Check subscription validity
        if tenant.SubscriptionID != nil {
            var subscription Subscription
            if err := db.First(&subscription, "id = ?", tenant.SubscriptionID).Error; err == nil {
                if subscription.Status == SubscriptionStatusExpired ||
                   subscription.Status == SubscriptionStatusCancelled {
                    c.JSON(403, gin.H{"error": "Subscription is not active"})
                    c.Abort()
                    return
                }

                // Check grace period
                if subscription.Status == SubscriptionStatusPastDue {
                    if subscription.GracePeriodEnds != nil &&
                       subscription.GracePeriodEnds.Before(time.Now()) {
                        c.JSON(403, gin.H{"error": "Subscription payment overdue"})
                        c.Abort()
                        return
                    }
                }
            }
        }

        // Inject tenant context
        c.Set("tenant", tenant)
        c.Set("userTenantRole", userTenant.Role)

        c.Next()
    }
}
```

### Query Pattern Enforcement

**CORRECT: Always filter by tenantID**
```go
// Get all products for current tenant
func GetProducts(c *gin.Context) {
    tenantID, _ := c.Get("tenantID")

    var products []Product
    err := db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
        Find(&products).Error

    // ... handle error and response
}

// Get specific product for current tenant
func GetProduct(c *gin.Context) {
    tenantID, _ := c.Get("tenantID")
    productID := c.Param("id")

    var product Product
    err := db.Where("tenant_id = ? AND id = ?", tenantID, productID).
        First(&product).Error

    // ... handle error and response
}
```

**WRONG: Missing tenantID filter (SECURITY VULNERABILITY)**
```go
// ❌ NEVER DO THIS - Cross-tenant data leakage risk
func GetProducts(c *gin.Context) {
    var products []Product
    err := db.Where("is_active = ?", true).Find(&products).Error
    // This will return products from ALL tenants!
}
```

### System Admin Override

**System admins can access all tenants for management:**
```go
func AdminTenantOverrideMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        isSystemAdmin, _ := c.Get("isSystemAdmin")

        if isSystemAdmin.(bool) {
            // Allow accessing any tenant via query parameter
            requestedTenantID := c.Query("tenant_id")
            if requestedTenantID != "" {
                // Validate tenant exists
                var tenant Tenant
                if err := db.First(&tenant, "id = ?", requestedTenantID).Error; err == nil {
                    c.Set("tenantID", requestedTenantID)
                    c.Set("tenant", tenant)
                    c.Set("isAdminOverride", true)
                }
            }
        }

        c.Next()
    }
}
```

### Tenant Switching Logic

**Switching Between User's Tenants:**
```go
func SwitchTenant(c *gin.Context) {
    userID, _ := c.Get("userID")

    var req struct {
        TenantID string `json:"tenantId" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }

    // Validate user has access to requested tenant
    var userTenant UserTenant
    err := db.Preload("Tenant.Company").
        Where("user_id = ? AND tenant_id = ? AND is_active = ?",
            userID, req.TenantID, true).
        First(&userTenant).Error

    if err != nil {
        c.JSON(403, gin.H{"error": "Access denied to this tenant"})
        return
    }

    // Validate tenant status
    if userTenant.Tenant.Status != TenantStatusActive &&
       userTenant.Tenant.Status != TenantStatusTrial {
        c.JSON(403, gin.H{"error": "Tenant is not active"})
        return
    }

    // Get user
    var user User
    db.First(&user, "id = ?", userID)

    // Generate new access token with new tenant context
    accessToken, err := GenerateAccessToken(&user, &userTenant)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate token"})
        return
    }

    // Create audit log
    oldTenantID, _ := c.Get("tenantID")
    ipAddress := c.ClientIP()
    userAgent := c.Request.UserAgent()
    LogTenantSwitch(&user, oldTenantID.(string), req.TenantID, ipAddress, userAgent)

    c.JSON(200, gin.H{
        "accessToken": accessToken,
        "activeTenant": gin.H{
            "tenantId":    userTenant.TenantID,
            "role":        userTenant.Role,
            "companyName": userTenant.Tenant.Company.Name,
            "status":      userTenant.Tenant.Status,
        },
    })
}
```

---

## Frontend Integration

### Redux Toolkit Setup

**Store Configuration:**
```typescript
// store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import { authApi } from './services/authApi';
import authReducer from './slices/authSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    [authApi.reducerPath]: authApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(authApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

**Auth Slice:**
```typescript
// store/slices/authSlice.ts
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface User {
  id: string;
  email: string;
  name: string;
  username: string;
  isSystemAdmin: boolean;
}

interface TenantContext {
  tenantId: string;
  role: 'OWNER' | 'ADMIN' | 'FINANCE' | 'SALES' | 'WAREHOUSE' | 'STAFF';
  companyName: string;
  status: 'TRIAL' | 'ACTIVE' | 'SUSPENDED' | 'CANCELLED' | 'EXPIRED';
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  activeTenant: TenantContext | null;
  availableTenants: TenantContext[];
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  user: null,
  accessToken: null,
  refreshToken: null,
  activeTenant: null,
  availableTenants: [],
  isAuthenticated: false,
  isLoading: false,
  error: null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{
        user: User;
        accessToken: string;
        refreshToken: string;
        activeTenant: TenantContext;
        availableTenants: TenantContext[];
      }>
    ) => {
      state.user = action.payload.user;
      state.accessToken = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      state.activeTenant = action.payload.activeTenant;
      state.availableTenants = action.payload.availableTenants;
      state.isAuthenticated = true;
      state.error = null;
    },
    setAccessToken: (state, action: PayloadAction<string>) => {
      state.accessToken = action.payload;
    },
    setActiveTenant: (state, action: PayloadAction<TenantContext>) => {
      state.activeTenant = action.payload;
    },
    logout: (state) => {
      state.user = null;
      state.accessToken = null;
      state.refreshToken = null;
      state.activeTenant = null;
      state.availableTenants = [];
      state.isAuthenticated = false;
      state.error = null;
    },
    setError: (state, action: PayloadAction<string>) => {
      state.error = action.payload;
      state.isLoading = false;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
});

export const {
  setCredentials,
  setAccessToken,
  setActiveTenant,
  logout,
  setError,
  clearError,
} = authSlice.actions;

export default authSlice.reducer;
```

### RTK Query API Configuration

**Auth API Service:**
```typescript
// store/services/authApi.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { RootState } from '../index';
import { setCredentials, setAccessToken, logout } from '../slices/authSlice';

const baseQuery = fetchBaseQuery({
  baseUrl: '/api/v1',
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).auth.accessToken;
    if (token) {
      headers.set('authorization', `Bearer ${token}`);
    }
    return headers;
  },
});

// Wrapper with automatic token refresh
const baseQueryWithReauth = async (args, api, extraOptions) => {
  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    // Try to refresh token
    const refreshToken = (api.getState() as RootState).auth.refreshToken;

    if (refreshToken) {
      const refreshResult = await baseQuery(
        { url: '/auth/refresh', method: 'POST', body: { refreshToken } },
        api,
        extraOptions
      );

      if (refreshResult.data) {
        // Store new tokens
        api.dispatch(setAccessToken(refreshResult.data.accessToken));

        // Retry original request with new token
        result = await baseQuery(args, api, extraOptions);
      } else {
        // Refresh failed, logout user
        api.dispatch(logout());
      }
    } else {
      api.dispatch(logout());
    }
  }

  return result;
};

export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User'],
  endpoints: (builder) => ({
    register: builder.mutation({
      query: (credentials) => ({
        url: '/auth/register',
        method: 'POST',
        body: credentials,
      }),
    }),

    login: builder.mutation({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setCredentials(data));
        } catch (err) {
          // Error handled by component
        }
      },
    }),

    logout: builder.mutation({
      query: (refreshToken) => ({
        url: '/auth/logout',
        method: 'POST',
        body: { refreshToken },
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          await queryFulfilled;
          dispatch(logout());
        } catch (err) {
          // Still logout on frontend even if API call fails
          dispatch(logout());
        }
      },
    }),

    refreshToken: builder.mutation({
      query: (refreshToken) => ({
        url: '/auth/refresh',
        method: 'POST',
        body: { refreshToken },
      }),
    }),

    verifyEmail: builder.mutation({
      query: (token) => ({
        url: '/auth/verify-email',
        method: 'POST',
        body: { token },
      }),
    }),

    forgotPassword: builder.mutation({
      query: (email) => ({
        url: '/auth/forgot-password',
        method: 'POST',
        body: { email },
      }),
    }),

    resetPassword: builder.mutation({
      query: ({ token, newPassword }) => ({
        url: '/auth/reset-password',
        method: 'POST',
        body: { token, newPassword },
      }),
    }),

    switchTenant: builder.mutation({
      query: (tenantId) => ({
        url: '/auth/switch-tenant',
        method: 'POST',
        body: { tenantId },
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setAccessToken(data.accessToken));
          dispatch(setActiveTenant(data.activeTenant));
        } catch (err) {
          // Error handled by component
        }
      },
    }),

    getCurrentUser: builder.query({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),
  }),
});

export const {
  useRegisterMutation,
  useLoginMutation,
  useLogoutMutation,
  useRefreshTokenMutation,
  useVerifyEmailMutation,
  useForgotPasswordMutation,
  useResetPasswordMutation,
  useSwitchTenantMutation,
  useGetCurrentUserQuery,
} = authApi;
```

### Automatic Token Refresh

**Token Refresh Logic:**
```typescript
// utils/tokenRefresh.ts
import { store } from '../store';
import { setAccessToken, logout } from '../store/slices/authSlice';
import jwt_decode from 'jwt-decode';

interface JWTPayload {
  exp: number;
  sub: string;
  tid: string;
  role: string;
}

export const setupTokenRefresh = () => {
  // Check token expiry every minute
  setInterval(() => {
    const state = store.getState();
    const { accessToken, refreshToken } = state.auth;

    if (!accessToken || !refreshToken) return;

    try {
      const decoded: JWTPayload = jwt_decode(accessToken);
      const expiresAt = decoded.exp * 1000; // Convert to milliseconds
      const now = Date.now();
      const timeUntilExpiry = expiresAt - now;

      // Refresh if token expires in less than 5 minutes
      if (timeUntilExpiry < 5 * 60 * 1000) {
        refreshAccessToken(refreshToken);
      }
    } catch (err) {
      console.error('Error decoding token:', err);
      store.dispatch(logout());
    }
  }, 60 * 1000); // Check every minute
};

const refreshAccessToken = async (refreshToken: string) => {
  try {
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });

    if (response.ok) {
      const data = await response.json();
      store.dispatch(setAccessToken(data.accessToken));

      // Update refresh token if rotation is enabled
      if (data.refreshToken) {
        // Update refresh token in state
        store.dispatch(setCredentials({
          ...store.getState().auth,
          refreshToken: data.refreshToken,
        }));
      }
    } else {
      // Refresh failed, logout
      store.dispatch(logout());
    }
  } catch (err) {
    console.error('Token refresh error:', err);
    store.dispatch(logout());
  }
};
```

### Protected Route Component

**Auth Guard:**
```typescript
// components/ProtectedRoute.tsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useSelector } from 'react-redux';
import type { RootState } from '../store';

interface ProtectedRouteProps {
  requiredRole?: 'OWNER' | 'ADMIN' | 'FINANCE' | 'SALES' | 'WAREHOUSE' | 'STAFF';
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ requiredRole }) => {
  const { isAuthenticated, activeTenant } = useSelector((state: RootState) => state.auth);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole && activeTenant) {
    const roleHierarchy = {
      OWNER: 6,
      ADMIN: 5,
      FINANCE: 4,
      SALES: 3,
      WAREHOUSE: 2,
      STAFF: 1,
    };

    if (roleHierarchy[activeTenant.role] < roleHierarchy[requiredRole]) {
      return <Navigate to="/unauthorized" replace />;
    }
  }

  return <Outlet />;
};
```

### Login Component Example

**Login Form with RTK Query:**
```typescript
// pages/Login.tsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLoginMutation } from '../store/services/authApi';

export const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [login, { isLoading, error }] = useLoginMutation();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await login({ email, password }).unwrap();
      navigate('/dashboard');
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <h2>Login</h2>

      {error && (
        <div className="error">
          {error.data?.error?.message || 'Login failed'}
        </div>
      )}

      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
      />

      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
      />

      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
};
```

### Tenant Switcher Component

**Tenant Selection Dropdown:**
```typescript
// components/TenantSwitcher.tsx
import React from 'react';
import { useSelector } from 'react-redux';
import { useSwitchTenantMutation } from '../store/services/authApi';
import type { RootState } from '../store';

export const TenantSwitcher: React.FC = () => {
  const { activeTenant, availableTenants } = useSelector((state: RootState) => state.auth);
  const [switchTenant, { isLoading }] = useSwitchTenantMutation();

  const handleTenantSwitch = async (tenantId: string) => {
    if (tenantId === activeTenant?.tenantId) return;

    try {
      await switchTenant(tenantId).unwrap();
      // Success - state updated by RTK Query mutation
    } catch (err) {
      console.error('Tenant switch failed:', err);
    }
  };

  return (
    <div className="tenant-switcher">
      <label>Active Company:</label>
      <select
        value={activeTenant?.tenantId || ''}
        onChange={(e) => handleTenantSwitch(e.target.value)}
        disabled={isLoading}
      >
        {availableTenants.map((tenant) => (
          <option key={tenant.tenantId} value={tenant.tenantId}>
            {tenant.companyName} ({tenant.role})
          </option>
        ))}
      </select>
    </div>
  );
};
```

---

## Middleware Stack

### Complete Middleware Chain

```
Request
  ↓
[1] CORS Middleware
  ↓
[2] Rate Limiter
  ↓
[3] Request Logger
  ↓
[4] JWT Validator (for protected routes)
  ↓
[5] Tenant Context (for protected routes)
  ↓
[6] Role Checker (for role-restricted routes)
  ↓
Handler
  ↓
Response
```

### Implementation

**Main Router Setup:**
```go
// cmd/api/router.go
func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Global middleware
    r.Use(CORSMiddleware())
    r.Use(RateLimitMiddleware())
    r.Use(RequestLoggerMiddleware())

    // Public routes
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", RegisterHandler)
        auth.POST("/login", LoginHandler)
        auth.POST("/verify-email", VerifyEmailHandler)
        auth.POST("/forgot-password", ForgotPasswordHandler)
        auth.POST("/reset-password", ResetPasswordHandler)
        auth.POST("/refresh", RefreshTokenHandler)
    }

    // Protected routes
    protected := r.Group("/api/v1/auth")
    protected.Use(JWTAuthMiddleware())
    {
        protected.POST("/logout", LogoutHandler)
        protected.GET("/me", GetCurrentUserHandler)
        protected.POST("/change-password", ChangePasswordHandler)
        protected.POST("/resend-verification", ResendVerificationHandler)
        protected.GET("/tenants", GetUserTenantsHandler)
        protected.POST("/switch-tenant", SwitchTenantHandler)
    }

    // Tenant-scoped routes
    api := r.Group("/api/v1")
    api.Use(JWTAuthMiddleware())
    api.Use(TenantContextMiddleware())
    {
        // Products (WAREHOUSE role required)
        products := api.Group("/products")
        products.Use(RequireRole(UserRoleWarehouse))
        {
            products.GET("", GetProductsHandler)
            products.POST("", CreateProductHandler)
            products.GET("/:id", GetProductHandler)
            products.PUT("/:id", UpdateProductHandler)
            products.DELETE("/:id", DeleteProductHandler)
        }

        // Invoices (FINANCE role required)
        invoices := api.Group("/invoices")
        invoices.Use(RequireRole(UserRoleFinance))
        {
            invoices.GET("", GetInvoicesHandler)
            invoices.POST("", CreateInvoiceHandler)
            // ... other invoice endpoints
        }
    }

    return r
}
```

**Middleware Implementations:**

```go
// CORS Middleware
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

// Request Logger Middleware
func RequestLoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        statusCode := c.Writer.Status()

        log.Printf("[%s] %s %s - %d - %v",
            c.Request.Method,
            path,
            query,
            statusCode,
            latency,
        )
    }
}

// Role-Based Authorization Middleware
func RequireRole(minRole UserRole) gin.HandlerFunc {
    roleHierarchy := map[UserRole]int{
        UserRoleOwner:     6,
        UserRoleAdmin:     5,
        UserRoleFinance:   4,
        UserRoleSales:     3,
        UserRoleWarehouse: 2,
        UserRoleStaff:     1,
    }

    return func(c *gin.Context) {
        // Check if system admin (has all permissions)
        isSystemAdmin, exists := c.Get("isSystemAdmin")
        if exists && isSystemAdmin.(bool) {
            c.Next()
            return
        }

        // Get user's role in current tenant
        role, exists := c.Get("userTenantRole")
        if !exists {
            c.JSON(403, gin.H{"error": "Role information missing"})
            c.Abort()
            return
        }

        userRole := role.(UserRole)

        // Check role hierarchy
        if roleHierarchy[userRole] < roleHierarchy[minRole] {
            c.JSON(403, gin.H{
                "error": fmt.Sprintf("Requires %s role or higher", minRole),
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## Error Handling

### Standard Error Response Format

```json
{
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": null,
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### Error Code Definitions

**Authentication Errors:**
```go
const (
    ErrInvalidCredentials      = "AUTH_INVALID_CREDENTIALS"
    ErrEmailNotVerified        = "AUTH_EMAIL_NOT_VERIFIED"
    ErrAccountLocked           = "AUTH_ACCOUNT_LOCKED"
    ErrAccountInactive         = "AUTH_ACCOUNT_INACTIVE"
    ErrTokenExpired            = "AUTH_TOKEN_EXPIRED"
    ErrTokenInvalid            = "AUTH_TOKEN_INVALID"
    ErrTenantAccessDenied      = "AUTH_TENANT_ACCESS_DENIED"
    ErrTenantSuspended         = "AUTH_TENANT_SUSPENDED"
    ErrInvalidRefreshToken     = "AUTH_INVALID_REFRESH_TOKEN"
    ErrPasswordTooWeak         = "AUTH_PASSWORD_TOO_WEAK"
    ErrEmailAlreadyExists      = "AUTH_EMAIL_ALREADY_EXISTS"
    ErrUsernameAlreadyExists   = "AUTH_USERNAME_ALREADY_EXISTS"
    ErrRateLimitExceeded       = "AUTH_RATE_LIMIT_EXCEEDED"
    ErrInvalidResetToken       = "AUTH_INVALID_RESET_TOKEN"
    ErrInvalidVerificationToken = "AUTH_INVALID_VERIFICATION_TOKEN"
)
```

### Error Response Helper

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code      string      `json:"code"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

func SendError(c *gin.Context, statusCode int, code, message string, details interface{}) {
    c.JSON(statusCode, ErrorResponse{
        Error: ErrorDetail{
            Code:      code,
            Message:   message,
            Details:   details,
            Timestamp: time.Now(),
        },
    })
}

// Usage examples
func LoginHandler(c *gin.Context) {
    // ... validation logic

    if invalidCredentials {
        SendError(c, 400, ErrInvalidCredentials,
            "Invalid email or password", nil)
        return
    }

    if !user.IsActive {
        SendError(c, 401, ErrEmailNotVerified,
            "Please verify your email before logging in", nil)
        return
    }
}
```

### Validation Error Handling

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func ValidateRegistration(req *RegisterRequest) []ValidationError {
    var errors []ValidationError

    // Email validation
    if !isValidEmail(req.Email) {
        errors = append(errors, ValidationError{
            Field:   "email",
            Message: "Invalid email format",
        })
    }

    // Password validation
    if err := ValidatePassword(req.Password); err != nil {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: err.Error(),
        })
    }

    // Username validation
    if len(req.Username) < 3 || len(req.Username) > 50 {
        errors = append(errors, ValidationError{
            Field:   "username",
            Message: "Username must be between 3 and 50 characters",
        })
    }

    return errors
}

func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        SendError(c, 400, "VALIDATION_ERROR", "Invalid request", err.Error())
        return
    }

    if validationErrors := ValidateRegistration(&req); len(validationErrors) > 0 {
        SendError(c, 400, "VALIDATION_ERROR", "Validation failed", validationErrors)
        return
    }

    // ... continue with registration
}
```

---

## Testing Strategy

### Unit Tests

**Password Service Tests:**
```go
// internal/services/auth/password_test.go
func TestHashPassword(t *testing.T) {
    password := "SecurePass123!"
    hash, err := HashPassword(password)

    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.NotEqual(t, password, hash)
}

func TestVerifyPassword(t *testing.T) {
    password := "SecurePass123!"
    hash, _ := HashPassword(password)

    err := VerifyPassword(hash, password)
    assert.NoError(t, err)

    err = VerifyPassword(hash, "WrongPassword")
    assert.Error(t, err)
}

func TestValidatePassword(t *testing.T) {
    tests := []struct {
        password    string
        expectError bool
    }{
        {"SecurePass123!", false},
        {"short", true},              // Too short
        {"nouppercase123!", true},    // No uppercase
        {"NOLOWERCASE123!", true},    // No lowercase
        {"NoNumbers!", true},         // No number
    }

    for _, tt := range tests {
        err := ValidatePassword(tt.password)
        if tt.expectError {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    }
}
```

**JWT Service Tests:**
```go
// internal/services/auth/jwt_test.go
func TestGenerateAccessToken(t *testing.T) {
    user := &User{
        ID:    "test_user_id",
        Email: "test@example.com",
        Name:  "Test User",
    }

    tenant := &UserTenant{
        TenantID: "test_tenant_id",
        Role:     UserRoleOwner,
    }

    token, err := GenerateAccessToken(user, tenant)

    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // Verify token can be decoded
    claims, err := ValidateAccessToken(token)
    assert.NoError(t, err)
    assert.Equal(t, user.ID, claims["sub"])
    assert.Equal(t, tenant.TenantID, claims["tid"])
    assert.Equal(t, string(tenant.Role), claims["role"])
}

func TestTokenExpiry(t *testing.T) {
    // Create token with very short expiry
    claims := jwt.MapClaims{
        "sub": "test_user",
        "exp": time.Now().Add(-1 * time.Hour).Unix(), // Already expired
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("test_secret"))

    _, err := ValidateAccessToken(tokenString)
    assert.Error(t, err)
}
```

### Integration Tests

**Login Flow Test:**
```go
// test/integration/auth_test.go
func TestLoginFlow(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer cleanupTestDB(db)

    // Create test user
    user := createTestUser(db, "test@example.com", "SecurePass123!")

    // Create test tenant
    tenant := createTestTenant(db)
    createUserTenant(db, user.ID, tenant.ID, UserRoleOwner)

    // Test login
    req := LoginRequest{
        Email:    "test@example.com",
        Password: "SecurePass123!",
    }

    resp, err := login(req)

    assert.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
    assert.NotEmpty(t, resp.RefreshToken)
    assert.Equal(t, user.Email, resp.User.Email)
    assert.Len(t, resp.Tenants, 1)
}

func TestLoginWithInvalidCredentials(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)

    user := createTestUser(db, "test@example.com", "SecurePass123!")

    req := LoginRequest{
        Email:    "test@example.com",
        Password: "WrongPassword",
    }

    _, err := login(req)

    assert.Error(t, err)
    assert.Equal(t, ErrInvalidCredentials, err.Code)
}
```

**Tenant Switching Test:**
```go
func TestTenantSwitching(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)

    user := createTestUser(db, "test@example.com", "SecurePass123!")
    tenant1 := createTestTenant(db)
    tenant2 := createTestTenant(db)

    createUserTenant(db, user.ID, tenant1.ID, UserRoleOwner)
    createUserTenant(db, user.ID, tenant2.ID, UserRoleAdmin)

    // Login
    loginResp, _ := login(LoginRequest{
        Email:    user.Email,
        Password: "SecurePass123!",
    })

    // Verify initial tenant
    assert.Equal(t, tenant1.ID, loginResp.ActiveTenant.TenantID)

    // Switch to tenant2
    switchResp, err := switchTenant(loginResp.AccessToken, tenant2.ID)

    assert.NoError(t, err)
    assert.Equal(t, tenant2.ID, switchResp.ActiveTenant.TenantID)
    assert.Equal(t, UserRoleAdmin, switchResp.ActiveTenant.Role)
}
```

### Security Tests

**Brute Force Protection Test:**
```go
func TestBruteForceProtection(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)

    user := createTestUser(db, "test@example.com", "SecurePass123!")
    ipAddress := "192.168.1.1"

    // Make 5 failed login attempts
    for i := 0; i < 5; i++ {
        _, err := loginWithIP(LoginRequest{
            Email:    user.Email,
            Password: "WrongPassword",
        }, ipAddress)

        assert.Error(t, err)
    }

    // 6th attempt should be blocked
    _, err := loginWithIP(LoginRequest{
        Email:    user.Email,
        Password: "SecurePass123!", // Correct password
    }, ipAddress)

    assert.Error(t, err)
    assert.Equal(t, ErrAccountLocked, err.Code)

    // Verify login attempts were recorded
    var count int64
    db.Model(&LoginAttempt{}).
        Where("email = ? AND ip_address = ?", user.Email, ipAddress).
        Count(&count)

    assert.Equal(t, int64(6), count)
}
```

**Token Revocation Test:**
```go
func TestTokenRevocation(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)

    user := createTestUser(db, "test@example.com", "SecurePass123!")
    tenant := createTestTenant(db)
    createUserTenant(db, user.ID, tenant.ID, UserRoleOwner)

    // Login
    loginResp, _ := login(LoginRequest{
        Email:    user.Email,
        Password: "SecurePass123!",
    })

    // Logout
    err := logout(loginResp.RefreshToken)
    assert.NoError(t, err)

    // Try to use revoked refresh token
    _, err = refreshToken(loginResp.RefreshToken)
    assert.Error(t, err)
    assert.Equal(t, ErrInvalidRefreshToken, err.Code)

    // Verify token is marked as revoked in DB
    var refreshTokenRecord RefreshToken
    db.Where("token = ?", loginResp.RefreshToken).First(&refreshTokenRecord)
    assert.True(t, refreshTokenRecord.IsRevoked)
}
```

---

## Deployment & Operations

### Environment Variables

```env
# Application
APP_ENV=production
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=erp_production
DB_USER=erp_user
DB_PASSWORD=<secure-password>
DB_SSL_MODE=require

# JWT Configuration
JWT_SECRET=<min-32-byte-random-string>  # Generate with: openssl rand -base64 32
JWT_ACCESS_TOKEN_EXPIRY=30m
JWT_REFRESH_TOKEN_EXPIRY=30d

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=<app-specific-password>
EMAIL_FROM=ERP System <noreply@example.com>
FRONTEND_URL=https://app.example.com

# Security Settings
# Argon2id parameters (optional, defaults are set in code)
ARGON2_MEMORY=65536          # 64 MB in KiB
ARGON2_ITERATIONS=3
ARGON2_PARALLELISM=4
ARGON2_SALT_LENGTH=16
ARGON2_KEY_LENGTH=32

MAX_LOGIN_ATTEMPTS=5
LOGIN_LOCKOUT_DURATION=15m
RATE_LIMIT_PER_MINUTE=100

# Session Settings
SESSION_TIMEOUT=30m
REFRESH_TOKEN_ROTATION=true

# CORS
ALLOWED_ORIGINS=https://app.example.com,https://www.example.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Database Migrations

**Run Migrations:**
```bash
# Development
make migrate-up

# Production
DB_HOST=prod-db.example.com make migrate-up

# Rollback
make migrate-down
```

**Migration Files:**
```sql
-- migrations/004_create_refresh_tokens.up.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
```

### Background Jobs

**Cron Jobs (using robfig/cron):**
```go
// cmd/worker/main.go
func setupCronJobs() {
    c := cron.New()

    // Cleanup expired tokens (hourly)
    c.AddFunc("@hourly", func() {
        CleanupExpiredTokens()
    })

    // Cleanup old login attempts (daily at 2 AM)
    c.AddFunc("0 2 * * *", func() {
        CleanupOldLoginAttempts()
    })

    // Send daily summary reports (daily at 6 AM)
    c.AddFunc("0 6 * * *", func() {
        SendDailySummaryReports()
    })

    c.Start()
}

func CleanupExpiredTokens() {
    // Delete expired refresh tokens
    result := db.Where("expires_at < ?", time.Now()).
        Delete(&RefreshToken{})

    log.Printf("Cleaned up %d expired refresh tokens", result.RowsAffected)

    // Delete expired/used verification tokens older than 7 days
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    result = db.Where("created_at < ? AND (is_used = ? OR expires_at < ?)",
        sevenDaysAgo, true, time.Now()).
        Delete(&EmailVerification{})

    log.Printf("Cleaned up %d old verification tokens", result.RowsAffected)

    // Delete expired/used password reset tokens older than 7 days
    result = db.Where("created_at < ? AND (is_used = ? OR expires_at < ?)",
        sevenDaysAgo, true, time.Now()).
        Delete(&PasswordReset{})

    log.Printf("Cleaned up %d old password reset tokens", result.RowsAffected)
}

func CleanupOldLoginAttempts() {
    // Keep login attempts for 30 days for security analysis
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    result := db.Where("created_at < ?", thirtyDaysAgo).
        Delete(&LoginAttempt{})

    log.Printf("Cleaned up %d old login attempts", result.RowsAffected)
}
```

### Monitoring & Alerting

**Prometheus Metrics:**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    loginAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_login_attempts_total",
            Help: "Total number of login attempts",
        },
        []string{"status"}, // success, failed, locked
    )

    tokenGeneration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "auth_token_generation_duration_seconds",
            Help: "Time taken to generate tokens",
        },
    )

    activeUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "auth_active_users",
            Help: "Number of currently authenticated users",
        },
    )
)

func init() {
    prometheus.MustRegister(loginAttempts)
    prometheus.MustRegister(tokenGeneration)
    prometheus.MustRegister(activeUsers)
}

// Usage in handlers
func LoginHandler(c *gin.Context) {
    // ... login logic

    if success {
        loginAttempts.WithLabelValues("success").Inc()
    } else {
        loginAttempts.WithLabelValues("failed").Inc()
    }
}
```

**Health Check Endpoint:**
```go
func HealthCheckHandler(c *gin.Context) {
    // Check database connection
    if err := db.DB().Ping(); err != nil {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "database": "down",
        })
        return
    }

    // Check Redis connection (if using for rate limiting)
    // ...

    c.JSON(200, gin.H{
        "status": "healthy",
        "database": "up",
        "timestamp": time.Now(),
    })
}
```

### Logging

**Structured Logging:**
```go
import "go.uber.org/zap"

var logger *zap.Logger

func InitLogger() {
    var err error
    if os.Getenv("APP_ENV") == "production" {
        logger, err = zap.NewProduction()
    } else {
        logger, err = zap.NewDevelopment()
    }

    if err != nil {
        panic(err)
    }
}

// Usage
func LoginHandler(c *gin.Context) {
    logger.Info("Login attempt",
        zap.String("email", req.Email),
        zap.String("ip", c.ClientIP()),
        zap.String("user_agent", c.Request.UserAgent()),
    )

    // ... login logic

    if err != nil {
        logger.Error("Login failed",
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return
    }

    logger.Info("Login successful",
        zap.String("user_id", user.ID),
        zap.String("tenant_id", activeTenant.TenantID),
    )
}
```

---

## Implementation Phases

### Phase 1: Core Authentication (Week 1-2)
**Deliverables:**
- [x] Database models (RefreshToken, EmailVerification, PasswordReset, LoginAttempt)
- [x] Password hashing service (argon2id)
- [x] JWT token generation and validation
- [x] Basic registration endpoint
- [x] Email verification flow
- [x] Login endpoint with token generation
- [x] Logout endpoint with token revocation
- [x] Audit logging integration

**Testing:**
- Unit tests for password hashing
- Unit tests for JWT generation
- Integration tests for registration flow
- Integration tests for login flow

**Success Criteria:**
- Users can register with email verification
- Users can login and receive JWT tokens
- Tokens can be validated successfully
- All login/logout events logged in audit trail

---

### Phase 2: Security Hardening (Week 2-3)
**Deliverables:**
- [x] Rate limiting middleware
- [x] Brute force protection (account lockout)
- [x] Password reset flow
- [x] Input validation and sanitization
- [x] CSRF protection (if using cookies)
- [x] Security middleware stack

**Testing:**
- Security tests for brute force protection
- Integration tests for password reset flow
- Rate limiting tests
- Input validation tests

**Success Criteria:**
- Account lockout after 5 failed attempts
- Rate limiting prevents abuse
- Password reset flow works end-to-end
- All inputs properly validated and sanitized

---

### Phase 3: Multi-Tenant Integration (Week 3-4)
**Deliverables:**
- [x] Tenant context middleware
- [x] Tenant switching endpoint
- [x] Multi-tenant access validation
- [x] Subscription status checks
- [x] Role-based authorization middleware
- [x] Cross-tenant security tests

**Testing:**
- Multi-tenant isolation tests
- Tenant switching tests
- Role-based access control tests
- Subscription validation tests

**Success Criteria:**
- Users can switch between tenants
- Tenant data completely isolated
- Subscription status properly enforced
- Role-based permissions working

---

### Phase 4: Frontend Integration (Week 4-5)
**Deliverables:**
- [x] Redux Toolkit store setup
- [x] RTK Query API configuration
- [x] Authentication state management
- [x] Automatic token refresh
- [x] Protected route components
- [x] Tenant switcher UI component
- [x] Login/Register forms
- [x] Password reset UI flow

**Testing:**
- E2E tests for complete auth flows
- Frontend component tests
- Integration tests with backend
- UX testing

**Success Criteria:**
- Seamless login/logout experience
- Automatic token refresh works
- Tenant switching in UI functional
- All error states properly handled

---

### Phase 5: Testing & Polish (Week 5-6)
**Deliverables:**
- [x] Comprehensive unit tests (80%+ coverage)
- [x] Integration tests for all flows
- [x] Security penetration tests
- [x] Performance optimization
- [x] Documentation (API docs, deployment guide)
- [x] Deployment configuration (Docker, env files)

**Testing:**
- Load testing (concurrent logins)
- Security audit
- Code review
- Documentation review

**Success Criteria:**
- 80%+ code coverage
- All security tests pass
- Performance meets SLAs
- Complete documentation
- Production-ready deployment

---

## Summary

This authentication system provides:

✅ **Security First**
- Argon2id password hashing (resistant to GPU/ASIC attacks)
- JWT with refresh token rotation
- Brute force protection
- Rate limiting
- Comprehensive audit logging

✅ **Multi-Tenant Architecture**
- Complete tenant isolation
- Per-tenant role-based access
- Subscription enforcement
- Seamless tenant switching

✅ **Developer Experience**
- Clear API design
- Comprehensive error handling
- Type-safe with Go structs
- Modular service architecture

✅ **User Experience**
- Automatic token refresh
- Email verification
- Password reset flow
- Remember me (long-lived refresh tokens)

✅ **Production Ready**
- Comprehensive testing
- Monitoring and alerting
- Background jobs
- Scalable architecture

**Timeline:** 5-6 weeks for complete MVP implementation

**Next Steps:**
1. Review and approve this design document
2. Set up development environment
3. Create database migrations
4. Begin Phase 1 implementation
5. Iterate with testing and feedback

---

**Document End**
