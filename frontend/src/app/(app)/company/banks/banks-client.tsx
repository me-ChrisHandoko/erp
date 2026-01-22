/**
 * Banks Client Component
 *
 * Client-side interactive component for bank account management.
 * Receives initial server-fetched data and handles:
 * - Interactive CRUD operations (add, edit, delete)
 * - RTK Query caching for subsequent requests
 * - Company switch handling with explicit refetch
 * - Pagination, search, and filtering
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { useGetBankAccountsQuery } from "@/store/services/companyApi";
import { BankAccountTable } from "@/components/company/bank-account-table";
import { BankAccountForm } from "@/components/company/bank-account-form";
import { ErrorDisplay } from "@/components/shared/error-display";
import { EmptyState } from "@/components/shared/empty-state";
import type {
  BankAccountResponse,
  BankAccountFilters,
  BankAccountListResponse,
} from "@/types/company.types";
import type { RootState } from "@/store";

interface BanksClientProps {
  initialData: BankAccountListResponse;
}

export function BanksClient({ initialData }: BanksClientProps) {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [filters, setFilters] = useState<BankAccountFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "bankName",
    sortOrder: "asc",
  });

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Fetch banks with RTK Query
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const {
    data: banksResponse,
    isLoading,
    error,
    refetch,
  } = useGetBankAccountsQuery(filters, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = banksResponse?.data || initialData.data;
  const pagination = banksResponse?.pagination || initialData.pagination;

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

  return (
    <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
      {/* Page Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Rekening Bank</h1>
          <p className="text-muted-foreground">
            Kelola rekening bank perusahaan untuk transaksi dan invoice
          </p>
        </div>
        <Button
          onClick={() => setIsAddDialogOpen(true)}
          disabled={isLoading}
          className="shrink-0"
        >
          <Plus className="mr-2 h-4 w-4" />
          Tambah Rekening
        </Button>
      </div>

      {/* Bank Accounts Card */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Subtle loading indicator for refetching */}
          {isLoading && displayData && displayData.length > 0 && (
            <div className="text-sm text-muted-foreground text-center py-2 mb-4">
              Memperbarui data...
            </div>
          )}

          {/* Loading State - Only show if no initial data */}
          {isLoading && !displayData && (
            <div className="space-y-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              title="Gagal Memuat Data"
              error={error}
              onRetry={refetch}
            />
          )}

          {/* Empty State */}
          {!isLoading && !error && displayData && displayData.length === 0 && (
            <EmptyState
              title="Belum Ada Rekening Bank"
              description="Tambahkan rekening bank pertama perusahaan Anda untuk mulai melakukan transaksi."
              action={{
                label: "Tambah Rekening Bank",
                onClick: () => setIsAddDialogOpen(true),
              }}
            />
          )}

          {/* Bank Accounts Table with Pagination */}
          {!error && displayData && displayData.length > 0 && (
            <div className="space-y-4">
              {/* Bank Accounts Table */}
              <BankAccountTable banks={displayData} />

              {/* Pagination */}
              {pagination && (
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  {/* 1. Summary - Record Data */}
                  <div className="text-sm text-muted-foreground text-center sm:text-left">
                    {(() => {
                      const start =
                        (pagination.page - 1) * pagination.limit + 1;
                      const end = Math.min(
                        pagination.page * pagination.limit,
                        pagination.total
                      );
                      return `Menampilkan ${start}-${end} dari ${pagination.total} item`;
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
                      disabled={pagination.page === 1}
                    >
                      &laquo;
                    </Button>

                    {/* Previous Page */}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.page - 1)}
                      disabled={pagination.page === 1}
                    >
                      &lsaquo;
                    </Button>

                    {/* Current Page Info */}
                    <span className="text-sm text-muted-foreground px-2">
                      Halaman {pagination.page} dari {pagination.totalPages}
                    </span>

                    {/* Next Page */}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.page + 1)}
                      disabled={pagination.page >= pagination.totalPages}
                    >
                      &rsaquo;
                    </Button>

                    {/* Last Page */}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.totalPages)}
                      disabled={pagination.page >= pagination.totalPages}
                    >
                      &raquo;
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Information Card */}
      <Card>
        <CardHeader>
          <CardTitle>Informasi Rekening Bank</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Rekening Utama</h4>
              <p className="text-sm text-muted-foreground">
                Rekening yang ditandai sebagai utama akan digunakan secara
                otomatis untuk transaksi dan invoice. Hanya satu rekening yang
                bisa menjadi rekening utama.
              </p>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Validasi Rekening</h4>
              <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
                <li>Nomor rekening minimal 8 digit, hanya angka</li>
                <li>Nama pemilik rekening minimal 3 karakter</li>
                <li>Minimal harus ada 1 rekening bank aktif</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Prefix Cek</h4>
              <p className="text-sm text-muted-foreground">
                Prefix cek digunakan untuk menghasilkan nomor cek otomatis
                (contoh: CHK-001, BNI-001). Field ini bersifat opsional.
              </p>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Menghapus Rekening</h4>
              <p className="text-sm text-muted-foreground">
                Anda tidak bisa menghapus rekening terakhir. Minimal harus ada 1
                rekening bank yang aktif di sistem.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Add Bank Dialog */}
      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Tambah Rekening Bank</DialogTitle>
            <DialogDescription>
              Isi formulir di bawah ini untuk menambahkan rekening bank baru
              perusahaan
            </DialogDescription>
          </DialogHeader>
          <BankAccountForm
            onSuccess={() => setIsAddDialogOpen(false)}
            onCancel={() => setIsAddDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}
