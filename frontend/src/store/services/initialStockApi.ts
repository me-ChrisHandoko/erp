/**
 * Initial Stock API Service
 *
 * RTK Query service for initial stock setup operations.
 * Handles warehouse stock initialization, Excel import, and validation.
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  InitialStockSetupRequest,
  InitialStockSetupResponse,
  InitialStockImportResult,
  WarehouseStockStatus,
} from "@/types/initial-stock.types";

export const initialStockApi = createApi({
  reducerPath: "initialStockApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["InitialStock", "WarehouseStatus"],
  endpoints: (builder) => ({
    /**
     * Get Warehouse Stock Status
     * Check if warehouses have initial stock setup
     */
    getWarehouseStockStatus: builder.query<WarehouseStockStatus[], void>({
      query: () => "/warehouses/stock-status",
      providesTags: ["WarehouseStatus"],
      transformResponse: (response: { warehouses: WarehouseStockStatus[] }) => {
        return response.warehouses || [];
      },
    }),

    /**
     * Submit Initial Stock Setup
     * Create warehouse_stock records and inventory_movements
     */
    submitInitialStock: builder.mutation<
      InitialStockSetupResponse,
      InitialStockSetupRequest
    >({
      query: (data) => ({
        url: "/warehouse-stocks/initial-setup",
        method: "POST",
        body: data,
      }),
      invalidatesTags: ["InitialStock", "WarehouseStatus"],
    }),

    /**
     * Validate Excel Import
     * Validate imported data before submission
     */
    validateExcelImport: builder.mutation<
      InitialStockImportResult,
      { warehouseId: string; file: File }
    >({
      query: ({ warehouseId, file }) => {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("warehouseId", warehouseId);

        return {
          url: "/warehouse-stocks/validate-import",
          method: "POST",
          body: formData,
        };
      },
    }),

    /**
     * Download Excel Template
     * Get template for bulk import
     */
    downloadExcelTemplate: builder.query<Blob, void>({
      query: () => ({
        url: "/warehouse-stocks/import-template",
        responseHandler: (response) => response.blob(),
      }),
    }),
  }),
});

export const {
  useGetWarehouseStockStatusQuery,
  useSubmitInitialStockMutation,
  useValidateExcelImportMutation,
  useLazyDownloadExcelTemplateQuery,
} = initialStockApi;
