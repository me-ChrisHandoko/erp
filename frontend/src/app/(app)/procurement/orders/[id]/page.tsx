/**
 * Purchase Order Detail Page
 *
 * Displays comprehensive purchase order information including:
 * - Order information (number, date, status)
 * - Supplier and warehouse details
 * - Line items with quantities and prices
 * - Workflow actions (confirm, cancel)
 */

"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  ShoppingCart,
  Edit,
  AlertCircle,
  ArrowLeft,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { OrderDetail } from "@/components/procurement/order-detail";
import {
  useGetPurchaseOrderQuery,
  useConfirmPurchaseOrderMutation,
  useCancelPurchaseOrderMutation,
} from "@/store/services/purchaseOrderApi";
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
import { toast } from "sonner";

export default function PurchaseOrderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.id as string;

  const permissions = usePermissions();
  const canEdit = permissions.canEdit("purchase-orders");
  const canConfirm = permissions.canEdit("purchase-orders");
  const canCancel = permissions.canDelete("purchase-orders");

  const { data, isLoading, error } = useGetPurchaseOrderQuery(orderId);

  // Confirm dialog state
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);

  // Cancel dialog state
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [cancellationNote, setCancellationNote] = useState("");

  // Mutations
  const [confirmOrder, { isLoading: isConfirming }] = useConfirmPurchaseOrderMutation();
  const [cancelOrder, { isLoading: isCancelling }] = useCancelPurchaseOrderMutation();

  const handleConfirmOrder = async () => {
    if (!data) return;

    try {
      await confirmOrder({ id: orderId }).unwrap();
      toast.success("PO Dikonfirmasi", {
        description: `${data.poNumber} berhasil dikonfirmasi`,
      });
      setConfirmDialogOpen(false);
    } catch (error: any) {
      toast.error("Gagal Mengkonfirmasi PO", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleCancelOrder = async () => {
    if (!data || !cancellationNote.trim()) return;

    try {
      await cancelOrder({
        id: orderId,
        data: { cancellationNote: cancellationNote.trim() },
      }).unwrap();
      toast.success("PO Dibatalkan", {
        description: `${data.poNumber} berhasil dibatalkan`,
      });
      setCancelDialogOpen(false);
      setCancellationNote("");
    } catch (error: any) {
      toast.error("Gagal Membatalkan PO", {
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
            { label: "Procurement", href: "/procurement" },
            { label: "Purchase Orders", href: "/procurement/orders" },
            { label: "Detail PO" },
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
            { label: "Procurement", href: "/procurement" },
            { label: "Purchase Orders", href: "/procurement/orders" },
            { label: "Detail PO" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data purchase order" : "Purchase order tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push("/procurement/orders")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar PO
          </Button>
        </div>
      </div>
    );
  }

  const order = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Procurement", href: "/procurement" },
          { label: "Purchase Orders", href: "/procurement/orders" },
          { label: "Detail PO" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <ShoppingCart className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">{order.poNumber}</h1>
            </div>
            <p className="text-muted-foreground">
              {order.supplier?.name || "Supplier"} &bull;{" "}
              {order.warehouse?.name || "Gudang"}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/procurement/orders")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>

            {/* Edit - Only for DRAFT */}
            {canEdit && order.status === "DRAFT" && (
              <Button
                variant="outline"
                className="shrink-0"
                onClick={() => router.push(`/procurement/orders/${orderId}/edit`)}
              >
                <Edit className="mr-2 h-4 w-4" />
                Edit PO
              </Button>
            )}

            {/* Confirm - Only for DRAFT */}
            {canConfirm && order.status === "DRAFT" && (
              <Button
                className="shrink-0 bg-blue-600 hover:bg-blue-700"
                onClick={() => setConfirmDialogOpen(true)}
              >
                <CheckCircle className="mr-2 h-4 w-4" />
                Konfirmasi PO
              </Button>
            )}

            {/* Cancel - Only for DRAFT or CONFIRMED */}
            {canCancel &&
              (order.status === "DRAFT" || order.status === "CONFIRMED") && (
                <Button
                  variant="destructive"
                  className="shrink-0"
                  onClick={() => setCancelDialogOpen(true)}
                >
                  <XCircle className="mr-2 h-4 w-4" />
                  Batalkan PO
                </Button>
              )}
          </div>
        </div>

        {/* Order Detail Component */}
        <OrderDetail order={order} />
      </div>

      {/* Confirm Dialog */}
      <AlertDialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Konfirmasi Purchase Order</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin mengkonfirmasi{" "}
              <span className="font-semibold">{order.poNumber}</span>?
              <br />
              <br />
              Setelah dikonfirmasi, PO tidak dapat diedit lagi dan menunggu
              penerimaan barang.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isConfirming}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmOrder}
              disabled={isConfirming}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isConfirming ? "Memproses..." : "Konfirmasi"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Cancel Dialog */}
      <AlertDialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Purchase Order</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin membatalkan{" "}
              <span className="font-semibold">{order.poNumber}</span>?
              <br />
              <br />
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="cancellationNote">
              Alasan Pembatalan <span className="text-destructive">*</span>
            </Label>
            <Input
              id="cancellationNote"
              value={cancellationNote}
              onChange={(e) => setCancellationNote(e.target.value)}
              placeholder="Masukkan alasan pembatalan..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isCancelling}>Kembali</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelOrder}
              disabled={isCancelling || !cancellationNote.trim()}
              className="bg-red-600 hover:bg-red-700"
            >
              {isCancelling ? "Memproses..." : "Batalkan PO"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
