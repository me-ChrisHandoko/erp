/**
 * Payments Client Component
 *
 * Client-side interactive component for payment management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, Wallet } from "lucide-react";
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
import { useListPaymentsQuery } from "@/store/services/paymentApi";
import { usePermissions } from "@/hooks/use-permissions";
import { PaymentsTable } from "@/components/payments/payments-table";
import type {
  PaymentFilters,
  PaymentListResponse,
  PaymentMethod,
} from "@/types/payment.types";
import type { RootState } from "@/store";
import { PAYMENT_METHOD_LABELS } from "@/types/payment.types";

interface PaymentsClientProps {
  initialData: PaymentListResponse;
}

export function PaymentsClient({ initialData }: PaymentsClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [paymentMethodFilter, setPaymentMethodFilter] = useState<
    PaymentMethod | undefined
  >(undefined);
  const [filters, setFilters] = useState<PaymentFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "paymentDate",
    sortOrder: "desc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreatePayments = permissions.canCreate("supplier-payments");
  const canEditPayments = permissions.canEdit("supplier-payments");

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch payments with filters
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    paymentMethod: paymentMethodFilter,
  };

  const {
    data: paymentsData,
    isLoading,
    error,
    refetch,
  } = useListPaymentsQuery(queryParams, {
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = paymentsData || initialData;

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
      page: 1,
    }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as PaymentFilters;
      }
      return {
        ...prev,
        sortBy: sortBy as PaymentFilters["sortBy"],
        sortOrder: "asc",
      } as PaymentFilters;
    });
  };

  const handlePaymentMethodFilterChange = (method: string | undefined) => {
    setPaymentMethodFilter(method as PaymentMethod | undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Daftar Pembayaran
          </h1>
          <p className="text-muted-foreground">Kelola pembayaran ke pemasok</p>
        </div>
        {canCreatePayments && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/procurement/payments/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Tambah Pembayaran
          </Button>
        )}
      </div>

      {/* Payments table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor pembayaran, pemasok, referensi..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Payment Method Filter */}
            <Select
              value={paymentMethodFilter || "all"}
              onValueChange={(value) =>
                handlePaymentMethodFilterChange(
                  value === "all" ? undefined : value
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[200px]">
                <SelectValue placeholder="Semua Metode" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Metode</SelectItem>
                {Object.entries(PAYMENT_METHOD_LABELS).map(([key, label]) => (
                  <SelectItem key={key} value={key}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {(paymentMethodFilter || search) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setDebouncedSearch("");
                  setPaymentMethodFilter(undefined);
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
              <LoadingSpinner size="lg" text="Memuat data pembayaran..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data pembayaran"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && displayData.data && (
            <>
              {displayData.data.length === 0 &&
              !search &&
              !paymentMethodFilter ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <Wallet className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada pembayaran
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan mencatat pembayaran pertama Anda
                  </p>
                  {canCreatePayments && (
                    <Button
                      onClick={() =>
                        router.push("/procurement/payments/create")
                      }
                    >
                      <Plus className="mr-2 h-4 w-4" />
                      Tambah Pembayaran
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

                  {/* Payments Table */}
                  <PaymentsTable
                    payments={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditPayments}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 ">
                      {/* Summary */}
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

                      {/* Page Size Selector */}
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

                      {/* Navigation Buttons */}
                      <div className="flex items-center justify-center sm:justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePageChange(1)}
                          disabled={displayData.pagination.page === 1}
                        >
                          &laquo;
                        </Button>
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
                        <span className="text-sm text-muted-foreground px-2">
                          Halaman {displayData.pagination.page} dari{" "}
                          {displayData.pagination.totalPages}
                        </span>
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
