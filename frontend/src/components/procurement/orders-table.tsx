/**
 * Purchase Orders Table Component
 *
 * Displays purchase orders in a sortable table with:
 * - Sortable columns (PO number, date, supplier, total, status)
 * - Status badges with workflow colors
 * - Action buttons (view, edit, confirm, cancel)
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  Pencil,
  ShoppingCart,
  MoreHorizontal,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { toast } from "sonner";
import {
  useConfirmPurchaseOrderMutation,
  useCancelPurchaseOrderMutation,
} from "@/store/services/purchaseOrderApi";
import type { PurchaseOrderResponse } from "@/types/purchase-order.types";
import {
  getStatusLabel,
  getStatusBadgeColor,
  formatCurrency,
  formatDate,
} from "@/types/purchase-order.types";

interface OrdersTableProps {
  orders: PurchaseOrderResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
  canConfirm: boolean;
  canCancel: boolean;
}

export function OrdersTable({
  orders,
  sortBy = "poDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
  canConfirm,
  canCancel,
}: OrdersTableProps) {
  // Cancel dialog state
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [orderToCancel, setOrderToCancel] = useState<PurchaseOrderResponse | null>(null);
  const [cancellationNote, setCancellationNote] = useState("");

  // Confirm dialog state
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);
  const [orderToConfirm, setOrderToConfirm] = useState<PurchaseOrderResponse | null>(null);

  // Mutations
  const [confirmOrder, { isLoading: isConfirming }] = useConfirmPurchaseOrderMutation();
  const [cancelOrder, { isLoading: isCancelling }] = useCancelPurchaseOrderMutation();

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

  const handleConfirmClick = (order: PurchaseOrderResponse) => {
    setOrderToConfirm(order);
    setConfirmDialogOpen(true);
  };

  const handleCancelClick = (order: PurchaseOrderResponse) => {
    setOrderToCancel(order);
    setCancellationNote("");
    setCancelDialogOpen(true);
  };

  const handleConfirmOrder = async () => {
    if (!orderToConfirm) return;

    try {
      await confirmOrder({ id: orderToConfirm.id }).unwrap();
      toast.success("PO Dikonfirmasi", {
        description: `${orderToConfirm.poNumber} berhasil dikonfirmasi`,
      });
      setConfirmDialogOpen(false);
      setOrderToConfirm(null);
    } catch (error: any) {
      toast.error("Gagal Mengkonfirmasi PO", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleCancelOrder = async () => {
    if (!orderToCancel || !cancellationNote.trim()) return;

    try {
      await cancelOrder({
        id: orderToCancel.id,
        data: { cancellationNote: cancellationNote.trim() },
      }).unwrap();
      toast.success("PO Dibatalkan", {
        description: `${orderToCancel.poNumber} berhasil dibatalkan`,
      });
      setCancelDialogOpen(false);
      setOrderToCancel(null);
      setCancellationNote("");
    } catch (error: any) {
      toast.error("Gagal Membatalkan PO", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
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
                  onClick={() => onSortChange("poNumber")}
                >
                  No. PO
                  <SortIcon column="poNumber" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("poDate")}
                >
                  Tanggal
                  <SortIcon column="poDate" />
                </Button>
              </TableHead>
              <TableHead>Supplier</TableHead>
              <TableHead>Gudang</TableHead>
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
              <TableHead className="text-center">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("status")}
                >
                  Status
                  <SortIcon column="status" />
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
                <TableCell colSpan={7}>
                  <EmptyState
                    icon={ShoppingCart}
                    title="Purchase Order tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              orders.map((order) => (
                <TableRow key={order.id}>
                  {/* PO Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    <Link
                      href={`/procurement/orders/${order.id}`}
                      className="hover:text-primary hover:underline"
                    >
                      {order.poNumber}
                    </Link>
                  </TableCell>

                  {/* Date */}
                  <TableCell>
                    <div className="text-sm">{formatDate(order.poDate)}</div>
                    {order.expectedDeliveryAt && (
                      <div className="text-xs text-muted-foreground">
                        ETA: {formatDate(order.expectedDeliveryAt)}
                      </div>
                    )}
                  </TableCell>

                  {/* Supplier */}
                  <TableCell>
                    {order.supplier ? (
                      <div>
                        <div className="font-medium">{order.supplier.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {order.supplier.code}
                        </div>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Warehouse */}
                  <TableCell>
                    {order.warehouse ? (
                      <div>
                        <div className="font-medium">{order.warehouse.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {order.warehouse.code}
                        </div>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Total Amount */}
                  <TableCell className="text-right font-medium">
                    {formatCurrency(order.totalAmount)}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge className={getStatusBadgeColor(order.status)}>
                      {getStatusLabel(order.status)}
                    </Badge>
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
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/procurement/orders/${order.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>

                        {/* Edit - Only for DRAFT */}
                        {canEdit && order.status === "DRAFT" && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/procurement/orders/${order.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit PO
                            </Link>
                          </DropdownMenuItem>
                        )}

                        {/* Confirm - Only for DRAFT */}
                        {canConfirm && order.status === "DRAFT" && (
                          <DropdownMenuItem
                            onClick={() => handleConfirmClick(order)}
                            className="cursor-pointer text-blue-600"
                          >
                            <CheckCircle className="mr-2 h-4 w-4" />
                            Konfirmasi PO
                          </DropdownMenuItem>
                        )}

                        {/* Cancel - Only for DRAFT (CONFIRMED requires detail page to check active GRN) */}
                        {canCancel && order.status === "DRAFT" && (
                          <DropdownMenuItem
                            onClick={() => handleCancelClick(order)}
                            className="cursor-pointer text-red-600"
                          >
                            <XCircle className="mr-2 h-4 w-4" />
                            Batalkan PO
                          </DropdownMenuItem>
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

      {/* Confirm Dialog */}
      <AlertDialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Konfirmasi Purchase Order</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin mengkonfirmasi{" "}
              <span className="font-semibold">{orderToConfirm?.poNumber}</span>?
              <br />
              <br />
              Setelah dikonfirmasi, PO tidak dapat diedit lagi dan menunggu
              penerimaan barang.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isConfirming}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmOrder}
              disabled={isConfirming}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isConfirming ? "Memproses..." : "Konfirmasi"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Cancel Dialog */}
      <AlertDialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Purchase Order</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin membatalkan{" "}
              <span className="font-semibold">{orderToCancel?.poNumber}</span>?
              <br />
              <br />
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="cancellationNote">
              Alasan Pembatalan <span className="text-destructive">*</span>
            </Label>
            <Input
              id="cancellationNote"
              value={cancellationNote}
              onChange={(e) => setCancellationNote(e.target.value)}
              placeholder="Masukkan alasan pembatalan..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isCancelling}>Kembali</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelOrder}
              disabled={isCancelling || !cancellationNote.trim()}
              className="bg-red-600 hover:bg-red-700"
            >
              {isCancelling ? "Memproses..." : "Batalkan PO"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
