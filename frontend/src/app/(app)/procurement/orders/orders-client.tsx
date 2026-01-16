/**
 * Purchase Orders Client Component
 *
 * Client-side interactive component for purchase order management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, ShoppingCart, Filter } from "lucide-react";
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
import { useListPurchaseOrdersQuery } from "@/store/services/purchaseOrderApi";
import { useListSuppliersQuery } from "@/store/services/supplierApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { usePermissions } from "@/hooks/use-permissions";
import { OrdersTable } from "@/components/procurement/orders-table";
import type {
  PurchaseOrderFilters,
  PurchaseOrderListResponse,
  PurchaseOrderStatus,
} from "@/types/purchase-order.types";
import { PURCHASE_ORDER_STATUS_OPTIONS } from "@/types/purchase-order.types";
import type { RootState } from "@/store";

interface OrdersClientProps {
  initialData: PurchaseOrderListResponse;
}

export function OrdersClient({ initialData }: OrdersClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<PurchaseOrderStatus | undefined>(
    undefined
  );
  const [supplierFilter, setSupplierFilter] = useState<string | undefined>(undefined);
  const [warehouseFilter, setWarehouseFilter] = useState<string | undefined>(undefined);
  const [filters, setFilters] = useState<PurchaseOrderFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "poDate",
    sortOrder: "desc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateOrders = permissions.canCreate('purchase-orders');
  const canEditOrders = permissions.canEdit('purchase-orders');
  const canConfirmOrders = permissions.canEdit('purchase-orders'); // Same as edit
  const canCancelOrders = permissions.canDelete('purchase-orders');

  // Fetch suppliers and warehouses for filter dropdowns
  const { data: suppliersData } = useListSuppliersQuery({ isActive: true, pageSize: 100 });
  const { data: warehousesData } = useListWarehousesQuery({ isActive: true, pageSize: 100 });

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch purchase orders with filters
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    status: statusFilter,
    supplierId: supplierFilter,
    warehouseId: warehouseFilter,
  };

  const {
    data: ordersData,
    isLoading,
    error,
    refetch,
  } = useListPurchaseOrdersQuery(queryParams, {
    // Skip query until company context is available
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = ordersData || initialData;

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
        } as PurchaseOrderFilters;
      }
      // New column, default to descending for dates/amounts
      return {
        ...prev,
        sortBy: sortBy as PurchaseOrderFilters["sortBy"],
        sortOrder: sortBy === "poDate" || sortBy === "totalAmount" ? "desc" : "asc",
      } as PurchaseOrderFilters;
    });
  };

  const handleStatusFilterChange = (status: string | undefined) => {
    setStatusFilter(
      status === "all" ? undefined : (status as PurchaseOrderStatus)
    );
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleSupplierFilterChange = (supplierId: string | undefined) => {
    setSupplierFilter(supplierId === "all" ? undefined : supplierId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleWarehouseFilterChange = (warehouseId: string | undefined) => {
    setWarehouseFilter(warehouseId === "all" ? undefined : warehouseId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const clearFilters = () => {
    setSearch("");
    setDebouncedSearch("");
    setStatusFilter(undefined);
    setSupplierFilter(undefined);
    setWarehouseFilter(undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const hasActiveFilters =
    statusFilter || supplierFilter || warehouseFilter || search;

  const suppliers = suppliersData?.data || [];
  const warehouses = warehousesData?.data || [];

  const handleCreateClick = () => {
    router.push("/procurement/orders/create");
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Purchase Orders</h1>
          <p className="text-muted-foreground">
            Kelola purchase order untuk pembelian barang dari supplier
          </p>
        </div>
        {canCreateOrders && (
          <Button className="shrink-0" onClick={handleCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Buat PO Baru
          </Button>
        )}
      </div>

      {/* Orders table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center flex-wrap">
            {/* Search */}
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor PO..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Status Filter */}
            <Select
              value={statusFilter || "all"}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[150px]">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                {PURCHASE_ORDER_STATUS_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Supplier Filter */}
            <Select
              value={supplierFilter || "all"}
              onValueChange={handleSupplierFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Supplier" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Supplier</SelectItem>
                {suppliers.map((supplier) => (
                  <SelectItem key={supplier.id} value={supplier.id}>
                    {supplier.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Warehouse Filter */}
            <Select
              value={warehouseFilter || "all"}
              onValueChange={handleWarehouseFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Gudang" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Gudang</SelectItem>
                {warehouses.map((warehouse) => (
                  <SelectItem key={warehouse.id} value={warehouse.id}>
                    {warehouse.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {hasActiveFilters && (
              <Button
                variant="outline"
                size="sm"
                onClick={clearFilters}
                className="w-full sm:w-auto"
              >
                Reset
              </Button>
            )}
          </div>

          {/* Loading State */}
          {isLoading && !displayData && (
            <div className="py-12">
              <LoadingSpinner size="lg" text="Memuat data purchase orders..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data purchase orders"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 && !hasActiveFilters ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <ShoppingCart className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada Purchase Order
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat purchase order pertama Anda
                  </p>
                  {canCreateOrders && (
                    <Button onClick={handleCreateClick}>
                      <Plus className="mr-2 h-4 w-4" />
                      Buat PO Baru
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

                  {/* Orders Table */}
                  <OrdersTable
                    orders={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditOrders}
                    canConfirm={canConfirmOrders}
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
