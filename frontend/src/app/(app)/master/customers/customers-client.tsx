/**
 * Customers Client Component
 *
 * Client-side interactive component for customer management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { Plus, Search, Users } from "lucide-react";
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
import { useListCustomersQuery } from "@/store/services/customerApi";
import { usePermissions } from "@/hooks/use-permissions";
import { CustomersTable } from "@/components/customers/customers-table";
import { CreateCustomerDialog } from "@/components/customers/create-customer-dialog";
import type { CustomerFilters, CustomerType, CustomerListResponse } from "@/types/customer.types";

interface CustomersClientProps {
  initialData: CustomerListResponse;
}

export function CustomersClient({ initialData }: CustomersClientProps) {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<CustomerType | undefined>(
    undefined
  );
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [filters, setFilters] = useState<CustomerFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "code",
    sortOrder: "asc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Compute permission checks ONCE at top level
  const canCreateCustomers = permissions.canCreate("customers");
  const canEditCustomers = permissions.canEdit("customers");

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch customers with filters
  // 🎯 KEY: Use initialData from server for first render
  const {
    data: customersData = initialData,
    isLoading,
    error,
    refetch,
  } = useListCustomersQuery(
    {
      ...filters,
      search: debouncedSearch || undefined,
      customerType: typeFilter,
      isActive: statusFilter,
    },
    {
      skip: false, // Always query (for updates)
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
        } as CustomerFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as CustomerFilters["sortBy"],
        sortOrder: "asc",
      } as CustomerFilters;
    });
  };

  const handleTypeFilterChange = (type: CustomerType | undefined) => {
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
          <h1 className="text-3xl font-bold tracking-tight">
            Daftar Pelanggan
          </h1>
          <p className="text-muted-foreground">
            Kelola data pelanggan untuk distribusi
          </p>
        </div>
        {canCreateCustomers && (
          <Button
            className="shrink-0"
            onClick={() => setIsCreateDialogOpen(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            Tambah Pelanggan
          </Button>
        )}
      </div>

      {/* Customers table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari kode atau nama pelanggan..."
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
                  value === "all" ? undefined : (value as CustomerType)
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Tipe</SelectItem>
                {customersData?.data &&
                  Array.from(
                    new Set(
                      customersData.data
                        .map((c) => c.customerType)
                        .filter((type): type is CustomerType => type != null)
                    )
                  ).map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
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
                  value === "all"
                    ? undefined
                    : value === "active"
                    ? true
                    : false
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[160px]">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="active">Aktif</SelectItem>
                <SelectItem value="inactive">Nonaktif</SelectItem>
              </SelectContent>
            </Select>

            {/* Reset Filters Button */}
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
          {isLoading && !customersData && (
            <div className="flex items-center justify-center py-12">
              <div className="text-center space-y-3">
                <LoadingSpinner size="lg" />
                <p className="text-sm text-muted-foreground">
                  Memuat data pelanggan...
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
                title="Gagal memuat data pelanggan"
              />
            </div>
          )}

          {/* Empty State (no data at all) */}
          {!isLoading &&
            !error &&
            customersData?.data &&
            customersData.data.length === 0 &&
            !hasActiveFilters && (
              <div className="py-12">
                <EmptyState
                  icon={Users}
                  title="Belum ada pelanggan"
                  description="Mulai dengan menambahkan pelanggan pertama Anda"
                  action={
                    canCreateCustomers
                      ? {
                          label: "Tambah Pelanggan",
                          onClick: () => setIsCreateDialogOpen(true),
                        }
                      : undefined
                  }
                />
              </div>
            )}

          {/* Data Display */}
          {!error && customersData?.data && customersData.data.length > 0 && (
            <>
              {/* Subtle loading indicator for refetching */}
              {isLoading && (
                <div className="text-sm text-muted-foreground text-center py-2">
                  Memperbarui data...
                </div>
              )}

              <CustomersTable
                customers={customersData.data}
                sortBy={filters.sortBy}
                sortOrder={filters.sortOrder}
                onSortChange={handleSortChange}
                canEdit={canEditCustomers}
              />

              {/* Pagination */}
              {customersData.pagination.totalPages > 1 && (
                <div className="mt-6 flex items-center justify-between border-t pt-4">
                  <div className="text-sm text-muted-foreground">
                    Halaman {customersData.pagination.page} dari{" "}
                    {customersData.pagination.totalPages} (
                    {customersData.pagination.totalItems} total)
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handlePageChange(customersData.pagination.page - 1)
                      }
                      disabled={customersData.pagination.page === 1}
                    >
                      Sebelumnya
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        handlePageChange(customersData.pagination.page + 1)
                      }
                      disabled={
                        customersData.pagination.page >=
                        customersData.pagination.totalPages
                      }
                    >
                      Selanjutnya
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Create Customer Dialog */}
      <CreateCustomerDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </div>
  );
}
