/**
 * Rejection Disposition Dialog
 *
 * Dialog component for managing rejection disposition (Odoo+M3 Model).
 * Allows users to set what happens to rejected items:
 * - PENDING_REPLACEMENT: Waiting for supplier to send replacement
 * - CREDIT_REQUESTED: Requesting credit/refund from supplier
 * - RETURNED: Items returned to supplier
 * - WRITTEN_OFF: Items written off as loss
 */

"use client";

import { useState, useEffect } from "react";
import { Package, AlertCircle, CheckCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  REJECTION_DISPOSITION_OPTIONS,
  getRejectionDispositionLabel,
  getRejectionDispositionColor,
  type GoodsReceiptItemResponse,
  type RejectionDispositionStatus,
} from "@/types/goods-receipt.types";

interface RejectionDispositionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  item: GoodsReceiptItemResponse | null;
  onSubmit: (data: { rejectionDisposition: RejectionDispositionStatus; dispositionNotes?: string }) => Promise<void>;
  isLoading: boolean;
}

export function RejectionDispositionDialog({
  open,
  onOpenChange,
  item,
  onSubmit,
  isLoading,
}: RejectionDispositionDialogProps) {
  const [disposition, setDisposition] = useState<RejectionDispositionStatus | "">("");
  const [notes, setNotes] = useState("");

  // Reset form when dialog opens/item changes
  useEffect(() => {
    if (open && item) {
      setDisposition(item.rejectionDisposition || "");
      setNotes(item.dispositionNotes || "");
    }
  }, [open, item]);

  const handleSubmit = async () => {
    if (!disposition || !item) return;

    await onSubmit({
      rejectionDisposition: disposition as RejectionDispositionStatus,
      dispositionNotes: notes.trim() || undefined,
    });

    // Reset form on success
    setDisposition("");
    setNotes("");
  };

  if (!item) return null;

  const rejectedQty = parseFloat(item.rejectedQty) || 0;
  const isAlreadyResolved = item.dispositionResolved;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Disposisi Barang Ditolak
          </DialogTitle>
          <DialogDescription>
            Tentukan tindakan untuk barang yang ditolak pada penerimaan ini.
          </DialogDescription>
        </DialogHeader>

        {/* Item Info */}
        <div className="rounded-lg border bg-muted/50 p-4">
          <div className="flex items-start justify-between">
            <div>
              <p className="font-medium">{item.product?.name || "Produk"}</p>
              <p className="text-sm text-muted-foreground">
                {item.product?.code || "N/A"}
              </p>
            </div>
            <Badge variant="destructive">
              {rejectedQty.toLocaleString("id-ID")}{" "}
              {item.productUnit?.unitName || item.product?.baseUnit || "unit"} ditolak
            </Badge>
          </div>
          {item.rejectionReason && (
            <div className="mt-3 text-sm">
              <span className="font-medium">Alasan penolakan:</span>{" "}
              <span className="text-muted-foreground">{item.rejectionReason}</span>
            </div>
          )}
        </div>

        {/* Already Resolved Warning */}
        {isAlreadyResolved && (
          <Alert className="border-green-500 bg-green-50">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <AlertDescription className="text-green-800">
              Disposisi ini sudah diselesaikan pada{" "}
              {item.dispositionResolvedAt
                ? new Date(item.dispositionResolvedAt).toLocaleDateString("id-ID", {
                    day: "2-digit",
                    month: "short",
                    year: "numeric",
                  })
                : "-"}
              {item.dispositionResolver && ` oleh ${item.dispositionResolver.fullName}`}
            </AlertDescription>
          </Alert>
        )}

        {/* Current Disposition (if set) */}
        {item.rejectionDisposition && (
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">Status saat ini:</span>
            <Badge className={getRejectionDispositionColor(item.rejectionDisposition)}>
              {getRejectionDispositionLabel(item.rejectionDisposition)}
            </Badge>
          </div>
        )}

        <div className="space-y-4 py-2">
          {/* Disposition Selection */}
          <div className="space-y-2">
            <Label htmlFor="disposition">
              Tindakan Disposisi <span className="text-destructive">*</span>
            </Label>
            <Select
              value={disposition}
              onValueChange={(val) => setDisposition(val as RejectionDispositionStatus)}
              disabled={isAlreadyResolved}
            >
              <SelectTrigger id="disposition">
                <SelectValue placeholder="Pilih tindakan disposisi..." />
              </SelectTrigger>
              <SelectContent>
                {REJECTION_DISPOSITION_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    <div className="flex items-center gap-2">
                      <div
                        className={`h-2 w-2 rounded-full ${opt.color.replace("text-", "bg-").split(" ")[0]}`}
                      />
                      {opt.label}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {disposition === "PENDING_REPLACEMENT" &&
                "Menunggu barang pengganti dari supplier"}
              {disposition === "CREDIT_REQUESTED" &&
                "Meminta kredit/pengembalian dana dari supplier"}
              {disposition === "RETURNED" && "Barang sudah dikembalikan ke supplier"}
              {disposition === "WRITTEN_OFF" && "Barang dihapuskan sebagai kerugian"}
            </p>
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="dispositionNotes">Catatan (Opsional)</Label>
            <Textarea
              id="dispositionNotes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Catatan tambahan tentang disposisi..."
              rows={3}
              disabled={isAlreadyResolved}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
            {isAlreadyResolved ? "Tutup" : "Batal"}
          </Button>
          {!isAlreadyResolved && (
            <Button
              onClick={handleSubmit}
              disabled={isLoading || !disposition}
            >
              {isLoading ? "Menyimpan..." : "Simpan Disposisi"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
