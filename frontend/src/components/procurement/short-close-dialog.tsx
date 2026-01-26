/**
 * Short Close Purchase Order Dialog
 *
 * Dialog component for closing a PO even when not fully delivered (SAP DCI Model).
 * Used when supplier cannot fulfill remaining quantity.
 */

"use client";

import { useState } from "react";
import { AlertTriangle } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { PurchaseOrderResponse } from "@/types/purchase-order.types";

// Common reasons for short closing a PO
const SHORT_CLOSE_REASONS = [
  { value: "SUPPLIER_DISCONTINUED", label: "Barang tidak lagi diproduksi/dijual" },
  { value: "SUPPLIER_UNABLE", label: "Supplier tidak dapat memenuhi sisa pesanan" },
  { value: "NO_LONGER_NEEDED", label: "Barang tidak lagi diperlukan" },
  { value: "QUALITY_ISSUES", label: "Masalah kualitas berulang" },
  { value: "PRICE_CHANGE", label: "Perubahan harga tidak dapat diterima" },
  { value: "OTHER", label: "Alasan lain" },
] as const;

interface ShortCloseDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  order: PurchaseOrderResponse;
  onSubmit: (data: { shortCloseReason: string }) => Promise<void>;
  isLoading: boolean;
}

export function ShortCloseDialog({
  open,
  onOpenChange,
  order,
  onSubmit,
  isLoading,
}: ShortCloseDialogProps) {
  const [reasonType, setReasonType] = useState<string>("");
  const [customReason, setCustomReason] = useState("");
  const [notes, setNotes] = useState("");

  // Calculate outstanding items
  const outstandingItems = order.items?.filter((item) => {
    const ordered = parseFloat(item.quantity) || 0;
    const received = parseFloat(item.receivedQty) || 0;
    return received < ordered;
  }) || [];

  const handleSubmit = async () => {
    let finalReason = reasonType === "OTHER"
      ? customReason.trim()
      : SHORT_CLOSE_REASONS.find(r => r.value === reasonType)?.label || reasonType;

    // Append notes to reason if provided
    if (notes.trim()) {
      finalReason = `${finalReason}. Catatan: ${notes.trim()}`;
    }

    if (!finalReason) return;

    await onSubmit({
      shortCloseReason: finalReason,
    });

    // Reset form on success
    setReasonType("");
    setCustomReason("");
    setNotes("");
  };

  const isValid = reasonType && (reasonType !== "OTHER" || customReason.trim());

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-lg">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            Short Close Purchase Order
          </AlertDialogTitle>
          <AlertDialogDescription>
            Anda akan menutup <span className="font-semibold">{order.poNumber}</span> meskipun
            tidak semua barang telah diterima. Tindakan ini tidak dapat dibatalkan.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {/* Outstanding Items Warning */}
        {outstandingItems.length > 0 && (
          <Alert variant="destructive" className="my-4">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              <p className="font-medium mb-2">
                {outstandingItems.length} item belum diterima sepenuhnya:
              </p>
              <ul className="list-disc list-inside text-sm space-y-1">
                {outstandingItems.slice(0, 5).map((item) => {
                  const ordered = parseFloat(item.quantity) || 0;
                  const received = parseFloat(item.receivedQty) || 0;
                  const outstanding = ordered - received;
                  return (
                    <li key={item.id}>
                      {item.product?.name || "Produk"}: {outstanding.toLocaleString("id-ID")}{" "}
                      {item.productUnit?.unitName || item.product?.baseUnit || "unit"} belum diterima
                    </li>
                  );
                })}
                {outstandingItems.length > 5 && (
                  <li>...dan {outstandingItems.length - 5} item lainnya</li>
                )}
              </ul>
            </AlertDescription>
          </Alert>
        )}

        <div className="space-y-4 py-4">
          {/* Reason Selection */}
          <div className="space-y-2">
            <Label htmlFor="reasonType">
              Alasan Short Close <span className="text-destructive">*</span>
            </Label>
            <Select value={reasonType} onValueChange={setReasonType}>
              <SelectTrigger id="reasonType">
                <SelectValue placeholder="Pilih alasan..." />
              </SelectTrigger>
              <SelectContent>
                {SHORT_CLOSE_REASONS.map((reason) => (
                  <SelectItem key={reason.value} value={reason.value}>
                    {reason.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Custom Reason (if OTHER selected) */}
          {reasonType === "OTHER" && (
            <div className="space-y-2">
              <Label htmlFor="customReason">
                Alasan Kustom <span className="text-destructive">*</span>
              </Label>
              <Input
                id="customReason"
                value={customReason}
                onChange={(e) => setCustomReason(e.target.value)}
                placeholder="Masukkan alasan short close..."
              />
            </div>
          )}

          {/* Additional Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Catatan Tambahan (Opsional)</Label>
            <Textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Catatan tambahan mengenai short close..."
              rows={3}
            />
          </div>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Batal</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleSubmit}
            disabled={isLoading || !isValid}
            className="bg-amber-600 hover:bg-amber-700"
          >
            {isLoading ? "Memproses..." : "Tutup PO (Short Close)"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
