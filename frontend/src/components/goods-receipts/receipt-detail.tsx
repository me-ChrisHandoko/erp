/**
 * Goods Receipt Detail Component
 *
 * Displays comprehensive goods receipt information including:
 * - Receipt information (GRN number, date, status)
 * - Supplier and warehouse details
 * - Line items with quantities
 * - Notes and timestamps
 */

"use client";

import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  PackageCheck,
  Building,
  Warehouse,
  Calendar,
  User,
  FileText,
  Package,
  ShoppingCart,
  MoreHorizontal,
  Settings,
  CheckCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  getGoodsReceiptStatusLabel,
  getGoodsReceiptStatusColor,
  getRejectionDispositionLabel,
  getRejectionDispositionColor,
  type GoodsReceiptResponse,
  type GoodsReceiptItemResponse,
} from "@/types/goods-receipt.types";

interface ReceiptDetailProps {
  receipt: GoodsReceiptResponse;
  onSetDisposition?: (item: GoodsReceiptItemResponse) => void;
  onResolveDisposition?: (item: GoodsReceiptItemResponse) => void;
}

// Format date to Indonesian locale
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });
};

// Format datetime to Indonesian locale
const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

export function ReceiptDetail({ receipt, onSetDisposition, onResolveDisposition }: ReceiptDetailProps) {
  const items = receipt.items || [];

  // Check if disposition actions are allowed (receipt must be ACCEPTED or PARTIAL and have rejected items)
  const canManageDisposition = receipt.status === "ACCEPTED" || receipt.status === "PARTIAL";

  return (
    <div className="space-y-6">
      {/* Receipt Status Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <PackageCheck className="h-8 w-8 text-muted-foreground" />
              <div>
                <h2 className="text-2xl font-bold">{receipt.grnNumber}</h2>
                <p className="text-sm text-muted-foreground">
                  Dibuat pada {formatDate(receipt.createdAt)}
                </p>
              </div>
            </div>
            <Badge className={`${getGoodsReceiptStatusColor(receipt.status)} text-sm px-4 py-1`}>
              {getGoodsReceiptStatusLabel(receipt.status)}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* Receipt Information */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* PO Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <ShoppingCart className="h-4 w-4" />
              Purchase Order
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {receipt.purchaseOrder ? (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Nomor PO</p>
                  <Link
                    href={`/procurement/orders/${receipt.purchaseOrderId}`}
                    className="font-mono text-primary hover:underline"
                  >
                    {receipt.purchaseOrder.poNumber}
                  </Link>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Tanggal PO</p>
                  <p className="font-medium">{formatDate(receipt.purchaseOrder.poDate)}</p>
                </div>
              </>
            ) : (
              <p className="text-muted-foreground">Data PO tidak tersedia</p>
            )}
          </CardContent>
        </Card>

        {/* Supplier Info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Building className="h-4 w-4" />
              Informasi Supplier
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {receipt.supplier ? (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Nama Supplier</p>
                  <p className="font-medium">{receipt.supplier.name}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Kode Supplier</p>
                  <p className="font-mono text-sm">{receipt.supplier.code}</p>
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
            {receipt.warehouse ? (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Nama Gudang</p>
                  <p className="font-medium">{receipt.warehouse.name}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Kode Gudang</p>
                  <p className="font-mono text-sm">{receipt.warehouse.code}</p>
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
              <p className="text-sm text-muted-foreground">Tanggal GRN</p>
              <p className="font-medium">{formatDate(receipt.grnDate)}</p>
            </div>
            {receipt.receivedAt && (
              <div>
                <p className="text-sm text-muted-foreground">Diterima</p>
                <p className="font-medium">{formatDateTime(receipt.receivedAt)}</p>
              </div>
            )}
            {receipt.inspectedAt && (
              <div>
                <p className="text-sm text-muted-foreground">Diperiksa</p>
                <p className="font-medium">{formatDateTime(receipt.inspectedAt)}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Receiver and Inspector Info */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Receiver */}
        {receipt.receiver && (
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <User className="h-4 w-4" />
                Penerima
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <p className="text-sm text-muted-foreground">Nama</p>
                <p className="font-medium">{receipt.receiver.fullName}</p>
              </div>
              {receipt.receivedAt && (
                <div>
                  <p className="text-sm text-muted-foreground">Diterima pada</p>
                  <p className="font-medium">{formatDateTime(receipt.receivedAt)}</p>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Inspector */}
        {receipt.inspector && (
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <User className="h-4 w-4" />
                Pemeriksa
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <p className="text-sm text-muted-foreground">Nama</p>
                <p className="font-medium">{receipt.inspector.fullName}</p>
              </div>
              {receipt.inspectedAt && (
                <div>
                  <p className="text-sm text-muted-foreground">Diperiksa pada</p>
                  <p className="font-medium">{formatDateTime(receipt.inspectedAt)}</p>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Supplier Documents */}
      {(receipt.supplierInvoice || receipt.supplierDONumber) && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <FileText className="h-4 w-4" />
              Dokumen Supplier
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {receipt.supplierInvoice && (
              <div>
                <p className="text-sm text-muted-foreground">No. Invoice Supplier</p>
                <p className="font-mono text-sm">{receipt.supplierInvoice}</p>
              </div>
            )}
            {receipt.supplierDONumber && (
              <div>
                <p className="text-sm text-muted-foreground">No. Surat Jalan Supplier</p>
                <p className="font-mono text-sm">{receipt.supplierDONumber}</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Line Items */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Item Barang
          </CardTitle>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground">
              Tidak ada item dalam penerimaan ini
            </p>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[50px]">No</TableHead>
                    <TableHead>Produk</TableHead>
                    <TableHead className="text-right">Qty Order</TableHead>
                    <TableHead className="text-right">Qty Diterima</TableHead>
                    <TableHead className="text-right">Qty Diterima OK</TableHead>
                    <TableHead className="text-right">Qty Ditolak</TableHead>
                    <TableHead>Disposisi</TableHead>
                    <TableHead>Batch / Kadaluarsa</TableHead>
                    <TableHead>Catatan</TableHead>
                    {canManageDisposition && onSetDisposition && <TableHead className="w-[50px]"></TableHead>}
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
                          {parseFloat(item.orderedQty).toLocaleString("id-ID")}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {item.productUnit?.unitName || item.product?.baseUnit || "-"}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="font-medium">
                          {parseFloat(item.receivedQty).toLocaleString("id-ID")}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        {parseFloat(item.acceptedQty) > 0 ? (
                          <Badge className="bg-green-100 text-green-800">
                            {parseFloat(item.acceptedQty).toLocaleString("id-ID")}
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        {parseFloat(item.rejectedQty) > 0 ? (
                          <Badge className="bg-red-100 text-red-800">
                            {parseFloat(item.rejectedQty).toLocaleString("id-ID")}
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {parseFloat(item.rejectedQty) > 0 ? (
                          <div className="space-y-1">
                            {item.rejectionDisposition ? (
                              <Badge className={getRejectionDispositionColor(item.rejectionDisposition)}>
                                {getRejectionDispositionLabel(item.rejectionDisposition)}
                              </Badge>
                            ) : (
                              <Badge variant="outline" className="text-muted-foreground">
                                Belum diatur
                              </Badge>
                            )}
                            {item.dispositionResolved && (
                              <div className="flex items-center gap-1 text-xs text-green-600">
                                <CheckCircle className="h-3 w-3" />
                                Selesai
                              </div>
                            )}
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {item.batchNumber ? (
                          <div>
                            <div className="font-mono text-sm">{item.batchNumber}</div>
                            {item.expiryDate && (
                              <div className="text-xs text-muted-foreground">
                                Exp: {formatDate(item.expiryDate)}
                              </div>
                            )}
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {item.rejectionReason ? (
                          <div className="text-sm text-red-600">
                            {item.rejectionReason}
                          </div>
                        ) : item.qualityNote ? (
                          <div className="text-sm">{item.qualityNote}</div>
                        ) : item.notes ? (
                          <div className="text-sm text-muted-foreground">{item.notes}</div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      {/* Disposition Actions */}
                      {canManageDisposition && onSetDisposition && (
                        <TableCell>
                          {parseFloat(item.rejectedQty) > 0 && (
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreHorizontal className="h-4 w-4" />
                                  <span className="sr-only">Menu aksi</span>
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => onSetDisposition(item)}>
                                  <Settings className="mr-2 h-4 w-4" />
                                  {item.rejectionDisposition ? "Ubah Disposisi" : "Atur Disposisi"}
                                </DropdownMenuItem>
                                {item.rejectionDisposition && !item.dispositionResolved && onResolveDisposition && (
                                  <DropdownMenuItem onClick={() => onResolveDisposition(item)}>
                                    <CheckCircle className="mr-2 h-4 w-4 text-green-600" />
                                    Selesaikan Disposisi
                                  </DropdownMenuItem>
                                )}
                              </DropdownMenuContent>
                            </DropdownMenu>
                          )}
                        </TableCell>
                      )}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Notes - show all notes sections */}
      {(receipt.notes || receipt.receiveNotes || receipt.inspectionNotes || receipt.acceptanceNotes || receipt.rejectionNotes) && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <FileText className="h-4 w-4" />
              Catatan
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {receipt.notes && (
              <div>
                <p className="text-sm font-medium text-muted-foreground mb-1">Catatan Umum</p>
                <p className="text-sm whitespace-pre-wrap bg-muted/30 p-3 rounded-md">{receipt.notes}</p>
              </div>
            )}
            {receipt.receiveNotes && (
              <div>
                <p className="text-sm font-medium text-blue-600 mb-1">Catatan Penerimaan</p>
                <p className="text-sm whitespace-pre-wrap bg-blue-50 p-3 rounded-md dark:bg-blue-900/20">{receipt.receiveNotes}</p>
              </div>
            )}
            {receipt.inspectionNotes && (
              <div>
                <p className="text-sm font-medium text-purple-600 mb-1">Catatan Inspeksi</p>
                <p className="text-sm whitespace-pre-wrap bg-purple-50 p-3 rounded-md dark:bg-purple-900/20">{receipt.inspectionNotes}</p>
              </div>
            )}
            {receipt.acceptanceNotes && (
              <div>
                <p className="text-sm font-medium text-green-600 mb-1">Catatan Penerimaan Akhir</p>
                <p className="text-sm whitespace-pre-wrap bg-green-50 p-3 rounded-md dark:bg-green-900/20">{receipt.acceptanceNotes}</p>
              </div>
            )}
            {receipt.rejectionNotes && (
              <div>
                <p className="text-sm font-medium text-red-600 mb-1">Catatan Penolakan</p>
                <p className="text-sm whitespace-pre-wrap bg-red-50 p-3 rounded-md dark:bg-red-900/20">{receipt.rejectionNotes}</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Status is REJECTED */}
      {receipt.status === "REJECTED" && (
        <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/10">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base text-red-700 dark:text-red-400">
              Barang Ditolak
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-red-700 dark:text-red-400">
              Penerimaan barang ini telah ditolak. Silakan cek alasan penolakan pada item.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
