/**
 * Approve Adjustment Dialog Component
 *
 * Confirmation dialog for approving a draft inventory adjustment.
 * Approving will apply the stock changes to the warehouse.
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
import { useApproveAdjustmentMutation } from "@/store/services/adjustmentApi";
import { useToast } from "@/hooks/use-toast";
import { AdjustmentTypeBadge } from "./adjustment-type-badge";
import {
  ADJUSTMENT_REASON_CONFIG,
  type InventoryAdjustment,
} from "@/types/adjustment.types";

interface ApproveAdjustmentDialogProps {
  adjustment: InventoryAdjustment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function ApproveAdjustmentDialog({
  adjustment,
  open,
  onOpenChange,
  onSuccess,
}: ApproveAdjustmentDialogProps) {
  const [notes, setNotes] = useState("");
  const { toast } = useToast();
  const [approveAdjustment, { isLoading }] = useApproveAdjustmentMutation();

  if (!adjustment) return null;

  const reasonConfig = ADJUSTMENT_REASON_CONFIG[adjustment.reason];

  const handleApprove = async () => {
    try {
      await approveAdjustment({
        id: adjustment.id,
        data: notes ? { notes } : undefined,
      }).unwrap();

      toast({
        title: "Penyesuaian Disetujui",
        description: `Penyesuaian ${adjustment.adjustmentNumber} telah disetujui dan stok telah diperbarui.`,
      });

      setNotes("");
      onOpenChange(false);
      onSuccess?.();
    } catch (error: any) {
      toast({
        variant: "destructive",
        title: "Gagal Menyetujui",
        description:
          error?.data?.message || "Terjadi kesalahan saat menyetujui penyesuaian.",
      });
    }
  };

  // Format currency
  const formatCurrency = (value: string | number) => {
    const numValue = typeof value === "string" ? parseFloat(value) : value;
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(numValue);
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Setujui Penyesuaian Stok?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Anda akan menyetujui penyesuaian stok berikut. Perubahan stok
                akan langsung diterapkan ke gudang.
              </p>

              <div className="bg-muted p-3 rounded-md space-y-2 text-sm">
                <div className="flex justify-between">
                  <span>No. Penyesuaian:</span>
                  <span className="font-mono font-medium">
                    {adjustment.adjustmentNumber}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Gudang:</span>
                  <span className="font-medium">
                    {adjustment.warehouse?.name}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Tipe:</span>
                  <AdjustmentTypeBadge type={adjustment.adjustmentType} />
                </div>
                <div className="flex justify-between">
                  <span>Alasan:</span>
                  <span className="font-medium">
                    {reasonConfig?.label || adjustment.reason}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Jumlah Item:</span>
                  <span className="font-medium">
                    {adjustment.totalItems || adjustment.items?.length || 0}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Total Nilai:</span>
                  <span className="font-medium">
                    {formatCurrency(adjustment.totalValue)}
                  </span>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="notes">Catatan Persetujuan (Opsional)</Label>
                <Textarea
                  id="notes"
                  placeholder="Tambahkan catatan persetujuan..."
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  rows={2}
                />
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Batal</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleApprove}
            disabled={isLoading}
            className="bg-green-600 hover:bg-green-700"
          >
            {isLoading ? "Menyetujui..." : "Setujui"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
