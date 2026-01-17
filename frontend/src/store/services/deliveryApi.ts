/**
 * Delivery API Service
 *
 * RTK Query service for delivery management including:
 * - Delivery CRUD operations
 * - Status updates and tracking
 * - Delivery listing with filters and pagination
 * - Proof of Delivery (POD) management
 *
 * Backend endpoints: /api/v1/deliveries
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  DeliveryResponse,
  DeliveryListResponse,
  CreateDeliveryPayload,
  UpdateDeliveryPayload,
  UpdateDeliveryStatusPayload,
  DeliveryFilters,
} from "@/types/delivery.types";
import type { ApiSuccessResponse } from "@/types/api";

export const deliveryApi = createApi({
  reducerPath: "deliveryApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Delivery", "DeliveryList"],
  endpoints: (builder) => ({
    // ==================== Delivery CRUD ====================

    /**
     * List Deliveries with Filters & Pagination
     * GET /api/v1/deliveries
     */
    listDeliveries: builder.query<DeliveryListResponse, DeliveryFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "deliveryDate",
          sort_order: f.sortOrder || "desc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.status) params.status = f.status;
        if (f.type) params.type = f.type;
        if (f.warehouseId) params.warehouse_id = f.warehouseId;
        if (f.customerId) params.customer_id = f.customerId;
        if (f.salesOrderId) params.sales_order_id = f.salesOrderId;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;

        return {
          url: "/deliveries",
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
                type: "Delivery" as const,
                id,
              })),
              { type: "DeliveryList" as const },
            ]
          : [{ type: "DeliveryList" as const }],
      transformResponse: (response: ApiSuccessResponse<DeliveryListResponse>) =>
        response.data,
    }),

    /**
     * Get Single Delivery
     * GET /api/v1/deliveries/:id
     */
    getDelivery: builder.query<DeliveryResponse, string>({
      query: (id) => `/deliveries/${id}`,
      providesTags: (result, error, id) => [{ type: "Delivery", id }],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Create Delivery
     * POST /api/v1/deliveries
     */
    createDelivery: builder.mutation<
      DeliveryResponse,
      CreateDeliveryPayload
    >({
      query: (delivery) => ({
        url: "/deliveries",
        method: "POST",
        body: delivery,
      }),
      invalidatesTags: [{ type: "DeliveryList" }],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Update Delivery
     * PUT /api/v1/deliveries/:id
     */
    updateDelivery: builder.mutation<
      DeliveryResponse,
      { id: string; data: UpdateDeliveryPayload }
    >({
      query: ({ id, data }) => ({
        url: `/deliveries/${id}`,
        method: "PUT",
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Update Delivery Status
     * PATCH /api/v1/deliveries/:id/status
     */
    updateDeliveryStatus: builder.mutation<
      DeliveryResponse,
      { id: string; data: UpdateDeliveryStatusPayload }
    >({
      query: ({ id, data }) => ({
        url: `/deliveries/${id}/status`,
        method: "PATCH",
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Delete Delivery
     * DELETE /api/v1/deliveries/:id
     */
    deleteDelivery: builder.mutation<void, string>({
      query: (id) => ({
        url: `/deliveries/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
    }),

    // ==================== Convenience Methods ====================

    /**
     * Start Delivery (Update status to IN_TRANSIT)
     * PATCH /api/v1/deliveries/:id/start
     */
    startDelivery: builder.mutation<
      DeliveryResponse,
      { id: string; departureTime?: string }
    >({
      query: ({ id, departureTime }) => ({
        url: `/deliveries/${id}/status`,
        method: "PATCH",
        body: {
          status: "IN_TRANSIT",
          departureTime: departureTime || new Date().toISOString(),
        },
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Complete Delivery (Update status to DELIVERED)
     * PATCH /api/v1/deliveries/:id/complete
     */
    completeDelivery: builder.mutation<
      DeliveryResponse,
      {
        id: string;
        receivedBy?: string;
        signatureUrl?: string;
        photoUrl?: string;
      }
    >({
      query: ({ id, receivedBy, signatureUrl, photoUrl }) => ({
        url: `/deliveries/${id}/status`,
        method: "PATCH",
        body: {
          status: "DELIVERED",
          receivedBy,
          receivedAt: new Date().toISOString(),
          signatureUrl,
          photoUrl,
        },
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Confirm Delivery (Update status to CONFIRMED)
     * PATCH /api/v1/deliveries/:id/confirm
     */
    confirmDelivery: builder.mutation<DeliveryResponse, string>({
      query: (id) => ({
        url: `/deliveries/${id}/status`,
        method: "PATCH",
        body: {
          status: "CONFIRMED",
        },
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),

    /**
     * Cancel Delivery
     * PATCH /api/v1/deliveries/:id/cancel
     */
    cancelDelivery: builder.mutation<
      DeliveryResponse,
      { id: string; notes?: string }
    >({
      query: ({ id, notes }) => ({
        url: `/deliveries/${id}/status`,
        method: "PATCH",
        body: {
          status: "CANCELLED",
          notes,
        },
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: "Delivery", id },
        { type: "DeliveryList" },
      ],
      transformResponse: (response: ApiSuccessResponse<DeliveryResponse>) =>
        response.data,
    }),
  }),
});

// Export hooks for usage in components
export const {
  useListDeliveriesQuery,
  useGetDeliveryQuery,
  useCreateDeliveryMutation,
  useUpdateDeliveryMutation,
  useUpdateDeliveryStatusMutation,
  useDeleteDeliveryMutation,
  useStartDeliveryMutation,
  useCompleteDeliveryMutation,
  useConfirmDeliveryMutation,
  useCancelDeliveryMutation,
} = deliveryApi;
