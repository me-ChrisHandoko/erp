/**
 * Delivery Detail Component
 *
 * Displays comprehensive delivery information in organized cards:
 * - Header with status and basic info
 * - Customer and warehouse information
 * - Sales order reference
 * - Delivery items table
 * - Tracking information (driver/expedition)
 * - Proof of delivery (POD)
 * - Timeline/history
 */

"use client";

import Link from "next/link";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
  Building2,
  Package,
  FileText,
  Truck,
  MapPin,
  Calendar,
  Clock,
  User,
  Phone,
  Image as ImageIcon,
  FileSignature,
} from "lucide-react";
import {
  getDeliveryStatusLabel,
  getDeliveryStatusColor,
  getDeliveryTypeLabel,
} from "@/types/delivery.types";
import type { DeliveryResponse } from "@/types/delivery.types";

interface DeliveryDetailProps {
  delivery: DeliveryResponse;
}

export function DeliveryDetail({ delivery }: DeliveryDetailProps) {
  // Format date to Indonesian locale
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
    }).format(date);
  };

  // Format datetime to Indonesian locale
  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Status and Basic Info Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>Informasi Pengiriman</CardTitle>
              <CardDescription>
                Detail status dan informasi dasar pengiriman
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Badge
                className={`${getDeliveryStatusColor(
                  delivery.status
                )} text-white`}
              >
                {getDeliveryStatusLabel(delivery.status)}
              </Badge>
              <Badge variant="outline">
                {getDeliveryTypeLabel(delivery.type)}
              </Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-3 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Nomor Pengiriman
              </dt>
              <dd className="text-sm font-mono font-semibold mt-1">
                {delivery.deliveryNumber}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                Tanggal Pengiriman
              </dt>
              <dd className="text-sm font-semibold mt-1">
                {formatDate(delivery.deliveryDate)}
              </dd>
            </div>
            {delivery.salesOrder && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <FileText className="h-4 w-4" />
                  Sales Order
                </dt>
                <dd className="text-sm mt-1">
                  <Link
                    href={`/sales/orders/${delivery.salesOrderId}`}
                    className="font-mono text-blue-600 hover:underline"
                  >
                    {delivery.salesOrder.soNumber}
                  </Link>
                  <div className="text-xs text-muted-foreground">
                    {formatDate(delivery.salesOrder.soDate)}
                  </div>
                </dd>
              </div>
            )}
            {delivery.notes && (
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-muted-foreground">
                  Catatan
                </dt>
                <dd className="text-sm mt-1">{delivery.notes}</dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>

      {/* Customer Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Customer
          </CardTitle>
        </CardHeader>
        <CardContent>
          {delivery.customer ? (
            <dl className="space-y-3">
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Nama
                </dt>
                <dd className="text-sm font-semibold mt-1">
                  {delivery.customer.name}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Kode Customer
                </dt>
                <dd className="text-sm font-mono mt-1">
                  {delivery.customer.code}
                </dd>
              </div>
              {delivery.customer.phone && (
                <div>
                  <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <Phone className="h-4 w-4" />
                    Telepon
                  </dt>
                  <dd className="text-sm mt-1">{delivery.customer.phone}</dd>
                </div>
              )}
              {delivery.deliveryAddress && (
                <div>
                  <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <MapPin className="h-4 w-4" />
                    Alamat Pengiriman
                  </dt>
                  <dd className="text-sm mt-1 whitespace-pre-line">
                    {delivery.deliveryAddress}
                  </dd>
                </div>
              )}
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">
              Informasi customer tidak tersedia
            </p>
          )}
        </CardContent>
      </Card>

      {/* Warehouse Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Gudang Asal
          </CardTitle>
        </CardHeader>
        <CardContent>
          {delivery.warehouse ? (
            <dl className="space-y-3">
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Nama Gudang
                </dt>
                <dd className="text-sm font-semibold mt-1">
                  {delivery.warehouse.name}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Kode Gudang
                </dt>
                <dd className="text-sm font-mono mt-1">
                  {delivery.warehouse.code}
                </dd>
              </div>
            </dl>
          ) : (
            <p className="text-sm text-muted-foreground">
              Informasi gudang tidak tersedia
            </p>
          )}
        </CardContent>
      </Card>

      {/* Tracking Info Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Truck className="h-5 w-5" />
            Informasi Pengiriman & Tracking
          </CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-3 sm:grid-cols-3">
            {delivery.driverName && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <User className="h-4 w-4" />
                  Nama Sopir
                </dt>
                <dd className="text-sm font-semibold mt-1">
                  {delivery.driverName}
                </dd>
              </div>
            )}
            {delivery.vehicleNumber && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <Truck className="h-4 w-4" />
                  Nomor Kendaraan
                </dt>
                <dd className="text-sm font-mono font-semibold mt-1">
                  {delivery.vehicleNumber}
                </dd>
              </div>
            )}
            {delivery.expeditionService && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Ekspedisi
                </dt>
                <dd className="text-sm font-semibold mt-1">
                  {delivery.expeditionService}
                </dd>
              </div>
            )}
            {delivery.ttnkNumber && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Nomor Resi
                </dt>
                <dd className="text-sm font-mono mt-1">{delivery.ttnkNumber}</dd>
              </div>
            )}
            {delivery.departureTime && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  Waktu Berangkat
                </dt>
                <dd className="text-sm mt-1">
                  {formatDateTime(delivery.departureTime)}
                </dd>
              </div>
            )}
            {delivery.arrivalTime && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  Waktu Tiba
                </dt>
                <dd className="text-sm text-green-600 font-semibold mt-1">
                  {formatDateTime(delivery.arrivalTime)}
                </dd>
              </div>
            )}
            {delivery.receivedBy && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <User className="h-4 w-4" />
                  Diterima Oleh
                </dt>
                <dd className="text-sm font-semibold mt-1">
                  {delivery.receivedBy}
                </dd>
              </div>
            )}
            {delivery.receivedAt && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  Waktu Diterima
                </dt>
                <dd className="text-sm text-green-600 font-semibold mt-1">
                  {formatDateTime(delivery.receivedAt)}
                </dd>
              </div>
            )}
            {delivery.signatureUrl && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <FileSignature className="h-4 w-4" />
                  Tanda Tangan
                </dt>
                <dd className="text-sm mt-1">
                  <a
                    href={delivery.signatureUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline"
                  >
                    Lihat Tanda Tangan
                  </a>
                </dd>
              </div>
            )}
            {delivery.photoUrl && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <ImageIcon className="h-4 w-4" />
                  Foto Pengiriman
                </dt>
                <dd className="text-sm mt-1">
                  <a
                    href={delivery.photoUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline"
                  >
                    Lihat Foto
                  </a>
                </dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>

      {/* Delivery Items Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Item Pengiriman
          </CardTitle>
          <CardDescription>
            Daftar barang yang dikirim dalam pengiriman ini
          </CardDescription>
        </CardHeader>
        <CardContent>
          {delivery.items && delivery.items.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Produk</TableHead>
                    <TableHead className="text-center">Unit</TableHead>
                    <TableHead className="text-right">Qty</TableHead>
                    <TableHead>Batch</TableHead>
                    <TableHead>Catatan</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {delivery.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        {item.product ? (
                          <div>
                            <div className="font-medium">
                              {item.product.name}
                            </div>
                            <div className="text-xs text-muted-foreground font-mono">
                              {item.product.code}
                            </div>
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-center">
                        {item.productUnit ? (
                          <Badge variant="secondary">
                            {item.productUnit.name}
                          </Badge>
                        ) : (
                          <Badge variant="outline">
                            {item.product?.baseUnit || "-"}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {Number(item.quantity).toLocaleString("id-ID")}
                      </TableCell>
                      <TableCell>
                        {item.batch ? (
                          <div className="text-sm">
                            <div className="font-mono">
                              {item.batch.batchNumber}
                            </div>
                            {item.batch.expiryDate && (
                              <div className="text-xs text-muted-foreground">
                                Exp: {formatDate(item.batch.expiryDate)}
                              </div>
                            )}
                          </div>
                        ) : (
                          <span className="text-muted-foreground text-sm">
                            -
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {item.notes || "-"}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              Tidak ada item pengiriman
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
