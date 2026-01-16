/**
 * Payments Table Component
 *
 * Displays payments in a sortable table with:
 * - Sortable columns (payment number, date, supplier, amount)
 * - Payment method badges
 * - Action buttons (view, edit)
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
  Pencil,
  Wallet,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { PaymentResponse } from "@/types/payment.types";
import { PAYMENT_METHOD_LABELS } from "@/types/payment.types";

interface PaymentsTableProps {
  payments: PaymentResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function PaymentsTable({
  payments,
  sortBy = "paymentDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
}: PaymentsTableProps) {
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

  // Format date to Indonesian format
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
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
                  onClick={() => onSortChange("paymentNumber")}
                >
                  No. Pembayaran
                  <SortIcon column="paymentNumber" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("paymentDate")}
                >
                  Tanggal
                  <SortIcon column="paymentDate" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("supplierName")}
                >
                  Pemasok
                  <SortIcon column="supplierName" />
                </Button>
              </TableHead>
              <TableHead className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("amount")}
                >
                  Jumlah
                  <SortIcon column="amount" />
                </Button>
              </TableHead>
              <TableHead className="text-center">Metode</TableHead>
              <TableHead>Referensi</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {payments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState
                    icon={Wallet}
                    title="Pembayaran tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              payments.map((payment) => (
                <TableRow key={payment.id}>
                  {/* Payment Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    {payment.paymentNumber}
                  </TableCell>

                  {/* Payment Date */}
                  <TableCell>
                    {formatDate(payment.paymentDate)}
                  </TableCell>

                  {/* Supplier */}
                  <TableCell>
                    <div className="font-medium">{payment.supplierName}</div>
                    {payment.supplierCode && (
                      <div className="text-sm text-muted-foreground">
                        {payment.supplierCode}
                      </div>
                    )}
                  </TableCell>

                  {/* Amount */}
                  <TableCell className="text-right font-medium">
                    Rp {Number(payment.amount).toLocaleString("id-ID")}
                  </TableCell>

                  {/* Payment Method */}
                  <TableCell className="text-center">
                    <Badge variant="secondary">
                      {PAYMENT_METHOD_LABELS[payment.paymentMethod]}
                    </Badge>
                  </TableCell>

                  {/* Reference */}
                  <TableCell>
                    {payment.reference ? (
                      <div className="text-sm">
                        {payment.reference}
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
                            href={`/procurement/payments/${payment.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/procurement/payments/${payment.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Pembayaran
                            </Link>
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
    </>
  );
}
