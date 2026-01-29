/**
 * Purchase Invoices Client Component
 *
 * Client-side interactive component for purchase invoice management.
 * Receives initial server-fetched data and handles:
 * - Interactive search, filters, pagination
 * - RTK Query caching for subsequent requests
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
import { Plus, Search, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
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
import {
  useListPurchaseInvoicesQuery,
  useDeletePurchaseInvoiceMutation,
  useSubmitInvoiceMutation,
  useApprovePurchaseInvoiceMutation,
  useRejectPurchaseInvoiceMutation,
} from "@/store/services/purchaseInvoiceApi";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { toast } from "sonner";
import { usePermissions } from "@/hooks/use-permissions";
import { PurchaseInvoicesTable } from "@/components/invoices/invoices-table";
import type {
  PurchaseInvoiceFilters,
  PurchaseInvoiceListResponse,
  PurchaseInvoiceResponse,
  PurchaseInvoiceStatus,
  PaymentStatus,
} from "@/types/purchase-invoice.types";
import type { RootState } from "@/store";

interface PurchaseInvoicesClientProps {
  initialData: PurchaseInvoiceListResponse;
}

export function PurchaseInvoicesClient({
  initialData,
}: PurchaseInvoicesClientProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    PurchaseInvoiceStatus | undefined
  >(undefined);
  const [paymentStatusFilter, setPaymentStatusFilter] = useState<
    PaymentStatus | undefined
  >(undefined);
  const [filters, setFilters] = useState<PurchaseInvoiceFilters>({
    page: 1,
    page_size: 20,
    sort_by: "invoiceDate",
    sort_order: "desc",
  });
  const [invoiceToDelete, setInvoiceToDelete] = useState<PurchaseInvoiceResponse | null>(null);
  const [invoiceToReject, setInvoiceToReject] = useState<PurchaseInvoiceResponse | null>(null);
  const [rejectionReason, setRejectionReason] = useState("");

  // Get permissions hook
  const permissions = usePermissions();

  // Mutations
  const [deleteInvoice, { isLoading: isDeleting }] = useDeletePurchaseInvoiceMutation();
  const [submitInvoice, { isLoading: isSubmitting }] = useSubmitInvoiceMutation();
  const [approveInvoice, { isLoading: isApproving }] = useApprovePurchaseInvoiceMutation();
  const [rejectInvoice, { isLoading: isRejecting }] = useRejectPurchaseInvoiceMutation();

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Compute permission checks ONCE at top level
  const canCreateInvoices = permissions.canCreate("purchase-invoices");
  const canEditInvoices = permissions.canEdit("purchase-invoices");
  const canApproveInvoices = permissions.canApprove("purchase-invoices");
  const canDeleteInvoices = permissions.canDelete("purchase-invoices");

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch invoices with filters
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const queryParams = {
    ...filters,
    search: debouncedSearch || undefined,
    status: statusFilter,
    payment_status: paymentStatusFilter,
  };

  const {
    data: invoicesData,
    isLoading,
    error,
    refetch,
  } = useListPurchaseInvoicesQuery(queryParams, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Use initialData as fallback only for first render before query completes
  const displayData = invoicesData || initialData;

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
      page_size: parseInt(newPageSize),
      page: 1, // Reset to page 1 when changing page size
    }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sort_by === sortBy) {
        return {
          ...prev,
          sort_order: prev.sort_order === "asc" ? "desc" : "asc",
        } as PurchaseInvoiceFilters;
      }
      // New column, default to descending
      return {
        ...prev,
        sort_by: sortBy as PurchaseInvoiceFilters["sort_by"],
        sort_order: "desc",
      } as PurchaseInvoiceFilters;
    });
  };

  const handleStatusFilterChange = (status: string | undefined) => {
    setStatusFilter(status as PurchaseInvoiceStatus | undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handlePaymentStatusFilterChange = (status: string | undefined) => {
    setPaymentStatusFilter(status as PaymentStatus | undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleClearFilters = () => {
    setSearch("");
    setDebouncedSearch("");
    setStatusFilter(undefined);
    setPaymentStatusFilter(undefined);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleCreateClick = () => {
    router.push("/procurement/invoices/create");
  };

  const handleDeleteClick = (invoice: PurchaseInvoiceResponse) => {
    setInvoiceToDelete(invoice);
  };

  const handleConfirmDelete = async () => {
    if (!invoiceToDelete) return;

    try {
      await deleteInvoice(invoiceToDelete.id).unwrap();
      toast.success("Faktur Berhasil Dihapus", {
        description: `Faktur ${invoiceToDelete.invoiceNumber} telah dihapus`,
      });
      setInvoiceToDelete(null);
      refetch();
    } catch (error: any) {
      toast.error("Gagal Menghapus Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menghapus faktur",
      });
    }
  };

  // Submit invoice for approval
  const handleSubmitInvoice = async (invoice: PurchaseInvoiceResponse) => {
    try {
      await submitInvoice(invoice.id).unwrap();
      toast.success("Faktur Berhasil Disubmit", {
        description: `Faktur ${invoice.invoiceNumber} telah disubmit untuk persetujuan`,
      });
      refetch();
    } catch (error: any) {
      toast.error("Gagal Submit Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat submit faktur",
      });
    }
  };

  // Approve invoice
  const handleApproveInvoice = async (invoice: PurchaseInvoiceResponse) => {
    try {
      await approveInvoice({ invoiceId: invoice.id }).unwrap();
      toast.success("Faktur Berhasil Disetujui", {
        description: `Faktur ${invoice.invoiceNumber} telah disetujui`,
      });
      refetch();
    } catch (error: any) {
      toast.error("Gagal Menyetujui Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menyetujui faktur",
      });
    }
  };

  // Open rejection dialog
  const handleRejectClick = (invoice: PurchaseInvoiceResponse) => {
    setInvoiceToReject(invoice);
    setRejectionReason("");
  };

  // Confirm rejection with reason
  const handleConfirmReject = async () => {
    if (!invoiceToReject) return;

    if (!rejectionReason.trim()) {
      toast.error("Alasan Penolakan Diperlukan", {
        description: "Silakan masukkan alasan penolakan faktur",
      });
      return;
    }

    try {
      await rejectInvoice({
        invoiceId: invoiceToReject.id,
        data: { reason: rejectionReason.trim() },
      }).unwrap();
      toast.success("Faktur Berhasil Ditolak", {
        description: `Faktur ${invoiceToReject.invoiceNumber} telah ditolak`,
      });
      setInvoiceToReject(null);
      setRejectionReason("");
      refetch();
    } catch (error: any) {
      toast.error("Gagal Menolak Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menolak faktur",
      });
    }
  };

  const hasActiveFilters = statusFilter || paymentStatusFilter || search;

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Faktur Pembelian
          </h1>
          <p className="text-muted-foreground">
            Kelola faktur pembelian dari supplier
          </p>
        </div>
        {canCreateInvoices && (
          <Button className="shrink-0" onClick={handleCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Faktur
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
                placeholder="Cari nomor faktur atau nama supplier..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Status Filter */}
            <Select
              value={statusFilter || "all"}
              onValueChange={(value) =>
                handleStatusFilterChange(value === "all" ? undefined : value)
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Semua Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="DRAFT">Draft</SelectItem>
                <SelectItem value="SUBMITTED">Submitted</SelectItem>
                <SelectItem value="APPROVED">Approved</SelectItem>
                <SelectItem value="REJECTED">Rejected</SelectItem>
                <SelectItem value="PAID">Paid</SelectItem>
                <SelectItem value="CANCELLED">Cancelled</SelectItem>
              </SelectContent>
            </Select>

            {/* Payment Status Filter */}
            <Select
              value={paymentStatusFilter || "all"}
              onValueChange={(value) =>
                handlePaymentStatusFilterChange(
                  value === "all" ? undefined : value
                )
              }
            >
              <SelectTrigger className="w-full sm:w-[180px]">
                <SelectValue placeholder="Status Pembayaran" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status Bayar</SelectItem>
                <SelectItem value="UNPAID">Belum Dibayar</SelectItem>
                <SelectItem value="PARTIAL">Dibayar Sebagian</SelectItem>
                <SelectItem value="PAID">Lunas</SelectItem>
                <SelectItem value="OVERDUE">Jatuh Tempo</SelectItem>
              </SelectContent>
            </Select>

            {/* Clear Filters Button */}
            {(statusFilter || paymentStatusFilter || search) && (
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
              <LoadingSpinner size="lg" text="Memuat data faktur..." />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data faktur pembelian"
              onRetry={refetch}
            />
          )}

          {/* Data Display */}
          {!error && displayData && (
            <>
              {(!displayData.data || displayData.data.length === 0) &&
              !hasActiveFilters ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <FileText className="mb-4 h-12 w-12 text-muted-foreground" />
                  <h3 className="mb-2 text-lg font-semibold">
                    Belum ada Faktur Pembelian
                  </h3>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Mulai dengan membuat faktur pembelian pertama Anda
                  </p>
                  {canCreateInvoices && (
                    <Button onClick={handleCreateClick}>
                      <Plus className="mr-2 h-4 w-4" />
                      Tambah Faktur
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

                  {/* Invoices Table */}
                  <PurchaseInvoicesTable
                    invoices={displayData.data || []}
                    sortBy={filters.sort_by}
                    sortOrder={filters.sort_order}
                    onSortChange={handleSortChange}
                    canEdit={canEditInvoices}
                    canApprove={canApproveInvoices}
                    canDelete={canDeleteInvoices}
                    onDelete={handleDeleteClick}
                    onSubmit={handleSubmitInvoice}
                    onApprove={handleApproveInvoice}
                    onReject={handleRejectClick}
                  />

                  {/* Pagination */}
                  {displayData?.pagination && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 ">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const pagination = displayData.pagination as any;
                          const page = pagination.page || 1;
                          const pageSize =
                            pagination.limit || pagination.page_size || 20;
                          const totalItems =
                            pagination.total || pagination.totalItems || 0;
                          const start = (page - 1) * pageSize + 1;
                          const end = Math.min(page * pageSize, totalItems);
                          return `Menampilkan ${start}-${end} dari ${totalItems} faktur`;
                        })()}
                      </div>

                      {/* 2. Page Size Selector */}
                      <div className="flex items-center justify-center sm:justify-start gap-2">
                        <span className="text-sm text-muted-foreground whitespace-nowrap">
                          Baris per Halaman
                        </span>
                        <Select
                          value={filters.page_size?.toString() || "20"}
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

                      {/* 3. Navigation Buttons */}
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
      {/* Delete Confirmation Dialog */}
      <AlertDialog open={!!invoiceToDelete} onOpenChange={(open) => !open && setInvoiceToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Faktur?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menghapus faktur <strong>{invoiceToDelete?.invoiceNumber}</strong>?
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? "Menghapus..." : "Hapus"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Reject Confirmation Dialog */}
      <AlertDialog open={!!invoiceToReject} onOpenChange={(open) => !open && setInvoiceToReject(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Tolak Faktur?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menolak faktur <strong>{invoiceToReject?.invoiceNumber}</strong>?
              Silakan masukkan alasan penolakan di bawah ini.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="rejection-reason" className="mb-2 block">
              Alasan Penolakan <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="rejection-reason"
              placeholder="Masukkan alasan penolakan faktur..."
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              rows={3}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmReject}
              disabled={isRejecting || !rejectionReason.trim()}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isRejecting ? "Menolak..." : "Tolak Faktur"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
