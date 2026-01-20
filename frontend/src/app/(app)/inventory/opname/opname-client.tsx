/**
 * Stock Opname Client Component
 *
 * Client-side interactive component for stock opname (physical inventory count) management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Status-based filtering and visualization
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { type DateRange } from "react-day-picker";
import { format } from "date-fns";
import {
  Search,
  ClipboardList,
  Plus,
  FileEdit,
  Clock,
  CheckCircle,
  CheckCircle2,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { DateRangePicker } from "@/components/ui/date-range-picker";
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
import { useListOpnamesQuery } from "@/store/services/opnameApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { usePermissions } from "@/hooks/use-permissions";
import { OpnameTable } from "@/components/opname/opname-table";
import type {
  StockOpnameFilters,
  StockOpnameListResponse,
  StockOpnameStatus,
} from "@/types/opname.types";
import type { RootState } from "@/store";

interface OpnameClientProps {
  initialData: StockOpnameListResponse;
}

export function OpnameClient({ initialData }: OpnameClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [warehouseFilter, setWarehouseFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined);
  const [filters, setFilters] = useState<StockOpnameFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "opnameNumber",
    sortOrder: "desc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateOpname = permissions.canCreate("stock-opname");
  const canEditOpname = permissions.canEdit("stock-opname");
  const canDeleteOpname = permissions.canDelete("stock-opname");
  const canApproveOpname = permissions.canApprove("stock-opname");

  // Fetch active warehouses for filter dropdown
  const { data: warehousesData } = useListWarehousesQuery(
    {
      isActive: true,
      pageSize: 100, // Get all active warehouses
      sortBy: "name",
      sortOrder: "asc",
    },
    {
      skip: !activeCompanyId,
    }
  );

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Build query params based on filters
  const queryParams: StockOpnameFilters = {
    ...filters,
    search: debouncedSearch || undefined,
    warehouseId: warehouseFilter !== "all" ? warehouseFilter : undefined,
    status:
      statusFilter !== "all" ? (statusFilter as StockOpnameStatus) : undefined,
    dateFrom: dateRange?.from ? format(dateRange.from, "yyyy-MM-dd") : undefined,
    dateTo: dateRange?.to ? format(dateRange.to, "yyyy-MM-dd") : undefined,
  };

  // Fetch stock opnames with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const {
    data: opnamesData,
    isLoading,
    error,
    refetch,
  } = useListOpnamesQuery(queryParams, {
    // Skip query until company context is available
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = opnamesData || initialData;

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
        } as StockOpnameFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as StockOpnameFilters["sortBy"],
        sortOrder: "asc",
      } as StockOpnameFilters;
    });
  };

  const handleWarehouseFilterChange = (warehouseId: string) => {
    setWarehouseFilter(warehouseId);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleStatusFilterChange = (status: string) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleClearFilters = () => {
    setSearch("");
    setDebouncedSearch("");
    setWarehouseFilter("all");
    setStatusFilter("all");
    setDateRange(undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // Use server-provided status counts (total counts, not per-page)
  const statusStats = displayData?.statusCounts
    ? {
        draft: displayData.statusCounts.draft,
        inProgress: displayData.statusCounts.inProgress,
        completed: displayData.statusCounts.completed,
        approved: displayData.statusCounts.approved,
      }
    : { draft: 0, inProgress: 0, completed: 0, approved: 0 };

  return (
    <div className="flex flex-1 flex-col gap-0 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-0 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Stock Opname</h1>
          <p className="text-muted-foreground">
            Kelola penghitungan fisik inventory dan penyesuaian stok
          </p>
        </div>
        {canCreateOpname && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/inventory/opname/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Buat Stock Opname
          </Button>
        )}
      </div>

      {/* Status Statistics Cards */}
      {displayData && displayData.data && (
        <div className="grid gap-1 md:grid-cols-4">
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
                      <span>Belum diproses</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <FileEdit className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* In Progress Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-blue-500 to-blue-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-blue-100">
                      In Progress
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.inProgress}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-blue-100">
                      <span>Sedang dihitung</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <Clock className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Completed Card */}
          <Card className="overflow-hidden border-none shadow-none hover:shadow-sm transition-shadow duration-300 rounded-xl">
            <CardContent className="p-0">
              <div className="relative bg-gradient-to-br from-orange-500 to-orange-600 p-4 rounded-xl">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-orange-100">
                      Completed
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.completed}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-orange-100">
                      <span>Menunggu approval</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <CheckCircle className="h-5 w-5 text-white" />
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
                      Approved
                    </p>
                    <p className="text-2xl font-bold text-white">
                      {statusStats.approved}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-green-100">
                      <span>Sudah disetujui</span>
                    </div>
                  </div>
                  <div className="rounded-full bg-white/20 p-2 backdrop-blur-sm">
                    <CheckCircle2 className="h-5 w-5 text-white" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Stock opname table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor opname (contoh: OPN-20260111-001)..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 bg-background"
              />
            </div>

            {/* Warehouse Filter */}
            <Select
              value={warehouseFilter}
              onValueChange={handleWarehouseFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[150px] bg-background">
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

            {/* Status Filter */}
            <Select
              value={statusFilter}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="w-full sm:w-[140px] bg-background">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="in_progress">In Progress</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
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
              warehouseFilter !== "all" ||
              statusFilter !== "all" ||
              dateRange) && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearFilters}
                className="w-full sm:w-auto"
              >
                Reset
              </Button>
            )}
          </div>

          {/* Loading State */}
          {isLoading && !displayData && (
            <div className="py-12">
              <LoadingSpinner size="lg" text="Memuat data stock opname..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data stock opname"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 &&
              !search &&
              warehouseFilter === "all" &&
              statusFilter === "all" &&
              !dateRange ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <ClipboardList className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada stock opname
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat stock opname pertama Anda
                  </p>
                  {canCreateOpname && (
                    <Button
                      onClick={() => router.push("/inventory/opname/create")}
                    >
                      <Plus className="mr-2 h-4 w-4" />
                      Buat Stock Opname
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

                  {/* Opname Table */}
                  <OpnameTable
                    opnames={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditOpname}
                    canDelete={canDeleteOpname}
                    canApprove={canApproveOpname}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-t pt-4 mt-6">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const page = displayData.pagination?.page || 1;
                          const pageSize =
                            displayData.pagination?.pageSize || 20;
                          const totalItems =
                            displayData.pagination?.totalItems || 0;
                          const start =
                            totalItems > 0 ? (page - 1) * pageSize + 1 : 0;
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
                          <SelectTrigger className="w-[70px] h-8 bg-background">
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
                          disabled={(displayData.pagination?.page || 1) === 1}
                        >
                          &laquo;
                        </Button>

                        {/* Previous Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(
                              (displayData.pagination?.page || 1) - 1
                            )
                          }
                          disabled={(displayData.pagination?.page || 1) === 1}
                        >
                          &lsaquo;
                        </Button>

                        {/* Current Page Info */}
                        <span className="text-sm text-muted-foreground px-2">
                          Halaman {displayData.pagination?.page || 1} dari{" "}
                          {displayData.pagination?.totalPages || 1}
                        </span>

                        {/* Next Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(
                              (displayData.pagination?.page || 1) + 1
                            )
                          }
                          disabled={
                            (displayData.pagination?.page || 1) >=
                            (displayData.pagination?.totalPages || 1)
                          }
                        >
                          &rsaquo;
                        </Button>

                        {/* Last Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(
                              displayData.pagination?.totalPages || 1
                            )
                          }
                          disabled={
                            (displayData.pagination?.page || 1) >=
                            (displayData.pagination?.totalPages || 1)
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
