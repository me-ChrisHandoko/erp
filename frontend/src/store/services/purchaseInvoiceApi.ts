/**
 * Purchase Invoice API Service
 *
 * RTK Query service for purchase invoice management including:
 * - Invoice CRUD operations
 * - Workflow actions (approve, reject)
 * - Payment recording
 * - Invoice listing with filters and pagination
 *
 * Backend endpoints: /api/v1/purchase-invoices
 */

import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from './authApi';
import type {
  PurchaseInvoiceResponse,
  PurchaseInvoiceListResponse,
  CreatePurchaseInvoiceRequest,
  UpdatePurchaseInvoiceRequest,
  ApprovePurchaseInvoiceRequest,
  RejectPurchaseInvoiceRequest,
  RecordPaymentRequest,
  PurchaseInvoiceFilters,
  PurchaseInvoicePaymentResponse,
} from '@/types/purchase-invoice.types';
import type { ApiSuccessResponse } from '@/types/api';

export const purchaseInvoiceApi = createApi({
  reducerPath: 'purchaseInvoiceApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['PurchaseInvoice', 'PurchaseInvoiceList', 'InvoicePayment'],
  endpoints: (builder) => ({
    // ==================== Purchase Invoice CRUD ====================

    /**
     * List Purchase Invoices with Filters & Pagination
     * GET /api/v1/purchase-invoices
     */
    listPurchaseInvoices: builder.query<PurchaseInvoiceListResponse, PurchaseInvoiceFilters | void>({
      query: (filters) => {
        const f = filters || {};
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.page_size || 20,
          sort_by: f.sort_by || 'invoiceDate',
          sort_order: f.sort_order || 'desc',
        };

        // Add optional filters
        if (f.search) params.search = f.search;
        if (f.status) params.status = f.status;
        if (f.payment_status) params.payment_status = f.payment_status;
        if (f.supplier_id) params.supplier_id = f.supplier_id;
        if (f.date_from) params.date_from = f.date_from;
        if (f.date_to) params.date_to = f.date_to;
        if (f.due_date_from) params.due_date_from = f.due_date_from;
        if (f.due_date_to) params.due_date_to = f.due_date_to;

        return {
          url: '/purchase-invoices',
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
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceListResponse>) =>
        response.data,
      providesTags: (result) =>
        result?.data
          ? [
              ...result.data.map(({ id }) => ({ type: 'PurchaseInvoice' as const, id })),
              { type: 'PurchaseInvoiceList' as const, id: 'LIST' },
            ]
          : [{ type: 'PurchaseInvoiceList' as const, id: 'LIST' }],
    }),

    /**
     * Get Purchase Invoice Details
     * GET /api/v1/purchase-invoices/:id
     */
    getPurchaseInvoice: builder.query<PurchaseInvoiceResponse, string>({
      query: (id) => `/purchase-invoices/${id}`,
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: 'PurchaseInvoice', id }],
    }),

    /**
     * Create Purchase Invoice
     * POST /api/v1/purchase-invoices
     */
    createPurchaseInvoice: builder.mutation<PurchaseInvoiceResponse, CreatePurchaseInvoiceRequest>({
      query: (data) => ({
        url: '/purchase-invoices',
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: [{ type: 'PurchaseInvoiceList', id: 'LIST' }],
    }),

    /**
     * Update Purchase Invoice
     * PUT /api/v1/purchase-invoices/:id
     */
    updatePurchaseInvoice: builder.mutation<
      PurchaseInvoiceResponse,
      { invoiceId: string; data: UpdatePurchaseInvoiceRequest }
    >({
      query: ({ invoiceId, data }) => ({
        url: `/purchase-invoices/${invoiceId}`,
        method: 'PUT',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { invoiceId }) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
      ],
    }),

    /**
     * Delete Purchase Invoice
     * DELETE /api/v1/purchase-invoices/:id
     */
    deletePurchaseInvoice: builder.mutation<void, string>({
      query: (invoiceId) => ({
        url: `/purchase-invoices/${invoiceId}`,
        method: 'DELETE',
      }),
      invalidatesTags: [{ type: 'PurchaseInvoiceList', id: 'LIST' }],
    }),

    // ==================== Workflow Actions ====================

    /**
     * Submit Invoice for Approval
     * POST /api/v1/purchase-invoices/:id/submit
     * Changes status from DRAFT to SUBMITTED
     */
    submitInvoice: builder.mutation<PurchaseInvoiceResponse, string>({
      query: (invoiceId) => ({
        url: `/purchase-invoices/${invoiceId}/submit`,
        method: 'POST',
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, invoiceId) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
      ],
    }),

    /**
     * Approve Purchase Invoice
     * POST /api/v1/purchase-invoices/:id/approve
     */
    approvePurchaseInvoice: builder.mutation<
      PurchaseInvoiceResponse,
      { invoiceId: string; data?: ApprovePurchaseInvoiceRequest }
    >({
      query: ({ invoiceId, data }) => ({
        url: `/purchase-invoices/${invoiceId}/approve`,
        method: 'POST',
        body: data || {},
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { invoiceId }) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
      ],
    }),

    /**
     * Reject Purchase Invoice
     * POST /api/v1/purchase-invoices/:id/reject
     */
    rejectPurchaseInvoice: builder.mutation<
      PurchaseInvoiceResponse,
      { invoiceId: string; data: RejectPurchaseInvoiceRequest }
    >({
      query: ({ invoiceId, data }) => ({
        url: `/purchase-invoices/${invoiceId}/reject`,
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { invoiceId }) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
      ],
    }),

    /**
     * Cancel Invoice
     * POST /api/v1/purchase-invoices/:id/cancel
     * Changes status to CANCELLED
     */
    cancelInvoice: builder.mutation<PurchaseInvoiceResponse, string>({
      query: (invoiceId) => ({
        url: `/purchase-invoices/${invoiceId}/cancel`,
        method: 'POST',
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, invoiceId) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
      ],
    }),

    // ==================== Payment Management ====================

    /**
     * Record Payment for Purchase Invoice
     * POST /api/v1/purchase-invoices/:id/payments
     */
    recordPayment: builder.mutation<
      PurchaseInvoicePaymentResponse,
      { invoiceId: string; data: RecordPaymentRequest }
    >({
      query: ({ invoiceId, data }) => ({
        url: `/purchase-invoices/${invoiceId}/payments`,
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<PurchaseInvoicePaymentResponse>) =>
        response.data,
      invalidatesTags: (result, error, { invoiceId }) => [
        { type: 'PurchaseInvoice', id: invoiceId },
        { type: 'PurchaseInvoiceList', id: 'LIST' },
        { type: 'InvoicePayment', id: 'LIST' },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Invoice CRUD
  useListPurchaseInvoicesQuery,
  useGetPurchaseInvoiceQuery,
  useCreatePurchaseInvoiceMutation,
  useUpdatePurchaseInvoiceMutation,
  useDeletePurchaseInvoiceMutation,
  // Workflow Actions
  useSubmitInvoiceMutation,
  useApprovePurchaseInvoiceMutation,
  useRejectPurchaseInvoiceMutation,
  useCancelInvoiceMutation,
  // Payment Management
  useRecordPaymentMutation,
} = purchaseInvoiceApi;
