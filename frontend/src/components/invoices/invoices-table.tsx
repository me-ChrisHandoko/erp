/**
 * Invoices Table Components
 *
 * Contains table components for both:
 * - Sales Invoices (InvoicesTable)
 * - Purchase Invoices (PurchaseInvoicesTable)
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
  FileText,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { PurchaseInvoiceResponse } from "@/types/purchase-invoice.types";
import type { InvoiceResponse } from "@/types/invoice.types";
import { useState } from "react";

// ============================================================================
// SALES INVOICES TABLE
// ============================================================================

interface InvoicesTableProps {
  data: InvoiceResponse[];
  isLoading: boolean;
  error: any;
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  onSortChange: (sortBy: string) => void;
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  canEdit: boolean;
  canDelete: boolean;
}

export function InvoicesTable({
  data,
  isLoading,
  error,
  currentPage,
  totalPages,
  onPageChange,
  onSortChange,
  sortBy = "invoiceDate",
  sortOrder = "desc",
  canEdit,
  canDelete,
}: InvoicesTableProps) {

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

  // Payment status badge helper
  const getPaymentStatusBadge = (status: string) => {
    const statusConfig = {
      UNPAID: { label: "Belum Dibayar", className: "bg-yellow-500 text-white hover:bg-yellow-600" },
      PARTIAL: { label: "Dibayar Sebagian", className: "bg-orange-500 text-white hover:bg-orange-600" },
      PAID: { label: "Lunas", className: "bg-green-500 text-white hover:bg-green-600" },
      OVERDUE: { label: "Jatuh Tempo", className: "bg-red-500 text-white hover:bg-red-600" },
    };
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.UNPAID;
    return <Badge className={config.className}>{config.label}</Badge>;
  };

  // Format date helper
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  // Check if due date is overdue
  const isOverdue = (dueDate: string, paymentStatus: string) => {
    if (paymentStatus === "PAID") return false;
    const due = new Date(dueDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return due < today;
  };

  if (error) {
    return (
      <div className="rounded-md border p-8">
        <EmptyState
          icon={FileText}
          title="Gagal memuat data faktur"
          description="Terjadi kesalahan saat memuat data. Silakan coba lagi."
        />
      </div>
    );
  }

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
                onClick={() => onSortChange("invoiceNumber")}
              >
                No. Faktur
                <SortIcon column="invoiceNumber" />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("invoiceDate")}
              >
                Tanggal
                <SortIcon column="invoiceDate" />
              </Button>
            </TableHead>
            <TableHead>Customer</TableHead>
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
            <TableHead className="text-center">Status Bayar</TableHead>
            <TableHead>Jatuh Tempo</TableHead>
            <TableHead className="w-[70px]">
              <span className="sr-only">Aksi</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={7} className="text-center">
                <div className="py-8">Memuat data...</div>
              </TableCell>
            </TableRow>
          ) : data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState
                    icon={FileText}
                    title="Faktur tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              data.map((invoice) => (
                <TableRow key={invoice.id}>
                  {/* Invoice Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    <Link
                      href={`/sales/invoices/${invoice.id}`}
                      className="hover:underline"
                    >
                      {invoice.invoiceNumber}
                    </Link>
                  </TableCell>

                  {/* Invoice Date */}
                  <TableCell>
                    {formatDate(invoice.invoiceDate)}
                  </TableCell>

                  {/* Customer */}
                  <TableCell>
                    <div className="font-medium">{invoice.customerName}</div>
                    {invoice.customerCode && (
                      <div className="text-sm text-muted-foreground">
                        {invoice.customerCode}
                      </div>
                    )}
                  </TableCell>

                  {/* Total Amount */}
                  <TableCell className="text-right font-medium">
                    Rp {Number(invoice.totalAmount).toLocaleString("id-ID")}
                  </TableCell>

                  {/* Payment Status */}
                  <TableCell className="text-center">
                    {getPaymentStatusBadge(invoice.paymentStatus)}
                  </TableCell>

                  {/* Due Date */}
                  <TableCell>
                    <div
                      className={
                        isOverdue(invoice.dueDate, invoice.paymentStatus)
                          ? "text-red-600 font-medium"
                          : ""
                      }
                    >
                      {formatDate(invoice.dueDate)}
                    </div>
                    {isOverdue(invoice.dueDate, invoice.paymentStatus) && (
                      <div className="text-xs text-red-600">Terlambat</div>
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
                            href={`/sales/invoices/${invoice.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && invoice.paymentStatus !== "PAID" && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/sales/invoices/${invoice.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Faktur
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
  );
}

// ============================================================================
// PURCHASE INVOICES TABLE
// ============================================================================

interface PurchaseInvoicesTableProps {
  invoices: PurchaseInvoiceResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
  canApprove: boolean;
}

export function PurchaseInvoicesTable({
  invoices,
  sortBy = "invoiceDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
}: PurchaseInvoicesTableProps) {
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

  // Invoice status badge helper
  const getInvoiceStatusBadge = (status: string) => {
    const statusConfig = {
      DRAFT: { label: "Draft", className: "bg-gray-500 text-white hover:bg-gray-600" },
      SUBMITTED: { label: "Submitted", className: "bg-blue-500 text-white hover:bg-blue-600" },
      APPROVED: { label: "Approved", className: "bg-green-500 text-white hover:bg-green-600" },
      REJECTED: { label: "Rejected", className: "bg-red-500 text-white hover:bg-red-600" },
      PAID: { label: "Paid", className: "bg-purple-500 text-white hover:bg-purple-600" },
      CANCELLED: { label: "Cancelled", className: "bg-gray-400 text-white hover:bg-gray-500" },
    };
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.DRAFT;
    return <Badge className={config.className}>{config.label}</Badge>;
  };

  // Payment status badge helper
  const getPaymentStatusBadge = (status: string) => {
    const statusConfig = {
      UNPAID: { label: "Belum Dibayar", className: "bg-yellow-500 text-white hover:bg-yellow-600" },
      PARTIAL: { label: "Dibayar Sebagian", className: "bg-orange-500 text-white hover:bg-orange-600" },
      PAID: { label: "Lunas", className: "bg-green-500 text-white hover:bg-green-600" },
      OVERDUE: { label: "Jatuh Tempo", className: "bg-red-500 text-white hover:bg-red-600" },
    };
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.UNPAID;
    return <Badge className={config.className}>{config.label}</Badge>;
  };

  // Format date helper
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  // Check if due date is overdue
  const isOverdue = (dueDate: string) => {
    const due = new Date(dueDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return due < today;
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
                  onClick={() => onSortChange("invoiceNumber")}
                >
                  No. Faktur
                  <SortIcon column="invoiceNumber" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("invoiceDate")}
                >
                  Tanggal
                  <SortIcon column="invoiceDate" />
                </Button>
              </TableHead>
              <TableHead>Supplier</TableHead>
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
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="text-center">Status Bayar</TableHead>
              <TableHead>Jatuh Tempo</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {invoices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8}>
                  <EmptyState
                    icon={FileText}
                    title="Faktur tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              invoices.map((invoice) => (
                <TableRow key={invoice.id}>
                  {/* Invoice Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    <Link
                      href={`/procurement/invoices/${invoice.id}`}
                      className="hover:underline"
                    >
                      {invoice.invoiceNumber}
                    </Link>
                  </TableCell>

                  {/* Invoice Date */}
                  <TableCell>
                    {formatDate(invoice.invoiceDate)}
                  </TableCell>

                  {/* Supplier */}
                  <TableCell>
                    <div className="font-medium">{invoice.supplierName}</div>
                    {invoice.supplierCode && (
                      <div className="text-sm text-muted-foreground">
                        {invoice.supplierCode}
                      </div>
                    )}
                  </TableCell>

                  {/* Total Amount */}
                  <TableCell className="text-right font-medium">
                    Rp {Number(invoice.totalAmount).toLocaleString("id-ID")}
                  </TableCell>

                  {/* Invoice Status */}
                  <TableCell className="text-center">
                    {getInvoiceStatusBadge(invoice.status)}
                  </TableCell>

                  {/* Payment Status */}
                  <TableCell className="text-center">
                    {getPaymentStatusBadge(invoice.paymentStatus)}
                  </TableCell>

                  {/* Due Date */}
                  <TableCell>
                    <div
                      className={
                        isOverdue(invoice.dueDate) && invoice.paymentStatus !== "PAID"
                          ? "text-red-600 font-medium"
                          : ""
                      }
                    >
                      {formatDate(invoice.dueDate)}
                    </div>
                    {isOverdue(invoice.dueDate) && invoice.paymentStatus !== "PAID" && (
                      <div className="text-xs text-red-600">Terlambat</div>
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
                            href={`/procurement/invoices/${invoice.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && invoice.status === "DRAFT" && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/procurement/invoices/${invoice.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Faktur
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
