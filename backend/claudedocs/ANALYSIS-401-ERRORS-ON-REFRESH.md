# Root Cause Analysis: 401 Unauthorized Errors on Page Refresh

**Date:** 2025-12-21
**Status:** ANALYSIS COMPLETE - Cosmetic UX Issue, Not a Security Bug
**Severity:** LOW (Cosmetic) - Application Functions Correctly
**Recommendation:** Optional UX Improvement

---

## Executive Summary

The 401 Unauthorized errors appearing in the console on page refresh are **EXPECTED BEHAVIOR** in a properly functioning JWT authentication system with auto-refresh capabilities. The application works correctly despite these errors because:

1. The access token expires after 24 hours (per `.env` configuration)
2. On page refresh with an expired token, RTK Query attempts API calls with the expired token
3. These calls fail with 401, triggering the automatic token refresh mechanism
4. The refresh succeeds using the refresh_token cookie (still valid for 7 days)
5. All API calls are automatically retried with the new token and succeed

**This is NOT a bug** - it's the authentication system working as designed. The errors are cosmetic console noise that could be reduced for better developer experience, but they indicate proper security behavior.

---

## Evidence Collection

### 1. Console Error Pattern

From the screenshot provided:
```
GET http://localhost:3002/api/v1/tenant 401 (Unauthorized)
GET http://localhost:3002/api/v1/tenant/users 401 (Unauthorized)

[Auth] Access token expired, attempting refresh...
[Auth] Initiating new token refresh (no refresh in progress)
[Auth] Token refresh successful
[Auth] Retrying original request with new token
```

### 2. Token Expiry Configuration

**Backend Configuration (`.env`):**
```env
JWT_EXPIRY=24h           # Access token expires after 24 hours
JWT_REFRESH_EXPIRY=168h  # Refresh token expires after 7 days (168 hours)
```

**Frontend Token Restoration Flow (`providers.tsx` lines 54-99):**
```typescript
// On app load, restore token from localStorage
const accessToken = localStorage.getItem("accessToken");

if (accessToken) {
  const decoded = jwtDecode<JWTPayload>(accessToken);
  const now = Date.now() / 1000;

  if (decoded.exp > now) {
    // Token still valid - restore session immediately
  } else {
    // Token expired - clear localStorage
    localStorage.removeItem("accessToken");
  }
}
```

### 3. API Query Auto-Execution Pattern

**Team Page (`company/team/page.tsx` lines 47-63):**
```typescript
// These queries execute IMMEDIATELY on component mount
const { data: tenant, isLoading, error } = useGetTenantQuery();
const { data: users } = useGetUsersQuery({ role: roleFilter });
```

**RTK Query Default Behavior:**
- `refetchOnMountOrArgChange: true` (default)
- `refetchOnReconnect: true` (default)
- `refetchOnFocus: true` (default)

These queries execute BEFORE the token refresh logic has a chance to run.

---

## Root Cause Analysis

### The Race Condition Explained

```
Page Refresh Timeline:
┌─────────────────────────────────────────────────────────────────┐
│ T=0ms    │ Browser loads page, React components mount             │
│ T=10ms   │ AuthInitializer reads localStorage                     │
│          │ → Finds expired access token (24h old)                 │
│          │ → Clears it from localStorage                          │
│          │                                                         │
│ T=15ms   │ Team Page component mounts                             │
│          │ → useGetTenantQuery() executes                         │
│          │ → useGetUsersQuery() executes                          │
│          │ → BOTH use empty/expired access token                  │
│          │ → BOTH fail with 401 Unauthorized ❌                   │
│          │                                                         │
│ T=20ms   │ baseQueryWithReauth intercepts 401 errors              │
│          │ → Initiates token refresh (POST /auth/refresh)         │
│          │ → Uses refresh_token from httpOnly cookie ✅           │
│          │                                                         │
│ T=120ms  │ Backend validates refresh_token                        │
│          │ → Generates new access token                           │
│          │ → Returns new token to frontend                        │
│          │                                                         │
│ T=130ms  │ Frontend updates Redux state with new token            │
│          │ → Retries GET /tenant → Success ✅                     │
│          │ → Retries GET /tenant/users → Success ✅               │
│          │                                                         │
│ T=150ms  │ Page renders correctly with data ✅                    │
└─────────────────────────────────────────────────────────────────┘
```

### Why This Happens

**Root Cause:** RTK Query queries execute on component mount BEFORE the AuthInitializer can detect token expiry and proactively refresh it.

**Key Files Involved:**

1. **`providers.tsx` (lines 49-99):** Token restoration logic runs but only checks expiry - doesn't proactively refresh
2. **`authApi.ts` (lines 86-146):** Token refresh logic is REACTIVE (waits for 401) not PROACTIVE
3. **`team/page.tsx` (lines 47-63):** Queries execute immediately without checking token validity first
4. **Backend middleware (`auth.go`):** Correctly rejects expired tokens with 401

### Why Application Still Works

**Multi-Layer Recovery System:**

1. **Layer 1: Token Refresh Mutex** (`authApi.ts` lines 67-122)
   - Prevents concurrent refresh attempts from multiple 401 errors
   - Single refresh handles all failed requests
   - Race condition protection works correctly ✅

2. **Layer 2: Automatic Retry** (`authApi.ts` lines 124-143)
   - After successful refresh, retries original requests
   - Uses new access token from Redux state
   - Requests succeed on retry ✅

3. **Layer 3: httpOnly Cookie Persistence**
   - refresh_token stored in secure httpOnly cookie
   - Survives page refresh, not accessible to JavaScript
   - Backend validates and issues new access token ✅

---

## Flow Diagram: Current vs Ideal Behavior

### Current Behavior (WITH 401 Errors)

```
┌─────────────┐
│ Page Load   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│ Read localStorage       │
│ - Has expired token     │
│ - Clear from storage    │
└──────┬──────────────────┘
       │
       ├────────────────────────────────┐
       │                                │
       ▼                                ▼
┌──────────────┐              ┌──────────────────┐
│ GET /tenant  │              │ GET /tenant/users │
│ (no token)   │              │ (no token)        │
└──────┬───────┘              └────────┬──────────┘
       │                               │
       ▼                               ▼
    ❌ 401                          ❌ 401
       │                               │
       └───────────┬───────────────────┘
                   │
                   ▼
         ┌────────────────────┐
         │ Token Refresh      │
         │ POST /auth/refresh │
         └────────┬───────────┘
                  │
                  ▼
               ✅ Success
                  │
         ┌────────┴────────┐
         │                 │
         ▼                 ▼
    Retry /tenant    Retry /tenant/users
         │                 │
         ▼                 ▼
      ✅ 200            ✅ 200
```

### Ideal Behavior (NO 401 Errors)

```
┌─────────────┐
│ Page Load   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│ Read localStorage       │
│ - Has expired token     │
│ - Proactively refresh   │ ⬅️ DIFFERENCE HERE
└──────┬──────────────────┘
       │
       ▼
┌────────────────────┐
│ POST /auth/refresh │
└────────┬───────────┘
         │
         ▼
      ✅ New Token
         │
         ▼
┌──────────────────────────┐
│ Wait for refresh before  │
│ executing queries        │ ⬅️ DIFFERENCE HERE
└──────┬───────────────────┘
       │
       ├────────────────────────────────┐
       │                                │
       ▼                                ▼
┌──────────────┐              ┌──────────────────┐
│ GET /tenant  │              │ GET /tenant/users │
│ (new token)  │              │ (new token)       │
└──────┬───────┘              └────────┬──────────┘
       │                               │
       ▼                               ▼
    ✅ 200                          ✅ 200
   (no retry needed)            (no retry needed)
```

---

## Investigation Evidence

### 1. Token Expiry Detection Logic

**File:** `frontend/src/components/providers.tsx`

**Issue:** Detection without proactive refresh
```typescript
// Lines 54-99
if (accessToken) {
  const decoded = jwtDecode<JWTPayload>(accessToken);
  const now = Date.now() / 1000;

  if (decoded.exp > now) {
    // ✅ Token valid - restore session
    dispatch(setCredentials({ ... }));
    setTokenRestored(true);
  } else {
    // ❌ Token expired - just clear it, don't refresh
    localStorage.removeItem("accessToken");
    // MISSING: Proactive refresh before queries execute
  }
}
```

**Gap Identified:** When token is expired, the code:
- ✅ Correctly detects expiry
- ✅ Clears expired token
- ❌ Does NOT proactively refresh before queries execute
- ❌ Lets queries fail and trigger reactive refresh

### 2. Query Execution Without Token Check

**File:** `frontend/src/app/(app)/company/team/page.tsx`

**Issue:** Queries execute immediately on mount
```typescript
// Lines 47-63
// NO token validity check before executing queries
const { data: tenant } = useGetTenantQuery();
const { data: users } = useGetUsersQuery({ role: roleFilter });

// These execute IMMEDIATELY when component mounts
// If token is expired/missing → 401 errors
```

**Gap Identified:** Queries don't wait for:
- Token restoration to complete
- Token refresh to complete (if expired)
- Authentication state to be ready

### 3. RTK Query Configuration

**File:** `frontend/src/store/index.ts`

**Current Configuration:**
```typescript
export const store = configureStore({
  reducer: {
    auth: authReducer,
    [authApi.reducerPath]: authApi.reducer,
    [companyApi.reducerPath]: companyApi.reducer,
    [tenantApi.reducerPath]: tenantApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(
      authApi.middleware,
      companyApi.middleware,
      tenantApi.middleware
    ),
});
```

**Missing:** No configuration to delay queries until auth ready

### 4. Backend Token Validation

**File:** `backend/internal/middleware/auth.go`

**Correct Behavior:**
```go
// Validates JWT token
// If expired → returns 401 (CORRECT)
// If valid → extracts claims and continues
c.Set("user_id", claims.UserID)
c.Set("tenant_id", claims.TenantID)
```

**Evidence:** Backend correctly rejects expired tokens with 401

---

## Why This Is NOT a Security Bug

### Security Mechanisms Working Correctly

1. **Access Token Expiry Enforced** ✅
   - 24-hour expiry properly validated
   - Expired tokens correctly rejected with 401
   - No unauthorized access with expired tokens

2. **Refresh Token Security** ✅
   - Stored in httpOnly cookie (not accessible to JavaScript)
   - 7-day expiry for long-term sessions
   - Properly validated on refresh endpoint

3. **CSRF Protection** ✅
   - CSRF token in cookie + header (double-submit pattern)
   - Prevents cross-site token theft

4. **Race Condition Prevention** ✅
   - Mutex prevents concurrent refresh attempts
   - Single refresh handles multiple 401s

5. **Automatic Logout on Refresh Failure** ✅
   - If refresh_token invalid/expired → logout user
   - No infinite retry loops

### What Attackers Cannot Exploit

- **Cannot bypass expiry:** Backend validates token expiry server-side
- **Cannot steal refresh token:** httpOnly cookie protects it
- **Cannot CSRF attack:** Double-submit cookie pattern prevents it
- **Cannot race condition attack:** Mutex prevents concurrent exploits

---

## Impact Assessment

### Functional Impact: NONE

- ✅ All API calls eventually succeed after token refresh
- ✅ User session restored correctly after page refresh
- ✅ Data loads correctly in all components
- ✅ No authentication bypass or security vulnerabilities

### User Experience Impact: MINIMAL

- ❌ Console shows error logs (only visible to developers)
- ❌ Slightly slower initial page load (~100-200ms for refresh)
- ✅ No visible errors to end users
- ✅ Application appears to work normally

### Developer Experience Impact: MODERATE

- ❌ Console clutter makes real errors harder to spot
- ❌ May cause confusion during debugging
- ❌ False alarms in error monitoring systems
- ✅ Clear log messages explain what's happening

---

## Recommendations

### Option 1: Accept Current Behavior (Recommended)

**Rationale:**
- Application functions correctly
- Security mechanisms working as designed
- Low priority cosmetic issue
- No impact on end users

**Action:** Document behavior, no code changes

### Option 2: Implement Proactive Token Refresh (Optional UX Improvement)

**Implementation Complexity:** Medium
**Risk:** Low
**Benefit:** Cleaner console logs, slightly faster page loads

**Changes Required:**

1. **`providers.tsx` - Add proactive refresh:**
   ```typescript
   // Detect expired token and refresh BEFORE queries execute
   if (decoded.exp < now) {
     await refreshAccessToken(); // Call /auth/refresh
     // Then restore session with new token
   }
   ```

2. **Add authentication state to Redux:**
   ```typescript
   // Track: 'initializing' | 'ready' | 'unauthenticated'
   authState: 'initializing'
   ```

3. **Delay query execution until auth ready:**
   ```typescript
   const authReady = useSelector(state => state.auth.authState === 'ready');
   const { data } = useGetTenantQuery(undefined, {
     skip: !authReady  // Wait for auth before executing
   });
   ```

**Estimated Effort:** 4-6 hours
**Testing Required:** Auth flow, token expiry, page refresh scenarios

### Option 3: Suppress Console Errors (Not Recommended)

**Why Not:**
- Hides legitimate errors
- Makes debugging harder
- Doesn't fix root cause
- Poor developer practice

---

## Conclusion

The 401 errors on page refresh are **cosmetic console noise from a properly functioning JWT authentication system**. They indicate:

1. ✅ Access tokens correctly expire after 24 hours
2. ✅ Expired tokens are properly rejected by backend
3. ✅ Automatic token refresh mechanism works correctly
4. ✅ Refresh tokens securely stored and validated
5. ✅ Failed requests automatically retried with new token

**This is EXPECTED BEHAVIOR in a secure JWT implementation.** The application works correctly; the errors are just visible evidence of the auto-refresh mechanism activating.

**Recommended Action:** Accept current behavior and document it. Implementing proactive refresh is a low-priority UX improvement that can be addressed later if console cleanliness becomes important.

---

## Related Files

**Frontend:**
- `/frontend/src/components/providers.tsx` - Token restoration logic
- `/frontend/src/store/services/authApi.ts` - Auto-refresh implementation
- `/frontend/src/app/(app)/company/team/page.tsx` - Example query usage
- `/frontend/src/store/slices/authSlice.ts` - Auth state management

**Backend:**
- `/backend/internal/middleware/auth.go` - JWT validation
- `/backend/internal/service/auth/auth_service.go` - Token generation/refresh
- `/backend/.env` - Token expiry configuration

**Configuration:**
- `JWT_EXPIRY=24h` - Access token lifetime
- `JWT_REFRESH_EXPIRY=168h` - Refresh token lifetime

---

**Analysis completed by:** Claude Code (Root Cause Analyst)
**Analysis method:** Systematic evidence collection, flow analysis, code review
**Confidence level:** 95% (High confidence based on comprehensive code review)
