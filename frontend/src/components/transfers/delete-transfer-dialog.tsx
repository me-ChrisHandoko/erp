/**
 * Delete Transfer Dialog Component
 *
 * Confirmation dialog for deleting a DRAFT transfer.
 * Permanently removes the transfer from the system.
 */

"use client";

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
import { useDeleteTransferMutation } from "@/store/services/transferApi";
import { AlertCircle, Trash } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { StockTransfer } from "@/types/transfer.types";
import { toast } from "sonner";

interface DeleteTransferDialogProps {
  transfer: StockTransfer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function DeleteTransferDialog({
  transfer,
  open,
  onOpenChange,
  onSuccess,
}: DeleteTransferDialogProps) {
  const [deleteTransfer, { isLoading, error }] = useDeleteTransferMutation();

  const handleDelete = async () => {
    if (!transfer) return;

    try {
      await deleteTransfer(transfer.id).unwrap();

      // Show success toast
      toast.success("✓ Transfer Berhasil Dihapus", {
        description: `Transfer ${transfer.transferNumber} telah dihapus dari sistem.`,
      });

      onOpenChange(false);
      onSuccess?.();
    } catch (err: any) {
      // Show error toast
      toast.error("Gagal Menghapus Transfer", {
        description: err?.data?.message || "Terjadi kesalahan saat menghapus transfer. Silakan coba lagi.",
      });
      console.error("Failed to delete transfer:", err);
    }
  };

  if (!transfer) return null;

  // Calculate total quantity
  const totalQuantity = transfer.items?.reduce(
    (sum, item) => sum + parseFloat(item.quantity || "0"),
    0
  ) || 0;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <Trash className="h-5 w-5 text-red-600" />
            Hapus Transfer
          </AlertDialogTitle>
          <AlertDialogDescription>
            Apakah Anda yakin ingin menghapus transfer ini? Aksi ini tidak dapat dibatalkan.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-4">
          {/* Transfer Information */}
          <div className="rounded-lg border bg-muted/50 p-4 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-muted-foreground">
                No. Transfer
              </span>
              <span className="font-mono font-semibold">
                {transfer.transferNumber}
              </span>
            </div>

            <div className="flex items-center justify-between pt-2 border-t">
              <span className="text-sm font-medium text-muted-foreground">
                Dari - Ke
              </span>
              <span className="text-sm font-medium">
                {transfer.sourceWarehouse?.name || "-"} →{" "}
                {transfer.destWarehouse?.name || "-"}
              </span>
            </div>

            <div className="flex items-center justify-between pt-2 border-t">
              <span className="text-sm font-medium text-muted-foreground">
                Total Item
              </span>
              <span className="font-semibold">
                {totalQuantity.toLocaleString("id-ID", {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 3,
                })}{" "}
                unit
              </span>
            </div>
          </div>

          {/* Warning */}
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Transfer ini akan dihapus secara permanen dari sistem.
              Data tidak dapat dipulihkan setelah dihapus.
            </AlertDescription>
          </Alert>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {"data" in error && error.data
                  ? (error.data as { message?: string })?.message ||
                    "Gagal menghapus transfer"
                  : "Gagal menghapus transfer"}
              </AlertDescription>
            </Alert>
          )}
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>
            Batal
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isLoading}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            {isLoading ? (
              <>
                <Trash className="mr-2 h-4 w-4 animate-pulse" />
                Menghapus...
              </>
            ) : (
              <>
                <Trash className="mr-2 h-4 w-4" />
                Hapus Transfer
              </>
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
