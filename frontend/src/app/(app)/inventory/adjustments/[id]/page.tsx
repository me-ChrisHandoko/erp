/**
 * Adjustment Detail Page
 *
 * Full-page view of inventory adjustment details with actions:
 * - Adjustment information (number, date, warehouse, status, type, reason)
 * - Items list with quantity before/after
 * - Status-based actions (Edit, Approve, Cancel, Delete)
 * - Audit trail
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import {
  Package,
  AlertCircle,
  Warehouse,
  Calendar,
  Edit,
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
import { AdjustmentStatusBadge } from "@/components/adjustments/adjustment-status-badge";
import { AdjustmentTypeBadge } from "@/components/adjustments/adjustment-type-badge";
import { ApproveAdjustmentDialog } from "@/components/adjustments/approve-adjustment-dialog";
import { CancelAdjustmentDialog } from "@/components/adjustments/cancel-adjustment-dialog";
import { DeleteAdjustmentDialog } from "@/components/adjustments/delete-adjustment-dialog";
import { useGetAdjustmentQuery } from "@/store/services/adjustmentApi";
import { usePermissions } from "@/hooks/use-permissions";
import { ADJUSTMENT_REASON_CONFIG } from "@/types/adjustment.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { useState } from "react";
import { cn } from "@/lib/utils";

export default function AdjustmentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const adjustmentId = params.id as string;

  const { data: adjustment, isLoading, error, refetch } = useGetAdjustmentQuery(adjustmentId);
  const permissions = usePermissions();

  // Action dialogs state
  const [isApproveDialogOpen, setIsApproveDialogOpen] = useState(false);
  const [isCancelDialogOpen, setIsCancelDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  // Permission checks
  const canEdit = permissions.canEdit('inventory-adjustments');
  const canDelete = permissions.canDelete('inventory-adjustments');
  const canApprove = permissions.can('approve', 'inventory-adjustments');

  const handleActionSuccess = () => {
    refetch();
  };

  const handleDeleteSuccess = () => {
    router.push("/inventory/adjustments");
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
            { label: "Detail Penyesuaian" },
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
  if (error || !adjustment) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
            { label: "Detail Penyesuaian" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data penyesuaian" : "Penyesuaian tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/inventory/adjustments")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Penyesuaian
          </Button>
        </div>
      </div>
    );
  }

  const reasonConfig = ADJUSTMENT_REASON_CONFIG[adjustment.reason];

  // Format currency
  const formatCurrency = (value: string | number) => {
    const numValue = typeof value === "string" ? parseFloat(value) : value;
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(numValue);
  };

  // Format quantity
  const formatQuantity = (value: string | number) => {
    const numValue = typeof value === "string" ? parseFloat(value) : value;
    return new Intl.NumberFormat("id-ID", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 3,
    }).format(numValue);
  };

  // Determine available actions based on status
  // Only DRAFT adjustments can be edited, approved, cancelled, or deleted
  // APPROVED adjustments cannot be cancelled because stock has already changed
  const canEditAdjustment = canEdit && adjustment.status === "DRAFT";
  const canApproveAdjustment = canApprove && adjustment.status === "DRAFT";
  const canCancelAdjustment = canApprove && adjustment.status === "DRAFT";
  const canDeleteAdjustment = canDelete && adjustment.status === "DRAFT";

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
          { label: adjustment.adjustmentNumber },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header with actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold tracking-tight">
                {adjustment.adjustmentNumber}
              </h1>
              <AdjustmentStatusBadge status={adjustment.status} />
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Calendar className="h-3.5 w-3.5" />
              {format(new Date(adjustment.adjustmentDate), "dd MMMM yyyy", { locale: localeId })}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-2">
            {canEditAdjustment && (
              <Button
                variant="outline"
                onClick={() => router.push(`/inventory/adjustments/${adjustmentId}/edit`)}
              >
                <Edit className="mr-2 h-4 w-4" />
                Edit
              </Button>
            )}
            {canApproveAdjustment && (
              <Button onClick={() => setIsApproveDialogOpen(true)}>
                <CheckCircle className="mr-2 h-4 w-4" />
                Setujui
              </Button>
            )}
            {canCancelAdjustment && (
              <Button
                variant="destructive"
                onClick={() => setIsCancelDialogOpen(true)}
              >
                <XCircle className="mr-2 h-4 w-4" />
                Batalkan
              </Button>
            )}
            {canDeleteAdjustment && (
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

        {/* Adjustment Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="h-5 w-5" />
              Informasi Penyesuaian
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Type and Reason */}
            <div className="flex items-center gap-2">
              <AdjustmentTypeBadge type={adjustment.adjustmentType} />
              <Badge variant="outline">{reasonConfig?.label || adjustment.reason}</Badge>
            </div>

            {/* Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Tanggal</p>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <p className="font-medium">
                    {format(new Date(adjustment.adjustmentDate), "dd MMMM yyyy", {
                      locale: localeId,
                    })}
                  </p>
                </div>
              </div>
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Gudang</p>
                <div className="flex items-center gap-2">
                  <Warehouse className="h-4 w-4 text-muted-foreground" />
                  <p className="font-medium">{adjustment.warehouse?.name || "-"}</p>
                </div>
              </div>
            </div>

            {/* Notes */}
            {adjustment.notes && (
              <>
                <Separator />
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">Catatan</p>
                  <p className="text-sm bg-muted p-3 rounded-md">{adjustment.notes}</p>
                </div>
              </>
            )}

            {/* Audit Trail */}
            {adjustment.approvedBy && (
              <>
                <Separator />
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">Disetujui Oleh</p>
                  <p className="text-sm">{adjustment.approvedBy}</p>
                  {adjustment.approvedAt && (
                    <p className="text-xs text-muted-foreground">
                      {format(new Date(adjustment.approvedAt), "dd MMM yyyy, HH:mm", {
                        locale: localeId,
                      })}
                    </p>
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
                Daftar Produk ({adjustment.items?.length || 0})
              </CardTitle>
              <Badge variant="secondary">
                Total: {formatCurrency(adjustment.totalValue)}
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
                    <TableHead className="text-right">Sebelum</TableHead>
                    <TableHead className="text-right">Penyesuaian</TableHead>
                    <TableHead className="text-right">Sesudah</TableHead>
                    <TableHead className="text-right">Nilai</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {adjustment.items?.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-mono text-sm">
                        {item.product?.code || "-"}
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <p className="font-medium">{item.product?.name || "-"}</p>
                          {item.notes && (
                            <p className="text-xs text-muted-foreground italic">
                              {item.notes}
                            </p>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {formatQuantity(item.quantityBefore)}
                      </TableCell>
                      <TableCell className="text-right">
                        <span
                          className={cn(
                            "font-medium",
                            adjustment.adjustmentType === "INCREASE"
                              ? "text-green-600"
                              : "text-red-600"
                          )}
                        >
                          {adjustment.adjustmentType === "INCREASE" ? "+" : "-"}
                          {formatQuantity(Math.abs(parseFloat(item.quantityAdjusted)))}
                        </span>
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {formatQuantity(item.quantityAfter)}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {formatCurrency(item.totalValue)}
                      </TableCell>
                    </TableRow>
                  ))}
                  {/* Total Row */}
                  <TableRow className="bg-muted/50 font-semibold">
                    <TableCell colSpan={5} className="text-right">
                      Total Nilai
                    </TableCell>
                    <TableCell className="text-right">
                      {formatCurrency(adjustment.totalValue)}
                    </TableCell>
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
          onClick={() => router.push("/inventory/adjustments")}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Kembali ke Daftar Penyesuaian
        </Button>
      </div>

      {/* Action Dialogs */}
      <ApproveAdjustmentDialog
        adjustment={adjustment}
        open={isApproveDialogOpen}
        onOpenChange={setIsApproveDialogOpen}
        onSuccess={handleActionSuccess}
      />

      <CancelAdjustmentDialog
        adjustment={adjustment}
        open={isCancelDialogOpen}
        onOpenChange={setIsCancelDialogOpen}
        onSuccess={handleActionSuccess}
      />

      <DeleteAdjustmentDialog
        adjustment={adjustment}
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onSuccess={handleDeleteSuccess}
      />
    </div>
  );
}
