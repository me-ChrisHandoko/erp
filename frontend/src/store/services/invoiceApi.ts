/**
 * Invoice API Service
 *
 * RTK Query service for sales invoice management including:
 * - Invoice CRUD operations
 * - Payment recording
 * - Invoice listing with filters and pagination
 *
 * Backend endpoints: /api/v1/invoices
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  InvoiceResponse,
  InvoiceListResponse,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  RecordPaymentRequest,
  InvoiceFilters,
} from "@/types/invoice.types";
import type { ApiSuccessResponse } from "@/types/api";

export const invoiceApi = createApi({
  reducerPath: "invoiceApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Invoice", "InvoiceList"],
  endpoints: (builder) => ({
    // ==================== Invoice CRUD ====================

    /**
     * List Invoices with Filters & Pagination
     * GET /api/v1/invoices
     */
    listInvoices: builder.query<InvoiceListResponse, InvoiceFilters | void>({
      query: (filters) => {
        const f = filters || {};
        // Build params object
        const params: Record<string, any> = {
          page: f.page || 1,
          page_size: f.pageSize || 20,
          sort_by: f.sortBy || "invoiceDate",
          sort_order: f.sortOrder || "desc",
        };

        // Only add optional params if they have values
        if (f.search) params.search = f.search;
        if (f.customerId) params.customer_id = f.customerId;
        if (f.paymentStatus) params.payment_status = f.paymentStatus;
        if (f.dateFrom) params.date_from = f.dateFrom;
        if (f.dateTo) params.date_to = f.dateTo;
        if (f.dueDateFrom) params.due_date_from = f.dueDateFrom;
        if (f.dueDateTo) params.due_date_to = f.dueDateTo;

        return {
          url: "/invoices",
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
                type: "Invoice" as const,
                id,
              })),
              { type: "InvoiceList", id: "LIST" },
            ]
          : [{ type: "InvoiceList", id: "LIST" }],
    }),

    /**
     * Get Invoice Details
     * GET /api/v1/invoices/:id
     */
    getInvoice: builder.query<InvoiceResponse, string>({
      query: (id) => `/invoices/${id}`,
      transformResponse: (response: ApiSuccessResponse<InvoiceResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Invoice", id }],
    }),

    /**
     * Create Invoice
     * POST /api/v1/invoices
     * Requires OWNER or ADMIN role
     */
    createInvoice: builder.mutation<InvoiceResponse, CreateInvoiceRequest>({
      query: (data) => ({
        url: "/invoices",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<InvoiceResponse>) =>
        response.data,
      invalidatesTags: [{ type: "InvoiceList", id: "LIST" }],
    }),

    /**
     * Update Invoice
     * PUT /api/v1/invoices/:id
     * Requires OWNER or ADMIN role
     */
    updateInvoice: builder.mutation<
      InvoiceResponse,
      { id: string; data: UpdateInvoiceRequest }
    >({
      query: ({ id, data }) => ({
        url: `/invoices/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<InvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Invoice", id },
        { type: "InvoiceList", id: "LIST" },
      ],
    }),

    /**
     * Delete Invoice (Soft Delete)
     * DELETE /api/v1/invoices/:id
     * Requires OWNER or ADMIN role
     */
    deleteInvoice: builder.mutation<void, string>({
      query: (id) => ({
        url: `/invoices/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Invoice", id },
        { type: "InvoiceList", id: "LIST" },
      ],
    }),

    // ==================== Payment Management ====================

    /**
     * Record Payment
     * POST /api/v1/invoices/:id/payments
     * Requires OWNER or ADMIN role
     */
    recordPayment: builder.mutation<
      InvoiceResponse,
      { invoiceId: string; data: RecordPaymentRequest }
    >({
      query: ({ invoiceId, data }) => ({
        url: `/invoices/${invoiceId}/payments`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<InvoiceResponse>) =>
        response.data,
      invalidatesTags: (result, error, { invoiceId }) => [
        { type: "Invoice", id: invoiceId },
        { type: "InvoiceList", id: "LIST" },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Invoice CRUD
  useListInvoicesQuery,
  useGetInvoiceQuery,
  useCreateInvoiceMutation,
  useUpdateInvoiceMutation,
  useDeleteInvoiceMutation,
  // Payment Management
  useRecordPaymentMutation,
} = invoiceApi;
