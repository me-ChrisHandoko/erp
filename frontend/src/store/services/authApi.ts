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
          console.warn('[Auth] User data fetch returned no data');
        }
      } catch (err) {
        console.error('[Auth] Failed to fetch user data after refresh:', err);
        // Continue anyway - token refresh was successful
      }

      // Retry the original request with new token
      console.log('[Auth] Retrying original request with new token');
      result = await baseQuery(args, api, extraOptions);
    } else {
      console.log('[Auth] Token refresh failed, logging out');

      // Refresh failed - logout user
      api.dispatch(logout());
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
