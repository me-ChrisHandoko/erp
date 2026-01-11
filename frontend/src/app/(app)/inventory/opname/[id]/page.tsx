/**
 * Stock Opname Detail Page - Server Component
 *
 * Displays comprehensive stock opname information including:
 * - Basic information (number, date, warehouse, status)
 * - Summary statistics (total items, quantities, differences)
 * - Item details with expected vs actual quantities
 * - Audit trail (created by, approved by)
 */

"use client";

import { use } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Edit, Trash2, CheckCircle, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useGetOpnameQuery, useDeleteOpnameMutation, useApproveOpnameMutation } from "@/store/services/opnameApi";
import { usePermissions } from "@/hooks/use-permissions";
import { useToast } from "@/hooks/use-toast";
import { OPNAME_STATUS_CONFIG } from "@/types/opname.types";
import { useState } from "react";

export default function OpnameDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const opnameId = params.id as string;
  const permissions = usePermissions();

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [approveDialogOpen, setApproveDialogOpen] = useState(false);

  const { data: opname, isLoading, error } = useGetOpnameQuery(opnameId);
  const [deleteOpname, { isLoading: isDeleting }] = useDeleteOpnameMutation();
  const [approveOpname, { isLoading: isApproving }] = useApproveOpnameMutation();

  const canEdit = permissions.canEdit("stock-opname");
  const canDelete = permissions.canDelete("stock-opname");
  const canApprove = permissions.canApprove("stock-opname");

  const handleDelete = async () => {
    try {
      await deleteOpname(opnameId).unwrap();
      toast({
        title: "Berhasil",
        description: "Stock opname berhasil dihapus",
      });
      router.push("/inventory/opname");
    } catch (error) {
      toast({
        title: "Gagal",
        description: "Gagal menghapus stock opname",
        variant: "destructive",
      });
    }
  };

  const handleApprove = async () => {
    try {
      await approveOpname({ id: opnameId }).unwrap();
      toast({
        title: "Berhasil",
        description: "Stock opname berhasil disetujui dan penyesuaian stok telah diterapkan",
      });
      setApproveDialogOpen(false);
    } catch (error) {
      toast({
        title: "Gagal",
        description: "Gagal menyetujui stock opname",
        variant: "destructive",
      });
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
    });
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatNumber = (value: string | number) => {
    const num = typeof value === "string" ? parseFloat(value) : value;
    return num.toLocaleString("id-ID", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    });
  };

  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Detail" },
          ]}
        />
        <div className="flex flex-1 items-center justify-center min-h-[400px]">
          <LoadingSpinner size="lg" text="Memuat data stock opname..." />
        </div>
      </div>
    );
  }

  if (error || !opname) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Detail" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <ErrorDisplay
            error={error}
            title="Gagal memuat data stock opname"
            onRetry={() => router.push("/inventory/opname")}
          />
        </div>
      </div>
    );
  }

  const statusConfig = OPNAME_STATUS_CONFIG[opname.status];

  return (
    <>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Detail" },
          ]}
        />

        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {/* Page Header */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-3">
                <FileText className="h-8 w-8 text-muted-foreground" />
                <div>
                  <h1 className="text-3xl font-bold tracking-tight">
                    {opname.opnameNumber}
                  </h1>
                  <p className="text-muted-foreground">
                    {formatDate(opname.opnameDate)} • {opname.warehouseName}
                  </p>
                </div>
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="outline"
                onClick={() => router.push("/inventory/opname")}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Kembali
              </Button>

              {canEdit && (opname.status === "draft" || opname.status === "in_progress") && (
                <Button
                  variant="outline"
                  onClick={() => router.push(`/inventory/opname/edit/${opname.id}`)}
                >
                  <Edit className="mr-2 h-4 w-4" />
                  Edit
                </Button>
              )}

              {canDelete && opname.status === "draft" && (
                <Button
                  variant="outline"
                  className="text-red-600 hover:text-red-700 border-red-200 hover:bg-red-50"
                  onClick={() => setDeleteDialogOpen(true)}
                  disabled={isDeleting}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Hapus
                </Button>
              )}

              {canApprove && opname.status === "completed" && (
                <Button
                  className="bg-green-600 hover:bg-green-700"
                  onClick={() => setApproveDialogOpen(true)}
                  disabled={isApproving}
                >
                  <CheckCircle className="mr-2 h-4 w-4" />
                  Approve
                </Button>
              )}
            </div>
          </div>

          {/* Summary Cards */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Badge variant={statusConfig.variant} className={statusConfig.className}>
                  {statusConfig.label}
                </Badge>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Total Item
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{opname.totalItems}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Expected Qty
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {formatNumber(opname.totalExpectedQty)}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Selisih
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div
                  className={`text-2xl font-bold ${
                    parseFloat(opname.totalDifference) > 0
                      ? "text-green-600"
                      : parseFloat(opname.totalDifference) < 0
                      ? "text-red-600"
                      : "text-muted-foreground"
                  }`}
                >
                  {parseFloat(opname.totalDifference) > 0 && "+"}
                  {formatNumber(opname.totalDifference)}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Items Table */}
          <Card>
            <CardHeader>
              <CardTitle>Detail Produk</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Kode</TableHead>
                      <TableHead>Nama Produk</TableHead>
                      <TableHead className="text-right">Expected Qty</TableHead>
                      <TableHead className="text-right">Actual Qty</TableHead>
                      <TableHead className="text-right">Selisih</TableHead>
                      <TableHead>Catatan</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {opname.items && opname.items.length > 0 ? (
                      opname.items.map((item) => {
                        const difference = parseFloat(item.difference);
                        return (
                          <TableRow key={item.id}>
                            <TableCell className="font-mono">
                              {item.productCode}
                            </TableCell>
                            <TableCell>{item.productName}</TableCell>
                            <TableCell className="text-right">
                              {formatNumber(item.expectedQty)}
                            </TableCell>
                            <TableCell className="text-right">
                              {formatNumber(item.actualQty)}
                            </TableCell>
                            <TableCell className="text-right">
                              <span
                                className={
                                  difference > 0
                                    ? "text-green-600 font-semibold"
                                    : difference < 0
                                    ? "text-red-600 font-semibold"
                                    : "text-muted-foreground"
                                }
                              >
                                {difference > 0 && "+"}
                                {formatNumber(item.difference)}
                              </span>
                            </TableCell>
                            <TableCell className="text-muted-foreground text-sm">
                              {item.notes || "-"}
                            </TableCell>
                          </TableRow>
                        );
                      })
                    ) : (
                      <TableRow>
                        <TableCell
                          colSpan={6}
                          className="text-center text-muted-foreground"
                        >
                          Tidak ada item
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>

          {/* Audit Trail */}
          <Card>
            <CardHeader>
              <CardTitle>Audit Trail</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid gap-3 md:grid-cols-2">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Dibuat oleh
                  </p>
                  <p className="text-sm">
                    {opname.createdByName || opname.createdBy}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatDateTime(opname.createdAt)}
                  </p>
                </div>
                {opname.approvedBy && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">
                      Disetujui oleh
                    </p>
                    <p className="text-sm">
                      {opname.approvedByName || opname.approvedBy}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {opname.approvedAt && formatDateTime(opname.approvedAt)}
                    </p>
                  </div>
                )}
              </div>
              {opname.notes && (
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Catatan
                  </p>
                  <p className="text-sm whitespace-pre-wrap">{opname.notes}</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Stock Opname?</AlertDialogTitle>
            <AlertDialogDescription>
              Tindakan ini tidak dapat dibatalkan. Stock opname{" "}
              <strong>{opname.opnameNumber}</strong> akan dihapus secara
              permanen dari sistem.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Hapus
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Approve Confirmation Dialog */}
      <AlertDialog open={approveDialogOpen} onOpenChange={setApproveDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Approve Stock Opname?</AlertDialogTitle>
            <AlertDialogDescription>
              Dengan menyetujui stock opname <strong>{opname.opnameNumber}</strong>,
              penyesuaian stok akan diterapkan ke sistem. Tindakan ini tidak
              dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleApprove}
              className="bg-green-600 text-white hover:bg-green-700"
            >
              Approve
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
