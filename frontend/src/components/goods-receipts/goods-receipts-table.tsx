/**
 * Goods Receipts Table Component
 *
 * Displays goods receipts in a sortable table with:
 * - Sortable columns (grnNumber, grnDate, status)
 * - Status badges with colors
 * - Quick action buttons based on status
 * - Responsive design
 */

"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
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
  MoreHorizontal,
  PackageCheck,
  ClipboardCheck,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { toast } from "sonner";
import { EmptyState } from "@/components/shared/empty-state";
import {
  getGoodsReceiptStatusLabel,
  getGoodsReceiptStatusColor,
  type GoodsReceiptResponse,
} from "@/types/goods-receipt.types";
import {
  useReceiveGoodsMutation,
  useInspectGoodsMutation,
  useAcceptGoodsMutation,
  useRejectGoodsMutation,
} from "@/store/services/goodsReceiptApi";

interface GoodsReceiptsTableProps {
  receipts: GoodsReceiptResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
}

export function GoodsReceiptsTable({
  receipts,
  sortBy = "grnDate",
  sortOrder = "desc",
  onSortChange,
}: GoodsReceiptsTableProps) {
  const router = useRouter();

  // State for reject dialog
  const [rejectDialogOpen, setRejectDialogOpen] = useState(false);
  const [selectedReceipt, setSelectedReceipt] = useState<GoodsReceiptResponse | null>(null);
  const [rejectionReason, setRejectionReason] = useState("");

  // Mutations for quick actions
  const [receiveGoods, { isLoading: isReceiving }] = useReceiveGoodsMutation();
  const [inspectGoods, { isLoading: isInspecting }] = useInspectGoodsMutation();
  const [acceptGoods, { isLoading: isAccepting }] = useAcceptGoodsMutation();
  const [rejectGoods, { isLoading: isRejecting }] = useRejectGoodsMutation();

  const isProcessing = isReceiving || isInspecting || isAccepting || isRejecting;

  // Quick action handlers
  const handleReceive = async (receipt: GoodsReceiptResponse) => {
    try {
      await receiveGoods({ id: receipt.id }).unwrap();
      toast.success("Barang Diterima", {
        description: `${receipt.grnNumber} berhasil diupdate ke status Diterima`,
      });
    } catch (error: unknown) {
      const err = error as { data?: { error?: { message?: string } } };
      toast.error("Gagal Menerima Barang", {
        description: err?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleInspect = async (receipt: GoodsReceiptResponse) => {
    // For inspect, we navigate to detail page to fill inspection data
    router.push(`/procurement/receipts/${receipt.id}`);
  };

  const handleAccept = async (receipt: GoodsReceiptResponse) => {
    try {
      await acceptGoods({ id: receipt.id }).unwrap();
      toast.success("Barang Disetujui", {
        description: `${receipt.grnNumber} berhasil disetujui dan stok diupdate`,
      });
    } catch (error: unknown) {
      const err = error as { data?: { error?: { message?: string } } };
      toast.error("Gagal Menyetujui Barang", {
        description: err?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const handleRejectClick = (receipt: GoodsReceiptResponse) => {
    setSelectedReceipt(receipt);
    setRejectionReason("");
    setRejectDialogOpen(true);
  };

  const handleRejectConfirm = async () => {
    if (!selectedReceipt || !rejectionReason.trim()) return;

    try {
      await rejectGoods({
        id: selectedReceipt.id,
        data: { rejectionReason: rejectionReason.trim() },
      }).unwrap();
      toast.success("Barang Ditolak", {
        description: `${selectedReceipt.grnNumber} berhasil ditolak`,
      });
      setRejectDialogOpen(false);
      setSelectedReceipt(null);
      setRejectionReason("");
    } catch (error: unknown) {
      const err = error as { data?: { error?: { message?: string } } };
      toast.error("Gagal Menolak Barang", {
        description: err?.data?.error?.message || "Terjadi kesalahan",
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

  // Format date to Indonesian locale
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  };

  // Format datetime to Indonesian locale
  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("grnNumber")}
              >
                No. GRN
                <SortIcon column="grnNumber" />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("grnDate")}
              >
                Tanggal
                <SortIcon column="grnDate" />
              </Button>
            </TableHead>
            <TableHead>No. PO</TableHead>
            <TableHead>Supplier</TableHead>
            <TableHead>Gudang</TableHead>
            <TableHead>Invoice / DO</TableHead>
            <TableHead className="text-center">Item</TableHead>
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
            <TableHead>Penerima / Pemeriksa</TableHead>
            <TableHead className="w-[70px]">
              <span className="sr-only">Aksi</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {receipts.length === 0 ? (
            <TableRow>
              <TableCell colSpan={10}>
                <EmptyState
                  icon={PackageCheck}
                  title="Penerimaan barang tidak ditemukan"
                  description="Coba sesuaikan pencarian atau filter Anda"
                />
              </TableCell>
            </TableRow>
          ) : (
            receipts.map((receipt) => (
              <TableRow key={receipt.id}>
                {/* GRN Number */}
                <TableCell className="font-mono text-sm font-medium">
                  {receipt.grnNumber}
                </TableCell>

                {/* GRN Date */}
                <TableCell>
                  <div>{formatDate(receipt.grnDate)}</div>
                  {receipt.receivedAt && (
                    <div className="text-xs text-muted-foreground">
                      Diterima: {formatDateTime(receipt.receivedAt)}
                    </div>
                  )}
                </TableCell>

                {/* PO Number */}
                <TableCell>
                  {receipt.purchaseOrder ? (
                    <Link
                      href={`/procurement/orders/${receipt.purchaseOrderId}`}
                      className="font-mono text-sm text-primary hover:underline"
                    >
                      {receipt.purchaseOrder.poNumber}
                    </Link>
                  ) : (
                    <span className="text-sm text-muted-foreground">-</span>
                  )}
                </TableCell>

                {/* Supplier */}
                <TableCell>
                  {receipt.supplier ? (
                    <div>
                      <div className="font-medium">{receipt.supplier.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {receipt.supplier.code}
                      </div>
                    </div>
                  ) : (
                    <span className="text-sm text-muted-foreground">-</span>
                  )}
                </TableCell>

                {/* Warehouse */}
                <TableCell>
                  {receipt.warehouse ? (
                    <div>
                      <div className="font-medium">{receipt.warehouse.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {receipt.warehouse.code}
                      </div>
                    </div>
                  ) : (
                    <span className="text-sm text-muted-foreground">-</span>
                  )}
                </TableCell>

                {/* Invoice / DO Supplier */}
                <TableCell>
                  {receipt.supplierInvoice || receipt.supplierDONumber ? (
                    <div className="space-y-0.5">
                      {receipt.supplierInvoice && (
                        <div className="text-sm">
                          <span className="text-muted-foreground text-xs">Inv: </span>
                          {receipt.supplierInvoice}
                        </div>
                      )}
                      {receipt.supplierDONumber && (
                        <div className="text-sm">
                          <span className="text-muted-foreground text-xs">DO: </span>
                          {receipt.supplierDONumber}
                        </div>
                      )}
                    </div>
                  ) : (
                    <span className="text-sm text-muted-foreground">-</span>
                  )}
                </TableCell>

                {/* Item Count */}
                <TableCell className="text-center">
                  <span className="font-medium">{receipt.itemCount}</span>
                </TableCell>

                {/* Status */}
                <TableCell className="text-center">
                  <Badge className={getGoodsReceiptStatusColor(receipt.status)}>
                    {getGoodsReceiptStatusLabel(receipt.status)}
                  </Badge>
                </TableCell>

                {/* Receiver / Inspector */}
                <TableCell>
                  <div className="space-y-1">
                    {/* Receiver info */}
                    {receipt.receiver ? (
                      <div className="text-sm">
                        <span className="text-muted-foreground text-xs">Penerima: </span>
                        {receipt.receiver.fullName}
                      </div>
                    ) : receipt.status !== "PENDING" ? (
                      <div className="text-sm text-muted-foreground">-</div>
                    ) : null}

                    {/* Inspector info - show for INSPECTED, ACCEPTED, PARTIAL, REJECTED */}
                    {receipt.inspector && ["INSPECTED", "ACCEPTED", "PARTIAL", "REJECTED"].includes(receipt.status) && (
                      <div className="text-sm">
                        <span className="text-muted-foreground text-xs">Pemeriksa: </span>
                        {receipt.inspector.fullName}
                        {receipt.inspectedAt && (
                          <div className="text-xs text-muted-foreground">
                            {formatDateTime(receipt.inspectedAt)}
                          </div>
                        )}
                      </div>
                    )}

                    {/* Show dash for PENDING status with no info */}
                    {receipt.status === "PENDING" && !receipt.receiver && (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
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
                        disabled={isProcessing}
                      >
                        <span className="sr-only">Open menu</span>
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                      <DropdownMenuSeparator />

                      {/* View Detail - Always available */}
                      <DropdownMenuItem asChild>
                        <Link
                          href={`/procurement/receipts/${receipt.id}`}
                          className="cursor-pointer"
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          Lihat Detail
                        </Link>
                      </DropdownMenuItem>

                      {/* Quick Actions based on Status */}
                      {receipt.status === "PENDING" && (
                        <>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleReceive(receipt)}
                            disabled={isProcessing}
                            className="text-blue-600 focus:text-blue-600"
                          >
                            <PackageCheck className="mr-2 h-4 w-4" />
                            Terima Barang
                          </DropdownMenuItem>
                        </>
                      )}

                      {receipt.status === "RECEIVED" && (
                        <>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleInspect(receipt)}
                            disabled={isProcessing}
                            className="text-purple-600 focus:text-purple-600"
                          >
                            <ClipboardCheck className="mr-2 h-4 w-4" />
                            Periksa Barang
                          </DropdownMenuItem>
                        </>
                      )}

                      {receipt.status === "INSPECTED" && (
                        <>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleAccept(receipt)}
                            disabled={isProcessing}
                            className="text-green-600 focus:text-green-600"
                          >
                            <CheckCircle className="mr-2 h-4 w-4" />
                            Setujui
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleRejectClick(receipt)}
                            disabled={isProcessing}
                            className="text-red-600 focus:text-red-600"
                          >
                            <XCircle className="mr-2 h-4 w-4" />
                            Tolak
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

      {/* Reject Dialog */}
      <AlertDialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Tolak Penerimaan Barang</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menolak{" "}
              <span className="font-semibold">{selectedReceipt?.grnNumber}</span>?
              <br />
              <br />
              Barang yang ditolak tidak akan masuk ke stok gudang.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="rejectionReason">
              Alasan Penolakan <span className="text-destructive">*</span>
            </Label>
            <Input
              id="rejectionReason"
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              placeholder="Masukkan alasan penolakan..."
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRejecting}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRejectConfirm}
              disabled={isRejecting || !rejectionReason.trim()}
              className="bg-red-600 hover:bg-red-700"
            >
              {isRejecting ? "Memproses..." : "Tolak"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
