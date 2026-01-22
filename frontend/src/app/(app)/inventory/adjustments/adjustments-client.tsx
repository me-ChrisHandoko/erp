/**
 * Adjustments Client Component
 *
 * Client-side interactive component for inventory adjustment management.
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
  ClipboardList,
  FileEdit,
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
import { useListAdjustmentsQuery } from "@/store/services/adjustmentApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { usePermissions } from "@/hooks/use-permissions";
import { AdjustmentsTable } from "@/components/adjustments/adjustments-table";
import { ApproveAdjustmentDialog } from "@/components/adjustments/approve-adjustment-dialog";
import { CancelAdjustmentDialog } from "@/components/adjustments/cancel-adjustment-dialog";
import { DeleteAdjustmentDialog } from "@/components/adjustments/delete-adjustment-dialog";
import {
  ADJUSTMENT_REASON_CONFIG,
  type AdjustmentFilters,
  type AdjustmentListResponse,
  type AdjustmentStatus,
  type AdjustmentType,
  type AdjustmentReason,
  type InventoryAdjustment,
} from "@/types/adjustment.types";
import type { RootState } from "@/store";

interface AdjustmentsClientProps {
  initialData: AdjustmentListResponse;
}

export function AdjustmentsClient({ initialData }: AdjustmentsClientProps) {
  const router = useRouter();

  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    AdjustmentStatus | undefined
  >(undefined);
  const [typeFilter, setTypeFilter] = useState<AdjustmentType | undefined>(
    undefined
  );
  const [reasonFilter, setReasonFilter] = useState<
    AdjustmentReason | undefined
  >(undefined);
  const [warehouseFilter, setWarehouseFilter] = useState<string | undefined>(
    undefined
  );
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined);

  // Action dialogs state (only action dialogs: Approve, Cancel, Delete)
  const [adjustmentToApprove, setAdjustmentToApprove] =
    useState<InventoryAdjustment | null>(null);
  const [isApproveDialogOpen, setIsApproveDialogOpen] = useState(false);
  const [adjustmentToCancel, setAdjustmentToCancel] =
    useState<InventoryAdjustment | null>(null);
  const [isCancelDialogOpen, setIsCancelDialogOpen] = useState(false);
  const [adjustmentToDelete, setAdjustmentToDelete] =
    useState<InventoryAdjustment | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const [filters, setFilters] = useState<AdjustmentFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "adjustmentNumber",
    sortOrder: "desc", // Latest first
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateAdjustments = permissions.canCreate("inventory-adjustments");
  const canEditAdjustments = permissions.canEdit("inventory-adjustments");
  const canDeleteAdjustments = permissions.canDelete("inventory-adjustments");
  const canApproveAdjustments = permissions.can(
    "approve",
    "inventory-adjustments"
  );

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

  // Fetch adjustments with filters
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    status: statusFilter,
    adjustmentType: typeFilter,
    reason: reasonFilter,
    warehouseId: warehouseFilter,
    dateFrom: dateRange?.from
      ? format(dateRange.from, "yyyy-MM-dd")
      : undefined,
    dateTo: dateRange?.to ? format(dateRange.to, "yyyy-MM-dd") : undefined,
  };

  const {
    data: adjustmentsData,
    isLoading,
    error,
    refetch,
  } = useListAdjustmentsQuery(queryParams, {
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = adjustmentsData || initialData;

  // Use server-provided status counts (total counts, not per-page)
  const statusStats = displayData?.statusCounts
    ? {
        draft: displayData.statusCounts.draft,
        approved: displayData.statusCounts.approved,
        cancelled: displayData.statusCounts.cancelled,
      }
    : { draft: 0, approved: 0, cancelled: 0 };

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
        } as AdjustmentFilters;
      }
      // New column, default to descending for adjustmentNumber, ascending for others
      return {
        ...prev,
        sortBy: sortBy as AdjustmentFilters["sortBy"],
        sortOrder: sortBy === "adjustmentNumber" ? "desc" : "asc",
      } as AdjustmentFilters;
    });
  };

  const handleStatusFilterChange = (status: string) => {
    setStatusFilter(
      status === "all" ? undefined : (status as AdjustmentStatus)
    );
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleTypeFilterChange = (type: string) => {
    setTypeFilter(type === "all" ? undefined : (type as AdjustmentType));
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleReasonFilterChange = (reason: string) => {
    setReasonFilter(
      reason === "all" ? undefined : (reason as AdjustmentReason)
    );
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleWarehouseFilterChange = (warehouseId: string) => {
    setWarehouseFilter(warehouseId === "all" ? undefined : warehouseId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // Action handlers
  const handleView = (adjustment: InventoryAdjustment) => {
    router.push(`/inventory/adjustments/${adjustment.id}`);
  };

  const handleApprove = (adjustment: InventoryAdjustment) => {
    setAdjustmentToApprove(adjustment);
    setIsApproveDialogOpen(true);
  };

  const handleCancel = (adjustment: InventoryAdjustment) => {
    setAdjustmentToCancel(adjustment);
    setIsCancelDialogOpen(true);
  };

  const handleEdit = (adjustment: InventoryAdjustment) => {
    router.push(`/inventory/adjustments/${adjustment.id}/edit`);
  };

  const handleDelete = (adjustment: InventoryAdjustment) => {
    setAdjustmentToDelete(adjustment);
    setIsDeleteDialogOpen(true);
  };

  const handleCreateClick = () => {
    router.push("/inventory/adjustments/create");
  };

  const handleActionSuccess = () => {
    refetch(); // Refetch adjustments list after successful action
  };

  const hasActiveFilters =
    search ||
    statusFilter ||
    typeFilter ||
    reasonFilter ||
    warehouseFilter ||
    dateRange;

  return (
    <div className="flex flex-1 flex-col gap-0 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-0 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Penyesuaian Stok
          </h1>
          <p className="text-muted-foreground">
            Kelola penyesuaian inventori untuk koreksi stok
          </p>
        </div>
        {canCreateAdjustments && (
          <Button className="shrink-0" onClick={handleCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Buat Penyesuaian
          </Button>
        )}
      </div>

      {/* Status Statistics Cards */}
      {displayData && displayData.data && (
        <div className="grid gap-3 md:grid-cols-3">
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
                      <span>Menunggu persetujuan</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <FileEdit className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Approved Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-green-500 to-green-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-green-100">
                      Disetujui
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.approved}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-green-100">
                      <span>Penyesuaian selesai</span>
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
                      <span>Penyesuaian batal</span>
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

      {/* Adjustments table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search Row - Full Width */}
          <div className="mb-4">
            <div className="relative w-full">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor penyesuaian atau produk..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>

          {/* Filters Row - Order: Date, Warehouse, Type, Reason, Status, Reset */}
          <div className="mb-6 flex flex-col sm:flex-row gap-3 sm:items-center">
            {/* Date Range Filter - Fixed width */}
            <DateRangePicker
              value={dateRange}
              onChange={(range) => {
                setDateRange(range);
                setFilters((prev) => ({ ...prev, page: 1 }));
              }}
              placeholder="Semua Tanggal"
              className="w-full sm:w-[220px] shrink-0"
            />

            {/* Warehouse Filter */}
            <Select
              value={warehouseFilter || "all"}
              onValueChange={handleWarehouseFilterChange}
            >
              <SelectTrigger className="w-full sm:flex-1">
                <SelectValue placeholder="Semua Gudang" />
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

            {/* Type Filter */}
            <Select
              value={typeFilter || "all"}
              onValueChange={handleTypeFilterChange}
            >
              <SelectTrigger className="w-full sm:flex-1">
                <SelectValue placeholder="Semua Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Tipe</SelectItem>
                <SelectItem value="INCREASE">Penambahan</SelectItem>
                <SelectItem value="DECREASE">Pengurangan</SelectItem>
              </SelectContent>
            </Select>

            {/* Reason Filter */}
            <Select
              value={reasonFilter || "all"}
              onValueChange={handleReasonFilterChange}
            >
              <SelectTrigger className="w-full sm:flex-1">
                <SelectValue placeholder="Semua Alasan" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Alasan</SelectItem>
                {Object.entries(ADJUSTMENT_REASON_CONFIG).map(
                  ([key, config]) => (
                    <SelectItem key={key} value={key}>
                      {config.label}
                    </SelectItem>
                  )
                )}
              </SelectContent>
            </Select>

            {/* Status Filter */}
            <Select
              value={statusFilter || "all"}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="w-full sm:flex-1">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="DRAFT">Draft</SelectItem>
                <SelectItem value="APPROVED">Disetujui</SelectItem>
                <SelectItem value="CANCELLED">Dibatalkan</SelectItem>
              </SelectContent>
            </Select>

            {/* Clear Filters Button - in same row */}
            {hasActiveFilters && (
              <Button
                variant="outline"
                size="default"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setStatusFilter(undefined);
                  setTypeFilter(undefined);
                  setReasonFilter(undefined);
                  setWarehouseFilter(undefined);
                  setDateRange(undefined);
                  setFilters((prev) => ({ ...prev, page: 1 }));
                }}
                className="shrink-0"
              >
                Reset Filter
              </Button>
            )}
          </div>

          {/* Loading State */}
          {isLoading && !displayData && (
            <div className="py-12">
              <LoadingSpinner size="lg" text="Memuat data penyesuaian..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data penyesuaian"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 && !hasActiveFilters ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <ClipboardList className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada penyesuaian
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat penyesuaian stok pertama Anda
                  </p>
                  {canCreateAdjustments && (
                    <Button onClick={handleCreateClick}>
                      <Plus className="mr-2 h-4 w-4" />
                      Buat Penyesuaian
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

                  {/* Adjustments Table */}
                  <AdjustmentsTable
                    adjustments={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    onView={handleView}
                    onApprove={handleApprove}
                    onCancel={handleCancel}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    canEdit={canEditAdjustments}
                    canDelete={canDeleteAdjustments}
                    canApprove={canApproveAdjustments}
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

      {/* Approve Adjustment Dialog */}
      <ApproveAdjustmentDialog
        adjustment={adjustmentToApprove}
        open={isApproveDialogOpen}
        onOpenChange={setIsApproveDialogOpen}
        onSuccess={handleActionSuccess}
      />

      {/* Cancel Adjustment Dialog */}
      <CancelAdjustmentDialog
        adjustment={adjustmentToCancel}
        open={isCancelDialogOpen}
        onOpenChange={setIsCancelDialogOpen}
        onSuccess={handleActionSuccess}
      />

      {/* Delete Adjustment Dialog */}
      <DeleteAdjustmentDialog
        adjustment={adjustmentToDelete}
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onSuccess={handleActionSuccess}
      />
    </div>
  );
}
