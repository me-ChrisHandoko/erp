/**
 * Goods Receipt API Service
 *
 * RTK Query service for goods receipt management including:
 * - Goods receipt CRUD operations
 * - Status workflow transitions (receive, inspect, accept, reject)
 * - Filtering and pagination
 *
 * Backend endpoints: /api/v1/goods-receipts
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  GoodsReceiptResponse,
  GoodsReceiptListResponse,
  CreateGoodsReceiptRequest,
  UpdateGoodsReceiptRequest,
  ReceiveGoodsRequest,
  InspectGoodsRequest,
  AcceptGoodsRequest,
  RejectGoodsRequest,
  UpdateRejectionDispositionRequest,
  ResolveDispositionRequest,
  GoodsReceiptFilters,
} from "@/types/goods-receipt.types";
import type { ApiSuccessResponse } from "@/types/api";

export const goodsReceiptApi = createApi({
  reducerPath: "goodsReceiptApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["GoodsReceipt", "GoodsReceiptList"],
  endpoints: (builder) => ({
    // ==================== Goods Receipt CRUD ====================

    /**
     * List Goods Receipts with Filters & Pagination
     * GET /api/v1/goods-receipts
     */
    listGoodsReceipts: builder.query<GoodsReceiptListResponse, GoodsReceiptFilters | void>({
      query: (filters) => {
        const f = filters || {};
        const params: Record<string, unknown> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "createdAt",
          sort_order: f.sortOrder || "desc",
        };

        if (f.search) params.search = f.search;
        if (f.status) params.status = f.status;
        if (f.purchaseOrderId) params.purchase_order_id = f.purchaseOrderId;
        if (f.supplierId) params.supplier_id = f.supplierId;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/goods-receipts",
          params,
        };
      },
      serializeQueryArgs: ({ queryArgs }) => {
        return JSON.stringify(queryArgs);
      },
      forceRefetch: ({ currentArg, previousArg }) => {
        return JSON.stringify(currentArg) !== JSON.stringify(previousArg);
      },
      keepUnusedDataFor: 0,
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "GoodsReceipt" as const,
                id,
              })),
              { type: "GoodsReceiptList", id: "LIST" },
            ]
          : [{ type: "GoodsReceiptList", id: "LIST" }],
    }),

    /**
     * Get Goods Receipt Details
     * GET /api/v1/goods-receipts/:id
     */
    getGoodsReceipt: builder.query<GoodsReceiptResponse, string>({
      query: (id) => `/goods-receipts/${id}`,
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "GoodsReceipt", id }],
    }),

    /**
     * Create Goods Receipt
     * POST /api/v1/goods-receipts
     * Requires OWNER or ADMIN role
     */
    createGoodsReceipt: builder.mutation<GoodsReceiptResponse, CreateGoodsReceiptRequest>({
      query: (data) => ({
        url: "/goods-receipts",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: [{ type: "GoodsReceiptList", id: "LIST" }],
    }),

    /**
     * Update Goods Receipt
     * PUT /api/v1/goods-receipts/:id
     * Requires OWNER or ADMIN role
     */
    updateGoodsReceipt: builder.mutation<
      GoodsReceiptResponse,
      { id: string; data: UpdateGoodsReceiptRequest }
    >({
      query: ({ id, data }) => ({
        url: `/goods-receipts/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    /**
     * Delete Goods Receipt
     * DELETE /api/v1/goods-receipts/:id
     * Requires OWNER or ADMIN role
     */
    deleteGoodsReceipt: builder.mutation<void, string>({
      query: (id) => ({
        url: `/goods-receipts/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    // ==================== Status Transitions ====================

    /**
     * Receive Goods (PENDING → RECEIVED)
     * POST /api/v1/goods-receipts/:id/receive
     */
    receiveGoods: builder.mutation<
      GoodsReceiptResponse,
      { id: string; data?: ReceiveGoodsRequest }
    >({
      query: ({ id, data }) => ({
        url: `/goods-receipts/${id}/receive`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    /**
     * Inspect Goods (RECEIVED → INSPECTED)
     * POST /api/v1/goods-receipts/:id/inspect
     */
    inspectGoods: builder.mutation<
      GoodsReceiptResponse,
      { id: string; data?: InspectGoodsRequest }
    >({
      query: ({ id, data }) => ({
        url: `/goods-receipts/${id}/inspect`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    /**
     * Accept Goods (INSPECTED → ACCEPTED/PARTIAL)
     * POST /api/v1/goods-receipts/:id/accept
     */
    acceptGoods: builder.mutation<
      GoodsReceiptResponse,
      { id: string; data?: AcceptGoodsRequest }
    >({
      query: ({ id, data }) => ({
        url: `/goods-receipts/${id}/accept`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    /**
     * Reject Goods (INSPECTED → REJECTED)
     * POST /api/v1/goods-receipts/:id/reject
     */
    rejectGoods: builder.mutation<
      GoodsReceiptResponse,
      { id: string; data: RejectGoodsRequest }
    >({
      query: ({ id, data }) => ({
        url: `/goods-receipts/${id}/reject`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "GoodsReceipt", id },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    // ==================== Rejection Disposition (Odoo+M3 Model) ====================

    /**
     * Update Rejection Disposition
     * PUT /api/v1/goods-receipts/:id/items/:itemId/disposition
     * Sets what happens to rejected items
     * Requires OWNER or ADMIN role
     */
    updateRejectionDisposition: builder.mutation<
      GoodsReceiptResponse,
      { goodsReceiptId: string; itemId: string; data: UpdateRejectionDispositionRequest }
    >({
      query: ({ goodsReceiptId, itemId, data }) => ({
        url: `/goods-receipts/${goodsReceiptId}/items/${itemId}/disposition`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { goodsReceiptId }) => [
        { type: "GoodsReceipt", id: goodsReceiptId },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),

    /**
     * Resolve Disposition
     * POST /api/v1/goods-receipts/:id/items/:itemId/resolve-disposition
     * Marks disposition as resolved (e.g., replacement received, credit applied)
     * Requires OWNER or ADMIN role
     */
    resolveDisposition: builder.mutation<
      GoodsReceiptResponse,
      { goodsReceiptId: string; itemId: string; data?: ResolveDispositionRequest }
    >({
      query: ({ goodsReceiptId, itemId, data }) => ({
        url: `/goods-receipts/${goodsReceiptId}/items/${itemId}/resolve-disposition`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<GoodsReceiptResponse>) =>
        response.data,
      invalidatesTags: (result, error, { goodsReceiptId }) => [
        { type: "GoodsReceipt", id: goodsReceiptId },
        { type: "GoodsReceiptList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Goods Receipt CRUD
  useListGoodsReceiptsQuery,
  useGetGoodsReceiptQuery,
  useCreateGoodsReceiptMutation,
  useUpdateGoodsReceiptMutation,
  useDeleteGoodsReceiptMutation,
  // Status Transitions
  useReceiveGoodsMutation,
  useInspectGoodsMutation,
  useAcceptGoodsMutation,
  useRejectGoodsMutation,
  // Rejection Disposition (Odoo+M3 Model)
  useUpdateRejectionDispositionMutation,
  useResolveDispositionMutation,
} = goodsReceiptApi;
