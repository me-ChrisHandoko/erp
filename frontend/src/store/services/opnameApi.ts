/**
 * Stock Opname API Service
 *
 * RTK Query service for stock opname (physical inventory count) management including:
 * - Stock opname CRUD operations
 * - Opname item management
 * - Approval workflow
 * - Listing with filters and pagination
 *
 * Backend endpoints: /api/v1/stock-opnames
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  StockOpname,
  StockOpnameResponse,
  StockOpnameListResponse,
  StockOpnameFilters,
  CreateStockOpnameRequest,
  UpdateStockOpnameRequest,
  ApproveStockOpnameRequest,
  StockOpnameItem,
  UpdateStockOpnameItemRequest,
} from "@/types/opname.types";
import type { ApiSuccessResponse } from "@/types/api";

export const opnameApi = createApi({
  reducerPath: "opnameApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["StockOpname", "StockOpnameList", "StockOpnameItem"],
  endpoints: (builder) => ({
    // ==================== Stock Opname CRUD ====================

    /**
     * List Stock Opnames with Filters & Pagination
     * GET /api/v1/stock-opnames
     */
    listOpnames: builder.query<StockOpnameListResponse, StockOpnameFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "opnameDate",
          sort_order: f.sortOrder || "desc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.status) params.status = f.status;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/stock-opnames",
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
                type: "StockOpname" as const,
                id,
              })),
              { type: "StockOpnameList", id: "LIST" },
            ]
          : [{ type: "StockOpnameList", id: "LIST" }],
    }),

    /**
     * Get Stock Opname Details
     * GET /api/v1/stock-opnames/:id
     */
    getOpname: builder.query<StockOpnameResponse, string>({
      query: (id) => `/stock-opnames/${id}`,
      transformResponse: (response: ApiSuccessResponse<StockOpnameResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "StockOpname", id }],
    }),

    /**
     * Create Stock Opname
     * POST /api/v1/stock-opnames
     */
    createOpname: builder.mutation<StockOpnameResponse, CreateStockOpnameRequest>({
      query: (body) => ({
        url: "/stock-opnames",
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiSuccessResponse<StockOpnameResponse>) =>
        response.data,
      invalidatesTags: [{ type: "StockOpnameList", id: "LIST" }],
    }),

    /**
     * Update Stock Opname
     * PUT /api/v1/stock-opnames/:id
     */
    updateOpname: builder.mutation<
      StockOpnameResponse,
      { id: string; data: UpdateStockOpnameRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-opnames/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<StockOpnameResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "StockOpname", id },
        { type: "StockOpnameList", id: "LIST" },
      ],
    }),

    /**
     * Delete Stock Opname
     * DELETE /api/v1/stock-opnames/:id
     * Only allowed for draft status
     */
    deleteOpname: builder.mutation<void, string>({
      query: (id) => ({
        url: `/stock-opnames/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "StockOpnameList", id: "LIST" }],
    }),

    /**
     * Approve Stock Opname
     * POST /api/v1/stock-opnames/:id/approve
     * Changes status to 'approved' and applies stock adjustments
     */
    approveOpname: builder.mutation<
      StockOpnameResponse,
      { id: string; data?: ApproveStockOpnameRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-opnames/${id}/approve`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<StockOpnameResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "StockOpname", id },
        { type: "StockOpnameList", id: "LIST" },
      ],
      // Note: Warehouse stocks will be invalidated by backend events or manual refetch
    }),

    // ==================== Stock Opname Items ====================

    /**
     * Add Item to Stock Opname
     * POST /api/v1/stock-opnames/:opnameId/items
     */
    addOpnameItem: builder.mutation<
      StockOpnameItem,
      { opnameId: string; productId: string; expectedQty: string; actualQty: string; notes?: string }
    >({
      query: ({ opnameId, ...body }) => ({
        url: `/stock-opnames/${opnameId}/items`,
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiSuccessResponse<StockOpnameItem>) =>
        response.data,
      invalidatesTags: (result, error, { opnameId }) => [
        { type: "StockOpname", id: opnameId },
        { type: "StockOpnameItem", id: "LIST" },
      ],
    }),

    /**
     * Update Opname Item
     * PUT /api/v1/stock-opnames/:opnameId/items/:itemId
     */
    updateOpnameItem: builder.mutation<
      StockOpnameItem,
      { opnameId: string; itemId: string; data: UpdateStockOpnameItemRequest }
    >({
      query: ({ opnameId, itemId, data }) => ({
        url: `/stock-opnames/${opnameId}/items/${itemId}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<StockOpnameItem>) =>
        response.data,
      invalidatesTags: (result, error, { opnameId, itemId }) => [
        { type: "StockOpname", id: opnameId },
        { type: "StockOpnameItem", id: itemId },
      ],
    }),

    /**
     * Delete Opname Item
     * DELETE /api/v1/stock-opnames/:opnameId/items/:itemId
     */
    deleteOpnameItem: builder.mutation<void, { opnameId: string; itemId: string }>({
      query: ({ opnameId, itemId }) => ({
        url: `/stock-opnames/${opnameId}/items/${itemId}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, { opnameId }) => [
        { type: "StockOpname", id: opnameId },
        { type: "StockOpnameItem", id: "LIST" },
      ],
    }),

    /**
     * Bulk Import Products to Opname
     * POST /api/v1/stock-opnames/:opnameId/import-products
     * Imports all products from warehouse with current stock as expected qty
     */
    importWarehouseProducts: builder.mutation<
      { itemsAdded: number },
      { opnameId: string; warehouseId: string }
    >({
      query: ({ opnameId, warehouseId }) => ({
        url: `/stock-opnames/${opnameId}/import-products`,
        method: "POST",
        body: { warehouseId },
      }),
      transformResponse: (response: ApiSuccessResponse<{ itemsAdded: number }>) =>
        response.data,
      invalidatesTags: (result, error, { opnameId }) => [
        { type: "StockOpname", id: opnameId },
        { type: "StockOpnameItem", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  useListOpnamesQuery,
  useGetOpnameQuery,
  useCreateOpnameMutation,
  useUpdateOpnameMutation,
  useDeleteOpnameMutation,
  useApproveOpnameMutation,
  useAddOpnameItemMutation,
  useUpdateOpnameItemMutation,
  useDeleteOpnameItemMutation,
  useImportWarehouseProductsMutation,
} = opnameApi;
