/**
 * Goods Receipts Table Component
 *
 * Displays goods receipts in a sortable table with:
 * - Sortable columns (grnNumber, grnDate, status)
 * - Status badges with colors
 * - Action buttons (view detail)
 * - Responsive design
 */

"use client";

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
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  MoreHorizontal,
  PackageCheck,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import {
  getGoodsReceiptStatusLabel,
  getGoodsReceiptStatusColor,
  type GoodsReceiptResponse,
} from "@/types/goods-receipt.types";

interface GoodsReceiptsTableProps {
  receipts: GoodsReceiptResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
}

export function GoodsReceiptsTable({
  receipts,
  sortBy = "createdAt",
  sortOrder = "desc",
  onSortChange,
}: GoodsReceiptsTableProps) {
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
            <TableHead>Penerima</TableHead>
            <TableHead className="w-[70px]">
              <span className="sr-only">Aksi</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {receipts.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8}>
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

                {/* Status */}
                <TableCell className="text-center">
                  <Badge className={getGoodsReceiptStatusColor(receipt.status)}>
                    {getGoodsReceiptStatusLabel(receipt.status)}
                  </Badge>
                </TableCell>

                {/* Receiver */}
                <TableCell>
                  {receipt.receiver ? (
                    <div className="text-sm">
                      {receipt.receiver.fullName}
                    </div>
                  ) : (
                    <span className="text-sm text-muted-foreground">-</span>
                  )}
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
                          href={`/procurement/receipts/${receipt.id}`}
                          className="cursor-pointer"
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          Lihat Detail
                        </Link>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
