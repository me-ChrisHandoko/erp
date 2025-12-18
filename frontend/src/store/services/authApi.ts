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
 * Base query configuration with authentication headers
 */
const baseQuery = fetchBaseQuery({
  baseUrl: `${process.env.NEXT_PUBLIC_API_URL}/api/v1`,
  credentials: 'include', // CRITICAL: Send cookies (refresh_token, csrf_token)
  prepareHeaders: (headers, { getState }) => {
    // Get access token from Redux state
    const token = (getState() as any).auth.accessToken;

    // Add Authorization header if token exists
    if (token) {
      headers.set('authorization', `Bearer ${token}`);
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
 * Base query with automatic token refresh on 401 errors
 * Implements the refresh token flow:
 * 1. Original request fails with 401
 * 2. Call /auth/refresh endpoint (refresh_token sent via cookie)
 * 3. Update access token in Redux
 * 4. Retry original request with new token
 * 5. If refresh fails, logout user
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

    // Attempt to refresh the token
    // Note: refreshToken is sent automatically via httpOnly cookie
    const refreshResult = await baseQuery(
      { url: '/auth/refresh', method: 'POST' },
      api,
      extraOptions
    );

    if (refreshResult.data) {
      console.log('[Auth] Token refresh successful');

      // Extract new access token from response envelope
      const refreshData = refreshResult.data as ApiSuccessResponse<RefreshTokenResponseData>;
      const newAccessToken = refreshData.data.accessToken;

      // Update Redux state with new access token
      api.dispatch(setAccessToken(newAccessToken));

      // Retry the original request with new token
      result = await baseQuery(args, api, extraOptions);
    } else {
      console.log('[Auth] Token refresh failed, logging out');

      // Refresh failed - logout user
      api.dispatch(logout());
    }
  }

  return result;
};

/**
 * Authentication API endpoints
 */
export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User', 'Tenants'],
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
          // Always logout on client side, even if API call fails
          dispatch(logout());
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
