/**
 * Ship Transfer Dialog Component
 *
 * Confirmation dialog for shipping a transfer (DRAFT → SHIPPED).
 * Records who shipped the transfer and when.
 */

"use client";

import { useState } from "react";
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
import { useShipTransferMutation } from "@/store/services/transferApi";
import { AlertCircle, Truck } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { StockTransfer } from "@/types/transfer.types";
import { toast } from "sonner";

interface ShipTransferDialogProps {
  transfer: StockTransfer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function ShipTransferDialog({
  transfer,
  open,
  onOpenChange,
  onSuccess,
}: ShipTransferDialogProps) {
  const [notes, setNotes] = useState("");
  const [shipTransfer, { isLoading, error }] = useShipTransferMutation();

  const handleShip = async () => {
    if (!transfer) return;

    try {
      const result = await shipTransfer({
        id: transfer.id,
        data: notes.trim() ? { notes: notes.trim() } : undefined,
      }).unwrap();

      // Show success toast
      toast.success("✓ Transfer Berhasil Dikirim", {
        description: `Transfer ${result.transferNumber} telah dikirim dari gudang asal.`,
      });

      // Reset form
      setNotes("");
      onOpenChange(false);
      onSuccess?.();
    } catch (err: any) {
      // Show error toast
      toast.error("Gagal Mengirim Transfer", {
        description: err?.data?.message || "Terjadi kesalahan saat mengirim transfer. Silakan coba lagi.",
      });
      console.error("Failed to ship transfer:", err);
    }
  };

  const handleCancel = () => {
    setNotes("");
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
            <Truck className="h-5 w-5 text-blue-600" />
            Kirim Transfer
          </DialogTitle>
          <DialogDescription>
            Konfirmasi pengiriman stok dari gudang asal ke gudang tujuan
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
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Dengan mengirim transfer ini, stok akan dikurangi dari gudang asal.
              Transfer akan menunggu konfirmasi penerimaan di gudang tujuan.
            </AlertDescription>
          </Alert>

          {/* Notes (Optional) */}
          <div className="space-y-2">
            <Label htmlFor="ship-notes">
              Catatan Pengiriman <span className="text-muted-foreground">(Opsional)</span>
            </Label>
            <Textarea
              id="ship-notes"
              placeholder="Tambahkan catatan pengiriman jika diperlukan..."
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
                    "Gagal mengirim transfer"
                  : "Gagal mengirim transfer"}
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
          <Button onClick={handleShip} disabled={isLoading}>
            {isLoading ? (
              <>
                <Truck className="mr-2 h-4 w-4 animate-pulse" />
                Mengirim...
              </>
            ) : (
              <>
                <Truck className="mr-2 h-4 w-4" />
                Kirim Transfer
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
