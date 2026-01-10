/**
 * Stock API Service
 *
 * RTK Query service for warehouse stock management including:
 * - Stock listing with filters and pagination
 * - Stock settings updates (min/max, location)
 * - Stock alerts and monitoring
 *
 * Backend endpoints: /api/v1/warehouse-stocks
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  WarehouseStockResponse,
  WarehouseStockListResponse,
  UpdateWarehouseStockRequest,
  StockFilters,
} from "@/types/stock.types";

export const stockApi = createApi({
  reducerPath: "stockApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Stock", "StockList"],
  endpoints: (builder) => ({
    // ==================== Stock Listing ====================

    /**
     * List Warehouse Stocks with Filters & Pagination
     * GET /api/v1/warehouse-stocks
     */
    listStocks: builder.query<WarehouseStockListResponse, StockFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          pageSize: f.pageSize || 20,
          sortBy: f.sortBy || "productCode",
          sortOrder: f.sortOrder || "asc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.warehouseID) params.warehouseID = f.warehouseID;
        if (f.productID) params.productID = f.productID;

        // Explicitly handle boolean filters
        if (f.lowStock !== undefined) params.lowStock = f.lowStock;
        if (f.zeroStock !== undefined) params.zeroStock = f.zeroStock;

        return {
          url: "/warehouse-stocks",
          params,
        };
      },
      // Force RTK Query to treat different filter values as different cache keys
      serializeQueryArgs: ({ queryArgs }) => {
        return JSON.stringify(queryArgs);
      },
      // Always refetch when arguments change
      forceRefetch: ({ currentArg, previousArg }) => {
        return JSON.stringify(currentArg) !== JSON.stringify(previousArg);
      },
      // Disable caching for filter changes
      keepUnusedDataFor: 0,
      transformResponse: (response: { stocks: WarehouseStockResponse[]; totalCount: number; page: number; pageSize: number; totalPages: number }) => {
        // Backend returns WarehouseStockListResponse directly (not wrapped in ApiSuccessResponse)
        // Transform to match our expected structure
        return {
          success: true,
          data: response.stocks,
          pagination: {
            page: response.page,
            pageSize: response.pageSize,
            totalItems: response.totalCount,
            totalPages: response.totalPages,
            hasMore: response.page < response.totalPages,
          },
        };
      },
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "Stock" as const,
                id,
              })),
              { type: "StockList", id: "LIST" },
            ]
          : [{ type: "StockList", id: "LIST" }],
    }),

    /**
     * Get Stock Details
     * GET /api/v1/warehouse-stocks/:id
     * Backend returns WarehouseStockResponse directly (not wrapped)
     */
    getStock: builder.query<WarehouseStockResponse, string>({
      query: (id) => `/warehouse-stocks/${id}`,
      providesTags: (result, error, id) => [{ type: "Stock", id }],
    }),

    /**
     * Update Warehouse Stock Settings
     * PUT /api/v1/warehouse-stocks/:id
     * Requires OWNER or ADMIN role
     * Backend returns WarehouseStockResponse directly (not wrapped)
     *
     * Note: This updates stock settings (min/max, location), not actual quantity.
     * Quantity changes are done via inventory movements.
     */
    updateStockSettings: builder.mutation<
      WarehouseStockResponse,
      { id: string; data: UpdateWarehouseStockRequest }
    >({
      query: ({ id, data }) => ({
        url: `/warehouse-stocks/${id}`,
        method: "PUT",
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Stock", id },
        { type: "StockList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  useListStocksQuery,
  useGetStockQuery,
  useUpdateStockSettingsMutation,
} = stockApi;
