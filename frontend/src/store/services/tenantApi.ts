/**
 * Tenant API Service
 *
 * RTK Query service for tenant and user management.
 * Provides team member listing, invitations, role updates, and removal.
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  Tenant,
  TenantUser,
  InviteUserRequest,
  UpdateUserRoleRequest,
  GetUsersFilters,
} from "@/types/tenant.types";
import type { ApiSuccessResponse } from "@/types/api";

export const tenantApi = createApi({
  reducerPath: "tenantApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Tenant", "Users"],
  endpoints: (builder) => ({
    // ==================== Tenant Info ====================

    /**
     * Get Tenant Details
     * GET /api/v1/tenant
     * Returns tenant info with subscription details
     */
    getTenant: builder.query<Tenant, void>({
      query: () => "/tenant",
      transformResponse: (response: ApiSuccessResponse<Tenant>) =>
        response.data,
      providesTags: ["Tenant"],
    }),

    // ==================== User Management ====================

    /**
     * Get Users with Filters
     * GET /api/v1/tenant/users?role=ADMIN&isActive=true
     */
    getUsers: builder.query<TenantUser[], GetUsersFilters | void>({
      query: (filters) => {
        const params = new URLSearchParams();
        if (filters?.role) params.append("role", filters.role);
        if (filters?.isActive !== undefined)
          params.append("isActive", String(filters.isActive));
        if (filters?.page) params.append("page", String(filters.page));
        if (filters?.limit) params.append("limit", String(filters.limit));

        return {
          url: `/tenant/users${params.toString() ? `?${params.toString()}` : ""}`,
        };
      },
      transformResponse: (response: ApiSuccessResponse<TenantUser[]>) =>
        response.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: "Users" as const, id })),
              { type: "Users", id: "LIST" },
            ]
          : [{ type: "Users", id: "LIST" }],
    }),

    /**
     * Invite User
     * POST /api/v1/tenant/users/invite
     * Sends invitation email with temporary password
     */
    inviteUser: builder.mutation<TenantUser, InviteUserRequest>({
      query: (data) => ({
        url: "/tenant/users/invite",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TenantUser>) =>
        response.data,
      invalidatesTags: [{ type: "Users", id: "LIST" }],
    }),

    /**
     * Update User Role
     * PUT /api/v1/tenant/users/:id/role
     * Cannot change OWNER role or remove last ADMIN
     */
    updateUserRole: builder.mutation<
      TenantUser,
      { userId: string; data: UpdateUserRoleRequest }
    >({
      query: ({ userId, data }) => ({
        url: `/tenant/users/${userId}/role`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TenantUser>) =>
        response.data,
      invalidatesTags: (result, error, { userId }) => [
        { type: "Users", id: userId },
        { type: "Users", id: "LIST" },
      ],
    }),

    /**
     * Remove User (Soft Delete)
     * DELETE /api/v1/tenant/users/:id
     * Cannot remove OWNER or last ADMIN
     */
    removeUser: builder.mutation<void, string>({
      query: (userId) => ({
        url: `/tenant/users/${userId}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "Users", id: "LIST" }],
    }),
  }),
});

export const {
  useGetTenantQuery,
  useGetUsersQuery,
  useInviteUserMutation,
  useUpdateUserRoleMutation,
  useRemoveUserMutation,
} = tenantApi;
