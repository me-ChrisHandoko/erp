/**
 * Transfer Detail Dialog Component
 *
 * Read-only view of stock transfer details including:
 * - Header information (number, date, warehouses, status)
 * - Items list with quantities
 * - Audit trail (shipped/received by and timestamps)
 * - Notes
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { TransferStatusBadge } from "./transfer-status-badge";
import {
  Package,
  Warehouse,
  Calendar,
  User,
  Clock,
  ArrowRight,
} from "lucide-react";
import type { StockTransfer } from "@/types/transfer.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";

interface TransferDetailDialogProps {
  transfer: StockTransfer | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function TransferDetailDialog({
  transfer,
  open,
  onOpenChange,
}: TransferDetailDialogProps) {
  if (!transfer) return null;

  // Format timestamp helper
  const formatTimestamp = (timestamp?: string) => {
    if (!timestamp) return "-";
    return format(new Date(timestamp), "dd MMM yyyy, HH:mm", {
      locale: localeId,
    });
  };

  // Format user ID for display (extract readable part)
  const formatUserId = (userId?: string) => {
    if (!userId) return "-";
    // Extract email username or first part before @ or UUID
    const emailMatch = userId.match(/^([^@]+)@/);
    if (emailMatch) return emailMatch[1];
    // For UUIDs or long IDs, show first 8 chars
    if (userId.length > 20) return userId.substring(0, 8);
    return userId;
  };

  // Calculate total quantity
  const totalQuantity =
    transfer.items?.reduce(
      (sum, item) => sum + parseFloat(item.quantity || "0"),
      0
    ) || 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader className="mt-5">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1">
              <DialogTitle className="text-xl font-bold">
                {transfer.transferNumber}
              </DialogTitle>
              <DialogDescription className="flex items-center gap-2 text-sm">
                <Calendar className="h-3.5 w-3.5" />
                {format(new Date(transfer.transferDate), "dd MMMM yyyy", {
                  locale: localeId,
                })}
              </DialogDescription>
            </div>
            <TransferStatusBadge status={transfer.status} />
          </div>
        </DialogHeader>

        <div className="space-y-6 pt-2">
          {/* Warehouse Info */}
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <Warehouse className="h-4 w-4" />
              Alur Transfer
            </div>
            <div className="grid gap-3">
              <div className="flex items-center gap-3 p-3 rounded-md border bg-muted/30">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <div className="text-xs text-muted-foreground whitespace-nowrap">
                    Dari:
                  </div>
                  <div className="font-medium truncate">
                    {transfer.sourceWarehouse?.name || "-"}
                  </div>
                  {transfer.sourceWarehouse?.code && (
                    <Badge
                      variant="outline"
                      className="text-[10px] font-mono flex-shrink-0"
                    >
                      {transfer.sourceWarehouse.code}
                    </Badge>
                  )}
                </div>
              </div>

              <div className="flex justify-center">
                <ArrowRight className="h-4 w-4 text-muted-foreground rotate-90" />
              </div>

              <div className="flex items-center gap-3 p-3 rounded-md border bg-muted/30">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <div className="text-xs text-muted-foreground whitespace-nowrap">
                    Ke:
                  </div>
                  <div className="font-medium truncate">
                    {transfer.destWarehouse?.name || "-"}
                  </div>
                  {transfer.destWarehouse?.code && (
                    <Badge
                      variant="outline"
                      className="text-[10px] font-mono flex-shrink-0"
                    >
                      {transfer.destWarehouse.code}
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          </div>

          <Separator />

          {/* Items Table */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Package className="h-4 w-4" />
                Daftar Item
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Total:</span>{" "}
                <span className="font-semibold">
                  {totalQuantity.toLocaleString("id-ID", {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 3,
                  })}{" "}
                  unit
                </span>
              </div>
            </div>

            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="w-[120px]">Kode</TableHead>
                    <TableHead>Nama Produk</TableHead>
                    <TableHead className="w-[100px]">Batch</TableHead>
                    <TableHead className="w-[100px] text-right">
                      Jumlah
                    </TableHead>
                    <TableHead className="w-[150px]">Catatan</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transfer.items && transfer.items.length > 0 ? (
                    transfer.items.map((item, index) => (
                      <TableRow key={index}>
                        <TableCell className="font-mono text-xs">
                          {item.product?.code || "-"}
                        </TableCell>
                        <TableCell className="font-medium">
                          {item.product?.name || "-"}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {item.batchId || "-"}
                        </TableCell>
                        <TableCell className="text-right font-semibold tabular-nums">
                          {parseFloat(item.quantity).toLocaleString("id-ID", {
                            minimumFractionDigits: 0,
                            maximumFractionDigits: 3,
                          })}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {item.notes || "-"}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={5}
                        className="text-center text-muted-foreground"
                      >
                        Tidak ada item
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </div>

          {/* Audit Trail - only for shipped/received transfers */}
          {(transfer.status === "SHIPPED" ||
            transfer.status === "RECEIVED") && (
            <>
              <Separator />
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  Riwayat
                </div>
                <div className="space-y-2">
                  {transfer.shippedAt && (
                    <div className="flex items-start gap-3 p-3 rounded-md border bg-muted/30">
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center gap-2">
                          <Badge
                            variant="secondary"
                            className="text-xs font-normal"
                          >
                            Dikirim
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {formatTimestamp(transfer.shippedAt)}
                          </span>
                        </div>
                        {transfer.shippedBy && (
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <User className="h-3 w-3" />
                            <span className="truncate">
                              {formatUserId(transfer.shippedBy)}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  {transfer.receivedAt && (
                    <div className="flex items-start gap-3 p-3 rounded-md border bg-muted/30">
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center gap-2">
                          <Badge
                            variant="secondary"
                            className="text-xs font-normal"
                          >
                            Diterima
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {formatTimestamp(transfer.receivedAt)}
                          </span>
                        </div>
                        {transfer.receivedBy && (
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <User className="h-3 w-3" />
                            <span className="truncate">
                              {formatUserId(transfer.receivedBy)}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}

          {/* Notes */}
          {transfer.notes && (
            <>
              <Separator />
              <div className="space-y-2">
                <div className="text-sm font-medium text-muted-foreground">
                  Catatan
                </div>
                <p className="text-sm p-3 rounded-md border bg-muted/30">
                  {transfer.notes}
                </p>
              </div>
            </>
          )}

          {/* Footer Metadata */}
          <div className="pt-2 border-t">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <Clock className="h-3 w-3" />
                Dibuat: {formatTimestamp(transfer.createdAt)}
              </div>
              <div className="flex items-center gap-1.5">
                <Clock className="h-3 w-3" />
                Diperbarui: {formatTimestamp(transfer.updatedAt)}
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
