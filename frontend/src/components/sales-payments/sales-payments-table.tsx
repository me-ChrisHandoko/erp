/**
 * Sales Payments Table Component
 *
 * Displays customer payments in a sortable table with:
 * - Sortable columns (payment number, date, customer, amount)
 * - Payment method badges
 * - Check status badges
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
  DollarSign,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { SalesPaymentResponse } from "@/types/sales-payment.types";
import { PAYMENT_METHOD_LABELS, CHECK_STATUS_LABELS } from "@/types/sales-payment.types";

interface SalesPaymentsTableProps {
  payments: SalesPaymentResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function SalesPaymentsTable({
  payments,
  sortBy = "paymentDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
}: SalesPaymentsTableProps) {
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

  // Format date
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
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
                  onClick={() => onSortChange("customerName")}
                >
                  Pelanggan
                  <SortIcon column="customerName" />
                </Button>
              </TableHead>
              <TableHead>Invoice</TableHead>
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
              <TableHead className="text-center">Status Cek/Giro</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {payments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8}>
                  <EmptyState
                    icon={DollarSign}
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
                    <div className="text-sm">{formatDate(payment.paymentDate)}</div>
                  </TableCell>

                  {/* Customer */}
                  <TableCell>
                    <div className="font-medium">{payment.customerName}</div>
                    {payment.customerCode && (
                      <div className="text-sm text-muted-foreground font-mono">
                        {payment.customerCode}
                      </div>
                    )}
                  </TableCell>

                  {/* Invoice Number */}
                  <TableCell>
                    <div className="font-mono text-sm">{payment.invoiceNumber}</div>
                  </TableCell>

                  {/* Amount */}
                  <TableCell className="text-right font-medium">
                    <div className="text-green-600 dark:text-green-400">
                      Rp {Number(payment.amount).toLocaleString("id-ID")}
                    </div>
                  </TableCell>

                  {/* Payment Method */}
                  <TableCell className="text-center">
                    <Badge
                      variant="outline"
                      className="bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800"
                    >
                      {PAYMENT_METHOD_LABELS[payment.paymentMethod]}
                    </Badge>
                  </TableCell>

                  {/* Check Status */}
                  <TableCell className="text-center">
                    {payment.checkStatus ? (
                      <Badge
                        className={
                          payment.checkStatus === 'CLEARED'
                            ? "bg-green-500 text-white hover:bg-green-600"
                            : payment.checkStatus === 'BOUNCED'
                            ? "bg-red-500 text-white hover:bg-red-600"
                            : payment.checkStatus === 'CANCELLED'
                            ? "bg-gray-500 text-white hover:bg-gray-600"
                            : "bg-yellow-500 text-white hover:bg-yellow-600"
                        }
                      >
                        {CHECK_STATUS_LABELS[payment.checkStatus]}
                      </Badge>
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
                            href={`/sales/payments/${payment.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/sales/payments/${payment.id}/edit`}
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
