/**
 * Purchase Invoice Detail Page
 *
 * Displays comprehensive invoice information including:
 * - Invoice header (number, dates, supplier)
 * - Invoice items with pricing
 * - Payment information
 * - Status and workflow actions
 * - Financial summary (subtotal, tax, total)
 */

"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { FileText, Edit, AlertCircle, ArrowLeft, Trash2, Send, CheckCircle, XCircle, Ban } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { InvoiceDetail } from "@/components/invoices/invoice-detail";
import {
  useGetPurchaseInvoiceQuery,
  useDeletePurchaseInvoiceMutation,
  useSubmitInvoiceMutation,
  useApprovePurchaseInvoiceMutation,
  useRejectPurchaseInvoiceMutation,
  useCancelInvoiceMutation,
} from "@/store/services/purchaseInvoiceApi";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
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

export default function PurchaseInvoiceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const invoiceId = params.id as string;
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showRejectDialog, setShowRejectDialog] = useState(false);
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [rejectionReason, setRejectionReason] = useState("");
  const [cancellationReason, setCancellationReason] = useState("");

  const { data, isLoading, error, refetch } = useGetPurchaseInvoiceQuery(invoiceId);
  const [deleteInvoice, { isLoading: isDeleting }] = useDeletePurchaseInvoiceMutation();
  const [submitInvoice, { isLoading: isSubmitting }] = useSubmitInvoiceMutation();
  const [approveInvoice, { isLoading: isApproving }] = useApprovePurchaseInvoiceMutation();
  const [rejectInvoice, { isLoading: isRejecting }] = useRejectPurchaseInvoiceMutation();
  const [cancelInvoice, { isLoading: isCancelling }] = useCancelInvoiceMutation();
  const permissions = usePermissions();
  const canDelete = permissions.canDelete("purchase-invoices");
  const canApprove = permissions.canApprove("purchase-invoices");

  const handleDelete = async () => {
    try {
      await deleteInvoice(invoiceId).unwrap();
      toast.success("Faktur Berhasil Dihapus", {
        description: `Faktur ${data?.invoiceNumber} telah dihapus`,
      });
      router.push("/procurement/invoices");
    } catch (error: any) {
      toast.error("Gagal Menghapus Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menghapus faktur",
      });
    }
    setShowDeleteDialog(false);
  };

  // Submit invoice for approval
  const handleSubmit = async () => {
    try {
      await submitInvoice(invoiceId).unwrap();
      toast.success("Faktur Berhasil Disubmit", {
        description: `Faktur ${data?.invoiceNumber} telah disubmit untuk persetujuan`,
      });
      refetch();
    } catch (error: any) {
      toast.error("Gagal Submit Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat submit faktur",
      });
    }
  };

  // Approve invoice
  const handleApprove = async () => {
    try {
      await approveInvoice({ invoiceId }).unwrap();
      toast.success("Faktur Berhasil Disetujui", {
        description: `Faktur ${data?.invoiceNumber} telah disetujui`,
      });
      refetch();
    } catch (error: any) {
      toast.error("Gagal Menyetujui Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menyetujui faktur",
      });
    }
  };

  // Reject invoice
  const handleReject = async () => {
    if (!rejectionReason.trim()) {
      toast.error("Alasan Penolakan Diperlukan", {
        description: "Silakan masukkan alasan penolakan faktur",
      });
      return;
    }

    try {
      await rejectInvoice({
        invoiceId,
        data: { reason: rejectionReason.trim() },
      }).unwrap();
      toast.success("Faktur Berhasil Ditolak", {
        description: `Faktur ${data?.invoiceNumber} telah ditolak`,
      });
      setShowRejectDialog(false);
      setRejectionReason("");
      refetch();
    } catch (error: any) {
      toast.error("Gagal Menolak Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat menolak faktur",
      });
    }
  };

  // Cancel approved invoice
  const handleCancel = async () => {
    if (!cancellationReason.trim()) {
      toast.error("Alasan Pembatalan Diperlukan", {
        description: "Silakan masukkan alasan pembatalan faktur",
      });
      return;
    }

    try {
      await cancelInvoice({
        invoiceId,
        data: { reason: cancellationReason.trim() },
      }).unwrap();
      toast.success("Faktur Berhasil Dibatalkan", {
        description: `Faktur ${data?.invoiceNumber} telah dibatalkan`,
      });
      setShowCancelDialog(false);
      setCancellationReason("");
      refetch();
    } catch (error: any) {
      toast.error("Gagal Membatalkan Faktur", {
        description: error?.data?.error?.message || error?.data?.message || "Terjadi kesalahan saat membatalkan faktur",
      });
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Detail Faktur" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !data) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Detail Faktur" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data faktur" : "Faktur tidak ditemukan"}
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  const invoice = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Faktur Pembelian", href: "/procurement/invoices" },
          { label: "Detail Faktur" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                Faktur {invoice.invoiceNumber}
              </h1>
            </div>
            <p className="text-muted-foreground">
              Supplier: <span className="font-semibold">{invoice.supplierName}</span>
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/procurement/invoices")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            {invoice.status === "DRAFT" && (
              <>
                <Button
                  className="shrink-0"
                  onClick={() => router.push(`/procurement/invoices/${invoiceId}/edit`)}
                >
                  <Edit className="mr-2 h-4 w-4" />
                  Edit Faktur
                </Button>
                <Button
                  variant="secondary"
                  className="shrink-0"
                  onClick={handleSubmit}
                  disabled={isSubmitting}
                >
                  <Send className="mr-2 h-4 w-4" />
                  {isSubmitting ? "Mengirim..." : "Submit untuk Persetujuan"}
                </Button>
              </>
            )}
            {canApprove && invoice.status === "SUBMITTED" && (
              <>
                <Button
                  variant="default"
                  className="shrink-0 bg-green-600 hover:bg-green-700"
                  onClick={handleApprove}
                  disabled={isApproving}
                >
                  <CheckCircle className="mr-2 h-4 w-4" />
                  {isApproving ? "Menyetujui..." : "Setujui Faktur"}
                </Button>
                <Button
                  variant="destructive"
                  className="shrink-0"
                  onClick={() => setShowRejectDialog(true)}
                >
                  <XCircle className="mr-2 h-4 w-4" />
                  Tolak Faktur
                </Button>
              </>
            )}
            {canApprove && invoice.status === "APPROVED" && Number(invoice.paidAmount || 0) === 0 && (
              <Button
                variant="destructive"
                className="shrink-0"
                onClick={() => setShowCancelDialog(true)}
              >
                <Ban className="mr-2 h-4 w-4" />
                Batalkan Faktur
              </Button>
            )}
            {canDelete && invoice.status === "DRAFT" && Number(invoice.paidAmount || 0) === 0 && (
              <Button
                variant="destructive"
                className="shrink-0"
                onClick={() => setShowDeleteDialog(true)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Hapus Faktur
              </Button>
            )}
          </div>
        </div>

        {/* Invoice Detail Component */}
        <InvoiceDetail invoice={invoice} />
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Faktur?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menghapus faktur <strong>{invoice.invoiceNumber}</strong>?
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? "Menghapus..." : "Hapus"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Reject Confirmation Dialog */}
      <AlertDialog open={showRejectDialog} onOpenChange={setShowRejectDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Tolak Faktur?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menolak faktur <strong>{invoice.invoiceNumber}</strong>?
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
            <AlertDialogCancel onClick={() => setRejectionReason("")}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleReject}
              disabled={isRejecting || !rejectionReason.trim()}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isRejecting ? "Menolak..." : "Tolak Faktur"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Cancel Confirmation Dialog */}
      <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Faktur?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin membatalkan faktur <strong>{invoice.invoiceNumber}</strong>?
              Faktur yang sudah disetujui akan dibatalkan dan kuantitas akan dikembalikan ke PO.
              Silakan masukkan alasan pembatalan di bawah ini.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="cancellation-reason" className="mb-2 block">
              Alasan Pembatalan <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="cancellation-reason"
              placeholder="Masukkan alasan pembatalan faktur..."
              value={cancellationReason}
              onChange={(e) => setCancellationReason(e.target.value)}
              rows={3}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setCancellationReason("")}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancel}
              disabled={isCancelling || !cancellationReason.trim()}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isCancelling ? "Membatalkan..." : "Batalkan Faktur"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
