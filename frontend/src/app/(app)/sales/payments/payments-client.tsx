/**
 * Sales Payments Client Component
 *
 * Client-side interactive component for customer payment management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, DollarSign, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { useListSalesPaymentsQuery } from "@/store/services/salesPaymentApi";
import { usePermissions } from "@/hooks/use-permissions";
import { SalesPaymentsTable } from "@/components/sales-payments/sales-payments-table";
import type { SalesPaymentFilters, SalesPaymentListResponse, PaymentMethod } from "@/types/sales-payment.types";
import type { RootState } from "@/store";
import { PAYMENT_METHOD, PAYMENT_METHOD_LABELS } from "@/types/sales-payment.types";

interface SalesPaymentsClientProps {
  initialData: SalesPaymentListResponse;
}

export function SalesPaymentsClient({ initialData }: SalesPaymentsClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [paymentMethodFilter, setPaymentMethodFilter] = useState<PaymentMethod | undefined>(
    undefined
  );
  const [dateFrom, setDateFrom] = useState<string>("");
  const [dateTo, setDateTo] = useState<string>("");
  const [filters, setFilters] = useState<SalesPaymentFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "paymentDate",
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
  const canCreatePayments = permissions.canCreate('payments');
  const canEditPayments = permissions.canEdit('payments');
  const canDeletePayments = permissions.canDelete('payments');

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch payments with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    paymentMethod: paymentMethodFilter,
    dateFrom: dateFrom || undefined,
    dateTo: dateTo || undefined,
  };

  const {
    data: paymentsData,
    isLoading,
    error,
    refetch,
  } = useListSalesPaymentsQuery(queryParams, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = paymentsData || initialData;

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
      page: 1 // Reset to page 1 when changing page size
    }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as SalesPaymentFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as SalesPaymentFilters["sortBy"],
        sortOrder: "asc",
      } as SalesPaymentFilters;
    });
  };

  const handlePaymentMethodFilterChange = (method: PaymentMethod | undefined) => {
    setPaymentMethodFilter(method);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  // Quick date filter functions
  const setDateRangeToToday = () => {
    const today = new Date().toISOString().split('T')[0];
    setDateFrom(today);
    setDateTo(today);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const setDateRangeToThisWeek = () => {
    const now = new Date();
    const dayOfWeek = now.getDay();
    const monday = new Date(now);
    monday.setDate(now.getDate() - (dayOfWeek === 0 ? 6 : dayOfWeek - 1));
    const sunday = new Date(monday);
    sunday.setDate(monday.getDate() + 6);

    setDateFrom(monday.toISOString().split('T')[0]);
    setDateTo(sunday.toISOString().split('T')[0]);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const setDateRangeToThisMonth = () => {
    const now = new Date();
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
    const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);

    setDateFrom(firstDay.toISOString().split('T')[0]);
    setDateTo(lastDay.toISOString().split('T')[0]);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Pembayaran Pelanggan
          </h1>
          <p className="text-muted-foreground">
            Kelola pembayaran dari pelanggan untuk invoice penjualan
          </p>
        </div>
        {canCreatePayments && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/sales/payments/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Catat Pembayaran
          </Button>
        )}
      </div>

      {/* Payments table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3">
            {/* First Row: Search and Payment Method */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              {/* Search */}
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Cari nomor pembayaran, pelanggan, invoice..."
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
                    value === "all" ? undefined : value as PaymentMethod
                  )
                }
              >
                <SelectTrigger className="w-full sm:w-[180px]">
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
            </div>

            {/* Second Row: Date Range Filter */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
              {/* Date Range Inputs */}
              <div className="flex flex-1 flex-col gap-3 sm:flex-row sm:items-end">
                {/* From Date */}
                <div className="flex-1 space-y-2">
                  <Label htmlFor="dateFrom" className="text-xs text-muted-foreground">
                    Tanggal Mulai
                  </Label>
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      id="dateFrom"
                      type="date"
                      value={dateFrom}
                      onChange={(e) => {
                        setDateFrom(e.target.value);
                        setFilters((prev) => ({ ...prev, page: 1 }));
                      }}
                      className="pl-9"
                    />
                  </div>
                </div>

                {/* To Date */}
                <div className="flex-1 space-y-2">
                  <Label htmlFor="dateTo" className="text-xs text-muted-foreground">
                    Tanggal Akhir
                  </Label>
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      id="dateTo"
                      type="date"
                      value={dateTo}
                      onChange={(e) => {
                        setDateTo(e.target.value);
                        setFilters((prev) => ({ ...prev, page: 1 }));
                      }}
                      className="pl-9"
                    />
                  </div>
                </div>
              </div>

              {/* Quick Filter Buttons */}
              <div className="flex flex-wrap gap-2 sm:flex-nowrap">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={setDateRangeToToday}
                  className="flex-1 sm:flex-none"
                >
                  Hari Ini
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={setDateRangeToThisWeek}
                  className="flex-1 sm:flex-none"
                >
                  Minggu Ini
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={setDateRangeToThisMonth}
                  className="flex-1 sm:flex-none"
                >
                  Bulan Ini
                </Button>
              </div>

              {/* Clear Filters Button */}
              {(paymentMethodFilter || search || dateFrom || dateTo) && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setSearch("");
                    setDebouncedSearch("");
                    setPaymentMethodFilter(undefined);
                    setDateFrom("");
                    setDateTo("");
                    setFilters((prev) => ({ ...prev, page: 1 }));
                  }}
                  className="w-full sm:w-auto"
                >
                  Reset
                </Button>
              )}
            </div>
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
              !paymentMethodFilter &&
              !dateFrom &&
              !dateTo ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <DollarSign className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada pembayaran
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan mencatat pembayaran pertama dari pelanggan
                  </p>
                  {canCreatePayments && (
                    <Button onClick={() => router.push("/sales/payments/create")}>
                      <Plus className="mr-2 h-4 w-4" />
                      Catat Pembayaran
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
                  <SalesPaymentsTable
                    payments={displayData.data}
                    sortBy={filters.sortBy}
                    sortOrder={filters.sortOrder}
                    onSortChange={handleSortChange}
                    canEdit={canEditPayments}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-t pt-4">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const pagination = displayData.pagination as any;
                          const page = pagination.page || 1;
                          const pageSize = pagination.limit || pagination.pageSize || 20;
                          const totalItems = pagination.total || pagination.totalItems || 0;
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
