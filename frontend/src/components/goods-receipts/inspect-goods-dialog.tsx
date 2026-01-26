/**
 * Inspect Goods Dialog Component
 *
 * Dialog for performing quality inspection on goods receipt items:
 * - Shows all items with product details
 * - Per-item input for accepted/rejected quantities
 * - Rejection reason (required when rejectedQty > 0)
 * - Quality notes (optional)
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { AlertCircle, CheckCircle, XCircle, Package } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import type {
  GoodsReceiptResponse,
  GoodsReceiptItemResponse,
  InspectGoodsRequest,
  UpdateGoodsReceiptItemRequest,
} from "@/types/goods-receipt.types";

interface InspectionItemState {
  id: string;
  acceptedQty: string;
  rejectedQty: string;
  rejectionReason: string;
  qualityNote: string;
}

interface InspectGoodsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  receipt: GoodsReceiptResponse;
  onSubmit: (data: InspectGoodsRequest) => Promise<void>;
  isLoading: boolean;
}

export function InspectGoodsDialog({
  open,
  onOpenChange,
  receipt,
  onSubmit,
  isLoading,
}: InspectGoodsDialogProps) {
  const [generalNotes, setGeneralNotes] = useState("");
  const [itemStates, setItemStates] = useState<InspectionItemState[]>([]);

  // Initialize item states when dialog opens or items change
  useEffect(() => {
    if (open && receipt.items) {
      const initialStates: InspectionItemState[] = receipt.items.map((item) => ({
        id: item.id,
        acceptedQty: item.receivedQty, // Default: accept all received
        rejectedQty: "0",
        rejectionReason: "",
        qualityNote: "",
      }));
      setItemStates(initialStates);
      setGeneralNotes("");
    }
  }, [open, receipt.items]);

  // Validation - returns true if valid, shows toast for errors
  const validateItems = (): boolean => {
    const validationErrors: string[] = [];

    itemStates.forEach((state, index) => {
      const item = receipt.items?.find((i) => i.id === state.id);
      if (!item) return;

      const productName = item.product?.name || `Item #${index + 1}`;
      const receivedQty = parseFloat(item.receivedQty) || 0;
      const acceptedQty = parseFloat(state.acceptedQty) || 0;
      const rejectedQty = parseFloat(state.rejectedQty) || 0;

      // Check accepted qty is valid
      if (acceptedQty < 0) {
        validationErrors.push(`${productName}: Jumlah diterima tidak boleh negatif`);
      }

      // Check rejected qty is valid
      if (rejectedQty < 0) {
        validationErrors.push(`${productName}: Jumlah ditolak tidak boleh negatif`);
      }

      // Check total doesn't exceed received
      if (acceptedQty + rejectedQty > receivedQty) {
        validationErrors.push(`${productName}: Total qty (${acceptedQty + rejectedQty}) melebihi jumlah diterima (${receivedQty})`);
      }

      // Check rejection reason is provided when rejectedQty > 0
      if (rejectedQty > 0 && !state.rejectionReason.trim()) {
        validationErrors.push(`${productName}: Alasan penolakan wajib diisi`);
      }
    });

    if (validationErrors.length > 0) {
      toast.error("Validasi Gagal", {
        description: validationErrors.join("\n"),
        duration: 5000,
      });
      return false;
    }

    return true;
  };

  // Calculate summary
  const summary = useMemo(() => {
    let totalItems = 0;
    let fullyAccepted = 0;
    let partiallyAccepted = 0;
    let fullyRejected = 0;

    itemStates.forEach((state) => {
      const item = receipt.items?.find((i) => i.id === state.id);
      if (!item) return;

      totalItems++;
      const receivedQty = parseFloat(item.receivedQty) || 0;
      const acceptedQty = parseFloat(state.acceptedQty) || 0;
      const rejectedQty = parseFloat(state.rejectedQty) || 0;

      if (acceptedQty === receivedQty && rejectedQty === 0) {
        fullyAccepted++;
      } else if (acceptedQty === 0 && rejectedQty > 0) {
        fullyRejected++;
      } else if (acceptedQty > 0 && rejectedQty > 0) {
        partiallyAccepted++;
      }
    });

    return { totalItems, fullyAccepted, partiallyAccepted, fullyRejected };
  }, [itemStates, receipt.items]);

  const handleSubmit = async () => {
    if (!validateItems()) return;

    const items: UpdateGoodsReceiptItemRequest[] = itemStates.map((state) => ({
      id: state.id,
      receivedQty: receipt.items?.find((i) => i.id === state.id)?.receivedQty || "0",
      acceptedQty: state.acceptedQty,
      rejectedQty: state.rejectedQty,
      rejectionReason: state.rejectionReason.trim() || undefined,
      qualityNote: state.qualityNote.trim() || undefined,
    }));

    await onSubmit({
      notes: generalNotes.trim() || undefined,
      items,
    });
  };

  const updateItemState = (
    id: string,
    field: keyof InspectionItemState,
    value: string
  ) => {
    setItemStates((prev) =>
      prev.map((item) =>
        item.id === id ? { ...item, [field]: value } : item
      )
    );
  };

  // Handle accepted qty change - auto-calculate rejected qty
  const handleAcceptedQtyChange = (id: string, value: string) => {
    const item = receipt.items?.find((i) => i.id === id);
    if (!item) return;

    const receivedQty = parseFloat(item.receivedQty) || 0;
    const newAccepted = parseFloat(value) || 0;

    // Calculate rejected qty (ensure non-negative)
    const newRejected = Math.max(0, receivedQty - newAccepted);

    setItemStates((prev) =>
      prev.map((state) =>
        state.id === id
          ? {
              ...state,
              acceptedQty: value,
              rejectedQty: newRejected.toString(),
              // Clear rejection reason if no rejection
              rejectionReason: newRejected === 0 ? "" : state.rejectionReason,
            }
          : state
      )
    );
  };

  // Handle rejected qty change - auto-calculate accepted qty
  const handleRejectedQtyChange = (id: string, value: string) => {
    const item = receipt.items?.find((i) => i.id === id);
    if (!item) return;

    const receivedQty = parseFloat(item.receivedQty) || 0;
    const newRejected = parseFloat(value) || 0;

    // Calculate accepted qty (ensure non-negative)
    const newAccepted = Math.max(0, receivedQty - newRejected);

    setItemStates((prev) =>
      prev.map((state) =>
        state.id === id
          ? {
              ...state,
              acceptedQty: newAccepted.toString(),
              rejectedQty: value,
              // Clear rejection reason if no rejection
              rejectionReason: newRejected === 0 ? "" : state.rejectionReason,
            }
          : state
      )
    );
  };

  // Handle "Terima Semua" for single item
  const handleAcceptAll = (id: string) => {
    const item = receipt.items?.find((i) => i.id === id);
    if (!item) return;
    updateItemState(id, "acceptedQty", item.receivedQty);
    updateItemState(id, "rejectedQty", "0");
    updateItemState(id, "rejectionReason", "");
  };

  // Handle "Tolak Semua" for single item
  const handleRejectAll = (id: string) => {
    const item = receipt.items?.find((i) => i.id === id);
    if (!item) return;
    updateItemState(id, "acceptedQty", "0");
    updateItemState(id, "rejectedQty", item.receivedQty);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader className="flex-shrink-0">
          <DialogTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Inspeksi Barang - {receipt.grnNumber}
          </DialogTitle>
          <DialogDescription>
            Periksa kualitas setiap item dan tentukan jumlah yang diterima atau ditolak.
          </DialogDescription>
        </DialogHeader>

        {/* Summary */}
        <div className="flex flex-wrap gap-2 py-2 border-b flex-shrink-0">
          <Badge variant="outline" className="gap-1">
            Total: {summary.totalItems} item
          </Badge>
          {summary.fullyAccepted > 0 && (
            <Badge className="gap-1 bg-green-100 text-green-800">
              <CheckCircle className="h-3 w-3" />
              Diterima penuh: {summary.fullyAccepted}
            </Badge>
          )}
          {summary.partiallyAccepted > 0 && (
            <Badge className="gap-1 bg-orange-100 text-orange-800">
              <AlertCircle className="h-3 w-3" />
              Sebagian: {summary.partiallyAccepted}
            </Badge>
          )}
          {summary.fullyRejected > 0 && (
            <Badge className="gap-1 bg-red-100 text-red-800">
              <XCircle className="h-3 w-3" />
              Ditolak: {summary.fullyRejected}
            </Badge>
          )}
        </div>

        {/* Items List - Scrollable area with fixed height */}
        <div className="overflow-y-auto min-h-[200px] max-h-[350px] pr-2">
          <div className="space-y-4 py-2">
            {receipt.items?.map((item, index) => {
              const state = itemStates.find((s) => s.id === item.id);
              if (!state) return null;

              const receivedQty = parseFloat(item.receivedQty) || 0;
              const acceptedQty = parseFloat(state.acceptedQty) || 0;
              const rejectedQty = parseFloat(state.rejectedQty) || 0;
              const hasRejection = rejectedQty > 0;

              return (
                <div
                  key={item.id}
                  className={cn(
                    "rounded-lg border p-4 space-y-4",
                    hasRejection && "border-orange-200 bg-orange-50/50"
                  )}
                >
                  {/* Item Header */}
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-muted-foreground">
                          #{index + 1}
                        </span>
                        <h4 className="font-semibold truncate">
                          {item.product?.name || "Produk"}
                        </h4>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        {item.product?.code} • {item.productUnit?.unitName || item.product?.baseUnit}
                      </p>
                    </div>
                    <div className="text-right shrink-0">
                      <div className="text-sm text-muted-foreground">Diterima</div>
                      <div className="text-lg font-bold">{item.receivedQty}</div>
                    </div>
                  </div>

                  {/* Quantity Inputs */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor={`accepted-${item.id}`} className="flex items-center gap-1">
                        <CheckCircle className="h-3 w-3 text-green-600" />
                        Jumlah Diterima
                      </Label>
                      <div className="flex gap-2">
                        <Input
                          id={`accepted-${item.id}`}
                          type="number"
                          min="0"
                          step="0.001"
                          value={state.acceptedQty}
                          onChange={(e) => handleAcceptedQtyChange(item.id, e.target.value)}
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => handleAcceptAll(item.id)}
                          className="shrink-0"
                        >
                          Semua
                        </Button>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor={`rejected-${item.id}`} className="flex items-center gap-1">
                        <XCircle className="h-3 w-3 text-red-600" />
                        Jumlah Ditolak
                      </Label>
                      <div className="flex gap-2">
                        <Input
                          id={`rejected-${item.id}`}
                          type="number"
                          min="0"
                          step="0.001"
                          value={state.rejectedQty}
                          onChange={(e) => handleRejectedQtyChange(item.id, e.target.value)}
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => handleRejectAll(item.id)}
                          className="shrink-0 text-red-600 hover:text-red-700"
                        >
                          Semua
                        </Button>
                      </div>
                    </div>
                  </div>

                  {/* Rejection Reason - Required if rejectedQty > 0 */}
                  {hasRejection && (
                    <div className="space-y-2">
                      <Label htmlFor={`reason-${item.id}`}>
                        Alasan Penolakan <span className="text-destructive">*</span>
                      </Label>
                      <Textarea
                        id={`reason-${item.id}`}
                        value={state.rejectionReason}
                        onChange={(e) => updateItemState(item.id, "rejectionReason", e.target.value)}
                        placeholder="Masukkan alasan penolakan..."
                        rows={2}
                      />
                    </div>
                  )}

                  {/* Quality Note - Optional */}
                  <div className="space-y-2">
                    <Label htmlFor={`quality-${item.id}`}>
                      Catatan Kualitas (Opsional)
                    </Label>
                    <Textarea
                      id={`quality-${item.id}`}
                      value={state.qualityNote}
                      onChange={(e) => updateItemState(item.id, "qualityNote", e.target.value)}
                      placeholder="Masukkan catatan hasil inspeksi kualitas..."
                      rows={2}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* General Inspection Notes */}
        <div className="space-y-2 border-t pt-4 flex-shrink-0">
          <Label htmlFor="generalNotes">Catatan Inspeksi (Opsional)</Label>
          <Textarea
            id="generalNotes"
            value={generalNotes}
            onChange={(e) => setGeneralNotes(e.target.value)}
            placeholder="Masukkan catatan hasil inspeksi secara keseluruhan..."
            rows={2}
          />
        </div>

        <DialogFooter className="flex-shrink-0">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Batal
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isLoading}
            className="bg-purple-600 hover:bg-purple-700"
          >
            {isLoading ? "Memproses..." : "Simpan Hasil Inspeksi"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
