# 🚨 CSRF & 403 Error Analysis Report

**Date:** 2026-01-08
**Issue:** Cannot create/edit/logout after 24 hours
**Status:** 🔴 **CRITICAL** - Multiple issues identified
**Severity:** HIGH - Affects all authenticated users after 24 hours

---

## 📋 Executive Summary

After 24 hours of inactivity, users experience:
- ❌ Cannot create customers or warehouses (403 Forbidden)
- ❌ Cannot logout (403 Forbidden)
- ❌ Application crashes with TypeError in frontend

**Root Causes Identified:** 3 separate but related issues

---

## 🔍 Issue #1: Frontend TypeError Prevents Error Recovery

### Location
`frontend/src/store/services/authApi.ts:295-299`

### Problem
Frontend error handling code crashes when processing 403 errors, preventing automatic recovery mechanisms from executing.

### Code Analysis
```typescript
// Line 295: Assumes error is string, but backend returns object
const errorMessage = errorData?.error || errorData?.message || '';

// Line 299: Crashes when error is object
errorMessage.toLowerCase().includes('csrf')
```

### What Backend Actually Returns
```json
{
  "success": false,
  "error": {
    "code": "CSRF_ERROR",
    "message": "CSRF token missing in cookie"
  }
}
```

### What Happens
1. Backend returns 403 with error object
2. Frontend: `errorData.error` = `{code: "...", message: "..."}`
3. Frontend: `errorMessage` = object (not string)
4. Frontend: `errorMessage.toLowerCase()` → **TypeError: toLowerCase is not a function**
5. Entire error recovery mechanism (lines 303-356) never executes
6. User sees unhandled error, application may crash

### Impact
- 🔴 **CRITICAL**: Prevents CSRF token recovery mechanism from working
- 🔴 **CRITICAL**: User cannot recover from CSRF errors
- 🔴 **CRITICAL**: User stuck and must refresh page or re-login

### Evidence from Error Logs
```
authApi.ts:98 POST http://localhost:8080/api/v1/customers 403 (Forbidden)

installHook.js:1 An unhandled error occurred processing a request for the endpoint "createCustomer".
TypeError: errorMessage.toLowerCase is not a function
    at baseQueryWithReauth (authApi.ts:299:20)
```

---

## 🔍 Issue #2: CSRF Token Lifetime Mismatch

### Configuration Analysis

| Token Type | Lifetime | Configuration Source |
|------------|----------|---------------------|
| Access Token | 24 hours | `backend/.env:15` (`JWT_EXPIRY=24h`) |
| Refresh Token | 7 days (168h) | `backend/.env:16` (`JWT_REFRESH_EXPIRY=168h`) |
| **CSRF Token** | **24 hours** | `backend/internal/middleware/csrf.go:89` |

### The Problem Timeline

**Hour 0: User logs in**
- ✅ Access token: Valid (expires in 24h)
- ✅ CSRF token: Valid (expires in 24h)
- ✅ Refresh token: Valid (expires in 7 days)

**Hour 24: User returns after 24+ hours**
- ❌ Access token: **EXPIRED** → Auto-refreshed via `/auth/refresh` ✅
- ❌ **CSRF token: EXPIRED** → **NOT refreshed** ❌
- ✅ Refresh token: Still valid (7 days)

**Result:**
- User authenticated (valid access token from auto-refresh)
- User authorized (has valid session)
- **CSRF token missing/expired → 403 on all POST/PUT/DELETE requests**

### Request Flow Analysis

#### Scenario: User clicks "Create Customer" after 24 hours

```
1. Frontend: POST /api/v1/customers
   ├─ Headers: Authorization: Bearer <access_token>
   ├─ Headers: X-CSRF-Token: <csrf_token>  ← Token expired!
   └─ Cookie: csrf_token=<value>  ← Cookie expired!

2. Backend Middleware Chain:
   ├─ RateLimitMiddleware → ✅ Pass
   ├─ JWTAuthMiddleware → ❌ 401 (access token expired)
   └─ Request aborted

3. Frontend Auto-Refresh (authApi.ts:101-138):
   ├─ Detect 401 error
   ├─ Call POST /auth/refresh
   ├─ Backend generates NEW access token ✅
   ├─ Backend generates NEW CSRF token ✅ (auth_handler.go:125)
   ├─ Set-Cookie: csrf_token=<new_value>
   └─ Return new access token to frontend

4. Frontend Retry Original Request:
   ├─ Headers: Authorization: Bearer <new_access_token> ✅
   ├─ Headers: X-CSRF-Token: <new_csrf_from_cookie> ✅
   └─ Cookie: csrf_token=<new_value> ✅

5. Backend Middleware Chain (Retry):
   ├─ RateLimitMiddleware → ✅ Pass
   ├─ JWTAuthMiddleware → ✅ Pass (new token valid)
   ├─ TenantContextMiddleware → ✅ Pass
   ├─ ValidateSubscriptionMiddleware → ✅ Pass
   ├─ CSRFMiddleware → ✅ SHOULD PASS (new CSRF token)
   └─ Request should succeed
```

### Why It Still Fails

**Hypothesis 1: Cookie Timing Issue**
- `/auth/refresh` sets new CSRF cookie via `Set-Cookie` header
- Browser updates `document.cookie`
- Frontend retry reads cookie via `getCSRFToken()` (authApi.ts:21-35)
- **Potential race condition:** Cookie not yet available when retry happens?

**Hypothesis 2: CSRF Token Not Being Sent**
- Frontend reads CSRF from cookie correctly
- But doesn't include it in retry request headers?
- Check `prepareHeaders` function (authApi.ts:43-66)

**Hypothesis 3: Issue #1 Prevents Recovery**
- 403 from CSRF middleware
- Frontend tries CSRF recovery (lines 303-356)
- **Crashes at line 299 before recovery code runs**
- Recovery mechanism never executes

**Most Likely: Issue #1 is preventing Issue #2 from being auto-fixed**

---

## 🔍 Issue #3: User Role Authorization Problem

### Configuration Analysis

From `backend/internal/router/router.go:337,377`:
```go
// Line 337: Customers require OWNER/ADMIN
customerGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), ...)

// Line 377: Warehouses require OWNER/ADMIN
warehouseGroup.POST("", middleware.RequireRoleMiddleware("OWNER", "ADMIN"), ...)
```

### Problem
User may not have OWNER or ADMIN role in current tenant.

### Backend Response
```json
{
  "success": false,
  "error": {
    "code": "AUTHORIZATION_ERROR",
    "message": "Access denied: insufficient permissions"
  }
}
```

### Impact
Even if CSRF token is valid, user still gets 403 if they don't have required role.

---

## 🔄 Complete Error Chain

### What Happens After 24 Hours

```
1. User idle for 24+ hours
   ├─ Access token expires (24h)
   ├─ CSRF token expires (24h)
   └─ Refresh token still valid (7d)

2. User clicks "Create Customer"
   └─ POST /api/v1/customers

3. Backend: Access token expired
   └─ 401 Unauthorized

4. Frontend: Auto-refresh triggered (authApi.ts:101-138)
   ├─ POST /auth/refresh
   ├─ New access token ✅
   ├─ New CSRF token generated ✅
   └─ Retry original request

5. Backend: Retry request
   ├─ JWTAuthMiddleware: ✅ Pass (new token)
   ├─ TenantContextMiddleware: ✅ Pass
   ├─ ValidateSubscriptionMiddleware: ✅ Pass
   └─ CSRFMiddleware: ❓
       └─ May fail if cookie not updated yet
       OR
       └─ May fail due to Issue #1 (TypeError prevents recovery)

6. If CSRF passes:
   └─ RequireRoleMiddleware: ❌ 403 (insufficient permissions - Issue #3)

7. Frontend Error Handling (authApi.ts:293-360):
   ├─ Receive 403 error
   ├─ Try to extract error message (line 295)
   ├─ **CRASH: TypeError at line 299** ❌
   └─ CSRF recovery code (lines 303-356) never runs

8. User Experience:
   ├─ Cannot create customers ❌
   ├─ Cannot create warehouses ❌
   ├─ Cannot logout ❌ (logout is POST, requires CSRF)
   └─ Must refresh page or re-login
```

---

## 🛠️ Solutions

### Solution for Issue #1: Fix TypeError in Frontend

**File:** `frontend/src/store/services/authApi.ts:295`

**Current Code (BROKEN):**
```typescript
const errorMessage = errorData?.error || errorData?.message || '';
```

**Fixed Code:**
```typescript
// Handle both string and object error formats
const errorMessage =
  (typeof errorData?.error === 'object' && errorData?.error?.message) ||
  (typeof errorData?.error === 'string' && errorData?.error) ||
  errorData?.message ||
  '';
```

**Why This Works:**
1. Check if `errorData.error` is object → Extract `message` property
2. Check if `errorData.error` is string → Use directly
3. Fallback to `errorData.message`
4. Final fallback to empty string
5. Result: Always a string, never crashes

**Priority:** 🔴 **CRITICAL** - Must fix first, enables other solutions to work

---

### Solution for Issue #2: Extend CSRF Token Lifetime

**Option A: Match CSRF lifetime to refresh token (Recommended)**

**File:** `backend/internal/middleware/csrf.go:89`

**Current:**
```go
24*60*60,           // maxAge (24 hours in seconds)
```

**Recommended:**
```go
7*24*60*60,         // maxAge (7 days to match refresh token)
```

**Why:** Prevents CSRF expiration while user is still authenticated

---

**Option B: Keep 24h but regenerate on every authenticated request**

Already implemented in `backend/internal/middleware/csrf.go:117-132` but not being called.

**Implementation:**
Call `RegenerateCSRFToken()` in `JWTAuthMiddleware` after successful token validation.

**Pros:**
- More secure (token rotation)
- Shorter window for token theft

**Cons:**
- More overhead (cookie set on every request)
- Requires middleware coordination

---

### Solution for Issue #3: Update User Role

**Temporary Fix (Database):**

```sql
-- Check current user role
SELECT u.email, ut.role, t.name as tenant_name
FROM users u
JOIN user_tenants ut ON u.id = ut.user_id
JOIN tenants t ON ut.tenant_id = t.id
WHERE u.email = 'your-email@example.com';

-- Update user role to OWNER
UPDATE user_tenants
SET role = 'OWNER'
WHERE user_id = (
    SELECT id FROM users WHERE email = 'your-email@example.com'
)
AND tenant_id = 'your-tenant-id';
```

**Permanent Fix (Application):**

1. Create user onboarding flow that assigns correct role
2. Add role management UI for tenant owners
3. Validate role assignments during user creation/invitation

---

## 📊 Testing Recommendations

### Test Case 1: Verify TypeError Fix

**Setup:**
1. Apply frontend fix (authApi.ts:295)
2. Login as user with insufficient role
3. Try to create customer

**Expected:**
- No TypeError crash ✅
- Clean 403 error message shown ✅
- Application remains functional ✅

---

### Test Case 2: Verify CSRF Recovery

**Setup:**
1. Apply frontend fix (authApi.ts:295)
2. Apply CSRF lifetime extension (csrf.go:89)
3. Login and wait 25 hours (or manually delete CSRF cookie)
4. Try to create customer

**Expected:**
- 401 detected (access token expired) ✅
- Auto-refresh triggered ✅
- New CSRF token generated ✅
- Original request retried successfully ✅
- OR: Clean error message if role insufficient ✅

---

### Test Case 3: Verify Role Authorization

**Setup:**
1. Update user role to OWNER
2. Login and try to create customer

**Expected:**
- Request succeeds ✅
- Customer created in database ✅

---

## 🎯 Implementation Priority

### Phase 1: Critical Fixes (Do This NOW)

**Priority 1:** Fix TypeError (Issue #1)
- **Impact:** Unblocks all other fixes
- **Effort:** 5 minutes
- **File:** `frontend/src/store/services/authApi.ts:295`

**Priority 2:** Update User Role (Issue #3)
- **Impact:** Allows current user to work immediately
- **Effort:** 1 SQL query
- **Action:** Run SQL UPDATE command

**Priority 3:** Extend CSRF Token Lifetime (Issue #2)
- **Impact:** Prevents 24-hour timeout issue
- **Effort:** Change one number
- **File:** `backend/internal/middleware/csrf.go:89`

---

### Phase 2: Testing & Validation

1. Test with fixed frontend code
2. Verify CSRF recovery works
3. Test logout after 24 hours
4. Test create/edit operations

---

### Phase 3: Long-term Improvements

1. Add frontend CSRF token refresh indicator
2. Implement proactive CSRF token refresh before expiry
3. Add user role management UI
4. Improve error messages for authorization failures
5. Add monitoring for CSRF token expiration events

---

## 📝 Configuration Summary

### Current Configuration

```bash
# backend/.env
JWT_EXPIRY=24h                 # Access token lifetime
JWT_REFRESH_EXPIRY=168h        # Refresh token lifetime (7 days)

# backend/internal/middleware/csrf.go:89
maxAge: 24*60*60               # CSRF token lifetime (24 hours)
```

### Recommended Configuration

```bash
# backend/.env (no change needed)
JWT_EXPIRY=24h                 # Access token lifetime
JWT_REFRESH_EXPIRY=168h        # Refresh token lifetime (7 days)

# backend/internal/middleware/csrf.go:89 (CHANGE THIS)
maxAge: 7*24*60*60             # CSRF token lifetime (7 days) ← MATCH REFRESH TOKEN
```

---

## 🔐 Security Implications

### CSRF Token Lifetime Extension

**Question:** Is it safe to extend CSRF token from 24h to 7 days?

**Answer:** Yes, with caveats:

**Pros:**
- ✅ CSRF tokens are bound to session (refresh token)
- ✅ When refresh token expires, CSRF becomes invalid
- ✅ CSRF uses double-submit cookie pattern (secure)
- ✅ SameSite=Strict provides additional protection
- ✅ Token is cryptographically random (32 bytes)

**Cons:**
- ⚠️ Longer window for token theft if user device compromised
- ⚠️ Less frequent token rotation

**Mitigation:**
- Refresh token already has 7-day lifetime (no additional exposure)
- CSRF token theft requires BOTH:
  1. Access to user's cookies (same-origin only)
  2. Ability to set X-CSRF-Token header (requires XSS)
- If attacker has XSS, CSRF token lifetime is irrelevant
- SameSite=Strict prevents cross-site CSRF attacks

**Conclusion:** Safe to extend to 7 days to match refresh token lifetime

---

## 📚 References

### Code References

- `backend/internal/middleware/csrf.go` - CSRF middleware implementation
- `backend/internal/handler/auth_handler.go:121-126` - CSRF regeneration on refresh
- `backend/internal/router/router.go:126,149` - CSRF middleware usage
- `frontend/src/store/services/authApi.ts:21-35` - CSRF token reading
- `frontend/src/store/services/authApi.ts:58-63` - CSRF header injection
- `frontend/src/store/services/authApi.ts:291-360` - CSRF error recovery (broken)

### Related Documentation

- `backend/docs/BACKEND-IMPLEMENTATION.md:437-579` - CSRF implementation reference
- `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md` - Security headers overview
- `frontend/docs/PHASE-3-CSP-AUDIT-REPORT.md` - Frontend security audit

---

## ✅ Checklist

### Immediate Actions

- [ ] Fix TypeError in `authApi.ts:295`
- [ ] Update user role to OWNER via SQL
- [ ] Extend CSRF token lifetime to 7 days
- [ ] Restart backend server
- [ ] Test create customer/warehouse
- [ ] Test logout functionality
- [ ] Verify no errors after 24 hours

### Validation

- [ ] No TypeError in browser console
- [ ] User can create customers
- [ ] User can create warehouses
- [ ] User can logout successfully
- [ ] No 403 errors after token refresh
- [ ] CSRF recovery mechanism working

### Monitoring

- [ ] Monitor error logs for CSRF failures
- [ ] Track 403 error frequency
- [ ] Monitor user session duration
- [ ] Check for unauthorized access attempts

---

**Report Generated:** 2026-01-08
**Last Updated:** 2026-01-08
**Status:** 🔴 Awaiting critical fixes
