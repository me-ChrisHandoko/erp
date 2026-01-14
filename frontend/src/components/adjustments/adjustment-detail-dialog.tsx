/**
 * Adjustment Detail Dialog Component
 *
 * Displays detailed information about an inventory adjustment including:
 * - Header information (number, date, warehouse, status)
 * - Type and reason
 * - List of adjusted items with quantities and values
 * - Audit trail (created by, approved by)
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { AdjustmentStatusBadge } from "./adjustment-status-badge";
import { AdjustmentTypeBadge } from "./adjustment-type-badge";
import {
  ADJUSTMENT_REASON_CONFIG,
  type InventoryAdjustment,
} from "@/types/adjustment.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";

interface AdjustmentDetailDialogProps {
  adjustment: InventoryAdjustment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AdjustmentDetailDialog({
  adjustment,
  open,
  onOpenChange,
}: AdjustmentDetailDialogProps) {
  if (!adjustment) return null;

  const reasonConfig = ADJUSTMENT_REASON_CONFIG[adjustment.reason];

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

  // Format quantity
  const formatQuantity = (value: string | number) => {
    const numValue = typeof value === "string" ? parseFloat(value) : value;
    return new Intl.NumberFormat("id-ID", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 3,
    }).format(numValue);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto overflow-x-hidden w-[95vw] sm:w-full">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            <span className="font-mono">{adjustment.adjustmentNumber}</span>
            <AdjustmentStatusBadge status={adjustment.status} />
          </DialogTitle>
          <DialogDescription>
            Detail penyesuaian stok inventori
          </DialogDescription>
        </DialogHeader>

        {/* Header Information */}
        <div className="rounded-lg border bg-background p-4 space-y-3">
          <div className="flex items-center justify-between">
            <AdjustmentTypeBadge type={adjustment.adjustmentType} />
            <Badge variant="outline">{reasonConfig?.label || adjustment.reason}</Badge>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Tanggal</p>
              <p className="text-sm font-medium">
                {format(new Date(adjustment.adjustmentDate), "dd MMMM yyyy", {
                  locale: localeId,
                })}
              </p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Gudang</p>
              <p className="text-sm font-medium">{adjustment.warehouse?.name || "-"}</p>
            </div>
          </div>

          {adjustment.notes && (
            <div className="space-y-1 pt-2 border-t">
              <p className="text-xs text-muted-foreground">Catatan</p>
              <p className="text-sm">{adjustment.notes}</p>
            </div>
          )}
        </div>

        {/* Items List */}
        <div className="py-4">
          <h4 className="font-semibold mb-3">Daftar Item ({adjustment.items?.length || 0})</h4>

          {adjustment.items && adjustment.items.length > 0 ? (
            <div className="space-y-2">
              {adjustment.items.map((item) => (
                <div
                  key={item.id}
                  className="p-3 rounded-lg border bg-background space-y-2"
                >
                  {/* Product Info */}
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <p className="font-mono text-xs text-muted-foreground">
                        {item.product?.code}
                      </p>
                      <p className="font-medium text-sm truncate">
                        {item.product?.name || "-"}
                      </p>
                    </div>
                    <div className="text-right flex-shrink-0">
                      <p className="font-semibold text-sm">
                        {formatCurrency(item.totalValue)}
                      </p>
                    </div>
                  </div>

                  {/* Quantity Details */}
                  <div className="grid grid-cols-3 gap-2 text-xs">
                    <div>
                      <p className="text-muted-foreground">Sebelum</p>
                      <p className="font-medium">{formatQuantity(item.quantityBefore)}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Penyesuaian</p>
                      <p className={`font-medium ${
                        adjustment.adjustmentType === "INCREASE"
                          ? "text-green-600"
                          : "text-red-600"
                      }`}>
                        {adjustment.adjustmentType === "INCREASE" ? "+" : "-"}
                        {formatQuantity(Math.abs(parseFloat(item.quantityAdjusted)))}
                      </p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Sesudah</p>
                      <p className="font-medium">{formatQuantity(item.quantityAfter)}</p>
                    </div>
                  </div>

                  {/* Price */}
                  <div className="text-xs text-muted-foreground">
                    Harga: {formatCurrency(item.unitCost)} × {formatQuantity(Math.abs(parseFloat(item.quantityAdjusted)))}
                  </div>

                  {/* Notes */}
                  {item.notes && (
                    <div className="text-xs text-muted-foreground italic">
                      Catatan: {item.notes}
                    </div>
                  )}
                </div>
              ))}

              {/* Total Row */}
              <div className="flex items-center justify-between p-3 rounded-lg bg-muted/50 font-semibold text-sm">
                <span>Total Nilai</span>
                <span>{formatCurrency(adjustment.totalValue)}</span>
              </div>
            </div>
          ) : (
            <div className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
              Tidak ada item
            </div>
          )}
        </div>

        <Separator />

        {/* Audit Trail */}
        <div className="py-4">
          <h4 className="font-semibold mb-3">Riwayat</h4>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Dibuat oleh</p>
              <p className="font-medium">
                {adjustment.createdByUser?.fullName || "-"}
              </p>
              <p className="text-xs text-muted-foreground">
                {format(new Date(adjustment.createdAt), "dd MMM yyyy HH:mm", {
                  locale: localeId,
                })}
              </p>
            </div>
            {adjustment.approvedBy && (
              <div>
                <p className="text-muted-foreground">Disetujui oleh</p>
                <p className="font-medium">
                  {adjustment.approvedByUser?.fullName || "-"}
                </p>
                {adjustment.approvedAt && (
                  <p className="text-xs text-muted-foreground">
                    {format(new Date(adjustment.approvedAt), "dd MMM yyyy HH:mm", {
                      locale: localeId,
                    })}
                  </p>
                )}
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
