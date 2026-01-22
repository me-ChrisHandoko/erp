/**
 * Suppliers Client Component
 *
 * Client-side interactive component for supplier management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, Building2 } from "lucide-react";
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
import { EmptyState } from "@/components/shared/empty-state";
import { useListSuppliersQuery } from "@/store/services/supplierApi";
import { usePermissions } from "@/hooks/use-permissions";
import { SuppliersTable } from "@/components/suppliers/suppliers-table";
import type {
  SupplierFilters,
  SupplierListResponse,
} from "@/types/supplier.types";
import type { RootState } from "@/store";

interface SuppliersClientProps {
  initialData: SupplierListResponse;
}

export function SuppliersClient({ initialData }: SuppliersClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<string | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );
  const [filters, setFilters] = useState<SupplierFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "code",
    sortOrder: "asc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateSuppliers = permissions.canCreate("suppliers");
  const canEditSuppliers = permissions.canEdit("suppliers");

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch suppliers with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    type: typeFilter,
    isActive: statusFilter,
  };

  const {
    data: suppliersData,
    isLoading,
    error,
    refetch,
  } = useListSuppliersQuery(queryParams, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = suppliersData || initialData;

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
        } as SupplierFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as SupplierFilters["sortBy"],
        sortOrder: "asc",
      } as SupplierFilters;
    });
  };

  const handleTypeFilterChange = (type: string | undefined) => {
    setTypeFilter(type);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleStatusFilterChange = (status: boolean | undefined) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleResetFilters = () => {
    setSearch("");
    setDebouncedSearch("");
    setTypeFilter(undefined);
    setStatusFilter(undefined);
    setFilters({
      page: 1,
      pageSize: 20,
      sortBy: "code",
      sortOrder: "asc",
    });
  };

  const hasActiveFilters = search || typeFilter || statusFilter !== undefined;

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Daftar Supplier</h1>
          <p className="text-muted-foreground">
            Kelola informasi supplier dan rekanan bisnis
          </p>
        </div>
        {canCreateSuppliers && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/master/suppliers/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Tambah Supplier
          </Button>
        )}
      </div>

      {/* Suppliers table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari kode, nama, atau email supplier..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Type Filter */}
            <Select
              value={typeFilter || "all"}
              onValueChange={(value) =>
                handleTypeFilterChange(value === "all" ? undefined : value)
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Tipe</SelectItem>
                {displayData?.data &&
                  Array.from(
                    new Set(displayData.data.map((s) => s.type).filter(Boolean))
                  ).map((type) => (
                    <SelectItem key={type} value={type as string}>
                      {(type as string).charAt(0) +
                        (type as string).slice(1).toLowerCase()}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>

            {/* Status Filter */}
            <Select
              value={
                statusFilter === undefined
                  ? "all"
                  : statusFilter
                  ? "active"
                  : "inactive"
              }
              onValueChange={(value) =>
                handleStatusFilterChange(
                  value === "all" ? undefined : value === "active"
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[150px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="active">Aktif</SelectItem>
                <SelectItem value="inactive">Nonaktif</SelectItem>
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {hasActiveFilters && (
              <Button
                variant="ghost"
                onClick={handleResetFilters}
                className="h-10 px-4"
              >
                Reset
              </Button>
            )}
          </div>

          {/* Loading State (only for refetching) */}
          {isLoading && !displayData && (
            <div className="flex items-center justify-center py-12">
              <div className="text-center space-y-3">
                <LoadingSpinner size="lg" />
                <p className="text-sm text-muted-foreground">
                  Memuat data supplier...
                </p>
              </div>
            </div>
          )}

          {/* Error State */}
          {error && !isLoading && (
            <div className="py-8">
              <ErrorDisplay
                error={error}
                onRetry={refetch}
                title="Gagal memuat data supplier"
              />
            </div>
          )}

          {/* Empty State (no data at all) */}
          {!isLoading &&
            !error &&
            displayData?.data &&
            displayData.data.length === 0 &&
            !hasActiveFilters && (
              <div className="py-12">
                <EmptyState
                  icon={Building2}
                  title="Belum ada supplier"
                  description="Mulai dengan menambahkan supplier pertama Anda"
                  action={
                    canCreateSuppliers
                      ? {
                          label: "Tambah Supplier",
                          onClick: () =>
                            router.push("/master/suppliers/create"),
                        }
                      : undefined
                  }
                />
              </div>
            )}

          {/* Data Display */}
          {!error && displayData?.data && displayData.data.length > 0 && (
            <>
              <div className="space-y-4">
                {/* Subtle loading indicator for refetching */}
                {isLoading && (
                  <div className="text-sm text-muted-foreground text-center py-2">
                    Memperbarui data...
                  </div>
                )}
                <SuppliersTable
                  suppliers={displayData.data}
                  sortBy={filters.sortBy}
                  sortOrder={filters.sortOrder}
                  onSortChange={handleSortChange}
                  canEdit={canEditSuppliers}
                />

                {/* Pagination */}
                {displayData?.pagination && (
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
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
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
