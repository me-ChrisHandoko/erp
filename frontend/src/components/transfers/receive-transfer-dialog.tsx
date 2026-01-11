/**
 * Receive Transfer Dialog Component
 *
 * Confirmation dialog for receiving a transfer (SHIPPED → RECEIVED).
 * Records who received the transfer and when.
 * Adds stock to destination warehouse.
 */

"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useReceiveTransferMutation } from "@/store/services/transferApi";
import { AlertCircle, CheckCircle } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { StockTransfer } from "@/types/transfer.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { toast } from "sonner";

interface ReceiveTransferDialogProps {
  transfer: StockTransfer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function ReceiveTransferDialog({
  transfer,
  open,
  onOpenChange,
  onSuccess,
}: ReceiveTransferDialogProps) {
  const [notes, setNotes] = useState("");
  const [receiveTransfer, { isLoading, error }] = useReceiveTransferMutation();

  // Reset form when modal is closed
  useEffect(() => {
    if (!open) {
      setNotes("");
    }
  }, [open]);

  const handleReceive = async () => {
    if (!transfer) return;

    try {
      const result = await receiveTransfer({
        id: transfer.id,
        data: notes.trim() ? { notes: notes.trim() } : undefined,
      }).unwrap();

      // Show success toast
      toast.success("✓ Transfer Berhasil Diterima", {
        description: `Transfer ${result.transferNumber} telah diterima di gudang tujuan.`,
      });

      // Success - close modal (reset will be handled by useEffect)
      onOpenChange(false);
      onSuccess?.();
    } catch (err: any) {
      // Show error toast
      toast.error("Gagal Menerima Transfer", {
        description: err?.data?.message || "Terjadi kesalahan saat menerima transfer. Silakan coba lagi.",
      });
      console.error("Failed to receive transfer:", err);
    }
  };

  const handleCancel = () => {
    // Close modal (reset will be handled by useEffect)
    onOpenChange(false);
  };

  if (!transfer) return null;

  // Calculate total quantity
  const totalQuantity = transfer.items?.reduce(
    (sum, item) => sum + parseFloat(item.quantity || "0"),
    0
  ) || 0;

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen && document.activeElement instanceof HTMLElement) {
          document.activeElement.blur();
        }
        onOpenChange(isOpen);
      }}
    >
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-600" />
            Terima Transfer
          </DialogTitle>
          <DialogDescription>
            Konfirmasi penerimaan stok di gudang tujuan
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Transfer Information */}
          <div className="rounded-lg border bg-muted/50 p-4 space-y-3">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  No. Transfer
                </p>
                <p className="font-mono font-semibold">
                  {transfer.transferNumber}
                </p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Total Item
                </p>
                <p className="font-semibold">
                  {totalQuantity.toLocaleString("id-ID", {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 3,
                  })}{" "}
                  unit
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4 pt-2 border-t">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Dari Gudang
                </p>
                <p className="font-medium">
                  {transfer.sourceWarehouse?.name || "-"}
                </p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Ke Gudang
                </p>
                <p className="font-medium">
                  {transfer.destWarehouse?.name || "-"}
                </p>
              </div>
            </div>

            {/* Shipping Information */}
            {transfer.shippedBy && (
              <div className="pt-2 border-t">
                <p className="text-sm font-medium text-muted-foreground mb-1">
                  Informasi Pengiriman
                </p>
                <div className="flex items-center justify-between text-sm">
                  <span>Dikirim oleh: <span className="font-medium">{transfer.shippedBy}</span></span>
                  {transfer.shippedAt && (
                    <span className="text-muted-foreground">
                      {format(new Date(transfer.shippedAt), "dd MMM yyyy, HH:mm", {
                        locale: localeId,
                      })}
                    </span>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Success Alert */}
          <Alert className="bg-green-50 border-green-200">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <AlertDescription className="text-green-800">
              Dengan menerima transfer ini, stok akan ditambahkan ke gudang tujuan.
              Transfer akan selesai dan tidak dapat diubah lagi.
            </AlertDescription>
          </Alert>

          {/* Notes (Optional) */}
          <div className="space-y-2">
            <Label htmlFor="receive-notes">
              Catatan Penerimaan <span className="text-muted-foreground">(Opsional)</span>
            </Label>
            <Textarea
              id="receive-notes"
              placeholder="Tambahkan catatan penerimaan jika diperlukan..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              disabled={isLoading}
            />
          </div>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {"data" in error && error.data
                  ? (error.data as { message?: string })?.message ||
                    "Gagal menerima transfer"
                  : "Gagal menerima transfer"}
              </AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={isLoading}
          >
            Batal
          </Button>
          <Button onClick={handleReceive} disabled={isLoading}>
            {isLoading ? (
              <>
                <CheckCircle className="mr-2 h-4 w-4 animate-pulse" />
                Memproses...
              </>
            ) : (
              <>
                <CheckCircle className="mr-2 h-4 w-4" />
                Terima Transfer
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
