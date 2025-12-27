// Multi-Company API Service using RTK Query
// Handles all multi-company related API calls
// PHASE 5: Frontend State Management

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { setAccessToken } from '../slices/authSlice';
import { setActiveCompany, setAvailableCompanies, setLoading } from '../slices/companySlice';
import type { RootState } from '../index';
import type {
  ApiSuccessResponse,
} from '@/types/api';
import type {
  AvailableCompany,
  ActiveCompany,
  SwitchCompanyRequest,
  SwitchCompanyResponse,
  GetAvailableCompaniesResponse,
  CompanyAccess,
  CompanyRole,
} from '@/types/company.types';
import { companyApi } from './companyApi';
import { companyUserApi } from './companyUserApi';

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
  credentials: 'include', // Send cookies (refresh_token, csrf_token)
  prepareHeaders: (headers, { getState }) => {
    const state = getState() as RootState;

    // Add Authorization header if token exists
    const token = state.auth.accessToken;
    if (token) {
      headers.set('authorization', `Bearer ${token}`);
    }

    // Add X-Company-ID header for multi-company context
    const activeCompanyId = state.company.activeCompany?.id;
    if (activeCompanyId) {
      headers.set('X-Company-ID', activeCompanyId);
    }

    // Add CSRF token header for POST/PUT/DELETE requests
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers.set('X-CSRF-Token', csrfToken);
    }

    return headers;
  },
});

/**
 * Calculate granular permissions based on role
 * Backend sends role, frontend derives permissions
 */
function calculatePermissions(role: CompanyRole): {
  canView: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canApprove: boolean;
} {
  switch (role) {
    case 'OWNER':
    case 'ADMIN':
      return {
        canView: true,
        canEdit: true,
        canDelete: true,
        canApprove: true,
      };
    case 'FINANCE':
      return {
        canView: true,
        canEdit: true,
        canDelete: false,
        canApprove: true,
      };
    case 'SALES':
    case 'WAREHOUSE':
      return {
        canView: true,
        canEdit: true,
        canDelete: false,
        canApprove: false,
      };
    case 'STAFF':
      return {
        canView: true,
        canEdit: false,
        canDelete: false,
        canApprove: false,
      };
    default:
      return {
        canView: false,
        canEdit: false,
        canDelete: false,
        canApprove: false,
      };
  }
}

/**
 * Transform available company to active company with access permissions
 */
function transformToActiveCompany(company: AvailableCompany): ActiveCompany {
  const permissions = calculatePermissions(company.role);

  const access: CompanyAccess = {
    companyId: company.id,
    tenantId: company.tenantId,
    role: company.role,
    accessTier: company.accessTier,
    hasAccess: company.accessTier > 0,
    ...permissions,
  };

  return {
    ...company,
    access,
  };
}

/**
 * Multi-Company API endpoints
 */
export const multiCompanyApi = createApi({
  reducerPath: 'multiCompanyApi',
  baseQuery,
  tagTypes: ['AvailableCompanies', 'ActiveCompany'],
  endpoints: (builder) => ({
    /**
     * Get available companies for user
     * GET /api/v1/companies
     */
    getAvailableCompanies: builder.query<AvailableCompany[], void>({
      query: () => '/companies',
      transformResponse: (response: ApiSuccessResponse<any[]>) => {
        // Map backend field names to frontend types
        // Backend sends: userRole, accessTier
        // Frontend expects: role, accessTier
        return response.data.map((company: any) => ({
          ...company,
          role: company.userRole || company.role, // Map userRole → role
        }));
      },
      providesTags: ['AvailableCompanies'],
      async onQueryStarted(_, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setAvailableCompanies(data));
        } catch (error) {
          console.error('Failed to fetch available companies:', error);
        }
      },
    }),

    /**
     * Switch active company context
     * POST /api/v1/auth/switch-company
     * Returns new access token with updated company context
     */
    switchCompany: builder.mutation<SwitchCompanyResponse, string>({
      query: (companyId) => ({
        url: '/auth/switch-company',
        method: 'POST',
        body: { company_id: companyId } as SwitchCompanyRequest,
      }),
      transformResponse: (response: ApiSuccessResponse<SwitchCompanyResponse>) =>
        response.data,
      async onQueryStarted(companyId, { dispatch, getState, queryFulfilled }) {
        try {
          dispatch(setLoading(true));
          const { data } = await queryFulfilled;
          const state = getState() as RootState;

          // Update access token with new company context
          dispatch(setAccessToken(data.access_token));

          // Find and set active company from available companies
          const availableCompanies = state.company.availableCompanies;
          const company = availableCompanies.find((c) => c.id === companyId);

          if (company) {
            const activeCompany = transformToActiveCompany(company);
            dispatch(
              setActiveCompany({
                company: activeCompany,
                persistToLocalStorage: true,
              })
            );
          }

          // Invalidate company-related caches to force refetch with new company context
          dispatch(companyApi.util.invalidateTags(['Company', 'Banks']));
          dispatch(companyUserApi.util.invalidateTags(['CompanyUsers']));

          console.log('✅ Company switched, invalidated Company, Banks, and CompanyUsers cache');
        } catch (error) {
          dispatch(setLoading(false));
          console.error('Failed to switch company:', error);
        }
      },
      invalidatesTags: ['ActiveCompany'],
    }),

    /**
     * Initialize company context after login
     * Automatically fetches available companies and sets first active one
     */
    initializeCompanyContext: builder.mutation<void, void>({
      queryFn: async (_, { dispatch }) => {
        try {
          console.log('🔄 [initializeCompanyContext] Starting...');

          // Fetch available companies
          console.log('📡 [initializeCompanyContext] Fetching companies...');
          const companiesResult = await dispatch(
            multiCompanyApi.endpoints.getAvailableCompanies.initiate()
          );

          console.log('📦 [initializeCompanyContext] Companies result:', companiesResult);

          if (companiesResult.data && companiesResult.data.length > 0) {
            console.log('✅ Received companies:', companiesResult.data);

            // Check localStorage for previously active company
            let targetCompanyId: string | null = null;
            if (typeof window !== 'undefined') {
              targetCompanyId = localStorage.getItem('activeCompanyId');
              console.log('💾 localStorage activeCompanyId:', targetCompanyId);
            }

            // Validate that stored company exists and is active
            const storedCompany = targetCompanyId
              ? companiesResult.data.find(
                  (c) => c.id === targetCompanyId && c.isActive
                )
              : null;

            // Use stored company if valid, otherwise use first active company
            const targetCompany =
              storedCompany ||
              companiesResult.data.find((c) => c.isActive) ||
              companiesResult.data[0];

            // Switch to target company
            console.log('🔄 Switching to company:', {
              id: targetCompany.id,
              name: targetCompany.name,
              isActive: targetCompany.isActive,
            });
            await dispatch(
              multiCompanyApi.endpoints.switchCompany.initiate(targetCompany.id)
            );

            // RTK Query requires { data: null } for void mutations, not { data: undefined }
            return { data: null };
          } else {
            // No companies available
            return {
              error: {
                status: 404,
                data: { message: 'No companies available' },
              } as FetchBaseQueryError,
            };
          }
        } catch (error) {
          return {
            error: {
              status: 'CUSTOM_ERROR',
              error: String(error),
            } as FetchBaseQueryError,
          };
        }
      },
    }),
  }),
});

// Export hooks for usage in components
export const {
  useGetAvailableCompaniesQuery,
  useSwitchCompanyMutation,
  useInitializeCompanyContextMutation,
} = multiCompanyApi;
