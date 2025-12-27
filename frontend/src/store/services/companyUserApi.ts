/**
 * Company User API Service
 *
 * RTK Query service for company-scoped user management.
 * Uses X-Company-ID header (automatically added from Redux state) for multi-company isolation.
 * Returns users who have UserCompanyRole for the active company only.
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  TenantUser,
  InviteUserRequest,
  UpdateUserRoleRequest,
  GetUsersFilters,
} from "@/types/tenant.types";
import type { ApiSuccessResponse } from "@/types/api";

export const companyUserApi = createApi({
  reducerPath: "companyUserApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["CompanyUsers"],
  endpoints: (builder) => ({
    // ==================== Company User Management ====================

    /**
     * Get Company Users
     * GET /api/v1/company/users?role=ADMIN&isActive=true
     *
     * Returns: Users with UserCompanyRole for the ACTIVE company
     * - Uses X-Company-ID header from Redux state (activeCompany)
     * - Filtered by company via CompanyContextMiddleware
     * - Cache invalidated when company switches
     */
    getCompanyUsers: builder.query<TenantUser[], GetUsersFilters | void>({
      query: (filters) => {
        const params = new URLSearchParams();
        if (filters?.role) params.append("role", filters.role);
        if (filters?.isActive !== undefined)
          params.append("isActive", String(filters.isActive));
        if (filters?.page) params.append("page", String(filters.page));
        if (filters?.limit) params.append("limit", String(filters.limit));

        return {
          url: `/company/users${params.toString() ? `?${params.toString()}` : ""}`,
        };
      },
      transformResponse: (response: ApiSuccessResponse<TenantUser[]>) =>
        response.data,
      // Include company ID in cache tags for proper invalidation on company switch
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: "CompanyUsers" as const, id })),
              { type: "CompanyUsers", id: "LIST" },
            ]
          : [{ type: "CompanyUsers", id: "LIST" }],
    }),

    /**
     * Invite User to Company
     * POST /api/v1/company/users/invite
     *
     * Creates UserCompanyRole for the active company
     * Sends invitation email with temporary password
     * Requires: OWNER/ADMIN role
     */
    inviteCompanyUser: builder.mutation<TenantUser, InviteUserRequest>({
      query: (data) => ({
        url: "/company/users/invite",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TenantUser>) =>
        response.data,
      invalidatesTags: [{ type: "CompanyUsers", id: "LIST" }],
    }),

    /**
     * Update User Role in Company
     * PUT /api/v1/company/users/:id/role
     *
     * Updates UserCompanyRole for the active company
     * Cannot change to OWNER role (company-level restriction)
     * Requires: OWNER/ADMIN role
     */
    updateCompanyUserRole: builder.mutation<
      TenantUser,
      { userId: string; data: UpdateUserRoleRequest }
    >({
      query: ({ userId, data }) => ({
        url: `/company/users/${userId}/role`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TenantUser>) =>
        response.data,
      invalidatesTags: (result, error, { userId }) => [
        { type: "CompanyUsers", id: userId },
        { type: "CompanyUsers", id: "LIST" },
      ],
    }),

    /**
     * Remove User from Company
     * DELETE /api/v1/company/users/:id
     *
     * Removes user's access to the active company (soft delete UserCompanyRole)
     * Does NOT delete the user account (user may have access to other companies)
     * Requires: OWNER/ADMIN role
     */
    removeCompanyUser: builder.mutation<void, string>({
      query: (userId) => ({
        url: `/company/users/${userId}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "CompanyUsers", id: "LIST" }],
    }),
  }),
});

export const {
  useGetCompanyUsersQuery,
  useInviteCompanyUserMutation,
  useUpdateCompanyUserRoleMutation,
  useRemoveCompanyUserMutation,
} = companyUserApi;
