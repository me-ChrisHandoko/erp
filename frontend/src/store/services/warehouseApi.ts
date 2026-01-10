/**
 * Warehouse API Service
 *
 * RTK Query service for warehouse management including:
 * - Warehouse CRUD operations
 * - Warehouse listing with filters and pagination
 *
 * Backend endpoints: /api/v1/warehouses
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  WarehouseResponse,
  WarehouseListResponse,
  CreateWarehouseRequest,
  UpdateWarehouseRequest,
  WarehouseFilters,
} from "@/types/warehouse.types";
import type { ApiSuccessResponse } from "@/types/api";

export const warehouseApi = createApi({
  reducerPath: "warehouseApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Warehouse", "WarehouseList"],
  endpoints: (builder) => ({
    // ==================== Warehouse CRUD ====================

    /**
     * List Warehouses with Filters & Pagination
     * GET /api/v1/warehouses
     */
    listWarehouses: builder.query<
      WarehouseListResponse,
      WarehouseFilters | void
    >({
      query: (filters) => {
        const params = filters || {};
        return {
          url: "/warehouses",
          params: {
            search: params.search,
            type: params.type,
            city: params.city,
            province: params.province,
            manager_id: params.managerID, // ← snake_case for backend
            is_active: params.isActive, // ← snake_case for backend
            page: params.page || 1,
            page_size: params.pageSize || 20, // ← snake_case for backend
            sort_by: params.sortBy || "code", // ← snake_case for backend
            sort_order: params.sortOrder || "asc", // ← snake_case for backend
          },
        };
      },
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "Warehouse" as const,
                id,
              })),
              { type: "WarehouseList", id: "LIST" },
            ]
          : [{ type: "WarehouseList", id: "LIST" }],
    }),

    /**
     * Get Warehouse Details
     * GET /api/v1/warehouses/:id
     */
    getWarehouse: builder.query<WarehouseResponse, string>({
      query: (id) => `/warehouses/${id}`,
      transformResponse: (response: ApiSuccessResponse<WarehouseResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Warehouse", id }],
    }),

    /**
     * Create Warehouse
     * POST /api/v1/warehouses
     * Requires OWNER or ADMIN role
     */
    createWarehouse: builder.mutation<
      WarehouseResponse,
      CreateWarehouseRequest
    >({
      query: (data) => ({
        url: "/warehouses",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<WarehouseResponse>) =>
        response.data,
      invalidatesTags: [{ type: "WarehouseList", id: "LIST" }],
    }),

    /**
     * Update Warehouse
     * PUT /api/v1/warehouses/:id
     * Requires OWNER or ADMIN role
     */
    updateWarehouse: builder.mutation<
      WarehouseResponse,
      { id: string; data: UpdateWarehouseRequest }
    >({
      query: ({ id, data }) => ({
        url: `/warehouses/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<WarehouseResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Warehouse", id },
        { type: "WarehouseList", id: "LIST" },
      ],
    }),

    /**
     * Delete Warehouse (Soft Delete)
     * DELETE /api/v1/warehouses/:id
     * Requires OWNER or ADMIN role
     */
    deleteWarehouse: builder.mutation<void, string>({
      query: (id) => ({
        url: `/warehouses/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Warehouse", id },
        { type: "WarehouseList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Warehouse CRUD
  useListWarehousesQuery,
  useGetWarehouseQuery,
  useCreateWarehouseMutation,
  useUpdateWarehouseMutation,
  useDeleteWarehouseMutation,
} = warehouseApi;
