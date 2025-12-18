# Frontend Authentication Implementation Guide

**Target:** React/TypeScript Frontend Developers  
**Version:** 1.0  
**Last Updated:** 2025-12-16  

---

## 🎯 Quick Start

Dokumen ini adalah **panduan implementasi frontend** untuk authentication system multi-tenant ERP menggunakan Redux Toolkit + RTK Query.

### Technology Stack

```
Framework:      React 18+
Language:       TypeScript
State:          Redux Toolkit
API Client:     RTK Query
Router:         React Router v6
Forms:          React Hook Form (optional)
Validation:     Zod / Yup (optional)
```

---

## 📦 Dependencies

```bash
npm install @reduxjs/toolkit react-redux
npm install react-router-dom
npm install jwt-decode

# Optional
npm install react-hook-form zod
```

---

## 🏗️ Redux Store Setup

### 1. Configure Store

```typescript
// src/store/index.ts
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

### 2. Auth Slice (State Management)

```typescript
// src/store/slices/authSlice.ts
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
    setCredentials: (state, action: PayloadAction<{
      user: User;
      accessToken: string;
      refreshToken: string;
      activeTenant: TenantContext;
      availableTenants: TenantContext[];
    }>) => {
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
  },
});

export const { setCredentials, setAccessToken, setActiveTenant, logout, setError } = authSlice.actions;
export default authSlice.reducer;
```

**Full implementation:** Lines 2020-2096 di `authentication-mvp-design.md`

---

## 🌐 RTK Query API Configuration

### Auth API Service dengan Auto-Refresh

```typescript
// src/store/services/authApi.ts
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

// Auto-refresh interceptor
const baseQueryWithReauth = async (args, api, extraOptions) => {
  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    // Try refresh token
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
        
        // Retry original request
        result = await baseQuery(args, api, extraOptions);
      } else {
        // Refresh failed, logout
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
          dispatch(logout()); // Logout even if API fails
        }
      },
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
  useLoginMutation,
  useLogoutMutation,
  useSwitchTenantMutation,
  useGetCurrentUserQuery,
} = authApi;
```

**Full implementation:** Lines 2097-2196 di `authentication-mvp-design.md`

---

## 🔄 Automatic Token Refresh

### Background Refresh Logic

```typescript
// src/utils/tokenRefresh.ts
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
      const expiresAt = decoded.exp * 1000;
      const now = Date.now();
      const timeUntilExpiry = expiresAt - now;

      // Refresh if expires in < 5 minutes
      if (timeUntilExpiry < 5 * 60 * 1000) {
        refreshAccessToken(refreshToken);
      }
    } catch (err) {
      console.error('Token decode error:', err);
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

      // Update refresh token if rotation enabled
      if (data.refreshToken) {
        // Update in state
      }
    } else {
      store.dispatch(logout());
    }
  } catch (err) {
    console.error('Token refresh error:', err);
    store.dispatch(logout());
  }
};
```

**Usage:**
```typescript
// src/App.tsx
import { setupTokenRefresh } from './utils/tokenRefresh';

function App() {
  useEffect(() => {
    setupTokenRefresh();
  }, []);

  return <Router>...</Router>;
}
```

**Full implementation:** Lines 2197-2260 di `authentication-mvp-design.md`

---

## 🛡️ Protected Routes

### Auth Guard Component

```typescript
// src/components/ProtectedRoute.tsx
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

### Router Setup

```typescript
// src/routes/index.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ProtectedRoute } from '../components/ProtectedRoute';
import { Login, Dashboard, Products, Invoices } from '../pages';

export const AppRoutes = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<Login />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />

        {/* Protected routes */}
        <Route element={<ProtectedRoute />}>
          <Route path="/dashboard" element={<Dashboard />} />
        </Route>

        {/* Role-restricted routes */}
        <Route element={<ProtectedRoute requiredRole="WAREHOUSE" />}>
          <Route path="/products" element={<Products />} />
        </Route>

        <Route element={<ProtectedRoute requiredRole="FINANCE" />}>
          <Route path="/invoices" element={<Invoices />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};
```

**Full implementation:** Lines 2261-2297 di `authentication-mvp-design.md`

---

## 🎨 UI Components

### Login Component

```typescript
// src/pages/Login.tsx
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

**Full implementation:** Lines 2298-2342 di `authentication-mvp-design.md`

### Tenant Switcher Component

```typescript
// src/components/TenantSwitcher.tsx
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

**Full implementation:** Lines 2343-2380 di `authentication-mvp-design.md`

---

## 📊 Error Handling

### Error Response Interface

```typescript
// src/types/api.ts
export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: unknown;
    timestamp: string;
  };
}

// Error codes
export const AuthErrorCodes = {
  INVALID_CREDENTIALS: 'AUTH_INVALID_CREDENTIALS',
  EMAIL_NOT_VERIFIED: 'AUTH_EMAIL_NOT_VERIFIED',
  ACCOUNT_LOCKED: 'AUTH_ACCOUNT_LOCKED',
  TOKEN_EXPIRED: 'AUTH_TOKEN_EXPIRED',
  TENANT_ACCESS_DENIED: 'AUTH_TENANT_ACCESS_DENIED',
  RATE_LIMIT_EXCEEDED: 'AUTH_RATE_LIMIT_EXCEEDED',
} as const;
```

### Error Display Component

```typescript
// src/components/ErrorMessage.tsx
import React from 'react';
import { ApiError } from '../types/api';

interface ErrorMessageProps {
  error: ApiError | null;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = ({ error }) => {
  if (!error) return null;

  const getMessage = (code: string) => {
    const messages: Record<string, string> = {
      AUTH_INVALID_CREDENTIALS: 'Email atau password salah',
      AUTH_EMAIL_NOT_VERIFIED: 'Silakan verifikasi email Anda terlebih dahulu',
      AUTH_ACCOUNT_LOCKED: 'Akun terkunci karena terlalu banyak percobaan gagal',
      AUTH_TOKEN_EXPIRED: 'Sesi Anda telah berakhir, silakan login kembali',
      AUTH_TENANT_ACCESS_DENIED: 'Anda tidak memiliki akses ke perusahaan ini',
      AUTH_RATE_LIMIT_EXCEEDED: 'Terlalu banyak permintaan, coba lagi nanti',
    };

    return messages[code] || error.error.message;
  };

  return (
    <div className="error-message">
      {getMessage(error.error.code)}
    </div>
  );
};
```

---

## 👥 Admin-Managed User Provisioning

**Security Model**: This ERP system uses admin-controlled user provisioning instead of self-service registration to ensure proper tenant isolation and access control.

### User Creation Workflow

```typescript
// Admin creates users through admin panel (to be implemented)
// No self-service registration for security and tenant isolation
```

**Process:**

1. **Tenant Owner/Admin** logs into admin panel
2. Navigates to User Management section
3. Creates new user with:
   - Email (pre-verified by admin)
   - Full name
   - Role: `OWNER`, `ADMIN`, `FINANCE`, `SALES`, `WAREHOUSE`, `STAFF`
   - Initial password (system-generated or manually set)
4. System sends email notification with credentials
5. User must change password on first login

**Benefits:**
- 🔒 **Enhanced Security**: No public registration endpoint prevents unauthorized access
- 🎯 **Access Control**: Admins control who joins their tenant
- 🏢 **Multi-Tenant Isolation**: Prevents unauthorized tenant creation
- ✅ **Pre-Verified Users**: Email addresses verified by admin entry

### Initial Tenant Setup

**Production SaaS Deployment:**
```bash
# Super admin creates first tenant via admin portal
# Or use CLI provisioning tool
```

**On-Premise Deployment:**
```bash
# Database seeding for initial tenant/user
# System admin configures first organization
```

### Self-Service Password Management

Users retain control over password recovery:

**Forgot Password Flow:**
```typescript
// User can reset forgotten passwords independently
<Route path="/forgot-password" element={<ForgotPassword />} />
<Route path="/reset-password" element={<ResetPassword />} />
```

**Process:**
1. User clicks "Forgot Password" on login page
2. Enters email address
3. Receives reset link via email
4. Sets new password with secure token
5. Redirected to login with new credentials

**Admin Password Reset:**
- Admins can force password reset for users
- Generates one-time reset link
- Sent directly to user's registered email

### UI Message for New Users

```typescript
// Login page already includes appropriate message
<p className="text-center text-sm text-muted-foreground">
  Belum punya akun?{" "}
  <a href="#" className="underline hover:text-primary">
    Hubungi administrator
  </a>
</p>
```

**Translation**: "Don't have an account? Contact administrator"

---

## 🧪 Testing

### Component Testing

```typescript
// src/pages/__tests__/Login.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { Login } from '../Login';
import { store } from '../../store';

describe('Login Component', () => {
  it('should render login form', () => {
    render(
      <Provider store={store}>
        <Login />
      </Provider>
    );

    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Password')).toBeInTheDocument();
  });

  it('should handle successful login', async () => {
    render(
      <Provider store={store}>
        <Login />
      </Provider>
    );

    fireEvent.change(screen.getByPlaceholderText('Email'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Password'), {
      target: { value: 'password123' },
    });

    fireEvent.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => {
      // Assert navigation to dashboard
    });
  });
});
```

### E2E Testing (Cypress/Playwright)

```typescript
// cypress/e2e/auth.cy.ts
describe('Authentication Flow', () => {
  it('should complete full login flow', () => {
    cy.visit('/login');
    cy.get('input[type="email"]').type('user@example.com');
    cy.get('input[type="password"]').type('SecurePass123!');
    cy.get('button[type="submit"]').click();

    cy.url().should('include', '/dashboard');
    cy.contains('Welcome').should('be.visible');
  });

  it('should switch tenants', () => {
    cy.login('user@example.com', 'SecurePass123!');
    cy.get('.tenant-switcher select').select('PT Example');

    cy.contains('PT Example').should('be.visible');
  });
});
```

---

## 📅 Frontend Implementation Timeline

### Phase 4: Frontend Integration (Week 4-5)

**Week 1:**
- ✅ Redux Toolkit store setup
- ✅ RTK Query API configuration
- ✅ Auth slice dengan TypeScript types
- ✅ Auto-refresh interceptor

**Week 2:**
- ✅ Protected route components
- ✅ Login form
- ✅ Tenant switcher UI
- ✅ Password reset flow
- ✅ Error handling & display
- ✅ E2E testing

**Deliverables:**
- Complete authentication UI (admin-managed user provisioning)
- Automatic token refresh
- Multi-tenant switching
- Error state handling
- E2E test coverage

---

## 📚 Reference Sections

| Topic | Line Numbers | File |
|-------|--------------|------|
| Redux Store Setup | 2020-2096 | authentication-mvp-design.md |
| RTK Query Config | 2097-2196 | authentication-mvp-design.md |
| Auto Token Refresh | 2197-2260 | authentication-mvp-design.md |
| Protected Routes | 2261-2297 | authentication-mvp-design.md |
| Login Component | 2298-2342 | authentication-mvp-design.md |
| Tenant Switcher | 2343-2380 | authentication-mvp-design.md |

---

## 🆘 Common Issues

**Q: Token refresh not working?**
A: Check `baseQueryWithReauth` implementation. Verify refreshToken is stored correctly.

**Q: Protected routes not redirecting?**
A: Ensure `isAuthenticated` state is properly set after login.

**Q: Tenant switch not updating UI?**
A: Check if `activeTenant` state is updated. Verify components subscribe to correct Redux state.

**Q: How do new users get access?**
A: This system uses admin-managed provisioning. New users must contact their tenant administrator to create an account. See "Admin-Managed User Provisioning" section above.

**Q: Why is there no registration page?**
A: For security and tenant isolation, user registration is controlled by tenant administrators. This prevents unauthorized account creation and ensures proper multi-tenant access control.

**Q: CORS errors in development?**
A: Configure proxy in `vite.config.ts` or `package.json`:
```json
"proxy": "http://localhost:8080"
```

---

**For complete implementation details, refer to: `authentication-mvp-design.md`**
