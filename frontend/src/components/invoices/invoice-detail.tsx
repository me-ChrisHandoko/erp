/**
 * Invoice Detail Component
 *
 * Comprehensive invoice information display with:
 * - Invoice header (number, dates, supplier, status)
 * - Invoice items table (products, quantities, prices)
 * - Financial summary (subtotal, tax, discount, total)
 * - Payment information and history
 * - Workflow action buttons
 * - Notes and remarks
 */

"use client";

import {
  FileText,
  Building2,
  Calendar,
  DollarSign,
  Package,
  Receipt,
  AlertCircle,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { PurchaseInvoiceResponse } from "@/types/purchase-invoice.types";

interface InvoiceDetailProps {
  invoice: PurchaseInvoiceResponse;
}

export function InvoiceDetail({ invoice }: InvoiceDetailProps) {
  // Format date helper
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  // Invoice status badge helper
  const getInvoiceStatusBadge = (status: string) => {
    const statusConfig = {
      DRAFT: {
        label: "Draft",
        className: "bg-gray-500 text-white hover:bg-gray-600",
      },
      SUBMITTED: {
        label: "Submitted",
        className: "bg-blue-500 text-white hover:bg-blue-600",
      },
      APPROVED: {
        label: "Approved",
        className: "bg-green-500 text-white hover:bg-green-600",
      },
      REJECTED: {
        label: "Rejected",
        className: "bg-red-500 text-white hover:bg-red-600",
      },
      PAID: {
        label: "Paid",
        className: "bg-purple-500 text-white hover:bg-purple-600",
      },
      CANCELLED: {
        label: "Cancelled",
        className: "bg-gray-400 text-white hover:bg-gray-500",
      },
    };
    const config =
      statusConfig[status as keyof typeof statusConfig] || statusConfig.DRAFT;
    return <Badge className={config.className}>{config.label}</Badge>;
  };

  // Payment status badge helper
  const getPaymentStatusBadge = (status: string) => {
    const statusConfig = {
      UNPAID: {
        label: "Belum Dibayar",
        className: "bg-yellow-500 text-white hover:bg-yellow-600",
      },
      PARTIAL: {
        label: "Dibayar Sebagian",
        className: "bg-orange-500 text-white hover:bg-orange-600",
      },
      PAID: {
        label: "Lunas",
        className: "bg-green-500 text-white hover:bg-green-600",
      },
      OVERDUE: {
        label: "Jatuh Tempo",
        className: "bg-red-500 text-white hover:bg-red-600",
      },
    };
    const config =
      statusConfig[status as keyof typeof statusConfig] ||
      statusConfig.UNPAID;
    return <Badge className={config.className}>{config.label}</Badge>;
  };

  // Check if overdue
  const isOverdue = () => {
    const due = new Date(invoice.dueDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return due < today && invoice.paymentStatus !== "PAID";
  };

  // Calculate total line item discount (diskon per item)
  const lineDiscount = invoice.items?.reduce((sum, item) => {
    return sum + Number(item.discountAmount || 0);
  }, 0) || 0;

  // Header-level discount (diskon faktur)
  const headerDiscount = Number(invoice.discountAmount || 0);

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Invoice Header Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Informasi Faktur
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* Invoice Number */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Nomor Faktur
              </p>
              <p className="font-mono text-lg font-semibold">
                {invoice.invoiceNumber}
              </p>
            </div>

            {/* Invoice Date */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Tanggal Faktur
              </p>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <p className="text-sm font-medium">
                  {formatDate(invoice.invoiceDate)}
                </p>
              </div>
            </div>

            {/* Due Date */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Jatuh Tempo
              </p>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <p
                  className={`text-sm font-medium ${
                    isOverdue() ? "text-red-600" : ""
                  }`}
                >
                  {formatDate(invoice.dueDate)}
                  {isOverdue() && (
                    <span className="ml-2 text-xs">(Terlambat)</span>
                  )}
                </p>
              </div>
            </div>

            {/* Supplier */}
            <div className="space-y-1 md:col-span-2">
              <p className="text-sm font-medium text-muted-foreground">
                Supplier
              </p>
              <div className="flex items-center gap-2">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="font-semibold">{invoice.supplierName}</p>
                  {invoice.supplierCode && (
                    <p className="text-xs text-muted-foreground">
                      {invoice.supplierCode}
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* Status */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Status
              </p>
              <div className="flex gap-2">
                {getInvoiceStatusBadge(invoice.status)}
                {getPaymentStatusBadge(invoice.paymentStatus)}
              </div>
            </div>
          </div>

          {/* Purchase Order Reference */}
          {invoice.poNumber && (
            <>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Referensi PO
                </p>
                <p className="font-mono text-sm">
                  {invoice.poNumber}
                </p>
              </div>
            </>
          )}

          {/* Notes */}
          {invoice.notes && (
            <>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Catatan
                </p>
                <p className="text-sm leading-relaxed">{invoice.notes}</p>
              </div>
            </>
          )}

          {/* Timestamps */}
          <Separator />
          <div className="grid gap-4 text-xs text-muted-foreground md:grid-cols-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Dibuat:{" "}
                {new Date(invoice.createdAt).toLocaleDateString("id-ID")}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Diperbarui:{" "}
                {new Date(invoice.updatedAt).toLocaleDateString("id-ID")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Items Table */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Item Faktur
          </CardTitle>
        </CardHeader>
        <CardContent>
          {invoice.items && invoice.items.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Produk</TableHead>
                    <TableHead className="text-center">Kuantitas</TableHead>
                    <TableHead className="text-right">Harga Satuan</TableHead>
                    <TableHead className="text-right">Diskon</TableHead>
                    <TableHead className="text-right">PPN</TableHead>
                    <TableHead className="text-right">Total</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {invoice.items.map((item, index) => (
                    <TableRow key={index}>
                      <TableCell className="font-medium">
                        <div>
                          <span>{item.productName}</span>
                        </div>
                        {item.productCode && (
                          <p className="text-xs text-muted-foreground">
                            {item.productCode}
                          </p>
                        )}
                      </TableCell>
                      <TableCell className="text-center">
                        {Number(item.quantity).toLocaleString("id-ID")}{" "}
                        {item.unitName}
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        Rp {Number(item.unitPrice).toLocaleString("id-ID")}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.discountAmount
                          ? `Rp ${Number(item.discountAmount).toLocaleString("id-ID")}`
                          : "-"}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.taxAmount
                          ? `Rp ${Number(item.taxAmount).toLocaleString("id-ID")}`
                          : "-"}
                      </TableCell>
                      <TableCell className="text-right font-semibold">
                        Rp {Number(item.lineTotal).toLocaleString("id-ID")}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Package className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>Belum ada item dalam faktur</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Financial Summary Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Ringkasan Keuangan
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Subtotal */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-muted-foreground">
              Subtotal
            </p>
            <p className="font-mono font-semibold">
              Rp {Number(invoice.subtotalAmount).toLocaleString("id-ID")}
            </p>
          </div>

          <Separator />

          {/* Diskon Item (dari masing-masing item) */}
          {lineDiscount > 0 && (
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-muted-foreground">
                Diskon Item
              </p>
              <p className="font-mono font-semibold text-green-600">
                - Rp {lineDiscount.toLocaleString("id-ID")}
              </p>
            </div>
          )}

          {/* Diskon Faktur (header-level discount) */}
          {headerDiscount > 0 && (
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-muted-foreground">
                Diskon Faktur
              </p>
              <p className="font-mono font-semibold text-green-600">
                - Rp {headerDiscount.toLocaleString("id-ID")}
              </p>
            </div>
          )}

          {/* Separator setelah diskon jika ada */}
          {(lineDiscount > 0 || headerDiscount > 0) && <Separator />}

          {/* Tax */}
          {invoice.taxAmount && Number(invoice.taxAmount) > 0 && (
            <>
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-muted-foreground">
                  PPN ({invoice.taxRate || 0}%)
                </p>
                <p className="font-mono font-semibold">
                  Rp {Number(invoice.taxAmount).toLocaleString("id-ID")}
                </p>
              </div>
              <Separator />
            </>
          )}

          {/* Total Amount */}
          <div className="flex items-center justify-between rounded-lg bg-muted/50 p-4">
            <p className="text-base font-bold">Total</p>
            <p className="text-2xl font-bold text-blue-600">
              Rp {Number(invoice.totalAmount).toLocaleString("id-ID")}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Payment Information Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Receipt className="h-5 w-5" />
            Informasi Pembayaran
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Payment Status */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-muted-foreground">
              Status Pembayaran
            </p>
            {getPaymentStatusBadge(invoice.paymentStatus)}
          </div>

          <Separator />

          {/* Amount Paid */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-muted-foreground">
              Jumlah Terbayar
            </p>
            <p className="font-mono font-semibold text-green-600">
              Rp {Number(invoice.paidAmount || 0).toLocaleString("id-ID")}
            </p>
          </div>

          <Separator />

          {/* Remaining Amount */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-muted-foreground">
              Sisa Pembayaran
            </p>
            <p
              className={`font-mono font-semibold ${
                Number(invoice.remainingAmount || 0) > 0
                  ? "text-orange-600"
                  : "text-green-600"
              }`}
            >
              Rp{" "}
              {Number(invoice.remainingAmount || 0).toLocaleString("id-ID")}
            </p>
          </div>

          {/* Overdue Alert */}
          {isOverdue() && (
            <>
              <Separator />
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  Faktur ini telah melewati jatuh tempo pembayaran
                </AlertDescription>
              </Alert>
            </>
          )}
        </CardContent>
      </Card>

      {/* Approval/Rejection Information */}
      {(invoice.approvedBy || invoice.rejectedBy) && (
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {invoice.approvedBy ? (
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              ) : (
                <XCircle className="h-5 w-5 text-red-600" />
              )}
              {invoice.approvedBy ? "Informasi Approval" : "Informasi Penolakan"}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              {invoice.approvedBy && (
                <>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-muted-foreground">
                      Disetujui Oleh
                    </p>
                    <p className="font-semibold">{invoice.approvedBy}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-muted-foreground">
                      Tanggal Disetujui
                    </p>
                    <p className="text-sm">
                      {invoice.approvedAt
                        ? formatDate(invoice.approvedAt)
                        : "-"}
                    </p>
                  </div>
                </>
              )}
              {invoice.rejectedBy && (
                <>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-muted-foreground">
                      Ditolak Oleh
                    </p>
                    <p className="font-semibold">{invoice.rejectedBy}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-muted-foreground">
                      Tanggal Ditolak
                    </p>
                    <p className="text-sm">
                      {invoice.rejectedAt
                        ? formatDate(invoice.rejectedAt)
                        : "-"}
                    </p>
                  </div>
                </>
              )}
            </div>

            {/* Rejection Reason */}
            {invoice.rejectedReason && (
              <>
                <Separator />
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">
                    Alasan Penolakan
                  </p>
                  <p className="text-sm leading-relaxed">
                    {invoice.rejectedReason}
                  </p>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
