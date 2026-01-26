/**
 * Delivery Tolerance API Service
 *
 * RTK Query service for delivery tolerance settings (SAP Model):
 * - Hierarchical tolerance configuration (Company > Category > Product)
 * - CRUD operations for tolerance settings
 * - Effective tolerance lookup for products
 *
 * Backend endpoints: /api/v1/delivery-tolerances
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  DeliveryToleranceResponse,
  DeliveryToleranceListResponse,
  EffectiveToleranceResponse,
  CreateDeliveryToleranceRequest,
  UpdateDeliveryToleranceRequest,
  DeliveryToleranceFilters,
} from "@/types/delivery-tolerance.types";
import type { ApiSuccessResponse } from "@/types/api";

export const deliveryToleranceApi = createApi({
  reducerPath: "deliveryToleranceApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["DeliveryTolerance", "DeliveryToleranceList", "EffectiveTolerance"],
  endpoints: (builder) => ({
    // ==================== Delivery Tolerance CRUD ====================

    /**
     * List Delivery Tolerances with Filters & Pagination
     * GET /api/v1/delivery-tolerances
     */
    listDeliveryTolerances: builder.query<
      DeliveryToleranceListResponse,
      DeliveryToleranceFilters | void
    >({
      query: (filters) => {
        const f = filters || {};
        const params: Record<string, unknown> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "level",
          sort_order: f.sortOrder || "asc",
        };

        if (f.level) params.level = f.level;
        if (f.categoryName) params.category_name = f.categoryName;
        if (f.productId) params.product_id = f.productId;
        if (f.isActive !== undefined) params.is_active = f.isActive;

        return {
          url: "/delivery-tolerances",
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
                type: "DeliveryTolerance" as const,
                id,
              })),
              { type: "DeliveryToleranceList", id: "LIST" },
            ]
          : [{ type: "DeliveryToleranceList", id: "LIST" }],
    }),

    /**
     * Get Delivery Tolerance Details
     * GET /api/v1/delivery-tolerances/:id
     */
    getDeliveryTolerance: builder.query<DeliveryToleranceResponse, string>({
      query: (id) => `/delivery-tolerances/${id}`,
      transformResponse: (response: ApiSuccessResponse<DeliveryToleranceResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "DeliveryTolerance", id }],
    }),

    /**
     * Create Delivery Tolerance
     * POST /api/v1/delivery-tolerances
     * Requires OWNER or ADMIN role
     */
    createDeliveryTolerance: builder.mutation<
      DeliveryToleranceResponse,
      CreateDeliveryToleranceRequest
    >({
      query: (data) => ({
        url: "/delivery-tolerances",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<DeliveryToleranceResponse>) =>
        response.data,
      invalidatesTags: [
        { type: "DeliveryToleranceList", id: "LIST" },
        { type: "EffectiveTolerance", id: "LIST" },
      ],
    }),

    /**
     * Update Delivery Tolerance
     * PUT /api/v1/delivery-tolerances/:id
     * Requires OWNER or ADMIN role
     */
    updateDeliveryTolerance: builder.mutation<
      DeliveryToleranceResponse,
      { id: string; data: UpdateDeliveryToleranceRequest }
    >({
      query: ({ id, data }) => ({
        url: `/delivery-tolerances/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<DeliveryToleranceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "DeliveryTolerance", id },
        { type: "DeliveryToleranceList", id: "LIST" },
        { type: "EffectiveTolerance", id: "LIST" },
      ],
    }),

    /**
     * Delete Delivery Tolerance
     * DELETE /api/v1/delivery-tolerances/:id
     * Requires OWNER or ADMIN role
     */
    deleteDeliveryTolerance: builder.mutation<void, string>({
      query: (id) => ({
        url: `/delivery-tolerances/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "DeliveryTolerance", id },
        { type: "DeliveryToleranceList", id: "LIST" },
        { type: "EffectiveTolerance", id: "LIST" },
      ],
    }),

    // ==================== Effective Tolerance ====================

    /**
     * Get Effective Tolerance for a Product
     * GET /api/v1/delivery-tolerances/effective?product_id=xxx
     * Resolves hierarchy: Product > Category > Company > Default
     */
    getEffectiveTolerance: builder.query<EffectiveToleranceResponse, string>({
      query: (productId) => ({
        url: "/delivery-tolerances/effective",
        params: { product_id: productId },
      }),
      transformResponse: (response: ApiSuccessResponse<EffectiveToleranceResponse>) =>
        response.data,
      providesTags: (result, error, productId) => [
        { type: "EffectiveTolerance", id: productId },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Delivery Tolerance CRUD
  useListDeliveryTolerancesQuery,
  useGetDeliveryToleranceQuery,
  useCreateDeliveryToleranceMutation,
  useUpdateDeliveryToleranceMutation,
  useDeleteDeliveryToleranceMutation,
  // Effective Tolerance
  useGetEffectiveToleranceQuery,
  useLazyGetEffectiveToleranceQuery,
} = deliveryToleranceApi;
