/**
 * Sales Payment API Service (Customer Payments)
 *
 * RTK Query service for customer payment management including:
 * - Payment CRUD operations
 * - Payment listing with filters and pagination
 * - Void payment workflow
 *
 * Backend endpoints: /api/v1/payments (Invoice Payments)
 */

import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from './authApi';
import type {
  SalesPaymentResponse,
  SalesPaymentListResponse,
  CreateSalesPaymentRequest,
  UpdateSalesPaymentRequest,
  VoidPaymentRequest,
  UpdateCheckStatusRequest,
  SalesPaymentFilters,
} from '@/types/sales-payment.types';
import type { ApiSuccessResponse } from '@/types/api';

export const salesPaymentApi = createApi({
  reducerPath: 'salesPaymentApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['SalesPayment', 'SalesPaymentList'],
  endpoints: (builder) => ({
    // ==================== Payment CRUD ====================

    /**
     * List Sales Payments with Filters & Pagination
     * GET /api/v1/payments
     */
    listSalesPayments: builder.query<SalesPaymentListResponse, SalesPaymentFilters | void>({
      query: (filters) => {
        const f = filters || {};
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || 'paymentDate',
          sort_order: f.sortOrder || 'desc',
        };

        // Add optional filters
        if (f.search) params.search = f.search;
        if (f.customerId) params.customer_id = f.customerId;
        if (f.invoiceId) params.invoice_id = f.invoiceId;
        if (f.paymentMethod) params.payment_method = f.paymentMethod;
        if (f.checkStatus) params.check_status = f.checkStatus;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;
        if (f.amountMin) params.amount_min = f.amountMin;
        if (f.amountMax) params.amount_max = f.amountMax;

        return {
          url: '/payments',
          params,
        };
      },
      // Force RTK Query to treat different filter values as different cache keys
      serializeQueryArgs: ({ queryArgs }) => {
        return JSON.stringify(queryArgs);
      },
      // Always refetch when arguments change (important for filter changes!)
      forceRefetch: ({ currentArg, previousArg }) => {
        return JSON.stringify(currentArg) !== JSON.stringify(previousArg);
      },
      // Disable caching for filter changes
      keepUnusedDataFor: 0,
      transformResponse: (response: ApiSuccessResponse<SalesPaymentListResponse>) =>
        response.data,
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({ type: 'SalesPayment' as const, id })),
              { type: 'SalesPaymentList' as const, id: 'LIST' },
            ]
          : [{ type: 'SalesPaymentList' as const, id: 'LIST' }],
    }),

    /**
     * Get Payment Details
     * GET /api/v1/payments/:id
     */
    getSalesPayment: builder.query<SalesPaymentResponse, string>({
      query: (id) => `/payments/${id}`,
      transformResponse: (response: ApiSuccessResponse<SalesPaymentResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: 'SalesPayment', id }],
    }),

    /**
     * Create Payment
     * POST /api/v1/payments
     * Requires OWNER or ADMIN role
     */
    createSalesPayment: builder.mutation<SalesPaymentResponse, CreateSalesPaymentRequest>({
      query: (data) => ({
        url: '/payments',
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesPaymentResponse>) =>
        response.data,
      invalidatesTags: [{ type: 'SalesPaymentList', id: 'LIST' }],
    }),

    /**
     * Update Payment
     * PUT /api/v1/payments/:id
     * Requires OWNER or ADMIN role
     */
    updateSalesPayment: builder.mutation<
      SalesPaymentResponse,
      { id: string; data: UpdateSalesPaymentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/payments/${id}`,
        method: 'PUT',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesPaymentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: 'SalesPayment', id },
        { type: 'SalesPaymentList', id: 'LIST' },
      ],
    }),

    /**
     * Void Payment (Delete - Same Day Only)
     * DELETE /api/v1/payments/:id
     * Requires OWNER or ADMIN role
     */
    voidSalesPayment: builder.mutation<void, { id: string; data?: VoidPaymentRequest }>({
      query: ({ id, data }) => ({
        url: `/payments/${id}`,
        method: 'DELETE',
        body: data || {},
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: 'SalesPayment', id },
        { type: 'SalesPaymentList', id: 'LIST' },
      ],
    }),

    /**
     * Update Check Status
     * PATCH /api/v1/payments/:id/check-status
     * Requires OWNER or ADMIN role
     */
    updateCheckStatus: builder.mutation<
      SalesPaymentResponse,
      { id: string; data: UpdateCheckStatusRequest }
    >({
      query: ({ id, data }) => ({
        url: `/payments/${id}/check-status`,
        method: 'PATCH',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<SalesPaymentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: 'SalesPayment', id },
        { type: 'SalesPaymentList', id: 'LIST' },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  useListSalesPaymentsQuery,
  useGetSalesPaymentQuery,
  useCreateSalesPaymentMutation,
  useUpdateSalesPaymentMutation,
  useVoidSalesPaymentMutation,
  useUpdateCheckStatusMutation,
} = salesPaymentApi;
