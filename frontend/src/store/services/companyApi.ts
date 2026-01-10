/**
 * Company API Service
 *
 * RTK Query service for company profile and bank account management.
 * Provides automatic caching, loading states, and optimistic updates.
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  CompanyResponse,
  UpdateCompanyRequest,
  BankAccountResponse,
  BankAccountFilters,
  BankAccountListResponse,
  AddBankAccountRequest,
  UpdateBankAccountRequest,
} from "@/types/company.types";
import type { ApiSuccessResponse, ApiPaginatedResponse } from "@/types/api";

export const companyApi = createApi({
  reducerPath: "companyApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Company", "Banks"],
  endpoints: (builder) => ({
    // ==================== Company Profile ====================

    /**
     * Get Company Profile
     * GET /api/v1/company
     */
    getCompany: builder.query<CompanyResponse, void>({
      query: () => "/company",
      transformResponse: (response: ApiSuccessResponse<CompanyResponse>) =>
        response.data,
      providesTags: ["Company"],
    }),

    /**
     * Update Company Profile
     * PUT /api/v1/company
     */
    updateCompany: builder.mutation<CompanyResponse, UpdateCompanyRequest>({
      query: (data) => ({
        url: "/company",
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<CompanyResponse>) =>
        response.data,
      invalidatesTags: ["Company"],
    }),

    /**
     * Upload Company Logo
     * POST /api/v1/company/logo
     */
    uploadLogo: builder.mutation<{ logoUrl: string }, FormData>({
      query: (formData) => ({
        url: "/company/logo",
        method: "POST",
        body: formData,
      }),
      transformResponse: (response: ApiSuccessResponse<{ logoUrl: string }>) =>
        response.data,
      invalidatesTags: ["Company"],
    }),

    // ==================== Bank Accounts ====================

    /**
     * Get Bank Accounts (with pagination and filters)
     * GET /api/v1/company/banks
     */
    getBankAccounts: builder.query<BankAccountListResponse, BankAccountFilters | void>({
      query: (filters) => {
        const params = new URLSearchParams();

        if (filters) {
          if (filters.search) params.append("search", filters.search);
          if (filters.isPrimary !== undefined) params.append("is_primary", String(filters.isPrimary));
          if (filters.isActive !== undefined) params.append("is_active", String(filters.isActive));
          if (filters.page) params.append("page", String(filters.page));
          if (filters.pageSize) params.append("page_size", String(filters.pageSize));
          if (filters.sortBy) params.append("sort_by", filters.sortBy);
          if (filters.sortOrder) params.append("sort_order", filters.sortOrder);
        }

        const queryString = params.toString();
        return queryString ? `/company/banks?${queryString}` : "/company/banks";
      },
      transformResponse: (response: ApiPaginatedResponse<BankAccountResponse>) => ({
        data: response.data,
        pagination: response.pagination,
      }),
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({ type: "Banks" as const, id })),
              { type: "Banks", id: "LIST" },
            ]
          : [{ type: "Banks", id: "LIST" }],
    }),

    /**
     * Add Bank Account
     * POST /api/v1/company/banks
     */
    addBankAccount: builder.mutation<BankAccountResponse, AddBankAccountRequest>({
      query: (data) => ({
        url: "/company/banks",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<BankAccountResponse>) =>
        response.data,
      invalidatesTags: [{ type: "Banks", id: "LIST" }],
    }),

    /**
     * Update Bank Account
     * PUT /api/v1/company/banks/:id
     */
    updateBankAccount: builder.mutation<
      BankAccountResponse,
      { id: string; data: UpdateBankAccountRequest }
    >({
      query: ({ id, data }) => ({
        url: `/company/banks/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<BankAccountResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Banks", id },
        { type: "Banks", id: "LIST" },
      ],
    }),

    /**
     * Delete Bank Account
     * DELETE /api/v1/company/banks/:id
     */
    deleteBankAccount: builder.mutation<void, string>({
      query: (id) => ({
        url: `/company/banks/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "Banks", id: "LIST" }],
    }),
  }),
});

export const {
  useGetCompanyQuery,
  useUpdateCompanyMutation,
  useUploadLogoMutation,
  useGetBankAccountsQuery,
  useAddBankAccountMutation,
  useUpdateBankAccountMutation,
  useDeleteBankAccountMutation,
} = companyApi;
