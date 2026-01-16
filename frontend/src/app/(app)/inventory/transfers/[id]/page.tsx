/**
 * Transfer Detail Page
 *
 * Full-page view of stock transfer details with actions:
 * - Transfer information (number, date, warehouses, status)
 * - Items list with quantities
 * - Status-based actions (Edit, Ship, Receive, Cancel, Delete)
 * - Audit trail
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import {
  Package,
  AlertCircle,
  Warehouse,
  Calendar,
  ArrowRight,
  Edit,
  Truck,
  CheckCircle,
  XCircle,
  Trash2,
  ArrowLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { TransferStatusBadge } from "@/components/transfers/transfer-status-badge";
import { ShipTransferDialog } from "@/components/transfers/ship-transfer-dialog";
import { ReceiveTransferDialog } from "@/components/transfers/receive-transfer-dialog";
import { CancelTransferDialog } from "@/components/transfers/cancel-transfer-dialog";
import { DeleteTransferDialog } from "@/components/transfers/delete-transfer-dialog";
import { useGetTransferQuery } from "@/store/services/transferApi";
import { usePermissions } from "@/hooks/use-permissions";
import type { StockTransfer } from "@/types/transfer.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { useState } from "react";

export default function TransferDetailPage() {
  const params = useParams();
  const router = useRouter();
  const transferId = params.id as string;

  const { data: transfer, isLoading, error, refetch } = useGetTransferQuery(transferId);
  const permissions = usePermissions();

  // Action dialogs state
  const [isShipDialogOpen, setIsShipDialogOpen] = useState(false);
  const [isReceiveDialogOpen, setIsReceiveDialogOpen] = useState(false);
  const [isCancelDialogOpen, setIsCancelDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  // Permission checks
  const canEdit = permissions.canEdit('stock-transfers');
  const canDelete = permissions.canDelete('stock-transfers');
  const canApprove = permissions.can('approve', 'stock-transfers');

  const handleActionSuccess = () => {
    refetch();
  };

  const handleDeleteSuccess = () => {
    router.push("/inventory/transfers");
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Transfer Gudang", href: "/inventory/transfers" },
            { label: "Detail Transfer" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </div>
    );
  }

  // Error state
  if (error || !transfer) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Transfer Gudang", href: "/inventory/transfers" },
            { label: "Detail Transfer" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data transfer" : "Transfer tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/inventory/transfers")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Transfer
          </Button>
        </div>
      </div>
    );
  }

  // Calculate total quantity
  const totalQuantity = transfer.items?.reduce(
    (sum, item) => sum + parseFloat(item.quantity || "0"),
    0
  ) || 0;

  // Format timestamp
  const formatTimestamp = (timestamp?: string) => {
    if (!timestamp) return "-";
    return format(new Date(timestamp), "dd MMM yyyy, HH:mm", { locale: localeId });
  };

  // Determine available actions based on status
  const canEditTransfer = canEdit && transfer.status === "DRAFT";
  const canShip = canApprove && transfer.status === "DRAFT";
  const canReceive = canApprove && transfer.status === "SHIPPED";
  const canCancel = canApprove && (transfer.status === "DRAFT" || transfer.status === "SHIPPED");
  const canDeleteTransfer = canDelete && transfer.status === "DRAFT";

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Transfer Gudang", href: "/inventory/transfers" },
          { label: transfer.transferNumber },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header with actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold tracking-tight">
                {transfer.transferNumber}
              </h1>
              <TransferStatusBadge status={transfer.status} />
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Calendar className="h-3.5 w-3.5" />
              {format(new Date(transfer.transferDate), "dd MMMM yyyy", { locale: localeId })}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-2">
            {canEditTransfer && (
              <Button
                variant="outline"
                onClick={() => router.push(`/inventory/transfers/${transferId}/edit`)}
              >
                <Edit className="mr-2 h-4 w-4" />
                Edit
              </Button>
            )}
            {canShip && (
              <Button onClick={() => setIsShipDialogOpen(true)}>
                <Truck className="mr-2 h-4 w-4" />
                Kirim
              </Button>
            )}
            {canReceive && (
              <Button onClick={() => setIsReceiveDialogOpen(true)}>
                <CheckCircle className="mr-2 h-4 w-4" />
                Terima
              </Button>
            )}
            {canCancel && (
              <Button
                variant="destructive"
                onClick={() => setIsCancelDialogOpen(true)}
              >
                <XCircle className="mr-2 h-4 w-4" />
                Batalkan
              </Button>
            )}
            {canDeleteTransfer && (
              <Button
                variant="destructive"
                onClick={() => setIsDeleteDialogOpen(true)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Hapus
              </Button>
            )}
          </div>
        </div>

        {/* Transfer Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="h-5 w-5" />
              Informasi Transfer
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Warehouses */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Dari Gudang</p>
                <div className="flex items-center gap-2">
                  <Warehouse className="h-4 w-4 text-muted-foreground" />
                  <p className="font-medium">{transfer.sourceWarehouse?.name || "-"}</p>
                </div>
              </div>
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Ke Gudang</p>
                <div className="flex items-center gap-2">
                  <Warehouse className="h-4 w-4 text-muted-foreground" />
                  <p className="font-medium">{transfer.destWarehouse?.name || "-"}</p>
                </div>
              </div>
            </div>

            {/* Notes */}
            {transfer.notes && (
              <>
                <Separator />
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">Catatan</p>
                  <p className="text-sm bg-muted p-3 rounded-md">{transfer.notes}</p>
                </div>
              </>
            )}

            {/* Audit Trail */}
            {(transfer.shippedBy || transfer.receivedBy) && (
              <>
                <Separator />
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {transfer.shippedBy && (
                    <div className="space-y-2">
                      <p className="text-sm text-muted-foreground">Dikirim Oleh</p>
                      <p className="text-sm">{transfer.shippedBy}</p>
                      <p className="text-xs text-muted-foreground">
                        {formatTimestamp(transfer.shippedAt)}
                      </p>
                    </div>
                  )}
                  {transfer.receivedBy && (
                    <div className="space-y-2">
                      <p className="text-sm text-muted-foreground">Diterima Oleh</p>
                      <p className="text-sm">{transfer.receivedBy}</p>
                      <p className="text-xs text-muted-foreground">
                        {formatTimestamp(transfer.receivedAt)}
                      </p>
                    </div>
                  )}
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Items List */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <Package className="h-5 w-5" />
                Daftar Produk ({transfer.items?.length || 0})
              </CardTitle>
              <Badge variant="secondary">
                Total: {totalQuantity.toLocaleString("id-ID", {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 3,
                })}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Kode</TableHead>
                    <TableHead>Nama Produk</TableHead>
                    <TableHead className="text-right">Jumlah</TableHead>
                    <TableHead>Catatan</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transfer.items?.map((item, index) => (
                    <TableRow key={index}>
                      <TableCell className="font-mono text-sm">
                        {item.product?.code || "-"}
                      </TableCell>
                      <TableCell className="font-medium">
                        {item.product?.name || "-"}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {parseFloat(item.quantity).toLocaleString("id-ID", {
                          minimumFractionDigits: 0,
                          maximumFractionDigits: 3,
                        })}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {item.notes || "-"}
                      </TableCell>
                    </TableRow>
                  ))}
                  {/* Total Row */}
                  <TableRow className="bg-muted/50 font-semibold">
                    <TableCell colSpan={2} className="text-right">
                      Total
                    </TableCell>
                    <TableCell className="text-right">
                      {totalQuantity.toLocaleString("id-ID", {
                        minimumFractionDigits: 0,
                        maximumFractionDigits: 3,
                      })}
                    </TableCell>
                    <TableCell></TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        {/* Back Button */}
        <Button
          variant="outline"
          className="w-fit"
          onClick={() => router.push("/inventory/transfers")}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Kembali ke Daftar Transfer
        </Button>
      </div>

      {/* Action Dialogs */}
      <ShipTransferDialog
        transfer={transfer}
        open={isShipDialogOpen}
        onOpenChange={setIsShipDialogOpen}
        onSuccess={handleActionSuccess}
      />

      <ReceiveTransferDialog
        transfer={transfer}
        open={isReceiveDialogOpen}
        onOpenChange={setIsReceiveDialogOpen}
        onSuccess={handleActionSuccess}
      />

      <CancelTransferDialog
        transfer={transfer}
        open={isCancelDialogOpen}
        onOpenChange={setIsCancelDialogOpen}
        onSuccess={handleActionSuccess}
      />

      <DeleteTransferDialog
        transfer={transfer}
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onSuccess={handleDeleteSuccess}
      />
    </div>
  );
}
