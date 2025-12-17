# Phase 3 MVP Implementation Analysis

**Analysis Date:** 2025-12-17
**Phase 1 Status:** ✅ 100% Complete
**Phase 2 Status:** ✅ 100% Complete
**Phase 3 Status:** ⚠️ 80% Complete (4/5 official requirements)
**Approach:** MVP with 80/20 prioritization

---

## 📊 Executive Summary

The backend authentication system has completed Phase 1 (Core Auth) and Phase 2 (Security Hardening) successfully. Phase 3 (Multi-Tenant Integration) is **80% complete** with most critical infrastructure already implemented. This analysis identifies the **20% of remaining work** needed to deliver **100% of Phase 3 value** through an MVP approach.

### Key Findings

✅ **Strengths (Already Implemented):**
- **Dual-layer tenant isolation** - Production-ready GORM callbacks + scopes
- **Subscription validation** - Integrated in login/refresh flows
- **Role-based access control** - RequireRoleMiddleware fully functional
- **Cross-tenant security tests** - Comprehensive test coverage
- **Tenant context middleware** - Automatic tenant session management

⚠️ **Gaps (Missing for Production):**
- Tenant switching endpoint (official Phase 3 requirement)
- Email verification handler (blocks user onboarding)
- GET /auth/me endpoint (common REST pattern)
- GET /auth/tenants endpoint (required for tenant switching UI)
- Change password endpoint (security best practice)

---

## 🎯 Phase 3 Implementation Status

### Official Requirements (From BACKEND-IMPLEMENTATION.md)

| # | Requirement | Status | Implementation | Production Ready |
|---|------------|--------|----------------|------------------|
| 1 | Tenant context middleware (subscription validation) | ✅ Complete | `middleware/auth.go:67-91`, `auth_service.go:103-110, 209-223` | ✅ YES |
| 2 | Tenant switching endpoint | ❌ Missing | TODO in `router.go:120` | ❌ NO - Phase 3 MVP |
| 3 | GORM scopes for tenant filtering (CRITICAL) | ✅ Complete | Dual-layer isolation in `database/tenant.go` | ✅ YES |
| 4 | Role-based authorization middleware | ✅ Complete | `middleware/auth.go:136-170` | ✅ YES |
| 5 | Cross-tenant security tests | ✅ Complete | `database/tenant_test.go`, `models/phase3_test.go` | ✅ YES |

**Summary:** 4/5 requirements complete (80%). Only tenant switching endpoint missing.

### Additional Missing Features (Router TODOs)

From analysis of `internal/router/router.go`:

| Line | Feature | Status | Priority | MVP Fit |
|------|---------|--------|----------|---------|
| 83 | Email verification handler | ❌ Missing | 🔴 HIGH | ✅ YES - Blocks onboarding |
| 118-119 | Change password endpoint | ❌ Missing | 🟡 MEDIUM | ⚠️ MAYBE - Nice to have |
| 120 | Switch tenant endpoint | ❌ Missing | 🔴 CRITICAL | ✅ YES - Official requirement |
| 124-130 | User profile endpoints | ❌ Missing | 🟡 MEDIUM | ❌ NO - Defer to Phase 4 |
| 133-140 | Tenant management routes | ❌ Missing | 🟢 LOW | ❌ NO - Defer to Phase 4 |
| 143-148 | Company profile routes | ❌ Missing | 🟢 LOW | ❌ NO - Defer to Phase 4 |

---

## 🎯 MVP Phase 3 Scope

### Critical Path (Must-Have for Production)

#### 🔴 Priority 1: Tenant Switching (2-3 hours)

**Impact:** Completes official Phase 3 requirements, enables multi-tenant usage
**Effort:** 2-3 hours
**Deliverable:** Users can switch between accessible tenants

**Implementation:**

1. **Service Methods** (`internal/service/auth/auth_service.go`)
   ```go
   // SwitchTenant switches user's active tenant
   func (s *AuthService) SwitchTenant(userID string, newTenantID string) (*SwitchTenantResponse, error)

   // GetUserTenants returns all tenants accessible to user
   func (s *AuthService) GetUserTenants(userID string) ([]TenantInfo, error)
   ```

2. **Handlers** (`internal/handler/auth_handler.go`)
   - POST `/api/v1/auth/switch-tenant` - Switch to different tenant
   - GET `/api/v1/auth/tenants` - List user's accessible tenants

3. **Business Logic:**
   - Validate user has access to target tenant (check user_tenants table)
   - Verify target tenant is ACTIVE or TRIAL with valid period
   - Generate new access token with updated tenantID claim
   - Keep same refresh token (user-scoped, not tenant-scoped)
   - Return new access token + active tenant info

4. **Routes** (`internal/router/router.go`)
   ```go
   protected.GET("/auth/tenants", authHandler.GetUserTenants)
   protected.POST("/auth/switch-tenant", authHandler.SwitchTenant)
   ```

5. **DTOs** (`internal/dto/auth_dto.go`)
   ```go
   type SwitchTenantRequest struct {
       TenantID string `json:"tenant_id" binding:"required"`
   }

   type SwitchTenantResponse struct {
       AccessToken  string     `json:"access_token"`
       ActiveTenant TenantInfo `json:"active_tenant"`
   }

   type TenantInfo struct {
       ID     string `json:"id"`
       Name   string `json:"name"`
       Role   string `json:"role"`
       Status string `json:"status"`
   }
   ```

**Security Considerations:**
- Validate user_tenant relationship exists and is_active = true
- Check tenant subscription status (ACTIVE or valid TRIAL period)
- Don't revoke old access token (short-lived 30min anyway)
- Log tenant switch events for audit trail

**Testing:**
- Unit tests for SwitchTenant service method
- Integration test: switch → verify new token → query with new tenant context
- Security test: attempt to switch to unauthorized tenant (should fail)
- Security test: verify old token still works until expiry

---

#### 🔴 Priority 2: Email Verification (1-2 hours)

**Impact:** Unblocks user onboarding, prevents unverified users from accessing system
**Effort:** 1-2 hours
**Deliverable:** Complete email verification workflow

**Implementation:**

1. **Service Method** (`internal/service/auth/auth_service.go`)
   ```go
   // VerifyEmail verifies user's email using token from verification email
   func (s *AuthService) VerifyEmail(token string) error
   ```

2. **Handler** (`internal/handler/auth_handler.go`)
   - POST `/api/v1/auth/verify-email` - Verify email with token

3. **Business Logic:**
   - Find email_verification record by token
   - Validate token not expired (24 hours from creation)
   - Validate token not already used (verified_at is NULL)
   - Update user: email_verified = true, email_verified_at = NOW()
   - Mark verification record as used (verified_at = NOW())
   - Return success response

4. **Login Integration:**
   - Update Login() method to check email_verified field
   - Return specific error if email not verified
   - Frontend should show "verify email" message with resend option

5. **Routes** (`internal/router/router.go`)
   ```go
   authGroup.POST("/verify-email", authHandler.VerifyEmail)
   ```

6. **DTOs** (`internal/dto/auth_dto.go`)
   ```go
   type VerifyEmailRequest struct {
       Token string `json:"token" binding:"required"`
   }

   type VerifyEmailResponse struct {
       Message string `json:"message"`
       Email   string `json:"email"`
   }
   ```

**Security Considerations:**
- Single-use tokens (mark as used after verification)
- 24-hour expiry window
- No user enumeration (generic success message)
- Rate limiting on verification attempts

**Testing:**
- Unit test: verify valid token
- Unit test: expired token should fail
- Unit test: already used token should fail
- Integration test: register → verify → login succeeds
- Integration test: register → login before verify → should fail

---

#### 🔴 Priority 3: Current User Profile (30 min)

**Impact:** Standard REST pattern, required by most frontends
**Effort:** 30 minutes
**Deliverable:** Users can fetch current profile and tenant info

**Implementation:**

1. **Handler** (`internal/handler/auth_handler.go`)
   - GET `/api/v1/auth/me` - Get current user profile

2. **Business Logic:**
   - Extract user_id and tenant_id from JWT claims (already in context)
   - Query user table for user details
   - Query user_tenants for current role
   - Query tenant table for tenant details
   - Return combined user + tenant info

3. **Routes** (`internal/router/router.go`)
   ```go
   protected.GET("/auth/me", authHandler.GetCurrentUser)
   ```

4. **DTOs** (`internal/dto/auth_dto.go`)
   ```go
   type CurrentUserResponse struct {
       User         UserInfo   `json:"user"`
       ActiveTenant TenantInfo `json:"active_tenant"`
       Tenants      []TenantInfo `json:"tenants"` // All accessible tenants
   }

   type UserInfo struct {
       ID            string `json:"id"`
       Email         string `json:"email"`
       FullName      string `json:"full_name"`
       Phone         string `json:"phone"`
       EmailVerified bool   `json:"email_verified"`
       IsSystemAdmin bool   `json:"is_system_admin"`
       LastLoginAt   string `json:"last_login_at"`
   }
   ```

**Testing:**
- Integration test: login → GET /me → verify user data matches
- Integration test: switch tenant → GET /me → verify active_tenant updated

---

### Recommended Additions (If Time Permits)

#### 🟡 Priority 4: Change Password (2 hours)

**Impact:** Security best practice, allows users to change password proactively
**Effort:** 2 hours
**Deliverable:** Users can change password while authenticated

**Implementation:**

1. **Service Method** (`internal/service/auth/auth_service.go`)
   ```go
   // ChangePassword allows authenticated user to change password
   func (s *AuthService) ChangePassword(userID string, oldPassword string, newPassword string) error
   ```

2. **Handler** (`internal/handler/auth_handler.go`)
   - POST `/api/v1/auth/change-password` - Change password

3. **Business Logic:**
   - Verify old password is correct
   - Validate new password meets requirements (use pkg/validator)
   - Hash new password with Argon2id
   - Update user password_hash in database
   - Revoke all refresh tokens (force re-login on all devices)
   - Log password change event

4. **Routes** (`internal/router/router.go`)
   ```go
   protected.POST("/auth/change-password", authHandler.ChangePassword)
   ```

5. **DTOs** (`internal/dto/auth_dto.go`)
   ```go
   type ChangePasswordRequest struct {
       OldPassword string `json:"old_password" binding:"required,min=8"`
       NewPassword string `json:"new_password" binding:"required,password_strength"`
   }
   ```

**Security Considerations:**
- Verify old password with constant-time comparison
- Revoke all refresh tokens to invalidate all sessions
- Rate limit password changes (max 3 per hour)
- Log password change events for audit

**Testing:**
- Unit test: change password with correct old password
- Unit test: change password with wrong old password (should fail)
- Unit test: change password with weak new password (should fail)
- Integration test: change password → old password no longer works → new password works

---

### Deferred to Phase 4 (Not MVP)

| Feature | Rationale for Deferral | Workaround |
|---------|------------------------|------------|
| **User Profile Update (PUT /me)** | Low priority - users rarely update profile. Can be done via admin interface if needed. | Users can contact admin for profile updates |
| **Tenant Management Routes** | Low usage - most tenants have single owner. Admin features can wait. | System admin can modify database directly if needed |
| **Company Profile Routes** | Very low frequency - set during registration, rarely changed. | Can be edited via SQL if needed, or wait for admin UI |

---

## 📅 MVP Implementation Timeline

### Day 1: Critical Endpoints (4-6 hours)

**Morning (2-3 hours):**
1. Implement SwitchTenant service method - 1 hour
2. Implement GetUserTenants service method - 30 min
3. Create SwitchTenant and GetUserTenants handlers - 1 hour
4. Add routes and DTOs - 30 min

**Afternoon (2-3 hours):**
1. Implement VerifyEmail service method - 1 hour
2. Create VerifyEmail handler - 30 min
3. Update Login to check email_verified - 30 min
4. Implement GetCurrentUser handler - 30 min
5. Add routes and DTOs - 30 min

**Deliverable:** 4 new working endpoints

---

### Day 2: Testing + Optional Features (2-4 hours)

**Morning (1-2 hours):**
1. Write unit tests for new service methods - 1 hour
2. Write integration tests for new endpoints - 1 hour

**Afternoon (1-2 hours, optional):**
1. Implement ChangePassword if time permits - 2 hours

**Deliverable:** Production-ready Phase 3 with tests

---

### Total Effort Estimate

- **Minimum (Critical only):** 4-6 hours
- **Recommended (Critical + Change Password):** 6-8 hours
- **Maximum (All features):** 8-10 hours

**Timeline:** 1-2 days for full MVP Phase 3 completion

---

## 🚨 Technical Debt & Risks

### Technical Debt Items

#### 1. Subscription Validation Location

**Current State:**
- Subscription checked in Login() and RefreshToken() methods
- NOT enforced at middleware level

**Issue:**
- If user logs in with active subscription, subscription expires during session
- Access token still valid for up to 30 minutes
- User can continue accessing system for 30-minute grace period

**Risk Level:** 🟡 MEDIUM (30-minute window is acceptable for MVP)

**Recommendation:**
- Accept as known limitation for MVP
- Document in deployment notes
- Consider adding subscription check to TenantContextMiddleware in Phase 4 if needed

**Mitigation:**
- Short access token expiry (30 min) minimizes risk
- Refresh token flow re-checks subscription every 30 min
- Can implement real-time subscription check in Phase 4 if needed

---

#### 2. Missing GET /auth/me Endpoint

**Current State:**
- Frontend stores user data from login response
- No way to refresh user state after login

**Issue:**
- Frontend state can become stale
- Common REST pattern expects /me endpoint
- Required for "refresh profile" functionality

**Risk Level:** 🟡 MEDIUM (UX issue, not security)

**Recommendation:**
- Implement as part of Phase 3 MVP (30 minutes)
- Standard pattern in most auth systems

---

#### 3. Email Verification Not Enforced

**Current State:**
- Registration creates email_verification record
- Email sending ready but commented out
- Login doesn't check email_verified field

**Issue:**
- Users can log in without verifying email
- Defeats purpose of email verification

**Risk Level:** 🔴 HIGH (Production blocker)

**Recommendation:**
- Implement VerifyEmail handler (Phase 3 MVP)
- Update Login to enforce email verification
- Uncomment email sending code when SMTP configured

---

#### 4. Tenant Switching Token Management

**Current State:**
- JWT contains tenantID claim
- No explicit mechanism for tenant switching

**Design Decision Needed:**
- **Option A (Recommended):** Generate new access token with different tenantID
  - Pros: Stateless, simple, follows JWT best practices
  - Cons: Multiple tokens per user (one per active session per tenant)

- **Option B:** Store "active tenant" in database, keep JWT same
  - Pros: Single token per user
  - Cons: Stateful, requires database query on every request, breaks JWT statelessness

**Recommendation:** Option A (stateless approach)
- Generate new access token with updated tenantID claim
- Keep same refresh token (user-scoped, not tenant-scoped)
- Don't revoke old access token (short-lived anyway)
- Frontend discards old token when switching

---

### Implementation Risks

#### Risk 1: Tenant Switching Security (HIGH)

**Problem:** User must have valid access to target tenant

**Validation Required:**
1. User-tenant relationship exists (`user_tenants` table)
2. Relationship is active (`is_active = true`)
3. Target tenant is not SUSPENDED or EXPIRED
4. Tenant subscription is valid (ACTIVE or TRIAL with valid period)

**Mitigation:**
```go
// Validation steps in SwitchTenant()
1. Query user_tenants WHERE user_id = ? AND tenant_id = ? AND is_active = true
2. If not found → return AuthorizationError("No access to tenant")
3. Query tenant WHERE id = ?
4. Check tenant.status IN ('ACTIVE', 'TRIAL')
5. If TRIAL → verify trial_ends_at > NOW()
6. Generate new access token with validated tenantID
```

**Test Coverage:**
- ✅ Switch to authorized tenant (success)
- ✅ Switch to unauthorized tenant (fail)
- ✅ Switch to SUSPENDED tenant (fail)
- ✅ Switch to expired TRIAL tenant (fail)

---

#### Risk 2: Email Verification Edge Cases (MEDIUM)

**Problem:** What happens if user tries various email verification scenarios?

**Edge Cases:**
1. User logs in before verifying email → Block with specific error message
2. User clicks verification link twice → Second click shows "already verified"
3. Verification token expires → Show "expired, resend verification email"
4. User requests password reset without verifying email → Allow (security requirement)

**Mitigation:**
```go
// Login flow
if !user.EmailVerified {
    return NewAuthenticationError("Email not verified. Check your inbox.")
}

// VerifyEmail flow
if verification.VerifiedAt != nil {
    return NewValidationError("Email already verified")
}
if verification.ExpiresAt.Before(time.Now()) {
    return NewValidationError("Verification link expired. Request new one.")
}
```

**Test Coverage:**
- ✅ Verify valid token (success)
- ✅ Verify expired token (fail)
- ✅ Verify used token (fail)
- ✅ Login before verification (fail)
- ✅ Login after verification (success)

---

#### Risk 3: Frontend Integration (LOW)

**Problem:** New endpoints need frontend coordination

**Impact:**
- Backend ready but frontend not using it
- Tenant switching UI not built
- Email verification flow incomplete

**Mitigation:**
- Provide comprehensive API documentation
- Coordinate with frontend team before deployment
- Test endpoints manually with Postman/curl
- Not blocking: Backend can be deployed independently

**Action Items:**
- Create API documentation for new endpoints
- Share with frontend team
- Schedule coordination meeting
- Test with Postman collection

---

## ✅ Phase 3 Success Criteria

### Functional Requirements

- ✅ Users can verify email after registration
- ✅ Users can switch between their accessible tenants
- ✅ Users can fetch current user profile and tenant info
- ✅ All existing security features remain intact
- ✅ Subscription validation works correctly
- ✅ Tenant isolation prevents cross-tenant data access

### Security Requirements

- ✅ Email verification prevents unverified logins
- ✅ Tenant switching validates user access before allowing switch
- ✅ Subscription status checked on tenant switch
- ✅ JWT tokens contain correct tenantID after switch
- ✅ No security regressions from Phase 1 or Phase 2

### Quality Requirements

- ✅ Unit test coverage ≥ 80% for new code
- ✅ Integration tests for all new endpoints
- ✅ Security tests for tenant switching scenarios
- ✅ API documentation updated
- ✅ No lint/formatting errors

### Operational Requirements

- ✅ All new endpoints documented
- ✅ Error responses follow consistent format
- ✅ Logging enabled for security events
- ✅ Deployment notes updated

---

## 📊 Implementation Checklist

### Day 1: Critical Endpoints

#### Morning Session (2-3 hours)

**Tenant Switching:**
- [ ] Create `SwitchTenantRequest` and `SwitchTenantResponse` DTOs
- [ ] Implement `SwitchTenant(userID, newTenantID)` service method
  - [ ] Validate user-tenant relationship
  - [ ] Check tenant status and subscription
  - [ ] Generate new access token with new tenantID
- [ ] Implement `GetUserTenants(userID)` service method
  - [ ] Query user_tenants table
  - [ ] Join with tenants for status
  - [ ] Return array of TenantInfo
- [ ] Create `SwitchTenant` handler
- [ ] Create `GetUserTenants` handler
- [ ] Add routes to protected group

**Test Manually:**
- [ ] GET /auth/tenants returns user's tenants
- [ ] POST /auth/switch-tenant with valid tenant_id succeeds
- [ ] POST /auth/switch-tenant with invalid tenant_id fails (403)
- [ ] New access token works with new tenant context

---

#### Afternoon Session (2-3 hours)

**Email Verification:**
- [ ] Create `VerifyEmailRequest` and `VerifyEmailResponse` DTOs
- [ ] Implement `VerifyEmail(token)` service method
  - [ ] Find email_verification by token
  - [ ] Validate not expired
  - [ ] Validate not already used
  - [ ] Update user.email_verified = true
  - [ ] Mark verification.verified_at = NOW()
- [ ] Create `VerifyEmail` handler
- [ ] Update `Login()` to check email_verified
- [ ] Add route to public auth group

**Test Manually:**
- [ ] POST /auth/verify-email with valid token succeeds
- [ ] POST /auth/verify-email with expired token fails
- [ ] POST /auth/verify-email with used token fails
- [ ] POST /auth/login before verification fails
- [ ] POST /auth/login after verification succeeds

**Current User Profile:**
- [ ] Create `CurrentUserResponse` DTO with UserInfo and TenantInfo
- [ ] Implement `GetCurrentUser` handler
  - [ ] Extract user_id and tenant_id from context
  - [ ] Query user details
  - [ ] Query user_tenants for role
  - [ ] Query tenant details
  - [ ] Call GetUserTenants for all accessible tenants
- [ ] Add route to protected group

**Test Manually:**
- [ ] GET /auth/me returns current user and active tenant
- [ ] GET /auth/me includes all accessible tenants
- [ ] Switch tenant → GET /auth/me shows new active tenant

---

### Day 2: Testing + Optional (2-4 hours)

#### Morning Session (1-2 hours)

**Unit Tests:**
- [ ] `TestSwitchTenant_Success`
- [ ] `TestSwitchTenant_UnauthorizedTenant`
- [ ] `TestSwitchTenant_InactiveTenant`
- [ ] `TestSwitchTenant_ExpiredSubscription`
- [ ] `TestGetUserTenants_MultipleTenantsReturned`
- [ ] `TestVerifyEmail_ValidToken`
- [ ] `TestVerifyEmail_ExpiredToken`
- [ ] `TestVerifyEmail_UsedToken`
- [ ] `TestLogin_EmailNotVerified`
- [ ] `TestGetCurrentUser_ReturnsUserAndTenant`

**Integration Tests:**
- [ ] Full tenant switching flow
- [ ] Full email verification flow
- [ ] GET /auth/me after tenant switch

**Run Test Suite:**
```bash
go test ./internal/service/auth/... -v
go test ./internal/handler/... -v
go test ./... -cover
```

---

#### Afternoon Session (1-2 hours, Optional)

**Change Password (If Time Permits):**
- [ ] Create `ChangePasswordRequest` DTO
- [ ] Implement `ChangePassword(userID, oldPassword, newPassword)` service method
  - [ ] Verify old password
  - [ ] Validate new password strength
  - [ ] Hash new password
  - [ ] Update user password_hash
  - [ ] Revoke all refresh tokens
- [ ] Create `ChangePassword` handler
- [ ] Add route to protected group
- [ ] Write unit tests
- [ ] Write integration test

**Final Checks:**
- [ ] All tests passing
- [ ] No lint errors: `golangci-lint run`
- [ ] API documentation updated
- [ ] Deployment notes created

---

## 📚 API Documentation Updates

### New Endpoints

#### POST /api/v1/auth/verify-email

**Description:** Verify user email with token from verification email

**Authentication:** None required

**Request:**
```json
{
  "token": "abc123def456..."
}
```

**Response (200 OK):**
```json
{
  "message": "Email verified successfully",
  "email": "user@example.com"
}
```

**Error Responses:**
- `400 Bad Request` - Token expired or already used
- `404 Not Found` - Invalid token

---

#### POST /api/v1/auth/switch-tenant

**Description:** Switch user's active tenant

**Authentication:** Required (Bearer token)

**Request:**
```json
{
  "tenant_id": "cm1tenant456"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "active_tenant": {
    "id": "cm1tenant456",
    "name": "PT Another Company",
    "role": "ADMIN",
    "status": "ACTIVE"
  }
}
```

**Error Responses:**
- `403 Forbidden` - User doesn't have access to tenant
- `403 Forbidden` - Tenant subscription expired
- `404 Not Found` - Tenant not found

---

#### GET /api/v1/auth/tenants

**Description:** Get all tenants accessible to current user

**Authentication:** Required (Bearer token)

**Response (200 OK):**
```json
{
  "tenants": [
    {
      "id": "cm1tenant123",
      "name": "PT Example Indonesia",
      "role": "OWNER",
      "status": "ACTIVE"
    },
    {
      "id": "cm1tenant456",
      "name": "PT Another Company",
      "role": "ADMIN",
      "status": "TRIAL"
    }
  ]
}
```

---

#### GET /api/v1/auth/me

**Description:** Get current authenticated user profile and tenant info

**Authentication:** Required (Bearer token)

**Response (200 OK):**
```json
{
  "user": {
    "id": "cm1user123",
    "email": "john@example.com",
    "full_name": "John Doe",
    "phone": "081234567890",
    "email_verified": true,
    "is_system_admin": false,
    "last_login_at": "2025-12-17T10:30:00Z"
  },
  "active_tenant": {
    "id": "cm1tenant123",
    "name": "PT Example Indonesia",
    "role": "OWNER",
    "status": "ACTIVE"
  },
  "tenants": [
    {
      "id": "cm1tenant123",
      "name": "PT Example Indonesia",
      "role": "OWNER",
      "status": "ACTIVE"
    }
  ]
}
```

---

#### POST /api/v1/auth/change-password (Optional)

**Description:** Change user password (requires current password)

**Authentication:** Required (Bearer token)

**Request:**
```json
{
  "old_password": "CurrentPass123!",
  "new_password": "NewSecurePass456!"
}
```

**Response (200 OK):**
```json
{
  "message": "Password changed successfully. Please log in again with your new password."
}
```

**Error Responses:**
- `401 Unauthorized` - Old password incorrect
- `400 Bad Request` - New password doesn't meet requirements

**Notes:**
- All refresh tokens are revoked after successful password change
- User must log in again with new password

---

## 🎓 Implementation Guidelines

### Code Quality Standards

**1. Error Handling:**
```go
// ✅ GOOD - Use custom error types
if !userTenant.IsActive {
    return nil, errors.NewAuthorizationError("Access to tenant has been revoked")
}

if tenant.Status == "SUSPENDED" {
    return nil, errors.NewSubscriptionError("Tenant subscription is suspended")
}

// ❌ BAD - Generic errors
if !userTenant.IsActive {
    return nil, fmt.Errorf("access revoked")
}
```

**2. Validation:**
```go
// ✅ GOOD - Explicit validation
func (s *AuthService) SwitchTenant(userID string, newTenantID string) error {
    // 1. Validate user-tenant relationship
    var userTenant UserTenant
    err := s.db.Where("user_id = ? AND tenant_id = ? AND is_active = ?",
        userID, newTenantID, true).First(&userTenant).Error
    if err != nil {
        return errors.NewAuthorizationError("You don't have access to this tenant")
    }

    // 2. Validate tenant status
    var tenant Tenant
    err = s.db.Where("id = ?", newTenantID).First(&tenant).Error
    if err != nil {
        return errors.NewNotFoundError("Tenant")
    }

    // 3. Check subscription
    if tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" {
        return errors.NewSubscriptionError("Tenant subscription is not active")
    }

    if tenant.Status == "TRIAL" && tenant.TrialEndsAt.Before(time.Now()) {
        return errors.NewSubscriptionError("Tenant trial period has expired")
    }

    // All validations passed
    return nil
}

// ❌ BAD - No validation
func (s *AuthService) SwitchTenant(userID string, newTenantID string) error {
    // Just generate new token without checking anything
}
```

**3. Testing:**
```go
// ✅ GOOD - Comprehensive test coverage
func TestSwitchTenant_UnauthorizedTenant(t *testing.T) {
    db := setupTestDB(t)
    service := setupTestAuthService(db)

    // User A and Tenant B (no relationship)
    userA := createTestUser(db, "user-a")
    tenantB := createTestTenant(db, "tenant-b")

    // Should fail
    _, err := service.SwitchTenant(userA.ID, tenantB.ID)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "don't have access")
}
```

---

## 🚀 Next Steps

### Immediate Actions (Before Starting Phase 3)

1. **Review This Analysis** (30 minutes)
   - Discuss with team
   - Prioritize features (critical vs optional)
   - Confirm timeline

2. **Create Feature Branch** (5 minutes)
   ```bash
   git checkout -b feature/phase-3-mvp
   ```

3. **Verify Development Environment** (15 minutes)
   - SMTP configured for email verification testing
   - Database migrations up to date
   - Redis running
   - Tests passing: `go test ./...`

4. **Coordinate with Frontend** (30 minutes)
   - Share new endpoint specifications
   - Discuss tenant switching UI
   - Confirm email verification flow

---

### Phase 3 Kickoff Checklist

- [ ] This analysis reviewed by team
- [ ] Feature scope agreed (critical only vs critical + optional)
- [ ] Feature branch created
- [ ] Development environment ready
- [ ] Frontend team notified
- [ ] API documentation template prepared
- [ ] Test database seeded with multi-tenant data

---

### Post-Phase 3 (Phase 4 Preview)

After Phase 3 completion, proceed to **Phase 4: Background Jobs & Operations** (Week 4-5):

Planned features:
- Background job scheduler (robfig/cron)
- Token cleanup jobs
- Email sending queue
- Audit log cleanup
- User profile update endpoints (deferred from Phase 3)
- Admin tenant management (deferred from Phase 3)

---

## 📈 Success Metrics

### Development Metrics
- [ ] All 3-4 MVP features implemented (tenant switch, email verify, GET /me, optional change password)
- [ ] Unit test coverage ≥ 80%
- [ ] Zero critical security vulnerabilities
- [ ] Zero high-severity code issues
- [ ] API documentation 100% complete

### Security Metrics
- [ ] Tenant switching validates access control
- [ ] Email verification prevents unverified logins
- [ ] Subscription status checked on tenant switch
- [ ] Zero authentication bypasses in security tests
- [ ] All existing Phase 2 security features intact

### Operational Metrics
- [ ] Email verification completion rate ≥ 90%
- [ ] Tenant switching works across all roles
- [ ] API response time <200ms (p95)
- [ ] Zero downtime during deployment

### User Experience Metrics
- [ ] Clear error messages for validation failures
- [ ] Tenant switching completes in <500ms
- [ ] Email verification flow completion time <1 minute
- [ ] GET /auth/me returns all necessary data for frontend

---

**Document Version:** 1.0
**Last Updated:** 2025-12-17
**Next Review:** After Phase 3 completion
**Owner:** Backend Team

---

## Appendix A: Quick Reference

### File Locations

**Service Layer:**
- `internal/service/auth/auth_service.go` - Add new methods here

**Handler Layer:**
- `internal/handler/auth_handler.go` - Add new handlers here

**DTO Layer:**
- `internal/dto/auth_dto.go` - Add request/response DTOs here

**Router Layer:**
- `internal/router/router.go` - Add routes here

**Tests:**
- `internal/service/auth/*_test.go` - Unit tests
- `test/integration/api/auth_test.go` - Integration tests

---

## Appendix B: Comparison with Phase 2

| Aspect | Phase 2 | Phase 3 MVP |
|--------|---------|-------------|
| **Effort** | 26-36 hours (5 days) | 4-8 hours (1-2 days) |
| **Complexity** | High (new security primitives) | Low (wiring existing services) |
| **Risk** | Medium (CSRF, email delivery) | Low (incremental additions) |
| **New Tables** | None (reused Phase 1 tables) | None (reused Phase 1 tables) |
| **Dependencies** | External (SMTP, Redis) | Minimal (existing auth flow) |
| **Testing** | Comprehensive (15+ tests) | Focused (8-10 tests) |
| **Deployment** | Breaking changes (CSRF) | Non-breaking (new endpoints) |

**Key Difference:** Phase 3 is much simpler because infrastructure exists. Phase 2 built the foundation, Phase 3 just adds missing endpoints.

---

**END OF ANALYSIS**
