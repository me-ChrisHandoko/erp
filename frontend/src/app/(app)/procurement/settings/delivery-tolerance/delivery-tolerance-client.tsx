/**
 * Delivery Tolerance Settings Client Component
 *
 * Manages delivery tolerance settings (SAP Model):
 * - List all tolerance configurations
 * - Create/edit tolerance settings
 * - Delete tolerance settings
 * - View effective tolerance for products
 */

"use client";

import { useState } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Settings,
  Pencil,
  Trash2,
  Building2,
  FolderTree,
  Package,
  AlertCircle,
  CheckCircle,
  Info,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import {
  useListDeliveryTolerancesQuery,
  useDeleteDeliveryToleranceMutation,
} from "@/store/services/deliveryToleranceApi";
import { usePermissions } from "@/hooks/use-permissions";
import { toast } from "sonner";
import {
  getToleranceLevelLabel,
  formatTolerancePercentage,
  type DeliveryToleranceResponse,
  type DeliveryToleranceLevel,
} from "@/types/delivery-tolerance.types";
import type { RootState } from "@/store";
import { ToleranceFormDialog } from "@/components/procurement/tolerance-form-dialog";

export function DeliveryToleranceClient() {
  const permissions = usePermissions();
  // Use system-config for settings permissions (OWNER/ADMIN only)
  const canCreate = permissions.canCreate("system-config");
  const canEdit = permissions.canEdit("system-config");
  const canDelete = permissions.canDelete("system-config");

  // Get activeCompanyId from Redux
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedTolerance, setSelectedTolerance] =
    useState<DeliveryToleranceResponse | null>(null);

  // Fetch tolerances
  const {
    data: tolerancesData,
    isLoading,
    error,
    refetch,
  } = useListDeliveryTolerancesQuery(
    { pageSize: 100, sortBy: "level", sortOrder: "asc" },
    { skip: !activeCompanyId }
  );

  // Delete mutation
  const [deleteTolerance, { isLoading: isDeleting }] =
    useDeleteDeliveryToleranceMutation();

  const handleEdit = (tolerance: DeliveryToleranceResponse) => {
    setSelectedTolerance(tolerance);
    setEditDialogOpen(true);
  };

  const handleDelete = (tolerance: DeliveryToleranceResponse) => {
    setSelectedTolerance(tolerance);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (!selectedTolerance) return;

    try {
      await deleteTolerance(selectedTolerance.id).unwrap();
      toast.success("Toleransi Dihapus", {
        description: "Pengaturan toleransi berhasil dihapus",
      });
      setDeleteDialogOpen(false);
      setSelectedTolerance(null);
    } catch (error: any) {
      toast.error("Gagal Menghapus", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const getLevelIcon = (level: DeliveryToleranceLevel) => {
    switch (level) {
      case "COMPANY":
        return <Building2 className="h-4 w-4 text-blue-600" />;
      case "CATEGORY":
        return <FolderTree className="h-4 w-4 text-amber-600" />;
      case "PRODUCT":
        return <Package className="h-4 w-4 text-green-600" />;
      default:
        return null;
    }
  };

  const getLevelBadgeColor = (level: DeliveryToleranceLevel) => {
    switch (level) {
      case "COMPANY":
        return "bg-blue-100 text-blue-800";
      case "CATEGORY":
        return "bg-amber-100 text-amber-800";
      case "PRODUCT":
        return "bg-green-100 text-green-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Group tolerances by level
  const companyTolerances =
    tolerancesData?.data.filter((t) => t.level === "COMPANY") || [];
  const categoryTolerances =
    tolerancesData?.data.filter((t) => t.level === "CATEGORY") || [];
  const productTolerances =
    tolerancesData?.data.filter((t) => t.level === "PRODUCT") || [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <ErrorDisplay
        error="Tidak dapat memuat pengaturan toleransi pengiriman"
        title="Gagal Memuat Data"
        onRetry={refetch}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Header with Create Button */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            Toleransi Pengiriman
          </h2>
          <p className="text-muted-foreground">
            Kelola toleransi untuk pengiriman kurang atau lebih dari pesanan
          </p>
        </div>
        {canCreate && (
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Toleransi
          </Button>
        )}
      </div>

      {/* Info Alert - Tolerance Hierarchy */}
      <Alert className="border-blue-200 bg-blue-50/50">
        <Info className="h-4 w-4 text-blue-600" />
        <AlertTitle className="text-blue-900">Hierarki Toleransi (SAP Model)</AlertTitle>
        <AlertDescription className="text-blue-800">
          <p className="mb-3">
            Sistem akan mencari pengaturan toleransi dengan urutan prioritas berikut:
          </p>
          <div className="flex flex-wrap items-center gap-2 mb-3">
            <Badge variant="outline" className="bg-green-100 text-green-800 border-green-300">
              <Package className="h-3 w-3 mr-1" />
              1. Produk
            </Badge>
            <span className="text-blue-400">→</span>
            <Badge variant="outline" className="bg-amber-100 text-amber-800 border-amber-300">
              <FolderTree className="h-3 w-3 mr-1" />
              2. Kategori
            </Badge>
            <span className="text-blue-400">→</span>
            <Badge variant="outline" className="bg-blue-100 text-blue-800 border-blue-300">
              <Building2 className="h-3 w-3 mr-1" />
              3. Perusahaan
            </Badge>
            <span className="text-blue-400">→</span>
            <Badge variant="outline" className="bg-gray-100 text-gray-700 border-gray-300">
              4. Default (0%)
            </Badge>
          </div>
          <p className="text-sm text-blue-700">
            Pengaturan yang lebih spesifik (level lebih tinggi) akan digunakan terlebih dahulu.
            Jika tidak ditemukan, sistem akan menggunakan toleransi 0%.
          </p>
        </AlertDescription>
      </Alert>

      {/* Company Level Tolerance */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-600" />
            <div>
              <CardTitle className="text-lg">Toleransi Perusahaan</CardTitle>
              <CardDescription>
                Toleransi default untuk semua produk di perusahaan
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {companyTolerances.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Building2 className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>Belum ada toleransi default perusahaan</p>
              {canCreate && (
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-2"
                  onClick={() => setCreateDialogOpen(true)}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Tambah Toleransi Perusahaan
                </Button>
              )}
            </div>
          ) : (
            <ToleranceTable
              tolerances={companyTolerances}
              onEdit={handleEdit}
              onDelete={handleDelete}
              canEdit={canEdit}
              canDelete={canDelete}
              getLevelIcon={getLevelIcon}
              getLevelBadgeColor={getLevelBadgeColor}
            />
          )}
        </CardContent>
      </Card>

      {/* Category Level Tolerance */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <FolderTree className="h-5 w-5 text-amber-600" />
            <div>
              <CardTitle className="text-lg">Toleransi Kategori</CardTitle>
              <CardDescription>
                Toleransi untuk kategori produk tertentu
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {categoryTolerances.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <FolderTree className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>Belum ada toleransi untuk kategori</p>
            </div>
          ) : (
            <ToleranceTable
              tolerances={categoryTolerances}
              onEdit={handleEdit}
              onDelete={handleDelete}
              canEdit={canEdit}
              canDelete={canDelete}
              getLevelIcon={getLevelIcon}
              getLevelBadgeColor={getLevelBadgeColor}
              showCategory
            />
          )}
        </CardContent>
      </Card>

      {/* Product Level Tolerance */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <Package className="h-5 w-5 text-green-600" />
            <div>
              <CardTitle className="text-lg">Toleransi Produk</CardTitle>
              <CardDescription>
                Toleransi khusus untuk produk tertentu (prioritas tertinggi)
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {productTolerances.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Package className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>Belum ada toleransi untuk produk tertentu</p>
            </div>
          ) : (
            <ToleranceTable
              tolerances={productTolerances}
              onEdit={handleEdit}
              onDelete={handleDelete}
              canEdit={canEdit}
              canDelete={canDelete}
              getLevelIcon={getLevelIcon}
              getLevelBadgeColor={getLevelBadgeColor}
              showProduct
            />
          )}
        </CardContent>
      </Card>

      {/* Create Dialog */}
      <ToleranceFormDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        mode="create"
      />

      {/* Edit Dialog */}
      <ToleranceFormDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        mode="edit"
        tolerance={selectedTolerance}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Toleransi</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menghapus pengaturan toleransi ini?
              {selectedTolerance && (
                <div className="mt-3 p-3 bg-muted rounded-lg">
                  <div className="flex items-center gap-2 mb-2">
                    {getLevelIcon(selectedTolerance.level)}
                    <Badge
                      className={getLevelBadgeColor(selectedTolerance.level)}
                    >
                      {getToleranceLevelLabel(selectedTolerance.level)}
                    </Badge>
                  </div>
                  {selectedTolerance.categoryName && (
                    <p className="text-sm">
                      Kategori: {selectedTolerance.categoryName}
                    </p>
                  )}
                  {selectedTolerance.product && (
                    <p className="text-sm">
                      Produk: {selectedTolerance.product.name}
                    </p>
                  )}
                  <p className="text-sm mt-1">
                    Kurang:{" "}
                    {formatTolerancePercentage(
                      selectedTolerance.underDeliveryTolerance
                    )}{" "}
                    | Lebih:{" "}
                    {selectedTolerance.unlimitedOverDelivery
                      ? "Tidak Terbatas"
                      : formatTolerancePercentage(
                          selectedTolerance.overDeliveryTolerance
                        )}
                  </p>
                </div>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={isDeleting}
              className="bg-red-600 hover:bg-red-700"
            >
              {isDeleting ? "Menghapus..." : "Hapus"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

// Separate component for tolerance table
interface ToleranceTableProps {
  tolerances: DeliveryToleranceResponse[];
  onEdit: (t: DeliveryToleranceResponse) => void;
  onDelete: (t: DeliveryToleranceResponse) => void;
  canEdit: boolean;
  canDelete: boolean;
  getLevelIcon: (level: DeliveryToleranceLevel) => React.ReactNode;
  getLevelBadgeColor: (level: DeliveryToleranceLevel) => string;
  showCategory?: boolean;
  showProduct?: boolean;
}

function ToleranceTable({
  tolerances,
  onEdit,
  onDelete,
  canEdit,
  canDelete,
  getLevelIcon,
  getLevelBadgeColor,
  showCategory,
  showProduct,
}: ToleranceTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            {showCategory && <TableHead>Kategori</TableHead>}
            {showProduct && <TableHead>Produk</TableHead>}
            <TableHead className="text-right">Toleransi Kurang</TableHead>
            <TableHead className="text-right">Toleransi Lebih</TableHead>
            <TableHead className="text-center">Status</TableHead>
            <TableHead>Catatan</TableHead>
            {(canEdit || canDelete) && (
              <TableHead className="w-[100px]">Aksi</TableHead>
            )}
          </TableRow>
        </TableHeader>
        <TableBody>
          {tolerances.map((tolerance) => (
            <TableRow key={tolerance.id}>
              {showCategory && (
                <TableCell className="font-medium">
                  {tolerance.categoryName || "-"}
                </TableCell>
              )}
              {showProduct && (
                <TableCell>
                  {tolerance.product ? (
                    <div>
                      <div className="font-medium">
                        {tolerance.product.name}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {tolerance.product.code}
                      </div>
                    </div>
                  ) : (
                    "-"
                  )}
                </TableCell>
              )}
              <TableCell className="text-right font-mono">
                <Badge variant="outline" className="text-red-600 border-red-300">
                  -{formatTolerancePercentage(tolerance.underDeliveryTolerance)}
                </Badge>
              </TableCell>
              <TableCell className="text-right font-mono">
                {tolerance.unlimitedOverDelivery ? (
                  <Badge className="bg-purple-100 text-purple-800">
                    Tidak Terbatas
                  </Badge>
                ) : (
                  <Badge
                    variant="outline"
                    className="text-green-600 border-green-300"
                  >
                    +{formatTolerancePercentage(tolerance.overDeliveryTolerance)}
                  </Badge>
                )}
              </TableCell>
              <TableCell className="text-center">
                {tolerance.isActive ? (
                  <Badge className="bg-green-100 text-green-800">
                    <CheckCircle className="h-3 w-3 mr-1" />
                    Aktif
                  </Badge>
                ) : (
                  <Badge variant="secondary">
                    <AlertCircle className="h-3 w-3 mr-1" />
                    Nonaktif
                  </Badge>
                )}
              </TableCell>
              <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                {tolerance.notes || "-"}
              </TableCell>
              {(canEdit || canDelete) && (
                <TableCell>
                  <div className="flex gap-1">
                    {canEdit && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onEdit(tolerance)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                    )}
                    {canDelete && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onDelete(tolerance)}
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
