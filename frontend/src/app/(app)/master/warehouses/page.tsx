/**
 * Warehouses List Page
 *
 * Displays and manages warehouse catalog with:
 * - Warehouse table with sorting & pagination
 * - Search and filter capabilities
 * - Create, edit actions (OWNER/ADMIN only)
 * - Multi-location warehouse management
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, Warehouse } from "lucide-react";
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
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { usePermissions } from "@/hooks/use-permissions";
import { WarehousesTable } from "@/components/warehouses/warehouses-table";
import { CreateWarehouseDialog } from "@/components/warehouses/create-warehouse-dialog";
import type {
  WarehouseFilters,
  WarehouseType,
} from "@/types/warehouse.types";
import { WAREHOUSE_TYPES } from "@/types/warehouse.types";
import type { RootState } from "@/store";

export default function WarehousesPage() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<WarehouseType | undefined>(
    undefined
  );
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [filters, setFilters] = useState<WarehouseFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "code",
    sortOrder: "asc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Compute permission checks ONCE at top level (Option 2 Plus pattern)
  const canCreateWarehouses = permissions.canCreate("warehouses");
  const canEditWarehouses = permissions.canEdit("warehouses");

  // 🔐 AUTH CHECK: Get authentication state
  const isAuthenticated = useSelector(
    (state: RootState) => state.auth.isAuthenticated
  );
  const accessToken = useSelector((state: RootState) => state.auth.accessToken);

  // 🔐 FIX #1: Get activeCompany from Redux to prevent race condition
  // This ensures we only fetch warehouses after company context is initialized
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // 🔐 AUTH CHECK: Redirect to login if not authenticated
  // This prevents infinite loading when accessing protected routes directly
  useEffect(() => {
    // Give auth system 2 seconds to restore from localStorage/cookie
    const authTimeout = setTimeout(() => {
      if (!isAuthenticated || !accessToken) {
        console.log(
          "[WarehousesPage] Not authenticated, redirecting to login..."
        );
        router.push("/login");
      }
    }, 2000);

    return () => clearTimeout(authTimeout);
  }, [isAuthenticated, accessToken, router]);

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch warehouses with filters
  // 🔐 FIX #1: Skip query until company context is ready to prevent 400 errors
  const {
    data: warehousesData,
    isLoading,
    error,
    refetch,
  } = useListWarehousesQuery(
    {
      ...filters,
      search: debouncedSearch || undefined,
      type: typeFilter,
      isActive: statusFilter,
    },
    {
      skip: !activeCompanyId, // Skip query until company context is available
    }
  );

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as WarehouseFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as WarehouseFilters["sortBy"],
        sortOrder: "asc",
      } as WarehouseFilters;
    });
  };

  const handleTypeFilterChange = (type: string | undefined) => {
    setTypeFilter(type as WarehouseType | undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleStatusFilterChange = (status: boolean | undefined) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // 🔐 FIX #1: Show loading state while company context is being initialized
  // This prevents the race condition where warehouses are fetched before company is selected
  if (!activeCompanyId) {
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Gudang" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <LoadingSpinner size="lg" />
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold">
                {!isAuthenticated
                  ? "Checking authentication..."
                  : "Initializing Company Context..."}
              </h3>
              <p className="text-sm text-muted-foreground">
                {!isAuthenticated
                  ? "Verifying your session, you will be redirected to login if needed..."
                  : "Please wait while we set up your company workspace"}
              </p>
            </div>
          </div>
        </div>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Gudang" },
          ]}
        />

        {/* Main content */}
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {/* Page title and actions */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-1">
              <h1 className="text-3xl font-bold tracking-tight">
                Daftar Gudang
              </h1>
              <p className="text-muted-foreground">
                Kelola gudang untuk manajemen inventori multi-lokasi
              </p>
            </div>
            {canCreateWarehouses && (
              <Button
                className="shrink-0"
                onClick={() => setIsCreateDialogOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                Tambah Gudang
              </Button>
            )}
          </div>

          {/* Warehouses table with search and filters */}
          <Card className="shadow-sm">
            <CardContent>
              {/* Search and Filters Row */}
              <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
                {/* Search */}
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Cari kode atau nama gudang..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="pl-9"
                  />
                </div>

                {/* Type Filter */}
                <Select
                  value={typeFilter || "all"}
                  onValueChange={(value) =>
                    handleTypeFilterChange(
                      value === "all" ? undefined : value
                    )
                  }
                >
                  <SelectTrigger className="w-full sm:w-[180px]">
                    <SelectValue placeholder="Semua Tipe" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua Tipe</SelectItem>
                    {WAREHOUSE_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
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
                {(typeFilter || statusFilter !== undefined || search) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSearch("");
                      setDebouncedSearch("");
                      setTypeFilter(undefined);
                      setStatusFilter(undefined);
                      setFilters((prev) => ({ ...prev, page: 1 }));
                    }}
                    className="w-full sm:w-auto"
                  >
                    Reset
                  </Button>
                )}
              </div>

              {isLoading && (
                <div className="py-12">
                  <LoadingSpinner size="lg" text="Memuat data gudang..." />
                </div>
              )}

              {error && (
                <ErrorDisplay
                  error={error}
                  title="Gagal memuat data gudang"
                  onRetry={refetch}
                />
              )}

              {!isLoading && !error && warehousesData && (
                <>
                  {warehousesData.data.length === 0 &&
                  !search &&
                  !typeFilter &&
                  statusFilter === undefined ? (
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <Warehouse className="mb-4 h-12 w-12 text-muted-foreground" />
                      <h3 className="mb-2 text-lg font-semibold">
                        Belum ada gudang
                      </h3>
                      <p className="mb-4 text-sm text-muted-foreground">
                        Mulai dengan menambahkan gudang pertama Anda
                      </p>
                      {canCreateWarehouses && (
                        <Button onClick={() => setIsCreateDialogOpen(true)}>
                          <Plus className="mr-2 h-4 w-4" />
                          Tambah Gudang
                        </Button>
                      )}
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {/* Warehouses Table */}
                      <WarehousesTable
                        warehouses={warehousesData.data}
                        sortBy={filters.sortBy}
                        sortOrder={filters.sortOrder}
                        onSortChange={handleSortChange}
                        canEdit={canEditWarehouses}
                      />

                      {/* Pagination */}
                      {warehousesData.pagination.totalPages > 1 && (
                        <div className="flex items-center justify-between border-t pt-4">
                          <div className="text-sm text-muted-foreground">
                            Halaman {warehousesData.pagination.page} dari{" "}
                            {warehousesData.pagination.totalPages} (
                            {warehousesData.pagination.total} gudang)
                          </div>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                handlePageChange(
                                  warehousesData.pagination.page - 1
                                )
                              }
                              disabled={warehousesData.pagination.page === 1}
                            >
                              Sebelumnya
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                handlePageChange(
                                  warehousesData.pagination.page + 1
                                )
                              }
                              disabled={
                                warehousesData.pagination.page >=
                                warehousesData.pagination.totalPages
                              }
                            >
                              Selanjutnya
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
      </div>

      {/* Create Warehouse Dialog */}
      <CreateWarehouseDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </ErrorBoundary>
  );
}
