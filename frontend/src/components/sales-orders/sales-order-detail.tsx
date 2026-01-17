/**
 * Sales Order Detail Component
 *
 * Displays complete sales order information including:
 * - Order header (number, date, status)
 * - Customer and warehouse details
 * - Order items with quantities and prices
 * - Financial summary
 * - Actions (edit, cancel, etc.)
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Pencil, XCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
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
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatCurrency, formatDate } from "@/lib/utils";
import { usePermissions } from "@/hooks/use-permissions";
import { useCancelSalesOrderMutation } from "@/store/services/salesOrderApi";
import { toast } from "sonner";
import type { SalesOrderResponse } from "@/types/sales-order.types";
import {
  SALES_ORDER_STATUS_LABELS,
  SALES_ORDER_STATUS_STYLES,
  canEditOrder,
  canCancelOrder,
} from "@/types/sales-order.types";

interface SalesOrderDetailProps {
  salesOrder: SalesOrderResponse;
}

export function SalesOrderDetail({ salesOrder }: SalesOrderDetailProps) {
  const router = useRouter();
  const permissions = usePermissions();
  const [cancelSalesOrder, { isLoading: isCancelling }] = useCancelSalesOrderMutation();

  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [cancelReason, setCancelReason] = useState("");

  const canEdit = permissions.canEdit("sales-orders");
  const canCancel = permissions.canDelete("sales-orders");

  const handleCancelOrder = async () => {
    if (!cancelReason.trim() || cancelReason.trim().length < 5) {
      toast.error("Alasan Tidak Valid", {
        description: "Mohon masukkan alasan pembatalan minimal 5 karakter",
      });
      return;
    }

    try {
      await cancelSalesOrder({
        id: salesOrder.id,
        reason: cancelReason,
      }).unwrap();

      toast.success("Pesanan Dibatalkan", {
        description: `Pesanan ${salesOrder.orderNumber} berhasil dibatalkan`,
      });

      setShowCancelDialog(false);
      setCancelReason("");

      // Refresh page to show updated status
      router.refresh();
    } catch (error: any) {
      toast.error("Gagal Membatalkan Pesanan", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          "Terjadi kesalahan saat membatalkan pesanan",
      });
    }
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Header with Back Button and Actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/sales/orders")}
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Kembali
            </Button>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">
            {salesOrder.orderNumber}
          </h1>
          <div className="flex items-center gap-2">
            <Badge className={SALES_ORDER_STATUS_STYLES[salesOrder.status]}>
              {SALES_ORDER_STATUS_LABELS[salesOrder.status]}
            </Badge>
            <span className="text-sm text-muted-foreground">
              {formatDate(salesOrder.orderDate)}
            </span>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2">
          {canEdit && canEditOrder(salesOrder.status) && (
            <Button
              onClick={() =>
                router.push(`/sales/orders/${salesOrder.id}/edit`)
              }
            >
              <Pencil className="mr-2 h-4 w-4" />
              Edit Pesanan
            </Button>
          )}
          {canCancel && canCancelOrder(salesOrder.status) && (
            <Button
              variant="destructive"
              onClick={() => setShowCancelDialog(true)}
              disabled={isCancelling}
            >
              {isCancelling ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Membatalkan...
                </>
              ) : (
                <>
                  <XCircle className="mr-2 h-4 w-4" />
                  Batalkan
                </>
              )}
            </Button>
          )}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {/* Customer Information */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Pelanggan</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div>
              <div className="text-sm text-muted-foreground">Nama</div>
              <div className="font-medium">{salesOrder.customerName}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Kode</div>
              <div className="font-medium font-mono text-sm">
                {salesOrder.customerCode}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Warehouse Information */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Gudang</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div>
              <div className="text-sm text-muted-foreground">Nama</div>
              <div className="font-medium">{salesOrder.warehouseName}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Kode</div>
              <div className="font-medium font-mono text-sm">
                {salesOrder.warehouseCode}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Order Dates */}
      <Card>
        <CardHeader>
          <CardTitle>Tanggal</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4">
            <div>
              <div className="text-sm text-muted-foreground">
                Tanggal Pesanan
              </div>
              <div className="font-medium">
                {formatDate(salesOrder.orderDate)}
              </div>
            </div>
            {salesOrder.requiredDate && (
              <div>
                <div className="text-sm text-muted-foreground">
                  Tanggal Dibutuhkan
                </div>
                <div className="font-medium">
                  {formatDate(salesOrder.requiredDate)}
                </div>
              </div>
            )}
            {salesOrder.deliveryDate && (
              <div>
                <div className="text-sm text-muted-foreground">
                  Tanggal Pengiriman
                </div>
                <div className="font-medium">
                  {formatDate(salesOrder.deliveryDate)}
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Order Items */}
      <Card>
        <CardHeader>
          <CardTitle>Item Pesanan</CardTitle>
        </CardHeader>
        <CardContent>
          {salesOrder.items && salesOrder.items.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>No</TableHead>
                    <TableHead>Produk</TableHead>
                    <TableHead>Unit</TableHead>
                    <TableHead className="text-right">Kuantitas</TableHead>
                    <TableHead className="text-right">Harga</TableHead>
                    <TableHead className="text-right">Diskon</TableHead>
                    <TableHead className="text-right">Total</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {salesOrder.items.map((item, index) => (
                    <TableRow key={item.id}>
                      <TableCell>{index + 1}</TableCell>
                      <TableCell>
                        <div className="font-medium">{item.productName}</div>
                        <div className="text-sm text-muted-foreground font-mono">
                          {item.productCode}
                        </div>
                      </TableCell>
                      <TableCell>{item.unitName}</TableCell>
                      <TableCell className="text-right">
                        {item.orderedQty}
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(item.unitPrice)}
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(item.discount)}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {formatCurrency(item.lineTotal)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              Tidak ada item dalam pesanan ini
            </div>
          )}
        </CardContent>
      </Card>

      {/* Financial Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Ringkasan Keuangan</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Subtotal</span>
            <span className="font-medium">
              {formatCurrency(salesOrder.subtotal)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Diskon</span>
            <span className="font-medium">
              {formatCurrency(salesOrder.discount)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">PPN (11%)</span>
            <span className="font-medium">
              {formatCurrency(salesOrder.tax)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Ongkos Kirim</span>
            <span className="font-medium">
              {formatCurrency(salesOrder.shippingCost)}
            </span>
          </div>
          <Separator />
          <div className="flex justify-between">
            <span className="text-base font-bold">Total</span>
            <span className="text-lg font-bold">
              {formatCurrency(salesOrder.totalAmount)}
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Notes */}
      {salesOrder.notes && (
        <Card>
          <CardHeader>
            <CardTitle>Catatan</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">
              {salesOrder.notes}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Cancel Confirmation Dialog */}
      <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Pesanan?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin membatalkan pesanan{" "}
              <strong>{salesOrder.orderNumber}</strong>? Tindakan ini tidak
              dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <div className="space-y-2 py-4">
            <Label htmlFor="cancelReason">
              Alasan Pembatalan <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="cancelReason"
              placeholder="Masukkan alasan pembatalan pesanan (minimal 5 karakter)..."
              value={cancelReason}
              onChange={(e) => setCancelReason(e.target.value)}
              rows={4}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground">
              Alasan pembatalan akan disimpan dalam riwayat pesanan
            </p>
          </div>

          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setShowCancelDialog(false);
                setCancelReason("");
              }}
              disabled={isCancelling}
            >
              Batal
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelOrder}
              disabled={isCancelling || cancelReason.trim().length < 5}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isCancelling ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Membatalkan...
                </>
              ) : (
                "Ya, Batalkan Pesanan"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
