/**
 * Supplier API Service
 *
 * RTK Query service for supplier management including:
 * - Supplier CRUD operations
 * - Supplier listing with filters and pagination
 * - Contact and business information management
 *
 * Backend endpoints: /api/v1/suppliers
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  SupplierResponse,
  SupplierListResponse,
  CreateSupplierRequest,
  UpdateSupplierRequest,
  SupplierFilters,
} from "@/types/supplier.types";
import type { ApiSuccessResponse } from "@/types/api";

export const supplierApi = createApi({
  reducerPath: "supplierApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Supplier", "SupplierList"],
  endpoints: (builder) => ({
    // ==================== Supplier CRUD ====================

    /**
     * List Suppliers with Filters & Pagination
     * GET /api/v1/suppliers
     */
    listSuppliers: builder.query<SupplierListResponse, SupplierFilters | void>(
      {
        query: (filters) => {
          const params = filters || {};
          return {
            url: "/suppliers",
            params: {
              search: params.search,
              type: params.type,
              city: params.city,
              province: params.province,
              is_active: params.isActive,
              page: params.page || 1,
              page_size: params.pageSize || 20,
              sort_by: params.sortBy || "code",
              sort_order: params.sortOrder || "asc",
            },
          };
        },
        providesTags: (result) =>
          result
            ? [
                ...result.data.map(({ id }) => ({
                  type: "Supplier" as const,
                  id,
                })),
                { type: "SupplierList", id: "LIST" },
              ]
            : [{ type: "SupplierList", id: "LIST" }],
      }
    ),

    /**
     * Get Supplier Details
     * GET /api/v1/suppliers/:id
     */
    getSupplier: builder.query<SupplierResponse, string>({
      query: (id) => `/suppliers/${id}`,
      transformResponse: (response: ApiSuccessResponse<SupplierResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Supplier", id }],
    }),

    /**
     * Create Supplier
     * POST /api/v1/suppliers
     * Requires OWNER or ADMIN role
     */
    createSupplier: builder.mutation<
      SupplierResponse,
      CreateSupplierRequest
    >({
      query: (data) => ({
        url: "/suppliers",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SupplierResponse>) =>
        response.data,
      invalidatesTags: [{ type: "SupplierList", id: "LIST" }],
    }),

    /**
     * Update Supplier
     * PUT /api/v1/suppliers/:id
     * Requires OWNER or ADMIN role
     */
    updateSupplier: builder.mutation<
      SupplierResponse,
      { id: string; data: UpdateSupplierRequest }
    >({
      query: ({ id, data }) => ({
        url: `/suppliers/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SupplierResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Supplier", id },
        { type: "SupplierList", id: "LIST" },
      ],
    }),

    /**
     * Delete Supplier (Soft Delete)
     * DELETE /api/v1/suppliers/:id
     * Requires OWNER or ADMIN role
     */
    deleteSupplier: builder.mutation<void, string>({
      query: (id) => ({
        url: `/suppliers/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Supplier", id },
        { type: "SupplierList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Supplier CRUD
  useListSuppliersQuery,
  useGetSupplierQuery,
  useCreateSupplierMutation,
  useUpdateSupplierMutation,
  useDeleteSupplierMutation,
} = supplierApi;
