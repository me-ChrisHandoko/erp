/**
 * Sales Order API Service
 *
 * RTK Query service for sales order management including:
 * - Sales order CRUD operations
 * - Order status transitions
 * - Sales order listing with filters and pagination
 * - Item management
 *
 * Backend endpoints: /api/v1/sales-orders
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  SalesOrderResponse,
  SalesOrderListResponse,
  CreateSalesOrderRequest,
  UpdateSalesOrderRequest,
  SalesOrderFilters,
} from "@/types/sales-order.types";
import type { ApiSuccessResponse } from "@/types/api";

export const salesOrderApi = createApi({
  reducerPath: "salesOrderApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["SalesOrder", "SalesOrderList"],
  endpoints: (builder) => ({
    // ==================== Sales Order CRUD ====================

    /**
     * List Sales Orders with Filters & Pagination
     * GET /api/v1/sales-orders
     */
    listSalesOrders: builder.query<
      SalesOrderListResponse,
      SalesOrderFilters | void
    >({
      query: (filters) => {
        const f = filters || {};
        // Build params object, explicitly handling values
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "orderDate",
          sort_order: f.sortOrder || "desc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.customerId) params.customer_id = f.customerId;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.status) params.status = f.status;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;
        if (f.requiredDateFrom)
          params.required_date_from = f.requiredDateFrom;
        if (f.requiredDateTo) params.required_date_to = f.requiredDateTo;

        return {
          url: "/sales-orders",
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
                type: "SalesOrder" as const,
                id,
              })),
              { type: "SalesOrderList", id: "LIST" },
            ]
          : [{ type: "SalesOrderList", id: "LIST" }],
    }),

    /**
     * Get Sales Order Details
     * GET /api/v1/sales-orders/:id
     */
    getSalesOrder: builder.query<SalesOrderResponse, string>({
      query: (id) => `/sales-orders/${id}`,
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "SalesOrder", id }],
    }),

    /**
     * Create Sales Order
     * POST /api/v1/sales-orders
     * Creates order in DRAFT status
     */
    createSalesOrder: builder.mutation<
      SalesOrderResponse,
      CreateSalesOrderRequest
    >({
      query: (data) => ({
        url: "/sales-orders",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: [{ type: "SalesOrderList", id: "LIST" }],
    }),

    /**
     * Update Sales Order
     * PUT /api/v1/sales-orders/:id
     * Only allowed for DRAFT status orders
     */
    updateSalesOrder: builder.mutation<
      SalesOrderResponse,
      { id: string; data: UpdateSalesOrderRequest }
    >({
      query: ({ id, data }) => ({
        url: `/sales-orders/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Delete Sales Order (Soft Delete)
     * DELETE /api/v1/sales-orders/:id
     * Only allowed for DRAFT status orders
     */
    deleteSalesOrder: builder.mutation<void, string>({
      query: (id) => ({
        url: `/sales-orders/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    // ==================== Status Transitions ====================

    /**
     * Submit Sales Order for Approval
     * POST /api/v1/sales-orders/:id/submit
     * Transition: DRAFT → PENDING
     */
    submitSalesOrder: builder.mutation<SalesOrderResponse, string>({
      query: (id) => ({
        url: `/sales-orders/${id}/submit`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Approve Sales Order
     * POST /api/v1/sales-orders/:id/approve
     * Transition: PENDING → APPROVED
     */
    approveSalesOrder: builder.mutation<SalesOrderResponse, string>({
      query: (id) => ({
        url: `/sales-orders/${id}/approve`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Reject Sales Order (back to draft)
     * POST /api/v1/sales-orders/:id/reject
     * Transition: PENDING → DRAFT
     */
    rejectSalesOrder: builder.mutation<SalesOrderResponse, string>({
      query: (id) => ({
        url: `/sales-orders/${id}/reject`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Cancel Sales Order
     * POST /api/v1/sales-orders/:id/cancel
     * Transition: any → CANCELLED
     * Not allowed from DELIVERED or COMPLETED
     */
    cancelSalesOrder: builder.mutation<
      SalesOrderResponse,
      { id: string; reason?: string }
    >({
      query: ({ id, reason }) => ({
        url: `/sales-orders/${id}/cancel`,
        method: "POST",
        body: reason ? { reason } : undefined,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Mark Order as Processing
     * POST /api/v1/sales-orders/:id/process
     * Transition: APPROVED → PROCESSING
     */
    processSalesOrder: builder.mutation<SalesOrderResponse, string>({
      query: (id) => ({
        url: `/sales-orders/${id}/process`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Ship Sales Order
     * POST /api/v1/sales-orders/:id/ship
     * Transition: PROCESSING → SHIPPED
     */
    shipSalesOrder: builder.mutation<
      SalesOrderResponse,
      { id: string; shippingNotes?: string }
    >({
      query: ({ id, shippingNotes }) => ({
        url: `/sales-orders/${id}/ship`,
        method: "POST",
        body: shippingNotes ? { shippingNotes } : undefined,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Mark Order as Delivered
     * POST /api/v1/sales-orders/:id/deliver
     * Transition: SHIPPED → DELIVERED
     */
    deliverSalesOrder: builder.mutation<
      SalesOrderResponse,
      { id: string; deliveryNotes?: string }
    >({
      query: ({ id, deliveryNotes }) => ({
        url: `/sales-orders/${id}/deliver`,
        method: "POST",
        body: deliveryNotes ? { deliveryNotes } : undefined,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),

    /**
     * Complete Sales Order (mark as fully invoiced)
     * POST /api/v1/sales-orders/:id/complete
     * Transition: DELIVERED → COMPLETED
     */
    completeSalesOrder: builder.mutation<SalesOrderResponse, string>({
      query: (id) => ({
        url: `/sales-orders/${id}/complete`,
        method: "POST",
      }),
      transformResponse: (response: ApiSuccessResponse<SalesOrderResponse>) =>
        response.data,
      invalidatesTags: (result, error, id) => [
        { type: "SalesOrder", id },
        { type: "SalesOrderList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Sales Order CRUD
  useListSalesOrdersQuery,
  useGetSalesOrderQuery,
  useCreateSalesOrderMutation,
  useUpdateSalesOrderMutation,
  useDeleteSalesOrderMutation,
  // Status Transitions
  useSubmitSalesOrderMutation,
  useApproveSalesOrderMutation,
  useRejectSalesOrderMutation,
  useCancelSalesOrderMutation,
  useProcessSalesOrderMutation,
  useShipSalesOrderMutation,
  useDeliverSalesOrderMutation,
  useCompleteSalesOrderMutation,
} = salesOrderApi;
