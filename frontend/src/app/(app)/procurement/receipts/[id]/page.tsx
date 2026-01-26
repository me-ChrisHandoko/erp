/**
 * Goods Receipt Detail Page
 *
 * Displays comprehensive goods receipt information including:
 * - Receipt information (GRN number, date, status)
 * - Supplier and warehouse details
 * - Line items with quantities
 * - Workflow actions (receive, inspect, accept, reject)
 */

"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  PackageCheck,
  AlertCircle,
  ArrowLeft,
  CheckCircle,
  XCircle,
  Search,
  ClipboardCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { ReceiptDetail } from "@/components/goods-receipts/receipt-detail";
import {
  useGetGoodsReceiptQuery,
  useReceiveGoodsMutation,
  useInspectGoodsMutation,
  useAcceptGoodsMutation,
  useRejectGoodsMutation,
  useUpdateRejectionDispositionMutation,
  useResolveDispositionMutation,
} from "@/store/services/goodsReceiptApi";
import { usePermissions } from "@/hooks/use-permissions";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { InspectGoodsDialog } from "@/components/goods-receipts/inspect-goods-dialog";
import { RejectionDispositionDialog } from "@/components/goods-receipts/rejection-disposition-dialog";
import { ResolveDispositionDialog } from "@/components/goods-receipts/resolve-disposition-dialog";
import type { InspectGoodsRequest, GoodsReceiptItemResponse, RejectionDispositionStatus } from "@/types/goods-receipt.types";

export default function GoodsReceiptDetailPage() {
  const params = useParams();
  const router = useRouter();
  const receiptId = params.id as string;

  const permissions = usePermissions();
  const canEdit = permissions.canEdit("goods-receipts");

  const { data, isLoading, error } = useGetGoodsReceiptQuery(receiptId);

  // Dialog states
  const [receiveDialogOpen, setReceiveDialogOpen] = useState(false);
  const [inspectDialogOpen, setInspectDialogOpen] = useState(false);
  const [acceptDialogOpen, setAcceptDialogOpen] = useState(false);
  const [rejectDialogOpen, setRejectDialogOpen] = useState(false);
  const [rejectionReason, setRejectionReason] = useState("");
  const [notes, setNotes] = useState("");

  // Disposition dialog states
  const [dispositionDialogOpen, setDispositionDialogOpen] = useState(false);
  const [resolveDispositionDialogOpen, setResolveDispositionDialogOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<GoodsReceiptItemResponse | null>(null);

  // Mutations
  const [receiveGoods, { isLoading: isReceiving }] = useReceiveGoodsMutation();
  const [inspectGoods, { isLoading: isInspecting }] = useInspectGoodsMutation();
  const [acceptGoods, { isLoading: isAccepting }] = useAcceptGoodsMutation();
  const [rejectGoods, { isLoading: isRejecting }] = useRejectGoodsMutation();
  const [updateRejectionDisposition, { isLoading: isSettingDisposition }] = useUpdateRejectionDispositionMutation();
  const [resolveDisposition, { isLoading: isResolvingDisposition }] = useResolveDispositionMutation();

  const handleReceiveGoods = async () => {
    if (!data) return;

    try {
      await receiveGoods({
        id: receiptId,
        data: notes ? { notes } : undefined,
      }).unwrap();
      toast.success("Barang Diterima", {
        description: `${data.grnNumber} berhasil diterima`,
      });
      setReceiveDialogOpen(false);
      setNotes("");
    } catch (error: any) {
      toast.error("Gagal Menerima Barang", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleInspectGoods = async (inspectData: InspectGoodsRequest) => {
    if (!data) return;

    try {
      await inspectGoods({
        id: receiptId,
        data: inspectData,
      }).unwrap();
      toast.success("Barang Diperiksa", {
        description: `${data.grnNumber} berhasil diperiksa`,
      });
      setInspectDialogOpen(false);
    } catch (error: any) {
      toast.error("Gagal Memeriksa Barang", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleAcceptGoods = async () => {
    if (!data) return;

    try {
      await acceptGoods({
        id: receiptId,
        data: notes ? { notes } : undefined,
      }).unwrap();
      toast.success("Barang Disetujui", {
        description: `${data.grnNumber} berhasil disetujui dan stok telah diperbarui`,
      });
      setAcceptDialogOpen(false);
      setNotes("");
    } catch (error: any) {
      toast.error("Gagal Menyetujui Barang", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleRejectGoods = async () => {
    if (!data || !rejectionReason.trim()) return;

    try {
      await rejectGoods({
        id: receiptId,
        data: { rejectionReason: rejectionReason.trim() },
      }).unwrap();
      toast.success("Barang Ditolak", {
        description: `${data.grnNumber} ditolak`,
      });
      setRejectDialogOpen(false);
      setRejectionReason("");
    } catch (error: any) {
      toast.error("Gagal Menolak Barang", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  // Handler for opening disposition dialog
  const handleOpenDispositionDialog = (item: GoodsReceiptItemResponse) => {
    setSelectedItem(item);
    setDispositionDialogOpen(true);
  };

  // Handler for opening resolve disposition dialog
  const handleOpenResolveDialog = (item: GoodsReceiptItemResponse) => {
    setSelectedItem(item);
    setResolveDispositionDialogOpen(true);
  };

  // Handler for setting disposition
  const handleSetDisposition = async (dispositionData: { rejectionDisposition: RejectionDispositionStatus; dispositionNotes?: string }) => {
    if (!data || !selectedItem) return;

    try {
      await updateRejectionDisposition({
        goodsReceiptId: receiptId,
        itemId: selectedItem.id,
        data: dispositionData,
      }).unwrap();
      toast.success("Disposisi Diatur", {
        description: `Disposisi untuk ${selectedItem.product?.name || "item"} berhasil diatur`,
      });
      setDispositionDialogOpen(false);
      setSelectedItem(null);
    } catch (error: any) {
      toast.error("Gagal Mengatur Disposisi", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  // Handler for resolving disposition
  const handleResolveDisposition = async (resolveData: { dispositionResolvedNotes?: string }) => {
    if (!data || !selectedItem) return;

    try {
      await resolveDisposition({
        goodsReceiptId: receiptId,
        itemId: selectedItem.id,
        data: resolveData,
      }).unwrap();
      toast.success("Disposisi Diselesaikan", {
        description: `Disposisi untuk ${selectedItem.product?.name || "item"} berhasil diselesaikan`,
      });
      setResolveDispositionDialogOpen(false);
      setSelectedItem(null);
    } catch (error: any) {
      toast.error("Gagal Menyelesaikan Disposisi", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
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
            { label: "Pembelian", href: "/procurement" },
            { label: "Penerimaan Barang", href: "/procurement/receipts" },
            { label: "Detail GRN" },
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
            { label: "Pembelian", href: "/procurement" },
            { label: "Penerimaan Barang", href: "/procurement/receipts" },
            { label: "Detail GRN" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data penerimaan barang" : "Penerimaan barang tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push("/procurement/receipts")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar GRN
          </Button>
        </div>
      </div>
    );
  }

  const receipt = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement" },
          { label: "Penerimaan Barang", href: "/procurement/receipts" },
          { label: "Detail GRN" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <PackageCheck className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">{receipt.grnNumber}</h1>
            </div>
            <p className="text-muted-foreground">
              {receipt.supplier?.name || "Supplier"} &bull;{" "}
              {receipt.warehouse?.name || "Gudang"}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/procurement/receipts")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>

            {/* Receive - Only for PENDING */}
            {canEdit && receipt.status === "PENDING" && (
              <Button
                className="shrink-0 bg-blue-600 hover:bg-blue-700"
                onClick={() => setReceiveDialogOpen(true)}
              >
                <PackageCheck className="mr-2 h-4 w-4" />
                Terima Barang
              </Button>
            )}

            {/* Inspect - Only for RECEIVED */}
            {canEdit && receipt.status === "RECEIVED" && (
              <Button
                className="shrink-0 bg-purple-600 hover:bg-purple-700"
                onClick={() => setInspectDialogOpen(true)}
              >
                <Search className="mr-2 h-4 w-4" />
                Periksa Barang
              </Button>
            )}

            {/* Accept - Only for INSPECTED */}
            {canEdit && receipt.status === "INSPECTED" && (
              <Button
                className="shrink-0 bg-green-600 hover:bg-green-700"
                onClick={() => setAcceptDialogOpen(true)}
              >
                <CheckCircle className="mr-2 h-4 w-4" />
                Setujui Barang
              </Button>
            )}

            {/* Reject - Only for INSPECTED */}
            {canEdit && receipt.status === "INSPECTED" && (
              <Button
                variant="destructive"
                className="shrink-0"
                onClick={() => setRejectDialogOpen(true)}
              >
                <XCircle className="mr-2 h-4 w-4" />
                Tolak Barang
              </Button>
            )}
          </div>
        </div>

        {/* Receipt Detail Component */}
        <ReceiptDetail
          receipt={receipt}
          onSetDisposition={canEdit ? handleOpenDispositionDialog : undefined}
          onResolveDisposition={canEdit ? handleOpenResolveDialog : undefined}
        />
      </div>

      {/* Receive Dialog */}
      <AlertDialog open={receiveDialogOpen} onOpenChange={setReceiveDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Terima Barang</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menerima barang untuk{" "}
              <span className="font-semibold">{receipt.grnNumber}</span>?
              <br />
              <br />
              Setelah diterima, status akan berubah menjadi RECEIVED.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="receiveNotes">Catatan Penerimaan (Opsional)</Label>
            <Textarea
              id="receiveNotes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Masukkan catatan penerimaan barang..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isReceiving}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleReceiveGoods}
              disabled={isReceiving}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isReceiving ? "Memproses..." : "Terima Barang"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Inspect Dialog */}
      <InspectGoodsDialog
        open={inspectDialogOpen}
        onOpenChange={setInspectDialogOpen}
        receipt={receipt}
        onSubmit={handleInspectGoods}
        isLoading={isInspecting}
      />

      {/* Accept Dialog */}
      <AlertDialog open={acceptDialogOpen} onOpenChange={setAcceptDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Setujui Barang</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menyetujui barang untuk{" "}
              <span className="font-semibold">{receipt.grnNumber}</span>?
              <br />
              <br />
              Setelah disetujui, stok gudang akan otomatis diperbarui.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="acceptNotes">Catatan Persetujuan (Opsional)</Label>
            <Textarea
              id="acceptNotes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Masukkan catatan persetujuan barang..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isAccepting}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleAcceptGoods}
              disabled={isAccepting}
              className="bg-green-600 hover:bg-green-700"
            >
              {isAccepting ? "Memproses..." : "Setujui Barang"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Reject Dialog */}
      <AlertDialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Tolak Barang</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menolak barang untuk{" "}
              <span className="font-semibold">{receipt.grnNumber}</span>?
              <br />
              <br />
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="rejectionReason">
              Alasan Penolakan <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="rejectionReason"
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              placeholder="Masukkan alasan penolakan..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRejecting}>Kembali</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRejectGoods}
              disabled={isRejecting || !rejectionReason.trim()}
              className="bg-red-600 hover:bg-red-700"
            >
              {isRejecting ? "Memproses..." : "Tolak Barang"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Rejection Disposition Dialog */}
      <RejectionDispositionDialog
        open={dispositionDialogOpen}
        onOpenChange={setDispositionDialogOpen}
        item={selectedItem}
        onSubmit={handleSetDisposition}
        isLoading={isSettingDisposition}
      />

      {/* Resolve Disposition Dialog */}
      <ResolveDispositionDialog
        open={resolveDispositionDialogOpen}
        onOpenChange={setResolveDispositionDialogOpen}
        item={selectedItem}
        onSubmit={handleResolveDisposition}
        isLoading={isResolvingDisposition}
      />
    </div>
  );
}
