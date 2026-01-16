/**
 * Goods Receipts Client Component
 *
 * Client-side interactive component for goods receipt management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { Search, PackageCheck, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useListGoodsReceiptsQuery } from "@/store/services/goodsReceiptApi";
import { GoodsReceiptsTable } from "@/components/goods-receipts/goods-receipts-table";
import {
  GOODS_RECEIPT_STATUS_OPTIONS,
  type GoodsReceiptFilters,
  type GoodsReceiptListResponse,
  type GoodsReceiptStatus,
} from "@/types/goods-receipt.types";
import type { RootState } from "@/store";

interface ReceiptsClientProps {
  initialData: GoodsReceiptListResponse;
}

export function ReceiptsClient({ initialData }: ReceiptsClientProps) {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<GoodsReceiptStatus | undefined>(
    undefined
  );
  const [filters, setFilters] = useState<GoodsReceiptFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "createdAt",
    sortOrder: "desc",
  });

  // Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch goods receipts with filters
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    status: statusFilter,
  };

  const {
    data: receiptsData,
    isLoading,
    error,
    refetch,
  } = useListGoodsReceiptsQuery(queryParams, {
    // Skip query until company context is available
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = receiptsData || initialData;

  // Explicit refetch when company changes
  useEffect(() => {
    if (activeCompanyId) {
      refetch();
    }
  }, [activeCompanyId, refetch]);

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
  };

  const handlePageSizeChange = (newPageSize: string) => {
    setFilters((prev) => ({
      ...prev,
      pageSize: parseInt(newPageSize),
      page: 1, // Reset to page 1 when changing page size
    }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as GoodsReceiptFilters;
      }
      // New column, default to descending for dates
      return {
        ...prev,
        sortBy: sortBy as GoodsReceiptFilters["sortBy"],
        sortOrder: sortBy === "createdAt" || sortBy === "grnDate" ? "desc" : "asc",
      } as GoodsReceiptFilters;
    });
  };

  const handleStatusFilterChange = (status: GoodsReceiptStatus | undefined) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Penerimaan Barang
          </h1>
          <p className="text-muted-foreground">
            Kelola penerimaan barang dari Purchase Order
          </p>
        </div>
      </div>

      {/* Goods Receipts table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor GRN atau nomor PO..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Status Filter */}
            <Select
              value={statusFilter || "all"}
              onValueChange={(value) =>
                handleStatusFilterChange(
                  value === "all" ? undefined : (value as GoodsReceiptStatus)
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                {GOODS_RECEIPT_STATUS_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {(statusFilter || search) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setStatusFilter(undefined);
                  setFilters((prev) => ({ ...prev, page: 1 }));
                }}
                className="w-full sm:w-auto"
              >
                Reset
              </Button>
            )}
          </div>

          {/* Loading State */}
          {isLoading && !displayData && (
            <div className="py-12">
              <LoadingSpinner size="lg" text="Memuat data penerimaan barang..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data penerimaan barang"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 && !search && !statusFilter ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <PackageCheck className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada penerimaan barang
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Penerimaan barang akan muncul setelah ada Purchase Order yang diproses
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {/* Subtle loading indicator for refetching */}
                  {isLoading && (
                    <div className="text-sm text-muted-foreground text-center py-2">
                      Memperbarui data...
                    </div>
                  )}

                  {/* Goods Receipts Table */}
                  <GoodsReceiptsTable
                    receipts={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-t pt-4">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const pagination = displayData.pagination;
                          const page = pagination.page || 1;
                          const pageSize = pagination.limit || 20;
                          const totalItems = pagination.total || 0;
                          const start = (page - 1) * pageSize + 1;
                          const end = Math.min(page * pageSize, totalItems);
                          return `Menampilkan ${start}-${end} dari ${totalItems} item`;
                        })()}
                      </div>

                      {/* 2. Page Size Selector */}
                      <div className="flex items-center justify-center sm:justify-start gap-2">
                        <span className="text-sm text-muted-foreground whitespace-nowrap">
                          Baris per Halaman
                        </span>
                        <Select
                          value={filters.pageSize?.toString() || "20"}
                          onValueChange={handlePageSizeChange}
                        >
                          <SelectTrigger className="w-[70px] h-8">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="10">10</SelectItem>
                            <SelectItem value="20">20</SelectItem>
                            <SelectItem value="50">50</SelectItem>
                            <SelectItem value="100">100</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      {/* 3. Navigation Buttons */}
                      <div className="flex items-center justify-center sm:justify-end gap-2">
                        {/* First Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePageChange(1)}
                          disabled={displayData.pagination.page === 1}
                        >
                          &laquo;
                        </Button>

                        {/* Previous Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(displayData.pagination.page - 1)
                          }
                          disabled={displayData.pagination.page === 1}
                        >
                          &lsaquo;
                        </Button>

                        {/* Current Page Info */}
                        <span className="text-sm text-muted-foreground px-2">
                          Halaman {displayData.pagination.page} dari{" "}
                          {displayData.pagination.totalPages}
                        </span>

                        {/* Next Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(displayData.pagination.page + 1)
                          }
                          disabled={
                            displayData.pagination.page >=
                            displayData.pagination.totalPages
                          }
                        >
                          &rsaquo;
                        </Button>

                        {/* Last Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(displayData.pagination.totalPages)
                          }
                          disabled={
                            displayData.pagination.page >=
                            displayData.pagination.totalPages
                          }
                        >
                          &raquo;
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
