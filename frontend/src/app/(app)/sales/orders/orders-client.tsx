/**
 * Sales Orders Client Component
 *
 * Client-side interactive component for sales order management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, ShoppingCart } from "lucide-react";
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
import { useListSalesOrdersQuery } from "@/store/services/salesOrderApi";
import { usePermissions } from "@/hooks/use-permissions";
import { SalesOrdersTable } from "@/components/sales-orders/sales-orders-table";
import type {
  SalesOrderFilters,
  SalesOrderListResponse,
  SalesOrderStatus,
} from "@/types/sales-order.types";
import { SALES_ORDER_STATUS_LABELS } from "@/types/sales-order.types";
import type { RootState } from "@/store";

interface OrdersClientProps {
  initialData: SalesOrderListResponse;
}

export function OrdersClient({ initialData }: OrdersClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [customerFilter, setCustomerFilter] = useState<string | undefined>(
    undefined
  );
  const [statusFilter, setStatusFilter] = useState<
    SalesOrderStatus | undefined
  >(undefined);
  const [warehouseFilter, setWarehouseFilter] = useState<string | undefined>(
    undefined
  );
  const [filters, setFilters] = useState<SalesOrderFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "orderDate",
    sortOrder: "desc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateOrders = permissions.canCreate("sales-orders");
  const canEditOrders = permissions.canEdit("sales-orders");
  const canCancelOrders = permissions.canDelete("sales-orders"); // Using delete permission for cancel

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch sales orders with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    customerId: customerFilter,
    status: statusFilter,
    warehouseId: warehouseFilter,
  };

  const {
    data: ordersData,
    isLoading,
    error,
    refetch,
  } = useListSalesOrdersQuery(queryParams, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = ordersData || initialData;

  // 🔑 CRITICAL: Explicit refetch when company changes
  // Cache invalidation alone doesn't trigger refetch for skipped queries
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
        } as SalesOrderFilters;
      }
      // New column, default to descending for dates, ascending for others
      return {
        ...prev,
        sortBy: sortBy as SalesOrderFilters["sortBy"],
        sortOrder: sortBy === "orderDate" ? "desc" : "asc",
      } as SalesOrderFilters;
    });
  };

  const handleCustomerFilterChange = (customer: string | undefined) => {
    setCustomerFilter(customer);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleStatusFilterChange = (status: SalesOrderStatus | undefined) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleWarehouseFilterChange = (warehouse: string | undefined) => {
    setWarehouseFilter(warehouse);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Pesanan Penjualan
          </h1>
          <p className="text-muted-foreground">
            Kelola pesanan penjualan dari pelanggan
          </p>
        </div>
        {canCreateOrders && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/sales/orders/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Buat Pesanan
          </Button>
        )}
      </div>

      {/* Sales orders table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters - Single Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor pesanan atau nama pelanggan..."
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
                  value === "all" ? undefined : (value as SalesOrderStatus)
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                {Object.entries(SALES_ORDER_STATUS_LABELS).map(
                  ([status, label]) => (
                    <SelectItem key={status} value={status}>
                      {label}
                    </SelectItem>
                  )
                )}
              </SelectContent>
            </Select>

            {/* Customer Filter - TODO: Populate from API */}
            <Select
              value={customerFilter || "all"}
              onValueChange={(value) =>
                handleCustomerFilterChange(
                  value === "all" ? undefined : value
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Pelanggan" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Pelanggan</SelectItem>
                {/* TODO: Map from customers API */}
              </SelectContent>
            </Select>

            {/* Warehouse Filter - TODO: Populate from API */}
            <Select
              value={warehouseFilter || "all"}
              onValueChange={(value) =>
                handleWarehouseFilterChange(
                  value === "all" ? undefined : value
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Gudang" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Gudang</SelectItem>
                {/* TODO: Map from warehouses API */}
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {(customerFilter ||
              statusFilter !== undefined ||
              warehouseFilter ||
              search) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setCustomerFilter(undefined);
                  setStatusFilter(undefined);
                  setWarehouseFilter(undefined);
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
              <LoadingSpinner size="lg" text="Memuat data pesanan..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data pesanan"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 &&
              !search &&
              !customerFilter &&
              !warehouseFilter &&
              statusFilter === undefined ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <ShoppingCart className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada pesanan
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat pesanan penjualan pertama Anda
                  </p>
                  {canCreateOrders && (
                    <Button onClick={() => router.push("/sales/orders/create")}>
                      <Plus className="mr-2 h-4 w-4" />
                      Buat Pesanan
                    </Button>
                  )}
                </div>
              ) : (
                <div className="space-y-4">
                  {/* Subtle loading indicator for refetching */}
                  {isLoading && (
                    <div className="text-sm text-muted-foreground text-center py-2">
                      Memperbarui data...
                    </div>
                  )}

                  {/* Sales Orders Table */}
                  <SalesOrdersTable
                    orders={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditOrders}
                    canCancel={canCancelOrders}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-t pt-4">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const pagination = displayData.pagination as any;
                          const page = pagination.page || 1;
                          const pageSize =
                            pagination.limit || pagination.pageSize || 20;
                          const totalItems =
                            pagination.total || pagination.totalItems || 0;
                          const start = (page - 1) * pageSize + 1;
                          const end = Math.min(page * pageSize, totalItems);
                          return `Menampilkan ${start}-${end} dari ${totalItems} item`;
                        })()}
                      </div>

                      {/* 2. Page Size Selector - Baris per Halaman */}
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

                      {/* 3. Navigation Buttons - << < Halaman > >> */}
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
