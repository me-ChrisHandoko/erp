/**
 * Sales Orders Table Component
 *
 * Displays sales orders in a sortable table with:
 * - Sortable columns (order number, date, customer, amount)
 * - Status badges (color-coded by status)
 * - Action buttons (view, edit, cancel)
 * - Responsive design
 */

"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  Pencil,
  ShoppingCart,
  MoreHorizontal,
  XCircle,
  Loader2,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { formatCurrency, formatDate } from "@/lib/utils";
import { useCancelSalesOrderMutation } from "@/store/services/salesOrderApi";
import { toast } from "sonner";
import type { SalesOrderResponse } from "@/types/sales-order.types";
import {
  SALES_ORDER_STATUS_LABELS,
  SALES_ORDER_STATUS_STYLES,
  canEditOrder,
  canCancelOrder,
} from "@/types/sales-order.types";

interface SalesOrdersTableProps {
  orders: SalesOrderResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
  canCancel: boolean;
}

export function SalesOrdersTable({
  orders,
  sortBy = "orderDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
  canCancel,
}: SalesOrdersTableProps) {
  const [cancelSalesOrder, { isLoading: isCancelling }] = useCancelSalesOrderMutation();
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<SalesOrderResponse | null>(null);
  const [cancelReason, setCancelReason] = useState("");

  const handleCancelClick = (order: SalesOrderResponse) => {
    setSelectedOrder(order);
    setShowCancelDialog(true);
  };

  const handleCancelOrder = async () => {
    if (!selectedOrder) return;

    if (!cancelReason.trim() || cancelReason.trim().length < 5) {
      toast.error("Alasan Tidak Valid", {
        description: "Mohon masukkan alasan pembatalan minimal 5 karakter",
      });
      return;
    }

    try {
      await cancelSalesOrder({
        id: selectedOrder.id,
        reason: cancelReason,
      }).unwrap();

      toast.success("Pesanan Dibatalkan", {
        description: `Pesanan ${selectedOrder.orderNumber} berhasil dibatalkan`,
      });

      setShowCancelDialog(false);
      setSelectedOrder(null);
      setCancelReason("");
    } catch (error: any) {
      toast.error("Gagal Membatalkan Pesanan", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          "Terjadi kesalahan saat membatalkan pesanan",
      });
    }
  };

  // Sort icon component
  const SortIcon = ({ column }: { column: string }) => {
    if (sortBy !== column) {
      return <ArrowUpDown className="ml-2 h-4 w-4 text-muted-foreground" />;
    }
    return sortOrder === "asc" ? (
      <ArrowUp className="ml-2 h-4 w-4" />
    ) : (
      <ArrowDown className="ml-2 h-4 w-4" />
    );
  };

  return (
    <>
      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("orderNumber")}
                >
                  Nomor Pesanan
                  <SortIcon column="orderNumber" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("orderDate")}
                >
                  Tanggal
                  <SortIcon column="orderDate" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("customerName")}
                >
                  Pelanggan
                  <SortIcon column="customerName" />
                </Button>
              </TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("totalAmount")}
                >
                  Total
                  <SortIcon column="totalAmount" />
                </Button>
              </TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {orders.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6}>
                  <EmptyState
                    icon={ShoppingCart}
                    title="Pesanan tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              orders.map((order) => (
                <TableRow key={order.id}>
                  {/* Order Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    {order.orderNumber}
                  </TableCell>

                  {/* Date */}
                  <TableCell>
                    <div className="font-medium">
                      {formatDate(order.orderDate)}
                    </div>
                    {order.requiredDate && (
                      <div className="text-xs text-muted-foreground">
                        Dibutuhkan: {formatDate(order.requiredDate)}
                      </div>
                    )}
                  </TableCell>

                  {/* Customer */}
                  <TableCell>
                    <div className="font-medium">{order.customerName}</div>
                    <div className="text-sm text-muted-foreground">
                      {order.customerCode}
                    </div>
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge className={SALES_ORDER_STATUS_STYLES[order.status]}>
                      {SALES_ORDER_STATUS_LABELS[order.status]}
                    </Badge>
                  </TableCell>

                  {/* Total Amount */}
                  <TableCell className="text-right font-medium">
                    <div>{formatCurrency(order.totalAmount)}</div>
                    <div className="text-xs text-muted-foreground">
                      {order.items?.length || 0} item
                    </div>
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                        >
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                        <DropdownMenuSeparator />

                        {/* View - always available */}
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/sales/orders/${order.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>

                        {/* Edit - only for DRAFT */}
                        {canEdit && canEditOrder(order.status) && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/sales/orders/${order.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Pesanan
                            </Link>
                          </DropdownMenuItem>
                        )}

                        {/* Cancel - not for final states */}
                        {canCancel && canCancelOrder(order.status) && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive focus:text-destructive"
                              onClick={() => handleCancelClick(order)}
                            >
                              <XCircle className="mr-2 h-4 w-4" />
                              Batalkan Pesanan
                            </DropdownMenuItem>
                          </>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Cancel Confirmation Dialog */}
      <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Pesanan?</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin membatalkan pesanan{" "}
              <strong>{selectedOrder?.orderNumber}</strong>? Tindakan ini tidak
              dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <div className="space-y-2 py-4">
            <Label htmlFor="cancelReasonTable">
              Alasan Pembatalan <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="cancelReasonTable"
              placeholder="Masukkan alasan pembatalan pesanan (minimal 5 karakter)..."
              value={cancelReason}
              onChange={(e) => setCancelReason(e.target.value)}
              rows={4}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground">
              Alasan pembatalan akan disimpan dalam riwayat pesanan
            </p>
          </div>

          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setShowCancelDialog(false);
                setSelectedOrder(null);
                setCancelReason("");
              }}
              disabled={isCancelling}
            >
              Batal
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelOrder}
              disabled={isCancelling || cancelReason.trim().length < 5}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isCancelling ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Membatalkan...
                </>
              ) : (
                "Ya, Batalkan Pesanan"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
