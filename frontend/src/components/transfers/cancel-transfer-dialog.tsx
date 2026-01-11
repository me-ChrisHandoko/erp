/**
 * Cancel Transfer Dialog Component
 *
 * Confirmation dialog for cancelling a transfer (SHIPPED → CANCELLED).
 * Returns stock to source warehouse.
 * Requires a reason for cancellation.
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
import { useCancelTransferMutation } from "@/store/services/transferApi";
import { AlertCircle, XCircle } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { StockTransfer } from "@/types/transfer.types";
import { toast } from "sonner";

interface CancelTransferDialogProps {
  transfer: StockTransfer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function CancelTransferDialog({
  transfer,
  open,
  onOpenChange,
  onSuccess,
}: CancelTransferDialogProps) {
  const [reason, setReason] = useState("");
  const [cancelTransfer, { isLoading, error }] = useCancelTransferMutation();

  // Reset form when modal is closed
  useEffect(() => {
    if (!open) {
      setReason("");
    }
  }, [open]);

  const handleCancel = async () => {
    if (!transfer) return;

    // Validation: reason is required
    if (!reason.trim()) {
      return;
    }

    try {
      const result = await cancelTransfer({
        id: transfer.id,
        data: {
          reason: reason.trim(),
        },
      }).unwrap();

      // Show success toast
      toast.success("✓ Transfer Berhasil Dibatalkan", {
        description: `Transfer ${result.transferNumber} telah dibatalkan dan stok dikembalikan ke gudang asal.`,
      });

      // Success - close modal (reset will be handled by useEffect)
      onOpenChange(false);
      onSuccess?.();
    } catch (err: any) {
      // Show error toast
      toast.error("Gagal Membatalkan Transfer", {
        description: err?.data?.message || "Terjadi kesalahan saat membatalkan transfer. Silakan coba lagi.",
      });
      console.error("Failed to cancel transfer:", err);
    }
  };

  const handleClose = () => {
    // Close modal (reset will be handled by useEffect)
    onOpenChange(false);
  };

  if (!transfer) return null;

  // Calculate total quantity
  const totalQuantity = transfer.items?.reduce(
    (sum, item) => sum + parseFloat(item.quantity || "0"),
    0
  ) || 0;

  const isReasonEmpty = !reason.trim();

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
            <XCircle className="h-5 w-5 text-red-600" />
            Batalkan Transfer
          </DialogTitle>
          <DialogDescription>
            Batalkan transfer dan kembalikan stok ke gudang asal
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
          </div>

          {/* Warning Alert */}
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <strong>Perhatian:</strong> Pembatalan transfer akan mengembalikan stok ke gudang asal.
              Aksi ini tidak dapat dibatalkan setelah dikonfirmasi.
            </AlertDescription>
          </Alert>

          {/* Reason (Required) */}
          <div className="space-y-2">
            <Label htmlFor="cancel-reason" className="flex items-center gap-1">
              Alasan Pembatalan
              <span className="text-red-600">*</span>
            </Label>
            <Textarea
              id="cancel-reason"
              placeholder="Jelaskan alasan pembatalan transfer ini..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={4}
              disabled={isLoading}
              className={isReasonEmpty ? "border-red-300" : ""}
            />
            {isReasonEmpty && reason !== "" && (
              <p className="text-sm text-red-600">
                Alasan pembatalan wajib diisi
              </p>
            )}
          </div>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {"data" in error && error.data
                  ? (error.data as { message?: string })?.message ||
                    "Gagal membatalkan transfer"
                  : "Gagal membatalkan transfer"}
              </AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleClose}
            disabled={isLoading}
          >
            Tutup
          </Button>
          <Button
            variant="destructive"
            onClick={handleCancel}
            disabled={isLoading || isReasonEmpty}
          >
            {isLoading ? (
              <>
                <XCircle className="mr-2 h-4 w-4 animate-pulse" />
                Memproses...
              </>
            ) : (
              <>
                <XCircle className="mr-2 h-4 w-4" />
                Batalkan Transfer
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
