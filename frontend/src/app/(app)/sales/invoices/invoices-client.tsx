/**
 * Invoices Client Component
 *
 * Client-side interactive component for sales invoice management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, Filter, X } from "lucide-react";
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
import { useListInvoicesQuery } from "@/store/services/invoiceApi";
import { usePermissions } from "@/hooks/use-permissions";
import { InvoicesTable } from "@/components/invoices/invoices-table";
import type { InvoiceFilters, InvoiceListResponse, PaymentStatus } from "@/types/invoice.types";
import type { RootState } from "@/store";

interface InvoicesClientProps {
  initialData: InvoiceListResponse;
}

export function InvoicesClient({ initialData }: InvoicesClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [paymentStatus, setPaymentStatus] = useState<PaymentStatus | "all">("all");
  const [filters, setFilters] = useState<InvoiceFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "invoiceDate",
    sortOrder: "desc",
  });

  // Get permissions hook
  const permissions = usePermissions();

  // Get activeCompanyId from Redux to trigger refetch on company switch
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateInvoices = permissions.canCreate('sales-invoices');
  const canEditInvoices = permissions.canEdit('sales-invoices');
  const canDeleteInvoices = permissions.canDelete('sales-invoices');

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch invoices with filters
  // Skip query until company context is ready
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    paymentStatus: paymentStatus !== "all" ? paymentStatus : undefined,
  };

  const {
    data: invoicesData,
    isLoading,
    error,
    refetch,
  } = useListInvoicesQuery(queryParams, {
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = invoicesData || initialData;

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
        } as InvoiceFilters;
      }
      return {
        ...prev,
        sortBy: sortBy as InvoiceFilters["sortBy"],
        sortOrder: "desc",
      } as InvoiceFilters;
    });
  };

  const handlePaymentStatusChange = (status: string) => {
    setPaymentStatus(status as PaymentStatus | "all");
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleClearFilters = () => {
    setSearch("");
    setDebouncedSearch("");
    setPaymentStatus("all");
    setFilters({
      page: 1,
      pageSize: 20,
      sortBy: "invoiceDate",
      sortOrder: "desc",
    });
  };

  // Check if any filters are active
  const hasActiveFilters = search || paymentStatus !== "all";

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Faktur Penjualan
          </h1>
          <p className="text-muted-foreground">
            Kelola faktur penjualan dan tagihan pelanggan
          </p>
        </div>
        {canCreateInvoices && (
          <Button
            className="shrink-0"
            onClick={() => router.push("/sales/invoices/create")}
          >
            <Plus className="mr-2 h-4 w-4" />
            Buat Faktur
          </Button>
        )}
      </div>

      {/* Invoices table with search and filters */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Search and Filters Row */}
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Cari nomor faktur atau nama pelanggan..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Payment Status Filter */}
            <Select
              value={paymentStatus}
              onValueChange={handlePaymentStatusChange}
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Status Pembayaran" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="UNPAID">Belum Dibayar</SelectItem>
                <SelectItem value="PARTIAL">Dibayar Sebagian</SelectItem>
                <SelectItem value="PAID">Lunas</SelectItem>
                <SelectItem value="OVERDUE">Jatuh Tempo</SelectItem>
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {hasActiveFilters && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearFilters}
                className="shrink-0"
              >
                <X className="mr-2 h-4 w-4" />
                Reset Filter
              </Button>
            )}
          </div>

          {/* Loading State */}
          {isLoading && !displayData && (
            <div className="flex items-center justify-center py-12">
              <LoadingSpinner size="lg" />
            </div>
          )}

          {/* Error State */}
          {error && !displayData && (
            <div className="py-8">
              <ErrorDisplay
                error={error}
                title="Gagal memuat data faktur"
                onRetry={() => refetch()}
              />
            </div>
          )}

          {/* Table */}
          {displayData && (
            <>
              <InvoicesTable
                data={displayData.data || []}
                isLoading={isLoading}
                error={error}
                currentPage={filters.page || 1}
                totalPages={displayData.pagination?.totalPages || 1}
                onPageChange={handlePageChange}
                onSortChange={handleSortChange}
                sortBy={filters.sortBy}
                sortOrder={filters.sortOrder}
                canEdit={canEditInvoices}
                canDelete={canDeleteInvoices}
              />

              {/* Pagination Controls */}
              {displayData.pagination && displayData.pagination.totalPages > 1 && (
                <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div className="text-sm text-muted-foreground">
                    Menampilkan {(((filters.page || 1) - 1) * (filters.pageSize || 20)) + 1} -{" "}
                    {Math.min((filters.page || 1) * (filters.pageSize || 20), displayData.pagination.total)} dari{" "}
                    {displayData.pagination.total} faktur
                  </div>
                  <div className="flex items-center gap-2">
                    <Select
                      value={(filters.pageSize || 20).toString()}
                      onValueChange={handlePageSizeChange}
                    >
                      <SelectTrigger className="h-8 w-[70px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="10">10</SelectItem>
                        <SelectItem value="20">20</SelectItem>
                        <SelectItem value="50">50</SelectItem>
                        <SelectItem value="100">100</SelectItem>
                      </SelectContent>
                    </Select>
                    <div className="flex gap-1">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange((filters.page || 1) - 1)}
                        disabled={(filters.page || 1) === 1 || isLoading}
                      >
                        Sebelumnya
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange((filters.page || 1) + 1)}
                        disabled={(filters.page || 1) >= displayData.pagination.totalPages || isLoading}
                      >
                        Selanjutnya
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
