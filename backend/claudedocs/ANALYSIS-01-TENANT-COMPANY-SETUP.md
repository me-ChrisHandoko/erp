# Analysis Report: 01-TENANT-COMPANY-SETUP Module
## Pre-Implementation MVP Analysis

**Analysis Date:** 2025-12-18
**Analysis Method:** Sequential Deep Thinking (--ultrathink)
**Analysis Scope:** Week 1 Foundation Setup Module
**Analyst:** Claude Code with Sequential MCP

---

## Executive Summary

**Overall Verdict:** ✅ **READY FOR IMPLEMENTATION** (with critical fixes)

**Readiness Score:** 7/10 (GOOD)

**Recommended Action:** Address 4 critical issues on Day 0, then proceed with Week 1 implementation.

**Key Findings:**
- ✅ Solid architectural foundation with multi-tenant support
- ✅ Well-structured API design following REST conventions
- ✅ Comprehensive business logic documented
- ⚠️ 4 critical security/functional issues require immediate attention
- ⚠️ Some transaction consistency gaps need addressing
- ✅ Realistic timeline with proper task breakdown

---

## Critical Issues (Must Fix Before Implementation)

### 🚨 Issue #1: SVG Upload Security Vulnerability
**Severity:** CRITICAL (XSS Risk)
**Impact:** Cross-site scripting attacks possible
**Current State:** Spec allows SVG uploads without sanitization

**Problem:**
```go
// Current spec mentions: "formats: jpg, png, svg"
// SVG files can contain embedded JavaScript
```

**Fix Required:**
```go
// Option 1: Disable SVG for MVP (RECOMMENDED)
allowedFormats := []string{"image/jpeg", "image/png"}

// Option 2: If SVG needed, sanitize rigorously
// - Validate magic bytes (not just extension)
// - Strip all script tags and event handlers
// - Use Content-Security-Policy headers
```

**Effort:** 1 hour
**Priority:** MUST FIX on Day 0

---

### 🚨 Issue #2: Subscription Status Validation Missing
**Severity:** CRITICAL (Billing Bypass)
**Impact:** Users with expired subscriptions can access system
**Current State:** TenantContextMiddleware doesn't validate subscription status

**Problem:**
- Spec mentions checking `Tenant.status` (TRIAL, ACTIVE, EXPIRED, etc.)
- Middleware implementation not shown
- Users could access tenant after subscription expires

**Fix Required:**
```go
// Add to middleware/tenant_context.go
func ValidateSubscriptionMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetString("tenant_id")

        var tenant models.Tenant
        db.Preload("Subscription").First(&tenant, "id = ?", tenantID)

        // Check tenant status
        if tenant.Status == models.TenantStatusExpired ||
           tenant.Status == models.TenantStatusSuspended {
            c.JSON(403, gin.H{"error": "Subscription expired or suspended"})
            c.Abort()
            return
        }

        // Check trial period
        if tenant.Status == models.TenantStatusTrial {
            if tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
                c.JSON(403, gin.H{"error": "Trial period expired"})
                c.Abort()
                return
            }
        }

        // Check grace period for past due
        if tenant.Status == models.TenantStatusPastDue {
            if tenant.Subscription != nil &&
               tenant.Subscription.GracePeriodEnds != nil &&
               time.Now().After(*tenant.Subscription.GracePeriodEnds) {
                c.JSON(403, gin.H{"error": "Subscription payment overdue"})
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

**Effort:** 2-3 hours
**Priority:** MUST FIX on Day 0

---

### 🚨 Issue #3: Email Service Integration Incomplete
**Severity:** CRITICAL (Functional Blocker)
**Impact:** User invitation flow broken without email
**Current State:** Code mentions "send invitation email" but implementation missing

**Problem:**
```go
// From spec line 1015-1017:
// user, invitationToken, err = s.authService.CreateUserWithInvitation(req.Email, req.FullName, req.Phone)
// But email sending implementation not shown
```

**Fix Required:**
1. Integrate with existing email service (should already exist from auth module)
2. Create invitation email template
3. Generate secure invitation token (with expiry)
4. Send email with activation link
5. Add retry logic for email failures

**Effort:** 3-4 hours
**Priority:** MUST FIX on Day 0

---

### 🚨 Issue #4: Minimum Bank Account Validation Missing
**Severity:** CRITICAL (Business Logic)
**Impact:** Could delete all bank accounts, breaking invoice generation
**Current State:** Spec says "Minimum 1 bank account required" but not enforced

**Problem:**
```go
// DELETE /api/v1/company/banks/:id
// No check for minimum bank account count
// User could delete last bank account
```

**Fix Required:**
```go
func (s *CompanyService) DeleteBankAccount(ctx context.Context, tenantID, bankID string) error {
    // Get company
    company, err := s.GetCompanyByTenantID(ctx, tenantID)
    if err != nil {
        return err
    }

    // Count active banks
    var bankCount int64
    s.db.Model(&models.CompanyBank{}).
        Where("company_id = ? AND is_active = ?", company.ID, true).
        Count(&bankCount)

    if bankCount <= 1 {
        return errors.New("cannot delete last bank account - minimum 1 required")
    }

    // Soft delete
    return s.db.Model(&models.CompanyBank{}).
        Where("id = ?", bankID).
        Update("is_active", false).Error
}
```

**Effort:** 30 minutes
**Priority:** MUST FIX on Day 0

---

## High Priority Issues (Fix During Week 1)

### Issue #5: NPWP Uniqueness Not Enforced at Database Level
**Severity:** HIGH
**Impact:** Two companies could use same NPWP (violates Indonesian law)

**Fix:**
```sql
-- Add to migration file
CREATE UNIQUE INDEX idx_companies_npwp ON companies(npwp)
WHERE npwp IS NOT NULL;
```

**Effort:** 10 minutes
**When:** Day 1 (with Company Profile implementation)

---

### Issue #6: Transaction Consistency Gaps
**Severity:** HIGH
**Impact:** Race conditions possible, data inconsistency risk

**Problems:**
1. `AddBankAccount` doesn't use transaction (unset primary + create new)
2. `RemoveUser` doesn't use transaction (check admin count + delete)
3. `UpdateCompany` doesn't use transaction

**Fix Pattern:**
```go
// Use db.Transaction helper for cleaner code
func (s *CompanyService) AddBankAccount(ctx context.Context, tenantID string, req *AddBankRequest) (*models.CompanyBank, error) {
    company, err := s.GetCompanyByTenantID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    var bank *models.CompanyBank

    err = s.db.Transaction(func(tx *gorm.DB) error {
        // Unset other primary banks if needed
        if req.IsPrimary {
            if err := tx.Model(&models.CompanyBank{}).
                Where("company_id = ?", company.ID).
                Update("is_primary", false).Error; err != nil {
                return err
            }
        }

        // Create new bank
        bank = &models.CompanyBank{
            CompanyID:     company.ID,
            BankName:      req.BankName,
            AccountNumber: req.AccountNumber,
            AccountName:   req.AccountName,
            BranchName:    req.BranchName,
            IsPrimary:     req.IsPrimary,
            IsActive:      true,
        }

        return tx.Create(bank).Error
    })

    return bank, err
}
```

**Effort:** 2-3 hours to refactor all operations
**When:** Day 1-4 (as implementing each service method)

---

### Issue #7: Audit Logging Missing
**Severity:** HIGH
**Impact:** No audit trail for sensitive RBAC operations

**Fix:**
```go
// After each RBAC change, create audit log
auditLog := &models.AuditLog{
    TenantID:      tenantID,
    UserID:        currentUserID,
    Action:        "USER_ROLE_CHANGED",
    ResourceType:  "USER_TENANT",
    ResourceID:    userTenantID,
    Changes:       fmt.Sprintf("Role changed from %s to %s", oldRole, newRole),
    IPAddress:     c.ClientIP(),
    UserAgent:     c.Request.UserAgent(),
}
db.Create(auditLog)
```

**Effort:** 2 hours
**When:** Day 5-7 (Tenant Management implementation)

---

### Issue #8: Rate Limiting Not Applied to Invitation Endpoint
**Severity:** MEDIUM-HIGH
**Impact:** Spam/abuse possible on invitation endpoint

**Fix:**
```go
// In router registration
inviteGroup := tenantGroup.Group("/users/invite")
inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5)) // 5 invites per minute
inviteGroup.POST("", tenantHandler.InviteUser)
```

**Effort:** 15 minutes
**When:** Day 5 (route registration)

---

## Medium Priority Issues (Can Address in Week 2)

### Issue #9: Idempotency Support Missing
**Severity:** MEDIUM
**Impact:** Duplicate requests could create duplicate records

**Recommendation:** Add idempotency middleware for POST operations
- Use `Idempotency-Key` header
- Store request hash in Redis with 24h TTL
- Return cached response if duplicate detected

**Effort:** 3-4 hours
**Defer to:** Week 2 optimization

---

### Issue #10: Race Condition in Primary Bank Selection
**Severity:** LOW-MEDIUM
**Impact:** Two simultaneous requests could both become primary

**Fix:** Use SELECT FOR UPDATE or serializable transaction isolation

**Effort:** 1 hour
**Defer to:** Week 2 if time permits

---

### Issue #11: Enhanced NPWP Validation (Check Digit)
**Severity:** LOW
**Impact:** Could accept invalid NPWP that passes format check

**Current:** Pattern validation only: `^\d{2}\.\d{3}\.\d{3}\.\d-\d{3}\.\d{3}$`
**Enhancement:** Add Luhn-like check digit algorithm validation

**Effort:** 2 hours (research + implementation)
**Defer to:** Phase 2

---

## API Design Quality Assessment

### ✅ Strengths
1. **RESTful Conventions:** Proper use of GET/POST/PUT/DELETE verbs
2. **Consistent Response Format:** Standardized success/error structure
3. **Proper Status Codes:** 200, 201, 400, 403, 404 correctly used
4. **Multi-Tenant Isolation:** X-Tenant-ID header pattern
5. **Security Headers:** CSRF token, Authorization bearer token
6. **Filtering Support:** Query parameters for role, isActive

### ⚠️ Improvements Needed
1. **Pagination Missing:** Company bank list has no pagination
2. **No Search/Filter:** Can't filter banks by name or isPrimary
3. **Bulk Operations Missing:** Can't invite multiple users at once
4. **Logo Metadata Missing:** Response should include size, format, dimensions

### 📋 Recommendations
```http
# Add pagination to bank list
GET /api/v1/company/banks?page=1&limit=20

# Add filtering
GET /api/v1/company/banks?bankName=BCA&isPrimary=true

# Enhanced logo response
{
  "logoUrl": "https://cdn.example.com/logo.png",
  "metadata": {
    "size": 245680,
    "format": "image/png",
    "width": 512,
    "height": 512,
    "uploadedAt": "2025-01-15T10:00:00Z"
  }
}
```

---

## Security Assessment

### ✅ Security Strengths
1. **Multi-Tenant Isolation:** Proper tenantID filtering in queries
2. **JWT Authentication:** Token-based auth with refresh tokens
3. **Role-Based Access Control:** RBAC middleware implementation
4. **CSRF Protection:** X-CSRF-Token validation
5. **Input Validation:** NPWP, phone, email format validation
6. **Soft Deletes:** Audit trail preserved with isActive flag
7. **Business Rule Protection:** Cannot delete OWNER, cannot invite duplicate OWNER

### 🚨 Security Vulnerabilities
1. **SVG Upload XSS** (CRITICAL) - Addressed in Issue #1
2. **Subscription Bypass** (CRITICAL) - Addressed in Issue #2
3. **Missing Rate Limiting** (HIGH) - Addressed in Issue #8
4. **Bank Account Plaintext** (MEDIUM) - Consider encryption at rest
5. **No Audit Logging** (HIGH) - Addressed in Issue #7

### 📊 Security Score: 6/10 → Target 9/10 after fixes

---

## Business Logic Assessment

### ✅ Business Logic Strengths
1. **One Company Per Tenant:** Enforced at application layer
2. **Primary Bank Auto-Unset:** Proper exclusive constraint
3. **OWNER Protection:** Cannot change/remove OWNER role
4. **Last ADMIN Protection:** Cannot remove last ADMIN
5. **User Invitation Flow:** Handles existing vs new users correctly
6. **Reactivation Logic:** Can reactivate deactivated users

### ⚠️ Business Logic Gaps
1. **PKP Validation:** Should enforce NPWP required if isPKP=true
2. **PPN Rate Range:** 0-100% too broad (constrain to 0-25%)
3. **Email Expiry:** Invitation tokens should expire (24-48h)
4. **Concurrent Updates:** Race conditions possible (addressed in Issue #6)

### 📋 Recommended Validations
```go
// Custom validator: PKP requires NPWP
func ValidatePKPStatus(req *CreateCompanyRequest) error {
    if req.IsPKP && (req.NPWP == nil || *req.NPWP == "") {
        return errors.New("NPWP is required for PKP entities")
    }
    return nil
}

// Constrain PPN rate
type CreateCompanyRequest struct {
    PPNRate decimal.Decimal `json:"ppnRate" validate:"required,gte=0,lte=25"`
}
```

---

## Database Schema Assessment

### ✅ Schema Completeness
- All required fields defined ✓
- Relationships correct (1:1, 1:many) ✓
- Nullable fields appropriate ✓
- Default values specified ✓
- Soft delete support (isActive) ✓

### ⚠️ Missing Indexes
```sql
-- Add these to migration file

-- Company: NPWP uniqueness (already covered in Issue #5)
CREATE UNIQUE INDEX idx_companies_npwp ON companies(npwp)
WHERE npwp IS NOT NULL;

-- UserTenant: Composite index for lookups
CREATE INDEX idx_user_tenant_lookup ON user_tenants(user_id, tenant_id, is_active);

-- CompanyBank: Primary bank lookup
CREATE INDEX idx_company_bank_primary ON company_banks(company_id, is_primary)
WHERE is_active = true;

-- Performance indexes
CREATE INDEX idx_companies_tenant ON companies(id) WHERE is_active = true;
CREATE INDEX idx_tenants_status ON tenants(status, trial_ends_at);
```

---

## Indonesian Compliance Assessment

### ✅ Compliance Strengths
1. **NPWP Format:** Validated with regex pattern ✓
2. **PKP Status:** Tracked with Faktur Pajak series ✓
3. **PPN Rate:** Configurable (11% default for 2025) ✓
4. **Entity Types:** CV, PT, UD, Firma enum ✓
5. **Bank Integration:** Local bank support (BCA, Mandiri, BRI) ✓
6. **Localization:** IDR currency, Asia/Jakarta timezone, id-ID locale ✓
7. **Phone Validation:** Indonesian format (+628xxx or 08xxx) ✓

### ⚠️ Compliance Gaps (Defer to Phase 2)
1. **NIB (Nomor Induk Berusaha):** Not tracked - required since 2020
2. **SIUP/TDP:** Business licenses not tracked (being phased out)
3. **KTP Validation:** Owner ID card not validated against NPWP

### 📊 Compliance Score: 8/10 (Good for MVP)

**Recommendation:** Current implementation sufficient for MVP. Add NIB field in Phase 2.

---

## Testing Strategy Assessment

### ✅ Test Coverage Strengths
- **Comprehensive Test Checklist:** 25 test scenarios documented
- **Happy Path Coverage:** All CRUD operations
- **Validation Testing:** Error cases included
- **Security Testing:** RBAC and multi-tenant isolation
- **Edge Cases:** Duplicate OWNER, last ADMIN removal

### ⚠️ Testing Gaps
1. **No Concurrency Tests:** Race conditions not tested
2. **No Performance Tests:** Pagination, large datasets
3. **Missing Email Tests:** Invitation flow integration
4. **No Subscription Tests:** Expired/suspended tenant access
5. **Missing Audit Tests:** Log creation verification

### 📋 Enhanced Test Checklist

#### Pre-Implementation Tests (Day 0)
- [ ] Subscription validation middleware blocks expired tenants
- [ ] Logo upload rejects SVG files (or sanitizes correctly)
- [ ] Email service sends invitation successfully
- [ ] Minimum bank account deletion prevented

#### Company Profile Tests (Day 1-2)
- [ ] Create company with valid data → 201 Created
- [ ] Create company with duplicate NPWP → 400 (unique constraint)
- [ ] Create company with invalid NPWP format → 400
- [ ] Create PKP company without NPWP → 400
- [ ] Update company profile → 200 OK
- [ ] Upload logo (jpg/png) → 200 OK with metadata
- [ ] Upload oversized file → 400 Bad Request

#### Bank Account Tests (Day 3-4)
- [ ] Add bank account → 201 Created
- [ ] Add primary bank → unsets other primary banks
- [ ] Update bank account → 200 OK
- [ ] Delete bank (multiple exist) → 200 OK
- [ ] Delete last bank → 400 (minimum 1 required)
- [ ] Concurrent primary bank creation → only 1 primary

#### Tenant Management Tests (Day 5-7)
- [ ] Get tenant details → 200 OK with subscription
- [ ] List users (all) → 200 OK
- [ ] List users (filter by role) → 200 OK with filtered results
- [ ] Invite new user → creates user + sends email + 201 Created
- [ ] Invite existing user → creates UserTenant link + 201 Created
- [ ] Invite OWNER → 400 Forbidden
- [ ] Non-admin invites user → 403 Forbidden
- [ ] Update user role (valid) → 200 OK + audit log created
- [ ] Update OWNER role → 400 Forbidden
- [ ] Promote to OWNER → 400 Forbidden
- [ ] Remove user (valid) → 200 OK + soft delete
- [ ] Remove OWNER → 400 Forbidden
- [ ] Remove last ADMIN → 400 Forbidden
- [ ] Concurrent admin removal → last ADMIN protected

#### Multi-Tenant Isolation Tests
- [ ] User A (Tenant 1) cannot access Tenant 2 company profile → 403
- [ ] User B (Tenant 2) cannot list Tenant 1 users → 403
- [ ] User C (no tenant access) cannot access any endpoint → 403
- [ ] Invalid X-Tenant-ID header → 403 Forbidden

#### Performance Tests (Week 2)
- [ ] List 1000 users with pagination → <500ms response
- [ ] List 100 bank accounts → <200ms response
- [ ] Concurrent user invitations (10 simultaneous) → all succeed

---

## Implementation Timeline Assessment

### Proposed Timeline (Week 1)
```
Day 0: Critical Fixes (NEW - ADD TO SCHEDULE)
├─ Subscription validation middleware (2-3h)
├─ Logo upload security (1h)
├─ Email integration (3-4h)
└─ Minimum bank validation (30m)

Day 1-2: Company Profile CRUD
├─ DTOs (2-3h)
├─ Service layer (4-6h)
├─ Handler (2-3h)
├─ Routes (1h)
├─ Validation (2-3h)
└─ Unit tests (4-5h)
Total: 15-21 hours ✓ Feasible

Day 3-4: Company Bank Management
├─ Service layer (3-4h)
├─ Handler (2-3h)
├─ Routes (1h)
└─ Integration tests (3-4h)
Total: 9-12 hours ✓ Feasible (1.5 days, buffer available)

Day 5-7: Tenant Management Module
├─ DTOs (2h)
├─ Service layer (6-8h)
├─ Handler (3-4h)
├─ Middleware (2-3h)
├─ Routes (1-2h)
├─ Unit tests (4-5h)
├─ Integration tests (4-5h)
└─ Multi-tenant tests (2-3h)
Total: 24-32 hours ✓ Feasible
```

### Revised Timeline Assessment
**Original:** 7 days (optimistic)
**Revised:** 8 days (realistic with Day 0 fixes)
**Status:** ✅ ACHIEVABLE

### Risk Factors
1. **Email Integration Complexity:** Could take longer than estimated
2. **Transaction Refactoring:** May reveal unexpected issues
3. **Multi-Tenant Testing:** Edge cases could surface
4. **Subscription Logic:** Business rules may need clarification

### Mitigation Strategies
- **Time-Box Development:** Strict daily limits
- **Daily Standups:** Identify blockers early
- **Defer Nice-to-Haves:** Focus on MVP features only
- **Parallel Testing:** Write tests while implementing

---

## MVP Scope Validation

### ✅ IN SCOPE (Appropriately Sized)
1. Single company profile per tenant ✓
2. Basic CRUD operations ✓
3. NPWP and PKP tracking ✓
4. Bank account management ✓
5. Invoice number format configuration ✓
6. User-tenant role management (6 roles) ✓
7. User invitation flow ✓
8. Role-based access control ✓

### ❌ OUT OF SCOPE (Correctly Deferred)
1. Multiple companies per tenant (use multi-tenant instead)
2. Advanced logo editing/cropping
3. Tax rate history tracking
4. Multi-currency support
5. Advanced business hours (holidays, shifts)
6. Per-module granular permissions
7. User activity dashboard
8. Custom role creation
9. NIB tracking (Phase 2)
10. Bank account encryption (Phase 2)

### 📊 Scope Assessment: ✅ APPROPRIATE FOR MVP

---

## Code Quality Recommendations

### Architectural Patterns

#### 1. Repository Pattern (Optional but Recommended)
```go
// Improves testability and separation of concerns
type CompanyRepository interface {
    Create(ctx context.Context, company *models.Company) error
    FindByTenantID(ctx context.Context, tenantID string) (*models.Company, error)
    Update(ctx context.Context, company *models.Company) error
}

type companyRepositoryImpl struct {
    db *gorm.DB
}

// Service uses repository, not db directly
type CompanyService struct {
    repo CompanyRepository
}
```

**Benefit:** Easier unit testing (mock repository)
**Effort:** 3-4 hours setup, reusable pattern
**Priority:** MEDIUM (nice-to-have for MVP)

---

#### 2. DTO Pattern (Already Planned ✓)
```go
// API layer (handler)
type CreateCompanyRequest struct { ... }
type CompanyResponse struct { ... }

// Domain layer (models)
type Company struct { ... }

// Converter/Mapper
func ToCompanyResponse(company *models.Company) *CompanyResponse {
    return &CompanyResponse{
        ID:   company.ID,
        Name: company.Name,
        // ... map all fields
    }
}
```

**Benefit:** API versioning, backward compatibility
**Status:** ✅ Already in spec

---

#### 3. Centralized Validation
```go
// validator/company_validator.go
type CompanyValidator struct {}

func (v *CompanyValidator) ValidateCreate(req *dto.CreateCompanyRequest) error {
    if err := v.ValidateNPWP(req.NPWP); err != nil {
        return err
    }
    if err := v.ValidatePKPStatus(req); err != nil {
        return err
    }
    return nil
}

func (v *CompanyValidator) ValidateNPWP(npwp *string) error {
    // Centralized NPWP validation logic
}
```

**Benefit:** Reusable validation, easier testing
**Effort:** 2-3 hours
**Priority:** MEDIUM (improve code quality)

---

#### 4. Error Handling Strategy
```go
// pkg/apperror/codes.go
type ErrorCode string

const (
    ErrCompanyNotFound      ErrorCode = "COMPANY_NOT_FOUND"
    ErrCompanyAlreadyExists ErrorCode = "COMPANY_ALREADY_EXISTS"
    ErrValidationFailed     ErrorCode = "VALIDATION_FAILED"
    ErrNPWPInvalid          ErrorCode = "NPWP_INVALID"
    ErrNPWPDuplicate        ErrorCode = "NPWP_DUPLICATE"
    // ... all error codes
)

// pkg/apperror/error.go
type AppError struct {
    Code       ErrorCode              `json:"code"`
    Message    string                 `json:"message"`
    Details    []ValidationError      `json:"details,omitempty"`
    StatusCode int                    `json:"-"`
}

func NewValidationError(details []ValidationError) *AppError {
    return &AppError{
        Code:       ErrValidationFailed,
        Message:    "Validation failed",
        Details:    details,
        StatusCode: http.StatusBadRequest,
    }
}
```

**Benefit:** Consistent error handling, i18n-ready
**Effort:** 2 hours
**Priority:** HIGH (improves API quality)

---

### Code Style Guidelines

1. **Naming Conventions:**
   - Use camelCase for JSON fields: `companyId`, `tenantId`
   - Use PascalCase for Go structs: `CompanyService`, `CreateRequest`
   - Use snake_case for database columns: `company_id`, `tenant_id`

2. **Comment Standards:**
   - Add godoc comments for all exported functions
   - Explain business rules inline
   - Document validation logic

3. **Function Length:**
   - Keep functions under 50 lines
   - Extract complex logic into helper functions
   - Use meaningful function names

4. **Error Handling:**
   - Always check errors immediately
   - Wrap errors with context: `fmt.Errorf("failed to create company: %w", err)`
   - Log errors before returning

---

## Risk Assessment & Mitigation

### RISK #1: Timeline Slippage
**Probability:** MEDIUM (40%)
**Impact:** HIGH (delays subsequent modules)

**Causes:**
- Underestimated complexity
- Technical blockers (email integration, transaction issues)
- Scope creep
- Testing reveals major bugs

**Mitigation:**
1. Add Day 0 for critical fixes (1 day buffer)
2. Time-box each task strictly
3. Daily progress tracking with blocker list
4. Defer optimization tasks to Week 2
5. **Fallback:** Reduce test coverage temporarily, catch up in Week 2

**Contingency Plan:**
- If Day 5 behind schedule, defer advanced RBAC tests to Week 2
- Focus on core functionality first, polish later

---

### RISK #2: Multi-Tenant Data Leakage
**Probability:** LOW (10%)
**Impact:** CRITICAL (security breach, regulatory violation)

**Causes:**
- Missing tenantID filter in query
- Middleware bypass
- Transaction isolation issue

**Mitigation:**
1. **Code Review Checklist:** Every query MUST include tenantID filter
2. **Integration Tests:** Specific cross-tenant access tests
3. **Middleware Enforcement:** Use TenantContextMiddleware on all routes
4. **Audit:** Manual review of all database queries before deployment
5. **Database-Level Protection:** Consider row-level security (RLS) in PostgreSQL

**Detection:**
```go
// Add to integration tests
func TestCrossTenantIsolation(t *testing.T) {
    // Create data for Tenant A
    tenantA := createTestTenant(t)
    companyA := createTestCompany(t, tenantA.ID)

    // Try to access from Tenant B
    tenantB := createTestTenant(t)
    userB := createTestUser(t, tenantB.ID)

    // Should return 403 or empty result
    resp := makeRequest(t, userB.Token, tenantB.ID, "/api/v1/company")
    assert.Equal(t, 403, resp.StatusCode)
}
```

---

### RISK #3: Subscription Billing Bypass
**Probability:** MEDIUM (30%)
**Impact:** HIGH (revenue loss, SaaS model broken)

**Causes:**
- Incomplete subscription validation
- Middleware not applied to all routes
- Grace period logic incorrect

**Mitigation:**
1. **Day 0 Implementation:** ValidateSubscriptionMiddleware (Issue #2)
2. **Comprehensive Testing:** Test with TRIAL, EXPIRED, SUSPENDED statuses
3. **Grace Period Logic:** Clearly defined and tested
4. **Monitoring:** Production alerts for subscription status changes

**Test Scenarios:**
```go
func TestSubscriptionValidation(t *testing.T) {
    tests := []struct {
        name           string
        tenantStatus   models.TenantStatus
        trialEndsAt    *time.Time
        gracePeriodEnds *time.Time
        expectedCode   int
    }{
        {
            name:         "Active subscription",
            tenantStatus: models.TenantStatusActive,
            expectedCode: 200,
        },
        {
            name:         "Expired trial",
            tenantStatus: models.TenantStatusTrial,
            trialEndsAt:  ptr(time.Now().Add(-1 * time.Hour)),
            expectedCode: 403,
        },
        {
            name:         "Past due within grace period",
            tenantStatus: models.TenantStatusPastDue,
            gracePeriodEnds: ptr(time.Now().Add(24 * time.Hour)),
            expectedCode: 200,
        },
        {
            name:         "Past due after grace period",
            tenantStatus: models.TenantStatusPastDue,
            gracePeriodEnds: ptr(time.Now().Add(-1 * time.Hour)),
            expectedCode: 403,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test subscription validation
        })
    }
}
```

---

### RISK #4: Email Service Failure
**Probability:** MEDIUM (30%)
**Impact:** MEDIUM (invitation flow broken, workaround possible)

**Causes:**
- External service downtime
- Configuration issues
- Rate limiting by email provider

**Mitigation:**
1. **Retry Logic:** Exponential backoff (3 retries)
2. **Fallback:** Manual invitation link generation by admin
3. **Queue System:** Async email processing with job queue
4. **Monitoring:** Email delivery success rate tracking

**Implementation:**
```go
func (s *TenantService) sendInvitationEmail(email, token string) error {
    maxRetries := 3
    var err error

    for i := 0; i < maxRetries; i++ {
        err = s.emailService.SendInvitation(email, token)
        if err == nil {
            return nil
        }

        // Exponential backoff
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }

    // Log failure for manual intervention
    log.Error().
        Str("email", email).
        Err(err).
        Msg("Failed to send invitation email after retries")

    return fmt.Errorf("failed to send invitation email: %w", err)
}
```

**Fallback Plan:**
- Admin can generate manual invitation link in UI
- User can be activated without email (admin action)

---

## Success Criteria (Week 1 Complete)

### Functional Completeness (Must Have)
- [ ] OWNER can create company profile with valid NPWP
- [ ] OWNER can add/update/delete bank accounts
- [ ] System prevents duplicate NPWP across tenants
- [ ] System enforces minimum 1 bank account
- [ ] Primary bank auto-unset works correctly
- [ ] OWNER/ADMIN can view tenant details with subscription status
- [ ] OWNER/ADMIN can invite users (creates account + sends email)
- [ ] OWNER/ADMIN can change user roles (with validation)
- [ ] OWNER/ADMIN can remove users (soft delete)
- [ ] System prevents removing OWNER or last ADMIN
- [ ] RequireRoleMiddleware blocks unauthorized access correctly

### Security Completeness (Must Have)
- [ ] Multi-tenant isolation verified (no cross-tenant data access)
- [ ] Subscription status validated on all protected routes
- [ ] Logo upload secured (no XSS vulnerabilities)
- [ ] CSRF protection working on all POST/PUT/DELETE requests
- [ ] Rate limiting applied to invitation endpoint
- [ ] Audit log created for all RBAC changes
- [ ] All database queries filter by tenantID

### Technical Completeness (Must Have)
- [ ] All endpoints return consistent error format
- [ ] Validation errors include field-level details
- [ ] Database transactions ensure data consistency
- [ ] Unit test coverage ≥80%
- [ ] All 25 integration test scenarios pass
- [ ] No N+1 query issues (use Preload appropriately)
- [ ] Database indexes created for performance

### Deployment Readiness (Must Have)
- [ ] Database migrations run cleanly (up and down)
- [ ] Seed data script populates test data
- [ ] Environment variables documented in .env.example
- [ ] Health check endpoint includes DB connectivity check
- [ ] API documentation generated (Swagger/OpenAPI)

### Performance Targets (Nice to Have)
- [ ] Company profile GET: <100ms response time
- [ ] User list with 100 users: <200ms response time
- [ ] Bank account CRUD: <150ms response time
- [ ] System handles 50 concurrent requests without errors

---

## Post-Week 1 Cleanup Tasks (Week 2)

### Code Quality
- [ ] Implement repository pattern (if time permits)
- [ ] Add comprehensive error code catalog
- [ ] Refactor validation into centralized validator
- [ ] Add godoc comments to all exported functions
- [ ] Code review with checklist

### Performance
- [ ] Add pagination to bank account list
- [ ] Implement caching for company profile (Redis)
- [ ] Optimize user list queries with proper indexes
- [ ] Load testing with 1000 concurrent users

### Security
- [ ] Bank account encryption at rest
- [ ] NPWP check digit validation
- [ ] Enhanced audit logging (include request/response)
- [ ] Security audit with OWASP checklist

### Documentation
- [ ] API documentation with examples
- [ ] Deployment guide with troubleshooting
- [ ] Developer onboarding guide
- [ ] Database schema diagram

---

## Implementation Checklist (Day-by-Day)

### Day 0: Critical Fixes (MUST COMPLETE FIRST)
- [ ] **Subscription Validation Middleware**
  - [ ] Create `middleware/subscription.go`
  - [ ] Implement ValidateSubscriptionMiddleware
  - [ ] Test with TRIAL, ACTIVE, EXPIRED, PAST_DUE statuses
  - [ ] Apply to all protected routes
  - [ ] Integration test: expired tenant blocked

- [ ] **Logo Upload Security**
  - [ ] Restrict to jpg/png only (disable SVG)
  - [ ] Validate file type via magic bytes
  - [ ] Implement 2MB size limit
  - [ ] Return metadata in response
  - [ ] Integration test: SVG rejected

- [ ] **Email Service Integration**
  - [ ] Review existing email service from auth module
  - [ ] Create invitation email template
  - [ ] Implement retry logic with exponential backoff
  - [ ] Test email sending (staging environment)
  - [ ] Add fallback: manual invitation link

- [ ] **Minimum Bank Account Validation**
  - [ ] Add count check in DeleteBankAccount
  - [ ] Return error if deleting last bank
  - [ ] Integration test: cannot delete last bank

- [ ] **Database Indexes**
  - [ ] Create migration: NPWP unique index
  - [ ] Create migration: UserTenant composite index
  - [ ] Create migration: CompanyBank primary index
  - [ ] Test migrations (up and down)

- [ ] **Audit Logging Setup**
  - [ ] Create audit log helper function
  - [ ] Test audit log creation
  - [ ] Verify AuditLog model ready

---

### Day 1-2: Company Profile Module

#### Day 1 Morning: Setup & DTOs
- [ ] Create directory structure:
  ```
  internal/
  ├── dto/
  │   └── company_dto.go
  ├── service/company/
  │   ├── company_service.go
  │   └── validation.go
  ├── handler/
  │   └── company_handler.go
  ```

- [ ] **Define DTOs** (`internal/dto/company_dto.go`)
  - [ ] CreateCompanyRequest struct
  - [ ] UpdateCompanyRequest struct
  - [ ] CompanyResponse struct
  - [ ] CompanyBankResponse struct
  - [ ] Add JSON tags (camelCase)
  - [ ] Add validation tags

- [ ] **Validation** (`internal/service/company/validation.go`)
  - [ ] ValidateNPWP function (regex + format)
  - [ ] ValidatePhoneNumber function (Indonesian format)
  - [ ] ValidatePKPStatus function (NPWP required if PKP)
  - [ ] Unit tests for validation functions

#### Day 1 Afternoon: Service Layer
- [ ] **CompanyService** (`internal/service/company/company_service.go`)
  - [ ] NewCompanyService constructor
  - [ ] GetCompanyByTenantID method
  - [ ] CreateCompany method (with transaction)
  - [ ] UpdateCompany method (partial updates)
  - [ ] Unit tests for service layer (mock db)

#### Day 2 Morning: Handler & Routes
- [ ] **CompanyHandler** (`internal/handler/company_handler.go`)
  - [ ] NewCompanyHandler constructor
  - [ ] GetCompany handler
  - [ ] CreateCompany handler
  - [ ] UpdateCompany handler
  - [ ] UploadLogo handler
  - [ ] Error handling with AppError
  - [ ] Unit tests for handlers (mock service)

- [ ] **Routes** (`internal/router/router.go`)
  - [ ] Register company routes:
    ```go
    companyGroup := protected.Group("/company")
    companyGroup.GET("", companyHandler.GetCompany)
    companyGroup.POST("", companyHandler.CreateCompany)
    companyGroup.PUT("", companyHandler.UpdateCompany)
    companyGroup.POST("/logo", companyHandler.UploadLogo)
    ```

#### Day 2 Afternoon: Testing
- [ ] **Integration Tests** (`tests/integration/company_test.go`)
  - [ ] Test: Create company (valid data) → 201
  - [ ] Test: Create company (duplicate NPWP) → 400
  - [ ] Test: Create company (invalid NPWP) → 400
  - [ ] Test: Create company (PKP without NPWP) → 400
  - [ ] Test: Get company profile → 200
  - [ ] Test: Get company (not found) → 404
  - [ ] Test: Update company → 200
  - [ ] Test: Upload logo (jpg) → 200 with metadata
  - [ ] Test: Upload logo (oversized) → 400
  - [ ] Test: Cross-tenant access → 403

---

### Day 3-4: Company Bank Management

#### Day 3 Morning: Service Layer
- [ ] **BankService Methods** (`internal/service/company/company_service.go`)
  - [ ] ListBankAccounts method (with filters)
  - [ ] AddBankAccount method (transaction-wrapped)
  - [ ] UpdateBankAccount method
  - [ ] DeleteBankAccount method (with min 1 validation)
  - [ ] Primary bank unset logic (in transaction)
  - [ ] Unit tests for bank methods

#### Day 3 Afternoon: Handler & Routes
- [ ] **BankHandler Methods** (`internal/handler/company_handler.go`)
  - [ ] ListBanks handler
  - [ ] AddBank handler
  - [ ] UpdateBank handler
  - [ ] DeleteBank handler
  - [ ] Unit tests for handlers

- [ ] **Routes** (`internal/router/router.go`)
  - [ ] Register bank routes:
    ```go
    bankGroup := companyGroup.Group("/banks")
    bankGroup.GET("", companyHandler.ListBanks)
    bankGroup.POST("", companyHandler.AddBank)
    bankGroup.PUT("/:id", companyHandler.UpdateBank)
    bankGroup.DELETE("/:id", companyHandler.DeleteBank)
    ```

#### Day 4 Morning: Testing
- [ ] **Integration Tests** (`tests/integration/company_bank_test.go`)
  - [ ] Test: List banks → 200 with pagination
  - [ ] Test: Add bank (valid) → 201
  - [ ] Test: Add bank (set primary) → unsets others
  - [ ] Test: Update bank → 200
  - [ ] Test: Delete bank (multiple exist) → 200
  - [ ] Test: Delete last bank → 400
  - [ ] Test: Concurrent primary bank creation → only 1 primary
  - [ ] Test: Transaction rollback on error

#### Day 4 Afternoon: Buffer & Refinement
- [ ] Code review Company + Bank modules
- [ ] Fix any issues found in testing
- [ ] Performance testing (load 100 banks)
- [ ] Documentation updates

---

### Day 5-7: Tenant Management Module

#### Day 5 Morning: DTOs & Middleware
- [ ] **Define DTOs** (`internal/dto/tenant_dto.go`)
  - [ ] TenantResponse struct
  - [ ] UserTenantResponse struct
  - [ ] InviteUserRequest struct
  - [ ] UpdateRoleRequest struct
  - [ ] Add JSON tags and validation

- [ ] **RequireRoleMiddleware** (`internal/middleware/role.go`)
  - [ ] Implement RequireRoleMiddleware function
  - [ ] Check user role in UserTenant
  - [ ] Validate multiple allowed roles
  - [ ] Set user_role in context
  - [ ] Unit tests for middleware

#### Day 5 Afternoon: TenantService (Part 1)
- [ ] **TenantService** (`internal/service/tenant/tenant_service.go`)
  - [ ] NewTenantService constructor
  - [ ] GetTenantByID method (with preload)
  - [ ] ListTenantUsers method (with filters)
  - [ ] Unit tests for read operations

#### Day 6 Morning: TenantService (Part 2)
- [ ] **TenantService** (continued)
  - [ ] InviteUserToTenant method
    - [ ] Check if user exists
    - [ ] Create user if needed (with email)
    - [ ] Create UserTenant link
    - [ ] Send invitation email
    - [ ] Handle reactivation
    - [ ] Audit log creation
  - [ ] Unit tests for invitation flow

#### Day 6 Afternoon: TenantService (Part 3)
- [ ] **TenantService** (continued)
  - [ ] UpdateUserRole method
    - [ ] Validation (cannot change OWNER, etc.)
    - [ ] Update role
    - [ ] Audit log creation
  - [ ] RemoveUserFromTenant method
    - [ ] Validation (cannot remove OWNER, last ADMIN)
    - [ ] Check admin count (in transaction)
    - [ ] Soft delete
    - [ ] Audit log creation
  - [ ] Unit tests for role management

#### Day 7 Morning: Handler & Routes
- [ ] **TenantHandler** (`internal/handler/tenant_handler.go`)
  - [ ] NewTenantHandler constructor
  - [ ] GetTenant handler
  - [ ] ListUsers handler
  - [ ] InviteUser handler (with rate limiting)
  - [ ] UpdateUserRole handler
  - [ ] RemoveUser handler
  - [ ] Unit tests for handlers

- [ ] **Routes** (`internal/router/router.go`)
  - [ ] Register tenant routes with RBAC:
    ```go
    tenantGroup := protected.Group("/tenant")
    tenantGroup.Use(middleware.RequireRoleMiddleware(db, models.UserRoleOwner, models.UserRoleAdmin))
    tenantGroup.GET("", tenantHandler.GetTenant)
    tenantGroup.GET("/users", tenantHandler.ListUsers)

    inviteGroup := tenantGroup.Group("/users/invite")
    inviteGroup.Use(middleware.RateLimitMiddleware(redisClient, 5))
    inviteGroup.POST("", tenantHandler.InviteUser)

    tenantGroup.PUT("/users/:id/role", tenantHandler.UpdateUserRole)
    tenantGroup.DELETE("/users/:id", tenantHandler.RemoveUser)
    ```

#### Day 7 Afternoon: Comprehensive Testing
- [ ] **Integration Tests** (`tests/integration/tenant_test.go`)
  - [ ] Test: Get tenant details → 200 with subscription
  - [ ] Test: List users (all) → 200
  - [ ] Test: List users (filter by role) → 200
  - [ ] Test: List users (filter by isActive) → 200
  - [ ] Test: Invite new user → creates user + email + 201
  - [ ] Test: Invite existing user → creates link + 201
  - [ ] Test: Invite OWNER → 400
  - [ ] Test: Non-admin invites → 403
  - [ ] Test: Update role (valid) → 200 + audit log
  - [ ] Test: Update OWNER role → 400
  - [ ] Test: Promote to OWNER → 400
  - [ ] Test: Remove user (valid) → 200 + soft delete
  - [ ] Test: Remove OWNER → 400
  - [ ] Test: Remove last ADMIN → 400
  - [ ] Test: Concurrent admin removal → protected
  - [ ] Test: Rate limiting on invitation → 429
  - [ ] Test: Cross-tenant user access → 403

- [ ] **Multi-Tenant Isolation Tests**
  - [ ] Test: Tenant A cannot access Tenant B data
  - [ ] Test: User with no tenant access blocked → 403
  - [ ] Test: Invalid X-Tenant-ID → 403

---

### End of Week 1: Final Validation
- [ ] All 25 integration tests passing
- [ ] Code coverage ≥80%
- [ ] Manual API testing with Postman/curl
- [ ] Database migrations tested (up/down)
- [ ] Seed data works correctly
- [ ] Security audit checklist completed
- [ ] Performance targets met
- [ ] Documentation updated
- [ ] Deployment checklist prepared

---

## Monitoring & Observability (Week 2 Setup)

### Metrics to Track
1. **Functional Metrics:**
   - Company profile creation rate
   - User invitation success rate
   - Email delivery success rate
   - API endpoint response times
   - Error rate by endpoint

2. **Security Metrics:**
   - Failed authentication attempts
   - Cross-tenant access attempts (should be 0)
   - Expired subscription access attempts
   - Rate limit violations

3. **Performance Metrics:**
   - Database query duration (P50, P95, P99)
   - API endpoint latency
   - Transaction failure rate
   - Concurrent request handling

### Recommended Tools
- **Logging:** Zerolog (already in use?)
- **Metrics:** Prometheus + Grafana
- **Tracing:** OpenTelemetry
- **Alerting:** PagerDuty or similar

---

## Documentation Requirements

### API Documentation (OpenAPI/Swagger)
- [ ] Generate from code annotations
- [ ] Include request/response examples
- [ ] Document error codes
- [ ] Add authentication flow
- [ ] Publish to API portal

### Developer Documentation
- [ ] Setup instructions
- [ ] Local development guide
- [ ] Testing guide
- [ ] Database migration guide
- [ ] Troubleshooting guide

### Deployment Documentation
- [ ] Environment variables
- [ ] Database setup
- [ ] Migration steps
- [ ] Health check endpoints
- [ ] Monitoring setup

---

## Conclusion

### Overall Assessment
**Specification Quality:** 8/10 (Excellent foundation)
**MVP Readiness:** 7/10 → 9/10 after critical fixes
**Implementation Risk:** MEDIUM (Manageable with proper planning)
**Timeline Feasibility:** ✅ ACHIEVABLE (8 days with buffer)

### Final Recommendation
**✅ PROCEED WITH IMPLEMENTATION** after completing Day 0 critical fixes:
1. Subscription validation middleware (2-3h)
2. Logo upload security (1h)
3. Email service integration (3-4h)
4. Minimum bank account validation (30m)
5. Database indexes (30m)

**Total Day 0 Effort:** ~8 hours (1 working day)

### Next Steps
1. **Immediate:** Review and approve this analysis
2. **Day 0:** Complete critical fixes and validations
3. **Day 1:** Start Company Profile implementation
4. **Daily:** Track progress against checklist
5. **End Week 1:** Complete success criteria validation

---

## Appendix: Code Templates

### A1: AppError Template
```go
// pkg/apperror/error.go
package apperror

import "net/http"

type ErrorCode string

const (
    ErrCompanyNotFound      ErrorCode = "COMPANY_NOT_FOUND"
    ErrCompanyAlreadyExists ErrorCode = "COMPANY_ALREADY_EXISTS"
    ErrValidationFailed     ErrorCode = "VALIDATION_FAILED"
    ErrNPWPInvalid          ErrorCode = "NPWP_INVALID"
    ErrNPWPDuplicate        ErrorCode = "NPWP_DUPLICATE"
    ErrBankAccountNotFound  ErrorCode = "BANK_ACCOUNT_NOT_FOUND"
    ErrCannotDeleteLastBank ErrorCode = "CANNOT_DELETE_LAST_BANK"
    ErrTenantNotFound       ErrorCode = "TENANT_NOT_FOUND"
    ErrUnauthorized         ErrorCode = "UNAUTHORIZED"
    ErrForbidden            ErrorCode = "FORBIDDEN"
    ErrInsufficientPermission ErrorCode = "INSUFFICIENT_PERMISSION"
    ErrUserAlreadyExists    ErrorCode = "USER_ALREADY_EXISTS"
    ErrCannotChangeOwnerRole ErrorCode = "CANNOT_CHANGE_OWNER_ROLE"
    ErrCannotRemoveOwner    ErrorCode = "CANNOT_REMOVE_OWNER"
    ErrCannotRemoveLastAdmin ErrorCode = "CANNOT_REMOVE_LAST_ADMIN"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type AppError struct {
    Code       ErrorCode         `json:"code"`
    Message    string            `json:"message"`
    Details    []ValidationError `json:"details,omitempty"`
    StatusCode int               `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

func NewNotFoundError(code ErrorCode, message string) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: http.StatusNotFound,
    }
}

func NewValidationError(details []ValidationError) *AppError {
    return &AppError{
        Code:       ErrValidationFailed,
        Message:    "Validation failed",
        Details:    details,
        StatusCode: http.StatusBadRequest,
    }
}

func NewForbiddenError(code ErrorCode, message string) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: http.StatusForbidden,
    }
}

func NewBadRequestError(code ErrorCode, message string) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: http.StatusBadRequest,
    }
}
```

### A2: Response Helper Template
```go
// pkg/response/response.go
package response

import (
    "github.com/gin-gonic/gin"
    "backend/pkg/apperror"
)

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   interface{} `json:"error,omitempty"`
}

type PaginatedResponse struct {
    Success    bool        `json:"success"`
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"totalPages"`
}

func Success(c *gin.Context, statusCode int, data interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Data:    data,
    })
}

func SuccessWithPagination(c *gin.Context, data interface{}, pagination Pagination) {
    c.JSON(200, PaginatedResponse{
        Success:    true,
        Data:       data,
        Pagination: pagination,
    })
}

func Error(c *gin.Context, err error) {
    if appErr, ok := err.(*apperror.AppError); ok {
        c.JSON(appErr.StatusCode, Response{
            Success: false,
            Error:   appErr,
        })
        return
    }

    // Generic internal server error
    c.JSON(500, Response{
        Success: false,
        Error: map[string]string{
            "code":    "INTERNAL_SERVER_ERROR",
            "message": "An unexpected error occurred",
        },
    })
}
```

### A3: Subscription Validation Middleware Template
```go
// internal/middleware/subscription.go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "backend/models"
    "backend/pkg/apperror"
    "backend/pkg/response"
)

func ValidateSubscriptionMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, exists := c.Get("tenant_id")
        if !exists {
            response.Error(c, apperror.NewForbiddenError(
                apperror.ErrForbidden,
                "Tenant context required",
            ))
            c.Abort()
            return
        }

        var tenant models.Tenant
        err := db.Preload("Subscription").
            First(&tenant, "id = ?", tenantID.(string)).Error

        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                response.Error(c, apperror.NewNotFoundError(
                    apperror.ErrTenantNotFound,
                    "Tenant not found",
                ))
            } else {
                response.Error(c, err)
            }
            c.Abort()
            return
        }

        // Check tenant status
        switch tenant.Status {
        case models.TenantStatusExpired:
            response.Error(c, apperror.NewForbiddenError(
                apperror.ErrForbidden,
                "Subscription expired. Please renew to continue.",
            ))
            c.Abort()
            return

        case models.TenantStatusSuspended:
            response.Error(c, apperror.NewForbiddenError(
                apperror.ErrForbidden,
                "Account suspended. Please contact support.",
            ))
            c.Abort()
            return

        case models.TenantStatusTrial:
            // Check if trial expired
            if tenant.TrialEndsAt != nil && time.Now().After(*tenant.TrialEndsAt) {
                response.Error(c, apperror.NewForbiddenError(
                    apperror.ErrForbidden,
                    "Trial period expired. Please subscribe to continue.",
                ))
                c.Abort()
                return
            }

        case models.TenantStatusPastDue:
            // Check if grace period expired
            if tenant.Subscription != nil &&
               tenant.Subscription.GracePeriodEnds != nil &&
               time.Now().After(*tenant.Subscription.GracePeriodEnds) {
                response.Error(c, apperror.NewForbiddenError(
                    apperror.ErrForbidden,
                    "Payment overdue. Please update your payment method.",
                ))
                c.Abort()
                return
            }

        case models.TenantStatusActive:
            // All good, proceed

        default:
            response.Error(c, apperror.NewForbiddenError(
                apperror.ErrForbidden,
                "Invalid tenant status",
            ))
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

**End of Analysis Report**

**Generated by:** Claude Code Sequential Analysis Engine
**Analysis Depth:** Ultrathink (25 reasoning steps)
**Total Analysis Time:** ~15 minutes
**Confidence Level:** HIGH (95%)
