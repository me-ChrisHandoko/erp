# Phase 4: Frontend Authentication Integration - MVP Analysis

**Analysis Date:** 2025-12-17
**Target:** Frontend Authentication Implementation
**Approach:** MVP (Minimum Viable Product)
**Framework:** Next.js 16 + Redux Toolkit + RTK Query
**Backend Status:** ✅ Authentication API Ready

---

## 🎯 Executive Summary

Phase 4 frontend authentication implementation is ready to begin. This analysis provides a comprehensive MVP roadmap that adapts the original React Router-based implementation guide to Next.js 16 App Router patterns, incorporating critical findings about backend API contracts (httpOnly cookies for refresh tokens, response envelope format, multi-tenant endpoints).

**Recommended Timeline:** 7 working days (2 weeks part-time)
**Risk Level:** LOW-MEDIUM
**Estimated Effort:** 11-16 hours
**Dependencies:** Backend API running on localhost:8080

---

## 📊 Current State Assessment

### ✅ Completed (Base Layout - Phase 3)

1. **Route Groups Architecture**
   - `(auth)` route group for public pages
   - `(app)` route group for authenticated pages
   - Clean separation of concerns

2. **UI Components**
   - Login page with Card UI (static, no functionality)
   - Dashboard layout with sidebar
   - Team switcher component (needs Redux integration)
   - Navigation structure (25 menu items)

3. **Dependencies Installed**
   - @reduxjs/toolkit v2.11.2 ✅
   - react-redux v9.2.0 ✅
   - jsonwebtoken v9.0.3 (not needed, will use jwt-decode)

### ❌ Not Started (Phase 4 Scope)

1. **Redux Store**
   - No src/store/ directory
   - No auth slice for state management
   - No RTK Query API configuration
   - No Redux Provider in root layout

2. **Authentication Logic**
   - Login page is static HTML only
   - No form submission handler
   - No error handling
   - No loading states
   - No redirect on success

3. **Route Protection**
   - No middleware for authentication checking
   - No redirect logic for unauthorized access
   - Protected routes accessible without login

4. **Token Management**
   - No token storage (localStorage/Redux)
   - No auto-refresh logic
   - No token expiry handling

---

## 🔍 Critical Findings

### Finding 1: Framework Mismatch (React Router vs Next.js)

**Issue:** FRONTEND-IMPLEMENTATION.md was written for React SPA with React Router v6, but this project uses Next.js 16 App Router.

**Impact:** Major patterns need adaptation:

| React Router Pattern | Next.js Pattern |
|---------------------|-----------------|
| `<BrowserRouter>` | Not needed (file-based routing) |
| `<Routes>, <Route>` | Not needed (pages in app/) |
| `useNavigate()` | `useRouter()` from 'next/navigation' |
| `<Navigate to="/path">` | `router.push('/path')` |
| `<Link to="/path">` | `<Link href="/path">` |
| `<ProtectedRoute>` wrapper | `middleware.ts` or layout checks |

**Resolution:** Adapt implementation guide to Next.js patterns (detailed in implementation plan).

---

### Finding 2: Cookie-Based Refresh Tokens (Critical!)

**Issue:** Backend uses httpOnly cookies for refresh tokens, NOT response body as shown in implementation guide.

**Verified from Backend API Documentation:**
```json
// Backend Login Response
{
  "status": "success",
  "data": {
    "user": { ... },
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    // ❌ NO refreshToken in response body!
  }
}

// Cookies Set by Backend
Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Max-Age=604800
Set-Cookie: csrf_token=...; Secure; SameSite=Strict; Max-Age=86400
```

**Impact:** Major changes to implementation:

1. **Auth Slice - Remove refreshToken:**
   ```typescript
   // ❌ WRONG (from implementation guide)
   interface AuthState {
     refreshToken: string | null; // CAN'T ACCESS httpOnly cookie!
   }

   // ✅ CORRECT (MVP adaptation)
   interface AuthState {
     accessToken: string | null; // Only this
   }
   ```

2. **RTK Query - Enable Credentials:**
   ```typescript
   const baseQuery = fetchBaseQuery({
     baseUrl: '/api/v1',
     credentials: 'include', // ✅ CRITICAL: Send cookies!
   });
   ```

3. **Auto-Refresh - Simplified:**
   ```typescript
   // ❌ WRONG (from implementation guide)
   const refreshResult = await baseQuery({
     url: '/auth/refresh',
     method: 'POST',
     body: { refreshToken } // Not accessible!
   });

   // ✅ CORRECT (cookie sent automatically)
   const refreshResult = await baseQuery({
     url: '/auth/refresh',
     method: 'POST'
     // No body needed, cookie sent via credentials: 'include'
   });
   ```

**Benefits:**
- ✅ More secure (XSS can't steal refresh token)
- ✅ Simpler code (no refresh token sync)
- ✅ Less localStorage management

---

### Finding 3: Response Format Envelope

**Issue:** Backend uses envelope format `{ status, data }`, not direct response.

**Backend Response Format:**
```typescript
// Success Response
{
  "status": "success",
  "data": {
    "user": { ... },
    "accessToken": "..."
  }
}

// Error Response
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email already registered"
  }
}
```

**Impact:** RTK Query response handling needs unwrapping:

```typescript
// Type definitions needed
interface ApiSuccessResponse<T> {
  status: "success";
  data: T;
}

interface ApiErrorResponse {
  status: "error";
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

// In mutation handler
async onQueryStarted(arg, { dispatch, queryFulfilled }) {
  try {
    const response = await queryFulfilled;
    const { data } = response.data; // Unwrap envelope
    dispatch(setCredentials(data));
  } catch (err) {
    // Error already unwrapped by RTK Query
  }
}
```

---

### Finding 4: Multi-Tenant Endpoints Confirmed

**Issue:** Implementation guide shows tenant switching, but needed to verify backend support.

**Verified Backend Endpoints:**
- ✅ `POST /api/v1/auth/switch-tenant` - Switch active tenant
- ✅ `GET /api/v1/auth/tenants` - Get user's available tenants
- ✅ JWT contains `tenantID` and `role` claims
- ✅ Middleware validates tenant access

**Impact:** Multi-tenant switching is fully supported. Can proceed with implementation as planned.

---

### Finding 5: Missing Dependency

**Issue:** `jwt-decode` not installed, but `jsonwebtoken` is.

**Analysis:**
- `jsonwebtoken` → Server-side Node.js library (not for browser)
- `jwt-decode` → Client-side browser library (lightweight, decode only)

**Resolution:** Install jwt-decode:
```bash
npm install jwt-decode
```

---

## 🎯 MVP Scope Definition

### MUST HAVE (Week 1-2 Priority)

**Critical Path Features:**

1. **Redux Store Setup** ⏱️ 2-3 hours
   - Configure Redux store with RTK
   - Create auth slice (simplified: no refreshToken)
   - Wrap root layout with Provider
   - Setup Redux DevTools

2. **RTK Query API Configuration** ⏱️ 3-4 hours
   - Create authApi service
   - Configure baseQuery with credentials
   - Implement login mutation
   - Implement logout mutation
   - Implement refresh interceptor (401 → refresh)
   - Add basic error handling

3. **Functional Login Page** ⏱️ 3-4 hours
   - Convert to client component
   - Add form state (email, password)
   - Integrate useLoginMutation hook
   - Display loading state during submission
   - Show error messages from API
   - Redirect to /dashboard on success

4. **Route Protection** ⏱️ 2-3 hours
   - Create src/middleware.ts
   - Check authentication state
   - Redirect unauthorized users to /login
   - Protect all `/dashboard/*` routes

5. **Token Persistence** ⏱️ 1-2 hours
   - Save accessToken to localStorage on login
   - Load accessToken from localStorage on app init
   - Clear localStorage on logout

6. **Basic Logout** ⏱️ 1-2 hours
   - Add logout mutation to RTK Query
   - Clear Redux state on logout
   - Clear localStorage
   - Redirect to /login

**Total Estimated Effort:** 12-18 hours → **2 working days**

---

### SHOULD HAVE (Nice to Have)

**Enhanced Features (if time permits):**

7. **Multi-Tenant Switching** ⏱️ 2-3 hours
   - Add switchTenant mutation
   - Add getTenants query
   - Integrate team-switcher.tsx with Redux
   - Display active tenant in sidebar

8. **Auto Token Refresh** ⏱️ 2-3 hours
   - Enhanced 401 interceptor
   - Request queuing during refresh
   - Prevent multiple simultaneous refreshes

**Total Additional Effort:** 4-6 hours → **1 working day**

---

### DEFERRED (Post-MVP)

**Non-Critical Features:**

- ❌ Password reset flow (separate feature)
- ❌ Email verification flow (not in MVP scope)
- ❌ Remember me functionality (nice-to-have)
- ❌ E2E automated testing (important but not blocking)
- ❌ Advanced error handling (localized messages, retry logic)
- ❌ Background token refresh timer (interceptor is sufficient)
- ❌ Role-based access control (just check isAuthenticated for MVP)

---

## 🏗️ Implementation Architecture

### Directory Structure

```
src/
├── app/
│   ├── layout.tsx                    # ✅ Root layout (add Provider)
│   ├── (auth)/
│   │   ├── layout.tsx               # ✅ Auth layout (already done)
│   │   └── login/
│   │       └── page.tsx             # 🔧 Make functional
│   └── (app)/
│       ├── layout.tsx               # ✅ App layout (already done)
│       └── dashboard/
│           └── page.tsx             # ✅ Protected page
├── store/                           # 🆕 CREATE
│   ├── index.ts                     # Store configuration
│   ├── slices/
│   │   └── authSlice.ts             # Auth state management
│   └── services/
│       └── authApi.ts               # RTK Query API
├── types/                           # 🆕 CREATE
│   └── api.ts                       # API types (User, Tenant, etc.)
├── utils/                           # 🆕 CREATE (or add to lib/)
│   └── localStorage.ts              # Token persistence
├── middleware.ts                    # 🆕 CREATE (Next.js middleware)
└── components/
    ├── team-switcher.tsx            # 🔧 Integrate with Redux
    └── ui/                          # ✅ Already exists
```

---

### Technology Stack Summary

| Layer | Technology | Purpose |
|-------|------------|---------|
| Framework | Next.js 16 | App Router, SSR, file-based routing |
| State Management | Redux Toolkit | Global auth state |
| API Client | RTK Query | Data fetching, caching, auto-refresh |
| Token Decode | jwt-decode | Client-side JWT parsing |
| Styling | Tailwind CSS | UI styling (already configured) |
| UI Components | shadcn/ui | Pre-built components (already installed) |

---

## 📋 Implementation Dependency Chain

**CRITICAL PATH (Must follow order):**

### Layer 1: Foundation (No Dependencies)
- ✅ Install jwt-decode
- ✅ Create TypeScript interfaces
- ✅ Create directory structure

### Layer 2: State Management (Depends on Layer 1)
- ✅ Create authSlice.ts
- ✅ Create store/index.ts
- ✅ Wrap root layout with Provider

### Layer 3: API Integration (Depends on Layer 2)
- ✅ Create authApi.ts with RTK Query
- ✅ Configure baseQuery with credentials
- ✅ Define login/logout/refresh endpoints

### Layer 4: UI Integration (Depends on Layer 3)
- ✅ Update login page with hooks
- ✅ Add error handling
- ✅ Add loading states
- ✅ Add redirect logic

### Layer 5: Protection (Depends on Layer 4)
- ✅ Create middleware.ts
- ✅ Check auth state
- ✅ Redirect unauthorized users

### Layer 6: Multi-Tenant (Depends on Layer 5)
- ✅ Integrate team-switcher
- ✅ Add switchTenant mutation
- ✅ Display active tenant

**⚠️ WARNING:** Cannot skip layers! Each layer builds on previous.

---

## 🚨 Potential Blockers & Mitigation

### Blocker 1: CORS Configuration

**Problem:** Frontend (localhost:3000) → Backend (localhost:8080) = CORS error

**Symptoms:**
```
Access to fetch at 'http://localhost:8080/api/v1/auth/login' from origin 'http://localhost:3000'
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present
```

**Mitigation:**

Backend must enable CORS with:
```go
// Backend CORS middleware
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
    AllowCredentials: true, // ✅ CRITICAL for cookies
}))
```

**Verification:**
- Check browser Network tab for CORS error
- Verify `Access-Control-Allow-Origin` header in response
- Verify `Access-Control-Allow-Credentials: true`

**Fallback:** Use Next.js API routes as proxy if CORS cannot be configured

---

### Blocker 2: Cookie SameSite Issues

**Problem:** Cookies not sent in cross-origin requests (localhost:3000 ↔ localhost:8080)

**Symptoms:**
- Login succeeds but refresh token cookie not sent in subsequent requests
- Auto-refresh fails with 401

**Root Cause:**
- `SameSite=Strict` blocks cross-origin cookie sending
- Localhost:3000 and localhost:8080 are different origins

**Mitigation:**

**Option A (Development Only):**
Backend sets cookies with:
```go
// Development only - relaxed SameSite
cookie := &http.Cookie{
    Name:     "refresh_token",
    Value:    token,
    HttpOnly: true,
    Secure:   false, // Allow HTTP in dev
    SameSite: http.SameSiteNoneMode, // Allow cross-origin
}
```

**Option B (Recommended):**
Run both on same origin using Next.js API routes as proxy:
```typescript
// pages/api/[...path].ts - Proxy all /api/* to backend
export default async function handler(req, res) {
  const response = await fetch(`http://localhost:8080${req.url}`, {
    method: req.method,
    headers: req.headers,
    body: req.body,
  });
  // Forward cookies
  response.headers.forEach((value, key) => {
    res.setHeader(key, value);
  });
  res.status(response.status).json(await response.json());
}
```

**Verification:**
- Check browser DevTools → Application → Cookies
- Verify refresh_token cookie exists
- Verify cookie is sent in Request Headers

---

### Blocker 3: Backend Not Running

**Problem:** Cannot develop/test without running backend server

**Mitigation:**

**Option A (Recommended for MVP):**
Start backend server locally:
```bash
cd ../backend
go run cmd/server/main.go
```

**Option B (If backend unavailable):**
Use Mock Service Worker (MSW) for development:
```bash
npm install -D msw
```

```typescript
// src/mocks/handlers.ts
import { rest } from 'msw';

export const handlers = [
  rest.post('/api/v1/auth/login', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        status: 'success',
        data: {
          user: { id: '1', email: 'test@example.com', fullName: 'Test User' },
          accessToken: 'mock-access-token',
        },
      }),
      ctx.cookie('refresh_token', 'mock-refresh-token', { httpOnly: true })
    );
  }),
];
```

**Recommendation:** Use Option A (real backend) for MVP to avoid mock data inconsistencies.

---

### Blocker 4: TypeScript Type Mismatches

**Problem:** Backend response format doesn't match frontend types

**Mitigation:**

1. **Console log actual API responses first:**
```typescript
const response = await queryFulfilled;
console.log('Actual API response:', response);
```

2. **Start with flexible types:**
```typescript
// Start flexible
interface LoginResponse {
  status: string;
  data?: any; // Temporary
  error?: any;
}

// Then narrow down after verification
interface LoginResponse {
  status: 'success';
  data: {
    user: User;
    accessToken: string;
  };
}
```

3. **Generate types from OpenAPI if available:**
```bash
# If backend has swagger.json
npx openapi-typescript http://localhost:8080/swagger.json -o src/types/api.ts
```

---

### Blocker 5: Environment Variables

**Problem:** No configuration for backend API URL

**Mitigation:**

Create `.env.local`:
```bash
# .env.local (gitignored)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Use in RTK Query:
```typescript
const baseQuery = fetchBaseQuery({
  baseUrl: process.env.NEXT_PUBLIC_API_URL + '/api/v1',
  credentials: 'include',
});
```

**Verification:**
```typescript
console.log('API URL:', process.env.NEXT_PUBLIC_API_URL);
```

**Note:** `NEXT_PUBLIC_` prefix required for client-side access

---

## 📅 7-Day Implementation Plan

### Pre-Implementation (Setup)

**Day 0: Environment Setup** ⏱️ 1-2 hours
- [ ] Install jwt-decode: `npm install jwt-decode`
- [ ] Create .env.local with backend URL
- [ ] Verify backend is running on localhost:8080
- [ ] Test manual API call to /auth/login via Postman/curl
- [ ] Verify CORS is configured correctly
- [ ] Create directory structure (store/, types/, utils/)

---

### Week 1: Foundation & Core Auth

**Day 1: Redux Setup** ⏱️ 2-3 hours

**Tasks:**
- [ ] Create `src/store/index.ts`
  - Configure Redux store with configureStore
  - Export RootState and AppDispatch types
  - Add Redux DevTools extension

- [ ] Create `src/store/slices/authSlice.ts`
  - Define AuthState interface (NO refreshToken)
  - Create authSlice with initialState
  - Add reducers: setCredentials, setAccessToken, logout, setError
  - Export actions and reducer

- [ ] Update `src/app/layout.tsx`
  - Import Provider from react-redux
  - Wrap children with `<Provider store={store}>`
  - Mark as client component if needed

- [ ] Verification:
  - [ ] Redux DevTools shows auth state
  - [ ] Dispatch actions manually from console
  - [ ] Verify state updates correctly

**Deliverables:**
- ✅ Working Redux store
- ✅ Auth slice with state management
- ✅ Provider wrapping app

---

**Day 2: RTK Query API** ⏱️ 3-4 hours

**Tasks:**
- [ ] Create `src/types/api.ts`
  - Define User interface
  - Define TenantContext interface
  - Define LoginRequest/LoginResponse
  - Define ApiSuccessResponse/ApiErrorResponse

- [ ] Create `src/store/services/authApi.ts`
  - Create baseQuery with fetchBaseQuery
  - Set baseUrl from environment variable
  - Set credentials: 'include' (for cookies)
  - Add prepareHeaders for Authorization header

  - Add baseQueryWithReauth interceptor:
    - Catch 401 errors
    - Call refresh endpoint
    - Update access token
    - Retry original request
    - Logout on refresh failure

  - Create authApi with createApi
  - Add login mutation
  - Add logout mutation
  - Add refresh mutation (for manual refresh)
  - Export hooks

- [ ] Update `src/store/index.ts`
  - Add authApi.reducer to store
  - Add authApi.middleware to middleware chain

- [ ] Verification:
  - [ ] Make test login API call from console
  - [ ] Verify cookies are set
  - [ ] Verify Redux state updates
  - [ ] Test 401 → refresh flow manually

**Deliverables:**
- ✅ RTK Query API configured
- ✅ Login/logout mutations working
- ✅ Auto-refresh on 401

---

**Day 3: Token Persistence** ⏱️ 1-2 hours

**Tasks:**
- [ ] Create `src/utils/localStorage.ts`
  ```typescript
  export const saveToken = (token: string) => {
    localStorage.setItem('accessToken', token);
  };

  export const getToken = (): string | null => {
    return localStorage.getItem('accessToken');
  };

  export const removeToken = () => {
    localStorage.removeItem('accessToken');
  };
  ```

- [ ] Update `src/store/slices/authSlice.ts`
  - Import localStorage utils
  - Update setCredentials to call saveToken
  - Update logout to call removeToken

- [ ] Create `src/app/providers.tsx` (client component)
  - Load token from localStorage on mount
  - Dispatch setCredentials if token exists
  - Verify token is not expired (decode with jwt-decode)

- [ ] Verification:
  - [ ] Login → Refresh page → Still authenticated
  - [ ] Check localStorage in DevTools
  - [ ] Logout → localStorage cleared

**Deliverables:**
- ✅ Token persistence working
- ✅ Authentication survives page refresh
- ✅ Logout clears storage

---

### Week 2: UI Integration & Protection

**Day 4: Functional Login Page** ⏱️ 3-4 hours

**Tasks:**
- [ ] Update `src/app/(auth)/login/page.tsx`
  - Add "use client" directive
  - Import useLoginMutation hook
  - Add useState for email/password
  - Add form submission handler
  - Add loading state display
  - Add error message display
  - Add redirect to /dashboard on success (using useRouter)

```typescript
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useLoginMutation } from "@/store/services/authApi";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [login, { isLoading, error }] = useLoginMutation();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login({ email, password }).unwrap();
      router.push("/dashboard");
    } catch (err) {
      // Error handled by RTK Query
    }
  };

  return (
    <Card>
      <form onSubmit={handleSubmit}>
        <CardHeader>
          <CardTitle>Login</CardTitle>
          <CardDescription>
            Masukkan email dan password untuk mengakses sistem ERP
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="text-sm text-red-600">
              {error?.data?.error?.message || "Login failed"}
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <Button className="w-full" size="lg" type="submit" disabled={isLoading}>
            {isLoading ? "Logging in..." : "Login"}
          </Button>
          <p className="text-center text-sm text-muted-foreground">
            Belum punya akun?{" "}
            <a href="#" className="underline hover:text-primary">
              Hubungi administrator
            </a>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
```

- [ ] Verification:
  - [ ] Enter credentials → Submit
  - [ ] See loading state during request
  - [ ] See error message on failure
  - [ ] Redirect to dashboard on success

**Deliverables:**
- ✅ Functional login form
- ✅ Error handling
- ✅ Loading states
- ✅ Redirect on success

---

**Day 5: Route Protection** ⏱️ 2-3 hours

**Tasks:**
- [ ] Create `src/middleware.ts` (Next.js middleware)

```typescript
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // Get token from localStorage is not possible in middleware
  // Use cookie or check auth state differently

  // For MVP: Check if refresh_token cookie exists
  const refreshToken = request.cookies.get("refresh_token");

  const isAuthPage = request.nextUrl.pathname.startsWith("/login");
  const isProtectedPage = request.nextUrl.pathname.startsWith("/dashboard");

  if (isProtectedPage && !refreshToken) {
    // Redirect to login if accessing protected page without auth
    return NextResponse.redirect(new URL("/login", request.url));
  }

  if (isAuthPage && refreshToken) {
    // Redirect to dashboard if accessing login while authenticated
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/login"],
};
```

**Alternative Approach (Layout-level):**
If middleware doesn't work well, use layout-level checks:

```typescript
// src/app/(app)/layout.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import type { RootState } from "@/store";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useSelector((state: RootState) => state.auth);
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return null; // or loading spinner
  }

  return <>{children}</>;
}
```

- [ ] Verification:
  - [ ] Access /dashboard without login → Redirect to /login
  - [ ] Login → Access /dashboard → Success
  - [ ] Logout → Try /dashboard → Redirect to /login

**Deliverables:**
- ✅ Protected routes working
- ✅ Redirect logic functional
- ✅ Unauthorized access prevented

---

**Day 6: Logout & Polish** ⏱️ 2-3 hours

**Tasks:**
- [ ] Add logout functionality to nav-user.tsx
  ```typescript
  "use client";

  import { useLogoutMutation } from "@/store/services/authApi";
  import { useRouter } from "next/navigation";

  export function NavUser() {
    const [logout] = useLogoutMutation();
    const router = useRouter();

    const handleLogout = async () => {
      try {
        await logout().unwrap();
        router.push("/login");
      } catch (err) {
        // Force logout even on API error
        router.push("/login");
      }
    };

    return (
      <DropdownMenuItem onClick={handleLogout}>
        Logout
      </DropdownMenuItem>
    );
  }
  ```

- [ ] Test full authentication flow:
  1. Start at /dashboard (not logged in) → Redirect to /login
  2. Login with valid credentials → Redirect to /dashboard
  3. Refresh page → Still authenticated
  4. Click logout → Redirect to /login
  5. Try accessing /dashboard → Redirect to /login

- [ ] Polish UI:
  - [ ] Add better error messages
  - [ ] Add success toast on login (optional)
  - [ ] Improve loading states
  - [ ] Add form validation

**Deliverables:**
- ✅ Logout working
- ✅ Full auth flow tested
- ✅ UI polished

---

**Day 7: Multi-Tenant (Optional)** ⏱️ 2-3 hours

**Tasks:**
- [ ] Add tenant endpoints to authApi:
  ```typescript
  getTenants: builder.query<GetTenantsResponse, void>({
    query: () => '/auth/tenants',
  }),

  switchTenant: builder.mutation<SwitchTenantResponse, string>({
    query: (tenantId) => ({
      url: '/auth/switch-tenant',
      method: 'POST',
      body: { tenantId },
    }),
    async onQueryStarted(arg, { dispatch, queryFulfilled }) {
      try {
        const { data } = await queryFulfilled;
        dispatch(setAccessToken(data.data.accessToken));
        dispatch(setActiveTenant(data.data.activeTenant));
      } catch (err) {
        // Handle error
      }
    },
  }),
  ```

- [ ] Update team-switcher.tsx:
  ```typescript
  "use client";

  import { useSelector } from "react-redux";
  import { useSwitchTenantMutation, useGetTenantsQuery } from "@/store/services/authApi";
  import type { RootState } from "@/store";

  export function TeamSwitcher() {
    const { activeTenant } = useSelector((state: RootState) => state.auth);
    const { data: tenants } = useGetTenantsQuery();
    const [switchTenant, { isLoading }] = useSwitchTenantMutation();

    const handleSwitch = async (tenantId: string) => {
      if (tenantId === activeTenant?.tenantId) return;
      await switchTenant(tenantId);
    };

    // ... rest of component
  }
  ```

- [ ] Verification:
  - [ ] Login with user who has multiple tenants
  - [ ] See tenant list in switcher
  - [ ] Switch tenant → New token → UI updates

**Deliverables:**
- ✅ Tenant switching working
- ✅ Active tenant displayed
- ✅ Multi-tenant flow complete

---

## ✅ MVP Completion Checklist

### Manual Testing

**Authentication Flow:**
- [ ] Navigate to /dashboard without login → Redirected to /login
- [ ] Login with valid credentials → Success → Redirected to /dashboard
- [ ] Login with invalid credentials → Error message displayed
- [ ] Refresh page after login → Still authenticated (token from localStorage)
- [ ] Click logout → Redirected to /login, state cleared
- [ ] Check localStorage → accessToken saved/cleared correctly
- [ ] Check cookies → refresh_token exists after login

**Token Management:**
- [ ] Make API call with valid token → Success
- [ ] Make API call with expired token → Auto-refresh → Success
- [ ] Logout → refresh_token cookie cleared
- [ ] Manual token expiry test (change exp claim) → Auto-refresh triggered

**Multi-Tenant (if implemented):**
- [ ] Login with multi-tenant user → See tenant list
- [ ] Switch tenant → New access token received
- [ ] UI updates to show active tenant
- [ ] API calls include new tenant context

**Redux State:**
- [ ] Open Redux DevTools → See auth state
- [ ] Login → State populated correctly
- [ ] Logout → State cleared
- [ ] No sensitive data logged to console

**Error Handling:**
- [ ] Network error → User-friendly message
- [ ] 401 error → Auto-refresh attempt → Logout on failure
- [ ] 500 error → Error message displayed
- [ ] Validation error → Field-specific messages

---

### Code Quality Checks

**TypeScript:**
- [ ] No `any` types (use proper interfaces)
- [ ] No TypeScript errors (`npm run build`)
- [ ] Proper type exports from authSlice and authApi

**Security:**
- [ ] No tokens logged to console (production)
- [ ] No sensitive data in error messages
- [ ] Refresh token never accessed from JavaScript
- [ ] HTTPS enforced in production (cookies Secure flag)

**Performance:**
- [ ] No unnecessary re-renders
- [ ] Memoized selectors if needed
- [ ] Proper dependency arrays in useEffect

**Code Organization:**
- [ ] Consistent file naming (camelCase vs kebab-case)
- [ ] Proper imports (@ alias used)
- [ ] No duplicate code

---

## 🎓 Learning Resources

### Next.js 16 App Router
- [Next.js Middleware](https://nextjs.org/docs/app/building-your-application/routing/middleware)
- [Next.js API Routes](https://nextjs.org/docs/app/building-your-application/routing/route-handlers)
- [Server vs Client Components](https://nextjs.org/docs/app/building-your-application/rendering/server-components)

### Redux Toolkit
- [RTK Query Tutorial](https://redux-toolkit.js.org/tutorials/rtk-query)
- [RTK Query Authentication](https://redux-toolkit.js.org/rtk-query/usage/customizing-queries#automatic-re-authorization-by-extending-fetchbasequery)
- [Redux DevTools](https://github.com/reduxjs/redux-devtools)

### Security
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://blog.logrocket.com/jwt-authentication-best-practices/)
- [Cookie Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies)

---

## 📊 Success Metrics

### Functionality
- ✅ User can login and access protected routes
- ✅ User stays authenticated after page refresh
- ✅ User can logout successfully
- ✅ Auto token refresh works on 401 errors
- ✅ Error messages displayed for failed login

### Performance
- ✅ Login response time < 2 seconds
- ✅ No unnecessary API calls
- ✅ No blocking operations on main thread
- ✅ Build succeeds without warnings

### Code Quality
- ✅ TypeScript strict mode passes
- ✅ ESLint passes with no errors
- ✅ No console.log in production code
- ✅ Proper error handling everywhere

### Security
- ✅ Refresh tokens in httpOnly cookies
- ✅ Access tokens not exposed in URLs
- ✅ No XSS vulnerabilities
- ✅ CSRF protection via cookies

---

## 🚀 Next Steps After MVP

### Phase 4.5: Polish & Enhancement
1. Add password reset flow
2. Add email verification flow
3. Add "Remember me" functionality
4. Improve error messages (i18n)
5. Add loading skeletons

### Phase 5: Testing & Quality
1. Write unit tests for authSlice
2. Write integration tests for RTK Query
3. Add E2E tests with Playwright
4. Add visual regression tests

### Phase 6: Advanced Features
1. Implement role-based access control
2. Add session timeout warnings
3. Add concurrent session detection
4. Add activity logging
5. Add security event notifications

---

## 📞 Support & Escalation

### Common Issues

**Issue:** Login succeeds but user redirected back to login
**Solution:** Check middleware.ts logic, verify token in cookies

**Issue:** Token refresh creates infinite loop
**Solution:** Check baseQueryWithReauth, ensure logout on refresh failure

**Issue:** CORS errors in development
**Solution:** Verify backend CORS config, use proxy if needed

**Issue:** Cookies not sent with requests
**Solution:** Verify `credentials: 'include'` in baseQuery

**Issue:** TypeScript errors after install
**Solution:** Run `npm install @types/jwt-decode`

---

## 🎯 Conclusion

Phase 4 MVP is well-defined and achievable in 7 working days with clear implementation path. Key success factors:

1. ✅ **Clear Scope**: Must-have vs nice-to-have features defined
2. ✅ **Adapted Architecture**: Next.js patterns, cookie-based tokens
3. ✅ **Risk Mitigation**: CORS, cookies, backend availability addressed
4. ✅ **Sequential Implementation**: Dependency chain prevents blockers
5. ✅ **Testing Strategy**: Manual testing checklist for validation

**Recommended Start Date:** After reviewing this analysis
**Estimated Completion:** 7 working days (2 calendar weeks part-time)
**Risk Level:** LOW-MEDIUM (mitigatable blockers identified)

---

**Document Version:** 1.0
**Last Updated:** 2025-12-17
**Next Review:** After Phase 4 completion
