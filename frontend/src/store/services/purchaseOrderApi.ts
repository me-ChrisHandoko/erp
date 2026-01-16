/**
 * Purchase Order API Service
 *
 * RTK Query service for purchase order management including:
 * - Purchase Order CRUD operations
 * - Order item management
 * - Status workflow (confirm, complete, cancel)
 * - Purchase order listing with filters and pagination
 *
 * Backend endpoints: /api/v1/purchase-orders
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  PurchaseOrderResponse,
  PurchaseOrderListResponse,
  CreatePurchaseOrderRequest,
  UpdatePurchaseOrderRequest,
  ConfirmPurchaseOrderRequest,
  CancelPurchaseOrderRequest,
  PurchaseOrderFilters,
} from "@/types/purchase-order.types";
import type { ApiSuccessResponse } from "@/types/api";

export const purchaseOrderApi = createApi({
  reducerPath: "purchaseOrderApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["PurchaseOrder", "PurchaseOrderList"],
  endpoints: (builder) => ({
    // ==================== Purchase Order CRUD ====================

    /**
     * List Purchase Orders with Filters & Pagination
     * GET /api/v1/purchase-orders
     */
    listPurchaseOrders: builder.query<PurchaseOrderListResponse, PurchaseOrderFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "poDate",
          sort_order: f.sortOrder || "desc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.supplierId) params.supplier_id = f.supplierId;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.status) params.status = f.status;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/purchase-orders",
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
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "PurchaseOrder" as const,
                id,
              })),
              { type: "PurchaseOrderList", id: "LIST" },
            ]
          : [{ type: "PurchaseOrderList", id: "LIST" }],
    }),

    /**
     * Get Purchase Order Details
     * GET /api/v1/purchase-orders/:id
     */
    getPurchaseOrder: builder.query<PurchaseOrderResponse, string>({
      query: (id) => `/purchase-orders/${id}`,
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "PurchaseOrder", id }],
    }),

    /**
     * Create Purchase Order
     * POST /api/v1/purchase-orders
     * Requires OWNER, ADMIN, or WAREHOUSE role
     */
    createPurchaseOrder: builder.mutation<PurchaseOrderResponse, CreatePurchaseOrderRequest>({
      query: (data) => ({
        url: "/purchase-orders",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      invalidatesTags: [{ type: "PurchaseOrderList", id: "LIST" }],
    }),

    /**
     * Update Purchase Order
     * PUT /api/v1/purchase-orders/:id
     * Only allowed for DRAFT status
     * Requires OWNER, ADMIN, or WAREHOUSE role
     */
    updatePurchaseOrder: builder.mutation<
      PurchaseOrderResponse,
      { id: string; data: UpdatePurchaseOrderRequest }
    >({
      query: ({ id, data }) => ({
        url: `/purchase-orders/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "PurchaseOrder", id },
        { type: "PurchaseOrderList", id: "LIST" },
      ],
    }),

    /**
     * Delete Purchase Order (Soft Delete)
     * DELETE /api/v1/purchase-orders/:id
     * Only allowed for DRAFT status
     * Requires OWNER or ADMIN role
     */
    deletePurchaseOrder: builder.mutation<void, string>({
      query: (id) => ({
        url: `/purchase-orders/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "PurchaseOrder", id },
        { type: "PurchaseOrderList", id: "LIST" },
      ],
    }),

    // ==================== Status Workflow ====================

    /**
     * Confirm Purchase Order
     * POST /api/v1/purchase-orders/:id/confirm
     * Changes status from DRAFT to CONFIRMED
     * Requires OWNER, ADMIN role
     */
    confirmPurchaseOrder: builder.mutation<
      PurchaseOrderResponse,
      { id: string; data?: ConfirmPurchaseOrderRequest }
    >({
      query: ({ id, data }) => ({
        url: `/purchase-orders/${id}/confirm`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "PurchaseOrder", id },
        { type: "PurchaseOrderList", id: "LIST" },
      ],
    }),

    /**
     * Complete Purchase Order
     * POST /api/v1/purchase-orders/:id/complete
     * Changes status from CONFIRMED to COMPLETED
     * Should be called after all goods are received
     * Requires OWNER, ADMIN, or WAREHOUSE role
     */
    completePurchaseOrder: builder.mutation<PurchaseOrderResponse, string>({
      query: (id) => ({
        url: `/purchase-orders/${id}/complete`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "PurchaseOrder", id },
        { type: "PurchaseOrderList", id: "LIST" },
      ],
    }),

    /**
     * Cancel Purchase Order
     * POST /api/v1/purchase-orders/:id/cancel
     * Changes status to CANCELLED
     * Requires OWNER or ADMIN role
     */
    cancelPurchaseOrder: builder.mutation<
      PurchaseOrderResponse,
      { id: string; data: CancelPurchaseOrderRequest }
    >({
      query: ({ id, data }) => ({
        url: `/purchase-orders/${id}/cancel`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "PurchaseOrder", id },
        { type: "PurchaseOrderList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Purchase Order CRUD
  useListPurchaseOrdersQuery,
  useGetPurchaseOrderQuery,
  useCreatePurchaseOrderMutation,
  useUpdatePurchaseOrderMutation,
  useDeletePurchaseOrderMutation,
  // Status Workflow
  useConfirmPurchaseOrderMutation,
  useCompletePurchaseOrderMutation,
  useCancelPurchaseOrderMutation,
} = purchaseOrderApi;
