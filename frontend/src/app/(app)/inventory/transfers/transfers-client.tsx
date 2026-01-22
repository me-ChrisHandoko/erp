/**
 * Transfers Client Component
 *
 * Client-side interactive component for stock transfer management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Status-based actions and workflow
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { type DateRange } from "react-day-picker";
import { format } from "date-fns";
import {
  Plus,
  Search,
  PackageOpen,
  FileEdit,
  Send,
  CheckCircle2,
  XCircle,
} from "lucide-react";
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
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useListTransfersQuery } from "@/store/services/transferApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { usePermissions } from "@/hooks/use-permissions";
import { TransfersTable } from "@/components/transfers/transfers-table";
import { ShipTransferDialog } from "@/components/transfers/ship-transfer-dialog";
import { ReceiveTransferDialog } from "@/components/transfers/receive-transfer-dialog";
import { CancelTransferDialog } from "@/components/transfers/cancel-transfer-dialog";
import { DeleteTransferDialog } from "@/components/transfers/delete-transfer-dialog";
import type {
  TransferFilters,
  TransferListResponse,
  StockTransferStatus,
  StockTransfer,
} from "@/types/transfer.types";
import type { RootState } from "@/store";

interface TransfersClientProps {
  initialData: TransferListResponse;
}

export function TransfersClient({ initialData }: TransfersClientProps) {
  const router = useRouter();

  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    StockTransferStatus | undefined
  >(undefined);
  const [sourceWarehouseFilter, setSourceWarehouseFilter] = useState<
    string | undefined
  >(undefined);
  const [destWarehouseFilter, setDestWarehouseFilter] = useState<
    string | undefined
  >(undefined);
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined);

  // Action dialogs state
  const [transferToShip, setTransferToShip] = useState<StockTransfer | null>(
    null
  );
  const [isShipDialogOpen, setIsShipDialogOpen] = useState(false);
  const [transferToReceive, setTransferToReceive] =
    useState<StockTransfer | null>(null);
  const [isReceiveDialogOpen, setIsReceiveDialogOpen] = useState(false);
  const [transferToCancel, setTransferToCancel] =
    useState<StockTransfer | null>(null);
  const [isCancelDialogOpen, setIsCancelDialogOpen] = useState(false);
  const [transferToDelete, setTransferToDelete] =
    useState<StockTransfer | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const [filters, setFilters] = useState<TransferFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "transferNumber",
    sortOrder: "desc", // Latest first
  });

  // Get permissions hook
  const permissions = usePermissions();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateTransfers = permissions.canCreate("stock-transfers");
  const canEditTransfers = permissions.canEdit("stock-transfers");
  const canDeleteTransfers = permissions.canDelete("stock-transfers");
  const canApproveTransfers = permissions.can("approve", "stock-transfers");

  // Fetch warehouses for filters
  const { data: warehousesData } = useListWarehousesQuery(
    { page: 1, pageSize: 100 }, // Get all warehouses for dropdown
    { skip: !activeCompanyId }
  );

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch transfers with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    status: statusFilter,
    sourceWarehouseId: sourceWarehouseFilter,
    destWarehouseId: destWarehouseFilter,
    dateFrom: dateRange?.from
      ? format(dateRange.from, "yyyy-MM-dd")
      : undefined,
    dateTo: dateRange?.to ? format(dateRange.to, "yyyy-MM-dd") : undefined,
  };

  const {
    data: transfersData,
    isLoading,
    error,
    refetch,
  } = useListTransfersQuery(queryParams, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = transfersData || initialData;

  // Use server-provided status counts (total counts, not per-page)
  const statusStats = displayData?.statusCounts
    ? {
        draft: displayData.statusCounts.draft,
        shipped: displayData.statusCounts.shipped,
        received: displayData.statusCounts.received,
        cancelled: displayData.statusCounts.cancelled,
      }
    : { draft: 0, shipped: 0, received: 0, cancelled: 0 };

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
        } as TransferFilters;
      }
      // New column, default to descending for transferNumber, ascending for others
      return {
        ...prev,
        sortBy: sortBy as TransferFilters["sortBy"],
        sortOrder: sortBy === "transferNumber" ? "desc" : "asc",
      } as TransferFilters;
    });
  };

  const handleStatusFilterChange = (status: string) => {
    setStatusFilter(
      status === "all" ? undefined : (status as StockTransferStatus)
    );
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleSourceWarehouseFilterChange = (warehouseId: string) => {
    setSourceWarehouseFilter(warehouseId === "all" ? undefined : warehouseId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleDestWarehouseFilterChange = (warehouseId: string) => {
    setDestWarehouseFilter(warehouseId === "all" ? undefined : warehouseId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // Action handlers
  const handleView = (transfer: StockTransfer) => {
    router.push(`/inventory/transfers/${transfer.id}`);
  };

  const handleShip = (transfer: StockTransfer) => {
    setTransferToShip(transfer);
    setIsShipDialogOpen(true);
  };

  const handleReceive = (transfer: StockTransfer) => {
    setTransferToReceive(transfer);
    setIsReceiveDialogOpen(true);
  };

  const handleCancel = (transfer: StockTransfer) => {
    setTransferToCancel(transfer);
    setIsCancelDialogOpen(true);
  };

  const handleEdit = (transfer: StockTransfer) => {
    router.push(`/inventory/transfers/${transfer.id}/edit`);
  };

  const handleDelete = (transfer: StockTransfer) => {
    setTransferToDelete(transfer);
    setIsDeleteDialogOpen(true);
  };

  const handleCreateClick = () => {
    router.push("/inventory/transfers/create");
  };

  const handleActionSuccess = () => {
    refetch(); // Refetch transfers list after successful action
  };

  return (
    <div className="flex flex-1 flex-col gap-0 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-0 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Transfer Gudang</h1>
          <p className="text-muted-foreground">
            Kelola transfer stok antar gudang
          </p>
        </div>
        {canCreateTransfers && (
          <Button className="shrink-0" onClick={handleCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Buat Transfer
          </Button>
        )}
      </div>

      {/* Status Statistics Cards */}
      {displayData && displayData.data && (
        <div className="grid gap-3 md:grid-cols-4">
          {/* Draft Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-gray-500 to-gray-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-gray-100">Draft</p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.draft}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-gray-100">
                      <span>Belum dikirim</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <FileEdit className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Shipped Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-blue-500 to-blue-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-blue-100">Dikirim</p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.shipped}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-blue-100">
                      <span>Dalam perjalanan</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <Send className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Received Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-green-500 to-green-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-green-100">
                      Diterima
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.received}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-green-100">
                      <span>Transfer selesai</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <CheckCircle2 className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Cancelled Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-red-500 to-red-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-red-100">
                      Dibatalkan
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.cancelled}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-red-100">
                      <span>Transfer dibatalkan</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <XCircle className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Transfers table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor transfer atau produk..."
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
                <SelectItem value="DRAFT">Draft</SelectItem>
                <SelectItem value="SHIPPED">Dikirim</SelectItem>
                <SelectItem value="RECEIVED">Diterima</SelectItem>
                <SelectItem value="CANCELLED">Dibatalkan</SelectItem>
              </SelectContent>
            </Select>

            {/* Source Warehouse Filter */}
            <Select
              value={sourceWarehouseFilter || "all"}
              onValueChange={handleSourceWarehouseFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[150px]">
                <SelectValue placeholder="Dari Gudang" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Gudang</SelectItem>
                {warehousesData?.data?.map((warehouse) => (
                  <SelectItem key={warehouse.id} value={warehouse.id}>
                    {warehouse.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Destination Warehouse Filter */}
            <Select
              value={destWarehouseFilter || "all"}
              onValueChange={handleDestWarehouseFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[150px]">
                <SelectValue placeholder="Ke Gudang" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Gudang</SelectItem>
                {warehousesData?.data?.map((warehouse) => (
                  <SelectItem key={warehouse.id} value={warehouse.id}>
                    {warehouse.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Date Range Filter */}
            <DateRangePicker
              value={dateRange}
              onChange={(range) => {
                setDateRange(range);
                setFilters((prev) => ({ ...prev, page: 1 }));
              }}
              placeholder="Pilih tanggal"
              className="w-full sm:w-[220px]"
            />

            {/* Clear Filters Button */}
            {(search ||
              statusFilter ||
              sourceWarehouseFilter ||
              destWarehouseFilter ||
              dateRange) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setStatusFilter(undefined);
                  setSourceWarehouseFilter(undefined);
                  setDestWarehouseFilter(undefined);
                  setDateRange(undefined);
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
              <LoadingSpinner size="lg" text="Memuat data transfer..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data transfer"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 &&
              !search &&
              !statusFilter &&
              !sourceWarehouseFilter &&
              !destWarehouseFilter ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <PackageOpen className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada transfer
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat transfer gudang pertama Anda
                  </p>
                  {canCreateTransfers && (
                    <Button onClick={handleCreateClick}>
                      <Plus className="mr-2 h-4 w-4" />
                      Buat Transfer
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

                  {/* Transfers Table */}
                  <TransfersTable
                    transfers={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    onView={handleView}
                    onShip={handleShip}
                    onReceive={handleReceive}
                    onCancel={handleCancel}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    canEdit={canEditTransfers}
                    canDelete={canDeleteTransfers}
                    canApprove={canApproveTransfers}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 ">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const { page, limit, total } = displayData.pagination;
                          const start = (page - 1) * limit + 1;
                          const end = Math.min(page * limit, total);
                          return `Menampilkan ${start}-${end} dari ${total} item`;
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

      {/* Ship Transfer Dialog */}
      <ShipTransferDialog
        transfer={transferToShip}
        open={isShipDialogOpen}
        onOpenChange={setIsShipDialogOpen}
        onSuccess={handleActionSuccess}
      />

      {/* Receive Transfer Dialog */}
      <ReceiveTransferDialog
        transfer={transferToReceive}
        open={isReceiveDialogOpen}
        onOpenChange={setIsReceiveDialogOpen}
        onSuccess={handleActionSuccess}
      />

      {/* Cancel Transfer Dialog */}
      <CancelTransferDialog
        transfer={transferToCancel}
        open={isCancelDialogOpen}
        onOpenChange={setIsCancelDialogOpen}
        onSuccess={handleActionSuccess}
      />

      {/* Delete Transfer Dialog */}
      <DeleteTransferDialog
        transfer={transferToDelete}
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onSuccess={handleActionSuccess}
      />
    </div>
  );
}
