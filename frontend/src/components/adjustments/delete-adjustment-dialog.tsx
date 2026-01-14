/**
 * Delete Adjustment Dialog Component
 *
 * Confirmation dialog for deleting a draft inventory adjustment.
 * Only DRAFT adjustments can be deleted.
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
import { useDeleteAdjustmentMutation } from "@/store/services/adjustmentApi";
import { useToast } from "@/hooks/use-toast";
import type { InventoryAdjustment } from "@/types/adjustment.types";

interface DeleteAdjustmentDialogProps {
  adjustment: InventoryAdjustment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function DeleteAdjustmentDialog({
  adjustment,
  open,
  onOpenChange,
  onSuccess,
}: DeleteAdjustmentDialogProps) {
  const { toast } = useToast();
  const [deleteAdjustment, { isLoading }] = useDeleteAdjustmentMutation();

  if (!adjustment) return null;

  const handleDelete = async () => {
    try {
      await deleteAdjustment(adjustment.id).unwrap();

      toast({
        title: "Penyesuaian Dihapus",
        description: `Penyesuaian ${adjustment.adjustmentNumber} telah dihapus.`,
      });

      onOpenChange(false);
      onSuccess?.();
    } catch (error: any) {
      toast({
        variant: "destructive",
        title: "Gagal Menghapus",
        description:
          error?.data?.message || "Terjadi kesalahan saat menghapus penyesuaian.",
      });
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Hapus Penyesuaian?</AlertDialogTitle>
          <AlertDialogDescription>
            Anda akan menghapus penyesuaian stok{" "}
            <span className="font-mono font-medium">
              {adjustment.adjustmentNumber}
            </span>
            . Tindakan ini tidak dapat dibatalkan.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Batal</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isLoading}
            className="bg-destructive hover:bg-destructive/90"
          >
            {isLoading ? "Menghapus..." : "Hapus"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
