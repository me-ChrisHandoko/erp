/**
 * Adjustment API Service
 *
 * RTK Query service for inventory adjustment management including:
 * - Adjustment CRUD operations
 * - Status transitions (approve, cancel)
 * - Adjustment listing with filters and pagination
 *
 * Backend endpoints: /api/v1/inventory-adjustments
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  AdjustmentResponse,
  AdjustmentListResponse,
  CreateAdjustmentRequest,
  UpdateAdjustmentRequest,
  ApproveAdjustmentRequest,
  CancelAdjustmentRequest,
  AdjustmentFilters,
} from "@/types/adjustment.types";
import type { ApiSuccessResponse } from "@/types/api";
import type { AuditLog } from "@/types/audit";

// Response type for audit logs by entity
interface AuditLogsByEntityResponse {
  success: boolean;
  data: AuditLog[];
}

export const adjustmentApi = createApi({
  reducerPath: "adjustmentApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Adjustment", "AdjustmentList", "Stock", "StockList", "AdjustmentAuditLog"],
  endpoints: (builder) => ({
    // ==================== Adjustment Listing ====================

    /**
     * List Adjustments with Filters & Pagination
     * GET /api/v1/inventory-adjustments
     */
    listAdjustments: builder.query<AdjustmentListResponse, AdjustmentFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "adjustmentNumber",
          sort_order: f.sortOrder || "desc", // Latest first
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.status) params.status = f.status;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.adjustmentType) params.adjustment_type = f.adjustmentType;
        if (f.reason) params.reason = f.reason;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/inventory-adjustments",
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
      transformResponse: (response: AdjustmentListResponse) =>
        response, // Return full response with data and pagination
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({
                type: "Adjustment" as const,
                id,
              })),
              { type: "AdjustmentList", id: "LIST" },
            ]
          : [{ type: "AdjustmentList", id: "LIST" }],
    }),

    /**
     * Get Adjustment Details
     * GET /api/v1/inventory-adjustments/:id
     */
    getAdjustment: builder.query<AdjustmentResponse, string>({
      query: (id) => `/inventory-adjustments/${id}`,
      transformResponse: (response: ApiSuccessResponse<AdjustmentResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Adjustment", id }],
    }),

    // ==================== Adjustment CRUD ====================

    /**
     * Create Adjustment (status: DRAFT)
     * POST /api/v1/inventory-adjustments
     */
    createAdjustment: builder.mutation<AdjustmentResponse, CreateAdjustmentRequest>({
      query: (data) => ({
        url: "/inventory-adjustments",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<AdjustmentResponse>) =>
        response.data,
      invalidatesTags: [{ type: "AdjustmentList", id: "LIST" }],
    }),

    /**
     * Update Adjustment (DRAFT only)
     * PUT /api/v1/inventory-adjustments/:id
     */
    updateAdjustment: builder.mutation<
      AdjustmentResponse,
      { id: string; data: UpdateAdjustmentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/inventory-adjustments/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<AdjustmentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Adjustment", id },
        { type: "AdjustmentList", id: "LIST" },
      ],
    }),

    /**
     * Delete Adjustment (DRAFT only)
     * DELETE /api/v1/inventory-adjustments/:id
     */
    deleteAdjustment: builder.mutation<void, string>({
      query: (id) => ({
        url: `/inventory-adjustments/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Adjustment", id },
        { type: "AdjustmentList", id: "LIST" },
      ],
    }),

    // ==================== Status Transitions ====================

    /**
     * Approve Adjustment (DRAFT → APPROVED)
     * POST /api/v1/inventory-adjustments/:id/approve
     *
     * IMPORTANT: This affects inventory in the warehouse
     */
    approveAdjustment: builder.mutation<
      AdjustmentResponse,
      { id: string; data?: ApproveAdjustmentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/inventory-adjustments/${id}/approve`,
        method: "POST",
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<AdjustmentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Adjustment", id },
        { type: "AdjustmentList", id: "LIST" },
        // Invalidate stock data since inventory changed
        { type: "Stock", id: "LIST" },
        { type: "StockList", id: "LIST" },
      ],
    }),

    /**
     * Cancel Adjustment (DRAFT → CANCELLED)
     * POST /api/v1/inventory-adjustments/:id/cancel
     */
    cancelAdjustment: builder.mutation<
      AdjustmentResponse,
      { id: string; data: CancelAdjustmentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/inventory-adjustments/${id}/cancel`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<AdjustmentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Adjustment", id },
        { type: "AdjustmentList", id: "LIST" },
      ],
    }),

    // ==================== Audit Logs ====================

    /**
     * Get Adjustment Audit Logs
     * GET /api/v1/audit-logs/inventory_adjustment/:adjustmentId
     * Returns audit trail for a specific inventory adjustment
     */
    getAdjustmentAuditLogs: builder.query<
      AuditLog[],
      { adjustmentId: string; limit?: number; offset?: number }
    >({
      query: ({ adjustmentId, limit = 50, offset = 0 }) => ({
        url: `/audit-logs/inventory_adjustment/${adjustmentId}`,
        params: { limit, offset },
      }),
      transformResponse: (response: AuditLogsByEntityResponse) => response.data,
      providesTags: (result, error, { adjustmentId }) => [
        { type: "AdjustmentAuditLog", id: adjustmentId },
      ],
    }),
  }),
});

// Export hooks for usage in functional components
export const {
  useListAdjustmentsQuery,
  useGetAdjustmentQuery,
  useCreateAdjustmentMutation,
  useUpdateAdjustmentMutation,
  useDeleteAdjustmentMutation,
  useApproveAdjustmentMutation,
  useCancelAdjustmentMutation,
  useGetAdjustmentAuditLogsQuery,
} = adjustmentApi;
