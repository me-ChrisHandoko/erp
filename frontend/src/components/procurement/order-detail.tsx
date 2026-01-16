/**
 * Purchase Order Detail Component
 *
 * Displays comprehensive purchase order information including:
 * - Order information (number, date, status)
 * - Supplier and warehouse details
 * - Line items with quantities and prices
 * - Totals and notes
 */

"use client";

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
import {
  ShoppingCart,
  Building,
  Warehouse,
  Calendar,
  User,
  FileText,
  Package,
} from "lucide-react";
import type { PurchaseOrderResponse } from "@/types/purchase-order.types";
import {
  getStatusLabel,
  getStatusBadgeColor,
  formatCurrency,
  formatDate,
} from "@/types/purchase-order.types";

interface OrderDetailProps {
  order: PurchaseOrderResponse;
}

export function OrderDetail({ order }: OrderDetailProps) {
  const items = order.items || [];

  return (
    <div className="space-y-6">
      {/* Order Status Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <ShoppingCart className="h-8 w-8 text-muted-foreground" />
              <div>
                <h2 className="text-2xl font-bold">{order.poNumber}</h2>
                <p className="text-sm text-muted-foreground">
                  Dibuat pada {formatDate(order.createdAt)}
                </p>
              </div>
            </div>
            <Badge className={`${getStatusBadgeColor(order.status)} text-sm px-4 py-1`}>
              {getStatusLabel(order.status)}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* Order Information */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Supplier Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Building className="h-4 w-4" />
              Informasi Supplier
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {order.supplier ? (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Nama Supplier</p>
                  <p className="font-medium">{order.supplier.name}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Kode Supplier</p>
                  <p className="font-mono text-sm">{order.supplier.code}</p>
                </div>
              </>
            ) : (
              <p className="text-muted-foreground">Data supplier tidak tersedia</p>
            )}
          </CardContent>
        </Card>

        {/* Warehouse Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Warehouse className="h-4 w-4" />
              Gudang Tujuan
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {order.warehouse ? (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Nama Gudang</p>
                  <p className="font-medium">{order.warehouse.name}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Kode Gudang</p>
                  <p className="font-mono text-sm">{order.warehouse.code}</p>
                </div>
              </>
            ) : (
              <p className="text-muted-foreground">Data gudang tidak tersedia</p>
            )}
          </CardContent>
        </Card>

        {/* Date Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Calendar className="h-4 w-4" />
              Tanggal
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-sm text-muted-foreground">Tanggal PO</p>
              <p className="font-medium">{formatDate(order.poDate)}</p>
            </div>
            {order.expectedDeliveryAt && (
              <div>
                <p className="text-sm text-muted-foreground">Estimasi Pengiriman</p>
                <p className="font-medium">{formatDate(order.expectedDeliveryAt)}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Requester Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <User className="h-4 w-4" />
              Pembuat
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {order.requester ? (
              <div>
                <p className="text-sm text-muted-foreground">Dibuat oleh</p>
                <p className="font-medium">{order.requester.fullName}</p>
              </div>
            ) : (
              <p className="text-muted-foreground">Data pembuat tidak tersedia</p>
            )}
            {order.approvedBy && order.approvedAt && (
              <div>
                <p className="text-sm text-muted-foreground">Disetujui pada</p>
                <p className="font-medium">{formatDate(order.approvedAt)}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Line Items */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Item Produk
          </CardTitle>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground">
              Tidak ada item dalam PO ini
            </p>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[50px]">No</TableHead>
                    <TableHead>Produk</TableHead>
                    <TableHead className="text-right">Qty</TableHead>
                    <TableHead className="text-right">Harga Satuan</TableHead>
                    <TableHead className="text-right">Diskon</TableHead>
                    <TableHead className="text-right">Subtotal</TableHead>
                    <TableHead className="text-right">Diterima</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item, index) => (
                    <TableRow key={item.id}>
                      <TableCell className="text-muted-foreground">
                        {index + 1}
                      </TableCell>
                      <TableCell>
                        {item.product ? (
                          <div>
                            <div className="font-medium">{item.product.name}</div>
                            <div className="text-xs text-muted-foreground">
                              {item.product.code}
                            </div>
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div>
                          {parseFloat(item.quantity).toLocaleString("id-ID")}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {item.productUnit?.unitName || item.product?.baseUnit || "-"}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(item.unitPrice)}
                      </TableCell>
                      <TableCell className="text-right">
                        {parseFloat(item.discountPct) > 0 ? (
                          <span>{parseFloat(item.discountPct)}%</span>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {formatCurrency(item.subtotal)}
                      </TableCell>
                      <TableCell className="text-right">
                        {parseFloat(item.receivedQty) > 0 ? (
                          <Badge
                            variant={
                              parseFloat(item.receivedQty) >=
                              parseFloat(item.quantity)
                                ? "default"
                                : "secondary"
                            }
                          >
                            {parseFloat(item.receivedQty).toLocaleString("id-ID")}
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Totals */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col items-end space-y-2">
            <div className="flex justify-between w-full max-w-xs">
              <span className="text-muted-foreground">Subtotal</span>
              <span className="font-medium">{formatCurrency(order.subtotal)}</span>
            </div>
            {parseFloat(order.discountAmount) > 0 && (
              <div className="flex justify-between w-full max-w-xs">
                <span className="text-muted-foreground">Diskon</span>
                <span className="font-medium text-red-600">
                  -{formatCurrency(order.discountAmount)}
                </span>
              </div>
            )}
            {parseFloat(order.taxAmount) > 0 && (
              <div className="flex justify-between w-full max-w-xs">
                <span className="text-muted-foreground">Pajak (PPN)</span>
                <span className="font-medium">{formatCurrency(order.taxAmount)}</span>
              </div>
            )}
            <Separator className="w-full max-w-xs" />
            <div className="flex justify-between w-full max-w-xs">
              <span className="text-lg font-semibold">Total</span>
              <span className="text-lg font-bold">
                {formatCurrency(order.totalAmount)}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notes */}
      {order.notes && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <FileText className="h-4 w-4" />
              Catatan
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm whitespace-pre-wrap">{order.notes}</p>
          </CardContent>
        </Card>
      )}

      {/* Cancellation Info */}
      {order.status === "CANCELLED" && order.cancellationNote && (
        <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/10">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base text-red-700 dark:text-red-400">
              Alasan Pembatalan
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-red-700 dark:text-red-400 whitespace-pre-wrap">
              {order.cancellationNote}
            </p>
            {order.cancelledAt && (
              <p className="text-xs text-red-500 mt-2">
                Dibatalkan pada {formatDate(order.cancelledAt)}
              </p>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
