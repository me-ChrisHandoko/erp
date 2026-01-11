/**
 * Transfer API Service
 *
 * RTK Query service for stock transfer management including:
 * - Transfer CRUD operations
 * - Status transitions (ship, receive, cancel)
 * - Transfer listing with filters and pagination
 *
 * Backend endpoints: /api/v1/stock-transfers
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  TransferResponse,
  TransferListResponse,
  CreateTransferRequest,
  UpdateTransferRequest,
  ShipTransferRequest,
  ReceiveTransferRequest,
  CancelTransferRequest,
  TransferFilters,
} from "@/types/transfer.types";
import type { ApiSuccessResponse } from "@/types/api";

export const transferApi = createApi({
  reducerPath: "transferApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Transfer", "TransferList", "Stock", "StockList"],
  endpoints: (builder) => ({
    // ==================== Transfer Listing ====================

    /**
     * List Transfers with Filters & Pagination
     * GET /api/v1/stock-transfers
     */
    listTransfers: builder.query<TransferListResponse, TransferFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "transferNumber",
          sort_order: f.sortOrder || "desc", // Latest first
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.status) params.status = f.status;
        if (f.sourceWarehouseId) params.source_warehouse_id = f.sourceWarehouseId;
        if (f.destWarehouseId) params.dest_warehouse_id = f.destWarehouseId;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/stock-transfers",
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
      transformResponse: (response: TransferListResponse) =>
        response, // Return full response with data and pagination
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({
                type: "Transfer" as const,
                id,
              })),
              { type: "TransferList", id: "LIST" },
            ]
          : [{ type: "TransferList", id: "LIST" }],
    }),

    /**
     * Get Transfer Details
     * GET /api/v1/stock-transfers/:id
     */
    getTransfer: builder.query<TransferResponse, string>({
      query: (id) => `/stock-transfers/${id}`,
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Transfer", id }],
    }),

    // ==================== Transfer CRUD ====================

    /**
     * Create Transfer (status: DRAFT)
     * POST /api/v1/stock-transfers
     */
    createTransfer: builder.mutation<TransferResponse, CreateTransferRequest>({
      query: (data) => ({
        url: "/stock-transfers",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      invalidatesTags: [{ type: "TransferList", id: "LIST" }],
    }),

    /**
     * Update Transfer (DRAFT only)
     * PUT /api/v1/stock-transfers/:id
     */
    updateTransfer: builder.mutation<
      TransferResponse,
      { id: string; data: UpdateTransferRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-transfers/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Transfer", id },
        { type: "TransferList", id: "LIST" },
      ],
    }),

    /**
     * Delete Transfer (DRAFT only)
     * DELETE /api/v1/stock-transfers/:id
     */
    deleteTransfer: builder.mutation<void, string>({
      query: (id) => ({
        url: `/stock-transfers/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Transfer", id },
        { type: "TransferList", id: "LIST" },
      ],
    }),

    // ==================== Status Transitions ====================

    /**
     * Ship Transfer (DRAFT → SHIPPED)
     * POST /api/v1/stock-transfers/:id/ship
     */
    shipTransfer: builder.mutation<
      TransferResponse,
      { id: string; data?: ShipTransferRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-transfers/${id}/ship`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Transfer", id },
        { type: "TransferList", id: "LIST" },
      ],
    }),

    /**
     * Receive Transfer (SHIPPED → RECEIVED)
     * POST /api/v1/stock-transfers/:id/receive
     *
     * IMPORTANT: This affects inventory in both warehouses
     */
    receiveTransfer: builder.mutation<
      TransferResponse,
      { id: string; data?: ReceiveTransferRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-transfers/${id}/receive`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Transfer", id },
        { type: "TransferList", id: "LIST" },
        // Invalidate stock data since inventory changed
        { type: "Stock", id: "LIST" },
        { type: "StockList", id: "LIST" },
      ],
    }),

    /**
     * Cancel Transfer (ANY → CANCELLED)
     * POST /api/v1/stock-transfers/:id/cancel
     */
    cancelTransfer: builder.mutation<
      TransferResponse,
      { id: string; data: CancelTransferRequest }
    >({
      query: ({ id, data }) => ({
        url: `/stock-transfers/${id}/cancel`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<TransferResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Transfer", id },
        { type: "TransferList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in functional components
export const {
  useListTransfersQuery,
  useGetTransferQuery,
  useCreateTransferMutation,
  useUpdateTransferMutation,
  useDeleteTransferMutation,
  useShipTransferMutation,
  useReceiveTransferMutation,
  useCancelTransferMutation,
} = transferApi;
