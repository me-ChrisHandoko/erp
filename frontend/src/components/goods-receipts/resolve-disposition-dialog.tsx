/**
 * Resolve Disposition Dialog
 *
 * Dialog component for marking a rejection disposition as resolved.
 * Used when the disposition action has been completed (e.g., replacement received,
 * credit applied, items returned, or written off).
 */

"use client";

import { useState } from "react";
import { CheckCircle } from "lucide-react";
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
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  getRejectionDispositionLabel,
  getRejectionDispositionColor,
  type GoodsReceiptItemResponse,
} from "@/types/goods-receipt.types";

interface ResolveDispositionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  item: GoodsReceiptItemResponse | null;
  onSubmit: (data: { dispositionResolvedNotes?: string }) => Promise<void>;
  isLoading: boolean;
}

export function ResolveDispositionDialog({
  open,
  onOpenChange,
  item,
  onSubmit,
  isLoading,
}: ResolveDispositionDialogProps) {
  const [notes, setNotes] = useState("");

  const handleSubmit = async () => {
    await onSubmit({
      dispositionResolvedNotes: notes.trim() || undefined,
    });
    setNotes("");
  };

  if (!item || !item.rejectionDisposition) return null;

  const rejectedQty = parseFloat(item.rejectedQty) || 0;

  // Get confirmation message based on disposition type
  const getConfirmationMessage = () => {
    switch (item.rejectionDisposition) {
      case "PENDING_REPLACEMENT":
        return "Apakah barang pengganti sudah diterima dari supplier?";
      case "CREDIT_REQUESTED":
        return "Apakah kredit/pengembalian dana sudah diterima dari supplier?";
      case "RETURNED":
        return "Apakah barang sudah dikonfirmasi dikembalikan ke supplier?";
      case "WRITTEN_OFF":
        return "Apakah Anda yakin ingin menyelesaikan penghapusan barang ini?";
      default:
        return "Apakah Anda yakin ingin menyelesaikan disposisi ini?";
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-lg">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-600" />
            Selesaikan Disposisi
          </AlertDialogTitle>
          <AlertDialogDescription>
            {getConfirmationMessage()}
          </AlertDialogDescription>
        </AlertDialogHeader>

        {/* Item Info */}
        <div className="rounded-lg border bg-muted/50 p-4 my-2">
          <div className="flex items-start justify-between">
            <div>
              <p className="font-medium">{item.product?.name || "Produk"}</p>
              <p className="text-sm text-muted-foreground">
                {item.product?.code || "N/A"}
              </p>
            </div>
            <div className="text-right">
              <Badge variant="destructive">
                {rejectedQty.toLocaleString("id-ID")}{" "}
                {item.productUnit?.unitName || item.product?.baseUnit || "unit"}
              </Badge>
              <div className="mt-1">
                <Badge className={getRejectionDispositionColor(item.rejectionDisposition)}>
                  {getRejectionDispositionLabel(item.rejectionDisposition)}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        <div className="py-2">
          <Label htmlFor="resolveNotes">Catatan Penyelesaian (Opsional)</Label>
          <Textarea
            id="resolveNotes"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Catatan tambahan tentang penyelesaian disposisi..."
            className="mt-2"
            rows={3}
          />
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Batal</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleSubmit}
            disabled={isLoading}
            className="bg-green-600 hover:bg-green-700"
          >
            {isLoading ? "Memproses..." : "Selesaikan Disposisi"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
