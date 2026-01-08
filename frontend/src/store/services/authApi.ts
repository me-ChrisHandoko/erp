// Authentication API Service using RTK Query
// Handles all authentication-related API calls with auto-refresh

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { setCredentials, setAccessToken, logout } from '../slices/authSlice';
import type {
  ApiSuccessResponse,
  LoginRequest,
  LoginResponseData,
  LogoutRequest,
  SwitchTenantRequest,
  SwitchTenantResponseData,
  GetTenantsResponseData,
  RefreshTokenResponseData,
} from '@/types/api';

/**
 * Helper function to get CSRF token from cookie
 */
function getCSRFToken(): string | null {
  if (typeof document === 'undefined') return null;

  const name = 'csrf_token=';
  const decodedCookie = decodeURIComponent(document.cookie);
  const cookieArray = decodedCookie.split(';');

  for (let cookie of cookieArray) {
    cookie = cookie.trim();
    if (cookie.indexOf(name) === 0) {
      return cookie.substring(name.length);
    }
  }
  return null;
}

/**
 * Helper function to check if refresh_token cookie exists
 * Used to prevent unnecessary refresh attempts when user is not authenticated
 */
function hasRefreshTokenCookie(): boolean {
  if (typeof document === 'undefined') return false;

  const name = 'refresh_token=';
  const decodedCookie = decodeURIComponent(document.cookie);
  return decodedCookie.includes(name);
}

/**
 * Base query configuration with authentication and company context headers
 */
const baseQuery = fetchBaseQuery({
  baseUrl: `${process.env.NEXT_PUBLIC_API_URL}/api/v1`,
  credentials: 'include', // CRITICAL: Send cookies (refresh_token, csrf_token)
  prepareHeaders: (headers, { getState }) => {
    const state = getState() as any;

    // Add Authorization header if token exists
    const token = state.auth.accessToken;
    if (token) {
      headers.set('authorization', `Bearer ${token}`);
    }

    // Add X-Company-ID header for multi-company context
    const activeCompanyId = state.company?.activeCompany?.id;
    if (activeCompanyId) {
      headers.set('X-Company-ID', activeCompanyId);
    }

    // Add CSRF token header for POST/PUT/DELETE requests
    // CSRF token is read from cookie and sent in header (double-submit pattern)
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers.set('X-CSRF-Token', csrfToken);
    }

    return headers;
  },
});

/**
 * Mutex for refresh token to prevent concurrent refresh attempts
 * This prevents race conditions when multiple API calls fail with 401 simultaneously
 */
let refreshTokenPromise: Promise<any> | null = null;

/**
 * Base query with automatic token refresh on 401 errors
 * Implements the refresh token flow with race condition prevention:
 * 1. Original request fails with 401
 * 2. Check if refresh is already in progress (mutex check)
 * 3. If yes, wait for existing refresh to complete and reuse result
 * 4. If no, initiate new refresh and store promise for concurrent requests
 * 5. Call /auth/refresh endpoint (refresh_token sent via cookie)
 * 6. Update access token in Redux
 * 7. Retry original request with new token
 * 8. If refresh fails, logout user
 *
 * Race Condition Fix:
 * - Multiple concurrent 401 errors will share the same refresh promise
 * - Only ONE /auth/refresh request is made regardless of concurrent failures
 * - All waiting requests reuse the result from the single refresh call
 */
const baseQueryWithReauth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  // Execute original request
  let result = await baseQuery(args, api, extraOptions);

  // Check if request failed with 401 Unauthorized
  if (result.error && result.error.status === 401) {
    // 🔐 FIX #1: Skip refresh for authentication endpoints to prevent infinite loop
    // Authentication endpoints (login, logout, register) should NOT trigger token refresh
    // because 401 from these endpoints means authentication failure, not expired token
    const url = typeof args === 'string' ? args : args.url;
    const isAuthEndpoint = url.includes('/auth/login') ||
                           url.includes('/auth/logout') ||
                           url.includes('/auth/register');

    if (isAuthEndpoint) {
      console.log('[Auth] 401 from authentication endpoint, skipping refresh to prevent loop');
      return result; // Return error directly without attempting refresh
    }

    // 🔐 FIX #3: Check if refresh_token cookie exists before attempting refresh
    // If no refresh_token cookie exists, user is not authenticated and refresh will fail
    // This prevents unnecessary network calls and infinite loops
    if (!hasRefreshTokenCookie()) {
      console.log('[Auth] No refresh_token cookie found, user not authenticated');
      console.log('[Auth] Forcing logout due to missing refresh token');
      api.dispatch(logout({ reason: 'session_expired' }));
      return result; // Return error without attempting refresh
    }

    console.log('[Auth] Access token expired, attempting refresh...');

    // RACE CONDITION PREVENTION: Check if refresh is already in progress
    if (!refreshTokenPromise) {
      console.log('[Auth] Initiating new token refresh (no refresh in progress)');

      // Create new refresh promise and store it for concurrent requests
      // Wrap in Promise.resolve to ensure it's a proper Promise with .finally()
      refreshTokenPromise = Promise.resolve(
        baseQuery(
          { url: '/auth/refresh', method: 'POST' },
          api,
          extraOptions
        )
      ).finally(() => {
        // Clear the promise when refresh completes (success or failure)
        // This allows new refreshes after current one finishes
        console.log('[Auth] Token refresh completed, clearing mutex');
        refreshTokenPromise = null;
      });
    } else {
      console.log('[Auth] Token refresh already in progress, waiting for result...');
    }

    // Wait for the refresh to complete (either new or existing)
    // Multiple concurrent 401s will all wait for the same promise
    const refreshResult = await refreshTokenPromise;

    if (refreshResult.data) {
      console.log('[Auth] Token refresh successful');

      // Extract new access token from response envelope
      const refreshData = refreshResult.data as ApiSuccessResponse<RefreshTokenResponseData>;
      const newAccessToken = refreshData.data.accessToken;

      // Update Redux state with new access token
      api.dispatch(setAccessToken(newAccessToken));

      // SOLUTION 2: Fetch user data after successful token refresh
      console.log('[Auth] Fetching user data after token refresh...');
      try {
        const userResult = await baseQuery(
          { url: '/auth/me', method: 'GET' },
          api,
          extraOptions
        );

        if (userResult.data) {
          const userData = userResult.data as ApiSuccessResponse<{ user: any; activeTenant: any }>;
          console.log('[Auth] User data fetched successfully:', {
            userId: userData.data.user.id,
            email: userData.data.user.email,
            fullName: userData.data.user.fullName,
          });

          // Update Redux with complete credentials
          api.dispatch(setCredentials({
            user: userData.data.user,
            accessToken: newAccessToken,
            activeTenant: userData.data.activeTenant,
            availableTenants: userData.data.activeTenant ? [userData.data.activeTenant] : [],
          }));

          console.log('[Auth] Redux state updated with user data after refresh');
        } else {
          // 🔐 FIX #3 Option A: /auth/me returned no data
          console.warn('[Auth] User data fetch returned no data');
          console.log('[Auth] Attempting fallback to existing user data...');

          const currentState = api.getState() as any;
          if (currentState.auth.user) {
            console.log('[Auth] Using existing user data from state');
            api.dispatch(setCredentials({
              user: currentState.auth.user,
              accessToken: newAccessToken,
              activeTenant: currentState.auth.activeTenant,
              availableTenants: currentState.auth.availableTenants,
            }));
          } else {
            console.error('[Auth] No existing user data available, forcing logout');
            api.dispatch(logout({ reason: 'session_expired' }));
          }
        }
      } catch (err) {
        console.error('[Auth] Failed to fetch user data after refresh:', err);

        // 🔐 FIX #3 Option A: Fallback to existing user data
        // Token refresh was successful, so user IS authenticated
        // Use existing user data to maintain session continuity
        console.log('[Auth] Attempting fallback to existing user data...');
        const currentState = api.getState() as any;
        if (currentState.auth.user) {
          console.log('[Auth] Using existing user data from state');
          api.dispatch(setCredentials({
            user: currentState.auth.user,
            accessToken: newAccessToken,
            activeTenant: currentState.auth.activeTenant,
            availableTenants: currentState.auth.availableTenants,
          }));
        } else {
          console.error('[Auth] No existing user data available, forcing logout');
          api.dispatch(logout({ reason: 'session_expired' }));
        }
      }

      // 🔐 FIX #4 Option B: Smart retry with company context check
      // Check if company context is available before retrying
      console.log('[Auth] Preparing to retry original request with new token');

      const currentState = api.getState() as any;
      let activeCompanyId = currentState.company?.activeCompany?.id;

      // If no company in state, try localStorage fallback (for returning users)
      if (!activeCompanyId && typeof window !== 'undefined') {
        try {
          const storedCompanyId = localStorage.getItem('activeCompanyId');

          // Validate company ID format (alphanumeric, hyphens, underscores only)
          if (storedCompanyId && /^[a-zA-Z0-9-_]+$/.test(storedCompanyId)) {
            activeCompanyId = storedCompanyId;
            console.log('[Auth] Using validated company ID from localStorage for retry');
          } else if (storedCompanyId) {
            // Invalid format - clear corrupted data (self-healing)
            console.warn('[Auth] Invalid company ID format in localStorage, clearing:', storedCompanyId);
            localStorage.removeItem('activeCompanyId');
          }
        } catch (error) {
          // localStorage access failed (disabled, full, or other error)
          console.warn('[Auth] localStorage access failed:', error);
        }
      }

      // Decide whether to retry based on company context availability
      if (!activeCompanyId) {
        // No company context available - skip retry
        // CompanyInitializer will run and user can navigate again after company selection
        console.warn('[Auth] No company context available, skipping automatic retry');
        console.log('[Auth] User should navigate again after company initialization completes');

        // Return a special error that frontend can handle gracefully
        // This is NOT a failure - it's expected for new users or first login
        result = {
          error: {
            status: 'COMPANY_CONTEXT_PENDING',
            data: {
              success: false,
              message: 'Company context is being initialized. Please retry your request after selecting a company.',
              code: 'COMPANY_CONTEXT_PENDING',
              shouldRetryAfterInit: true,
            },
          },
        } as any;
      } else {
        // Company context available - safe to retry
        console.log('[Auth] Company context available, retrying request with company ID:', activeCompanyId);

        // Check if we need to manually inject X-Company-ID header
        // (prepareHeaders will add it if company is in state, but might not be set yet)
        const needsManualHeader = !currentState.company?.activeCompany?.id && activeCompanyId;

        if (needsManualHeader) {
          // Manually inject X-Company-ID header for this retry
          // This handles the case where localStorage has company but Redux state doesn't yet
          console.log('[Auth] Manually injecting X-Company-ID header for retry');

          const argsWithCompanyHeader = typeof args === 'string'
            ? { url: args, headers: { 'X-Company-ID': activeCompanyId } }
            : {
                ...args,
                headers: {
                  ...(args.headers || {}),
                  'X-Company-ID': activeCompanyId,
                },
              };

          result = await baseQuery(argsWithCompanyHeader, api, extraOptions);
        } else {
          // Company is in state, prepareHeaders will handle it
          result = await baseQuery(args, api, extraOptions);
        }
      }
    } else {
      console.log('[Auth] Token refresh failed (session expired), logging out');

      // Refresh failed - logout user with session expired reason
      api.dispatch(logout({ reason: 'session_expired' }));
    }
  }

  // 🔐 HYBRID SOLUTION PART 2: Reactive 403 CSRF Error Handler
  // Handle CSRF token errors with automatic recovery
  if (result.error && result.error.status === 403) {
    const errorData = result.error.data as any;

    // 🔐 FIX: Handle both string and object error formats from backend
    // Backend can return error as string OR as object {code, message}
    const errorMessage =
      (typeof errorData?.error === 'object' && errorData?.error?.message) ||
      (typeof errorData?.error === 'string' && errorData?.error) ||
      errorData?.message ||
      '';

    // Check if it's a CSRF-related error
    const isCSRFError =
      errorMessage.toLowerCase().includes('csrf') ||
      errorMessage.toLowerCase().includes('token') ||
      errorMessage.toLowerCase().includes('forbidden');

    if (isCSRFError) {
      console.log('[Auth] 🚨 403 CSRF error detected:', errorMessage);
      console.log('[Auth] 🔄 Attempting recovery via token refresh...');

      try {
        // RACE CONDITION PREVENTION: Check if refresh is already in progress
        if (!refreshTokenPromise) {
          console.log('[Auth] Initiating token refresh for CSRF recovery');

          // Create new refresh promise
          refreshTokenPromise = Promise.resolve(
            baseQuery(
              { url: '/auth/refresh', method: 'POST' },
              api,
              extraOptions
            )
          ).finally(() => {
            console.log('[Auth] CSRF recovery refresh completed');
            refreshTokenPromise = null;
          });
        } else {
          console.log('[Auth] Token refresh already in progress for CSRF recovery');
        }

        // Wait for refresh to complete
        const refreshResult = await refreshTokenPromise;

        if (refreshResult.data) {
          console.log('[Auth] ✅ Token refresh successful, CSRF token regenerated');

          // Extract new access token
          const refreshData = refreshResult.data as ApiSuccessResponse<RefreshTokenResponseData>;
          const newAccessToken = refreshData.data.accessToken;

          // Update Redux state with new access token
          api.dispatch(setAccessToken(newAccessToken));

          // Retry original request with new CSRF token
          console.log('[Auth] 🔄 Retrying original request with new CSRF token...');
          result = await baseQuery(args, api, extraOptions);

          if (result.data) {
            console.log('[Auth] ✅ Request succeeded after CSRF recovery');
          } else if (result.error) {
            console.error('[Auth] ❌ Request still failed after CSRF recovery:', result.error);
          }
        } else {
          console.error('[Auth] ❌ Token refresh failed during CSRF recovery');
          console.log('[Auth] Unable to recover from CSRF error, user may need to re-login');
        }
      } catch (error) {
        console.error('[Auth] ❌ CSRF recovery error:', error);
        console.log('[Auth] Fallback: User may need to reload page or re-login');
      }
    } else {
      console.log('[Auth] 403 error (not CSRF-related):', errorMessage);
    }
  }

  return result;
};

// Export baseQueryWithReauth for use in other API services
export { baseQueryWithReauth };

/**
 * Authentication API endpoints
 */
export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User', 'Tenants', 'AvailableCompanies', 'Company', 'Banks', 'CompanyUsers'],
  endpoints: (builder) => ({
    /**
     * Login endpoint
     * POST /auth/login
     */
    login: builder.mutation<ApiSuccessResponse<LoginResponseData>, LoginRequest>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;

          // Extract data from response envelope
          const loginData = data.data;

          // Update Redux state with user credentials
          dispatch(
            setCredentials({
              user: loginData.user,
              accessToken: loginData.accessToken,
              activeTenant: null, // Will be set after fetching tenants
              availableTenants: [],
            })
          );

          console.log('[Auth] Login successful');
        } catch (err) {
          console.error('[Auth] Login failed:', err);
        }
      },
      invalidatesTags: ['User', 'Tenants'],
    }),

    /**
     * Logout endpoint
     * POST /auth/logout
     * Note: Refresh token is sent via httpOnly cookie
     */
    logout: builder.mutation<ApiSuccessResponse<{ message: string }>, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          await queryFulfilled;
          console.log('[Auth] Logout successful');
        } catch (err) {
          console.error('[Auth] Logout API call failed:', err);
        } finally {
          // Always logout on client side first, even if API call fails
          // This will trigger the resetAllApiStates middleware to clear all RTK Query cache
          dispatch(logout());

          console.log('[Auth] Logout dispatched, cache will be cleared by middleware');
        }
      },
      invalidatesTags: ['User', 'Tenants'],
    }),

    /**
     * Switch tenant endpoint
     * POST /auth/switch-tenant
     */
    switchTenant: builder.mutation<
      ApiSuccessResponse<SwitchTenantResponseData>,
      SwitchTenantRequest
    >({
      query: (request) => ({
        url: '/auth/switch-tenant',
        method: 'POST',
        body: request,
      }),
      async onQueryStarted(arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;

          // Extract data from response envelope
          const switchData = data.data;

          // Update access token (contains new tenant context)
          dispatch(setAccessToken(switchData.accessToken));

          console.log('[Auth] Tenant switch successful');
        } catch (err) {
          console.error('[Auth] Tenant switch failed:', err);
        }
      },
      invalidatesTags: ['User'],
    }),

    /**
     * Get user's available tenants
     * GET /auth/tenants
     */
    getTenants: builder.query<ApiSuccessResponse<GetTenantsResponseData>, void>({
      query: () => '/auth/tenants',
      providesTags: ['Tenants'],
    }),

    /**
     * Get current user info
     * GET /auth/me
     */
    getCurrentUser: builder.query<ApiSuccessResponse<{ user: any; activeTenant: any }>, void>({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),
  }),
});

// Export hooks for usage in components
export const {
  useLoginMutation,
  useLogoutMutation,
  useSwitchTenantMutation,
  useGetTenantsQuery,
  useGetCurrentUserQuery,
} = authApi;
