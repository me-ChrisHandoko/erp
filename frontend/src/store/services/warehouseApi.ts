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
        if (f.type) params.type = f.type;
        if (f.city) params.city = f.city;
        if (f.province) params.province = f.province;
        if (f.managerID) params.managerID = f.managerID;

        // Explicitly handle boolean filters (include false values!)
        // Note: Backend warehouse DTO uses camelCase, not snake_case
        if (f.isActive !== undefined) params.isActive = f.isActive;

        return {
          url: "/warehouses",
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
