/**
 * Stock Client Component
 *
 * Client-side interactive component for warehouse stock management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Stock status visualization and alerts
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { Search, Package, AlertTriangle } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useListStocksQuery } from "@/store/services/stockApi";
import { usePermissions } from "@/hooks/use-permissions";
import { StockTable } from "@/components/stock/stock-table";
import type { StockFilters, WarehouseStockListResponse } from "@/types/stock.types";
import type { RootState } from "@/store";

interface StockClientProps {
  initialData: WarehouseStockListResponse;
}

export function StockClient({ initialData }: StockClientProps) {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [stockFilter, setStockFilter] = useState<string>("all");
  const [filters, setFilters] = useState<StockFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "productCode",
    sortOrder: "asc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission check - users with OWNER or ADMIN can edit stock settings
  const canEditStock = permissions.canEdit('warehouse-stocks') ||
                       permissions.canEdit('warehouses') ||
                       permissions.role === 'OWNER' ||
                       permissions.role === 'ADMIN';

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Build query params based on filters
  const queryParams: StockFilters = {
    ...filters,
    search: debouncedSearch || undefined,
    lowStock: stockFilter === "lowStock" ? true : undefined,
    zeroStock: stockFilter === "zeroStock" ? true : undefined,
  };

  // Fetch stocks with filters
  const {
    data: stocksData,
    isLoading,
    error,
    refetch,
  } = useListStocksQuery(queryParams, {
    // Skip query until company context is available
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = stocksData || initialData;

  // 🔑 CRITICAL: Explicit refetch when company changes
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
      page: 1,
    }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as StockFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as StockFilters["sortBy"],
        sortOrder: "asc",
      } as StockFilters;
    });
  };

  const handleStockFilterChange = (filter: string) => {
    setStockFilter(filter);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // Calculate stock statistics
  const stockStats = displayData?.data
    ? {
        total: displayData.data.length,
        low: displayData.data.filter((s) =>
          Number(s.quantity) < Number(s.minimumStock)
        ).length,
        zero: displayData.data.filter((s) => Number(s.quantity) === 0).length,
      }
    : { total: 0, low: 0, zero: 0 };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and stats */}
      <div className="flex flex-col gap-4">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Stok Gudang</h1>
          <p className="text-muted-foreground">
            Kelola stok produk di seluruh gudang
          </p>
        </div>

        {/* Stock Statistics */}
        {displayData && displayData.data.length > 0 && (
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <Package className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">
                    Total Item
                  </span>
                </div>
                <p className="text-2xl font-bold mt-2">
                  {displayData.pagination.totalItems}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-yellow-500" />
                  <span className="text-sm text-muted-foreground">
                    Stok Rendah
                  </span>
                </div>
                <p className="text-2xl font-bold mt-2 text-yellow-500">
                  {stockStats.low}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-red-500" />
                  <span className="text-sm text-muted-foreground">
                    Stok Habis
                  </span>
                </div>
                <p className="text-2xl font-bold mt-2 text-red-500">
                  {stockStats.zero}
                </p>
              </CardContent>
            </Card>
          </div>
        )}
      </div>

      {/* Stock table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari kode atau nama produk..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Stock Filter */}
            <Select
              value={stockFilter}
              onValueChange={handleStockFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Filter Stok" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Stok</SelectItem>
                <SelectItem value="lowStock">Stok Rendah</SelectItem>
                <SelectItem value="zeroStock">Stok Habis</SelectItem>
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {(search || stockFilter !== "all") && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setStockFilter("all");
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
              <LoadingSpinner size="lg" text="Memuat data stok..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data stok"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 && !search && stockFilter === "all" ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <Package className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada data stok
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Data stok akan muncul setelah produk ditambahkan ke gudang
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

                  {/* Stock Table */}
                  <StockTable
                    stocks={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditStock}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex items-center justify-between border-t pt-4">
                      <div className="text-sm text-muted-foreground">
                        {(() => {
                          const { page, pageSize, totalItems } = displayData.pagination;
                          const start = (page - 1) * pageSize + 1;
                          const end = Math.min(page * pageSize, totalItems);
                          return `Menampilkan ${start}-${end} dari ${totalItems} item`;
                        })()}
                      </div>
                      <div className="flex items-center gap-2">
                        {/* Page Size Selector */}
                        <div className="flex items-center gap-2">
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
