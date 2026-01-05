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
      query: (filters = {}) => ({
        url: "/customers",
        params: {
          search: filters.search,
          customer_type: filters.customerType,
          is_active: filters.isActive,
          page: filters.page || 1,
          page_size: filters.pageSize || 20,
          sort_by: filters.sortBy || "code",
          sort_order: filters.sortOrder || "asc",
        },
      }),
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
