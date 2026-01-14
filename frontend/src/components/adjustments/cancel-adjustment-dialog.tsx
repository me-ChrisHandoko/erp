/**
 * Cancel Adjustment Dialog Component
 *
 * Confirmation dialog for cancelling a draft inventory adjustment.
 * Requires a cancellation reason.
 */

"use client";

import { useState } from "react";
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
import { useCancelAdjustmentMutation } from "@/store/services/adjustmentApi";
import { useToast } from "@/hooks/use-toast";
import type { InventoryAdjustment } from "@/types/adjustment.types";

interface CancelAdjustmentDialogProps {
  adjustment: InventoryAdjustment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function CancelAdjustmentDialog({
  adjustment,
  open,
  onOpenChange,
  onSuccess,
}: CancelAdjustmentDialogProps) {
  const [reason, setReason] = useState("");
  const { toast } = useToast();
  const [cancelAdjustment, { isLoading }] = useCancelAdjustmentMutation();

  if (!adjustment) return null;

  const handleCancel = async () => {
    if (!reason.trim() || reason.trim().length < 3) {
      toast({
        variant: "destructive",
        title: "Alasan Diperlukan",
        description: "Mohon masukkan alasan pembatalan (minimal 3 karakter).",
      });
      return;
    }

    try {
      await cancelAdjustment({
        id: adjustment.id,
        data: { reason: reason.trim() },
      }).unwrap();

      toast({
        title: "Penyesuaian Dibatalkan",
        description: `Penyesuaian ${adjustment.adjustmentNumber} telah dibatalkan.`,
      });

      setReason("");
      onOpenChange(false);
      onSuccess?.();
    } catch (error: any) {
      toast({
        variant: "destructive",
        title: "Gagal Membatalkan",
        description:
          error?.data?.message || "Terjadi kesalahan saat membatalkan penyesuaian.",
      });
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Batalkan Penyesuaian?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Anda akan membatalkan penyesuaian stok{" "}
                <span className="font-mono font-medium">
                  {adjustment.adjustmentNumber}
                </span>
                . Penyesuaian yang dibatalkan tidak dapat dikembalikan.
              </p>

              <div className="space-y-2">
                <Label htmlFor="reason">
                  Alasan Pembatalan <span className="text-destructive">*</span>
                </Label>
                <Textarea
                  id="reason"
                  placeholder="Masukkan alasan pembatalan..."
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  rows={3}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Minimal 3 karakter
                </p>
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Kembali</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleCancel}
            disabled={isLoading || reason.trim().length < 3}
            className="bg-destructive hover:bg-destructive/90"
          >
            {isLoading ? "Membatalkan..." : "Batalkan Penyesuaian"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
