/**
 * Customer API Service
 *
 * RTK Query service for customer management including:
 * - Customer CRUD operations
 * - Customer listing with filters and pagination
 * - Customer type management
 *
 * Backend endpoints: /api/v1/customers
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  CustomerResponse,
  CustomerListResponse,
  CustomerDetailResponse,
  CreateCustomerRequest,
  UpdateCustomerRequest,
  CustomerFilters,
} from "@/types/customer.types";
import type { ApiSuccessResponse } from "@/types/api";

export const customerApi = createApi({
  reducerPath: "customerApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Customer", "CustomerList"],
  endpoints: (builder) => ({
    // ==================== Customer CRUD ====================

    /**
     * List Customers with Filters & Pagination
     * GET /api/v1/customers
     */
    listCustomers: builder.query<CustomerListResponse, CustomerFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object, explicitly handling boolean values
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "code",
          sort_order: f.sortOrder || "asc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.customerType) params.customer_type = f.customerType;

        // Explicitly handle boolean filters (include false values!)
        if (f.isActive !== undefined) params.is_active = f.isActive;

        return {
          url: "/customers",
          params,
        };
      },
      // Force RTK Query to treat different filter values as different cache keys
      serializeQueryArgs: ({ queryArgs }) => {
        // Create unique cache key that includes all filter values
        return JSON.stringify(queryArgs);
      },
      // Always refetch when arguments change (important for filter changes!)
      forceRefetch: ({ currentArg, previousArg }) => {
        return JSON.stringify(currentArg) !== JSON.stringify(previousArg);
      },
      // Disable caching for filter changes
      keepUnusedDataFor: 0,
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "Customer" as const,
                id,
              })),
              { type: "CustomerList", id: "LIST" },
            ]
          : [{ type: "CustomerList", id: "LIST" }],
    }),

    /**
     * Get Customer Details
     * GET /api/v1/customers/:id
     */
    getCustomer: builder.query<CustomerResponse, string>({
      query: (id) => `/customers/${id}`,
      transformResponse: (response: ApiSuccessResponse<CustomerResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Customer", id }],
    }),

    /**
     * Create Customer
     * POST /api/v1/customers
     * Requires OWNER or ADMIN role
     */
    createCustomer: builder.mutation<CustomerResponse, CreateCustomerRequest>({
      query: (data) => ({
        url: "/customers",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<CustomerResponse>) =>
        response.data,
      invalidatesTags: [{ type: "CustomerList", id: "LIST" }],
    }),

    /**
     * Update Customer
     * PUT /api/v1/customers/:id
     * Requires OWNER or ADMIN role
     */
    updateCustomer: builder.mutation<
      CustomerResponse,
      { id: string; data: UpdateCustomerRequest }
    >({
      query: ({ id, data }) => ({
        url: `/customers/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<CustomerResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Customer", id },
        { type: "CustomerList", id: "LIST" },
      ],
    }),

    /**
     * Delete Customer (Soft Delete)
     * DELETE /api/v1/customers/:id
     * Requires OWNER or ADMIN role
     */
    deleteCustomer: builder.mutation<void, string>({
      query: (id) => ({
        url: `/customers/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Customer", id },
        { type: "CustomerList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Customer CRUD
  useListCustomersQuery,
  useGetCustomerQuery,
  useCreateCustomerMutation,
  useUpdateCustomerMutation,
  useDeleteCustomerMutation,
} = customerApi;
