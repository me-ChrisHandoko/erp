/**
 * Payment API Service
 *
 * RTK Query service for supplier payment management including:
 * - Payment CRUD operations
 * - Payment listing with filters and pagination
 * - Approval workflow
 *
 * Backend endpoints: /api/v1/supplier-payments
 */

import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from './authApi';
import type {
  PaymentResponse,
  PaymentListResponse,
  CreatePaymentRequest,
  UpdatePaymentRequest,
  ApprovePaymentRequest,
  PaymentFilters,
} from '@/types/payment.types';
import type { ApiSuccessResponse } from '@/types/api';

export const paymentApi = createApi({
  reducerPath: 'paymentApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Payment', 'PaymentList'],
  endpoints: (builder) => ({
    // ==================== Payment CRUD ====================

    /**
     * List Payments with Filters & Pagination
     * GET /api/v1/supplier-payments
     */
    listPayments: builder.query<PaymentListResponse, PaymentFilters | void>({
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
        if (f.supplierId) params.supplier_id = f.supplierId;
        if (f.paymentMethod) params.payment_method = f.paymentMethod;
        if (f.status) params.status = f.status;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;
        if (f.amountMin) params.amount_min = f.amountMin;
        if (f.amountMax) params.amount_max = f.amountMax;

        return {
          url: '/supplier-payments',
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
      transformResponse: (response: ApiSuccessResponse<PaymentListResponse>) =>
        response.data,
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({ type: 'Payment' as const, id })),
              { type: 'PaymentList' as const, id: 'LIST' },
            ]
          : [{ type: 'PaymentList' as const, id: 'LIST' }],
    }),

    /**
     * Get Payment Details
     * GET /api/v1/supplier-payments/:id
     */
    getPayment: builder.query<PaymentResponse, string>({
      query: (id) => `/supplier-payments/${id}`,
      transformResponse: (response: ApiSuccessResponse<PaymentResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: 'Payment', id }],
    }),

    /**
     * Create Payment
     * POST /api/v1/supplier-payments
     * Requires OWNER or ADMIN role
     */
    createPayment: builder.mutation<PaymentResponse, CreatePaymentRequest>({
      query: (data) => ({
        url: '/supplier-payments',
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PaymentResponse>) =>
        response.data,
      invalidatesTags: [{ type: 'PaymentList', id: 'LIST' }],
    }),

    /**
     * Update Payment
     * PUT /api/v1/supplier-payments/:id
     * Requires OWNER or ADMIN role
     */
    updatePayment: builder.mutation<
      PaymentResponse,
      { id: string; data: UpdatePaymentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/supplier-payments/${id}`,
        method: 'PUT',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PaymentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: 'Payment', id },
        { type: 'PaymentList', id: 'LIST' },
      ],
    }),

    /**
     * Delete Payment (Soft Delete)
     * DELETE /api/v1/supplier-payments/:id
     * Requires OWNER or ADMIN role
     */
    deletePayment: builder.mutation<void, string>({
      query: (id) => ({
        url: `/supplier-payments/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'Payment', id },
        { type: 'PaymentList', id: 'LIST' },
      ],
    }),

    /**
     * Approve Payment
     * POST /api/v1/supplier-payments/:id/approve
     * Requires OWNER or ADMIN role
     */
    approvePayment: builder.mutation<
      PaymentResponse,
      { id: string; data?: ApprovePaymentRequest }
    >({
      query: ({ id, data }) => ({
        url: `/supplier-payments/${id}/approve`,
        method: 'POST',
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<PaymentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: 'Payment', id },
        { type: 'PaymentList', id: 'LIST' },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  useListPaymentsQuery,
  useGetPaymentQuery,
  useCreatePaymentMutation,
  useUpdatePaymentMutation,
  useDeletePaymentMutation,
  useApprovePaymentMutation,
} = paymentApi;
