/**
 * Edit Stock Opname Page - Client Component
 *
 * Page for editing existing stock opname.
 * Only draft and in_progress status can be edited.
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState, useMemo } from "react";
import { ArrowLeft, Save, Trash2, Plus, Package, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useToast } from "@/hooks/use-toast";
import {
  useGetOpnameQuery,
  useUpdateOpnameMutation,
  useBatchUpdateOpnameItemsMutation,
  useAddOpnameItemMutation,
  useDeleteOpnameItemMutation,
} from "@/store/services/opnameApi";
import { useListProductsQuery } from "@/store/services/productApi";
import { useListStocksQuery } from "@/store/services/stockApi";
import { useSelector } from "react-redux";
import type { RootState } from "@/store";
import type { StockOpnameItem } from "@/types/opname.types";

interface OpnameItemForm extends StockOpnameItem {
  isDirty?: boolean;
}

export default function EditOpnamePage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const opnameId = params.id as string;

  const [opnameDate, setOpnameDate] = useState<string>("");
  const [notes, setNotes] = useState<string>("");
  const [status, setStatus] = useState<"draft" | "in_progress" | "completed">("draft");
  const [items, setItems] = useState<OpnameItemForm[]>([]);
  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<OpnameItemForm | null>(null);

  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  const { data: opname, isLoading, error, refetch } = useGetOpnameQuery(opnameId);
  const [updateOpname, { isLoading: isUpdatingOpname }] = useUpdateOpnameMutation();
  const [batchUpdateItems, { isLoading: isUpdatingItems }] = useBatchUpdateOpnameItemsMutation();
  const [addItem, { isLoading: isAddingItem }] = useAddOpnameItemMutation();
  const [deleteItem, { isLoading: isDeletingItem }] = useDeleteOpnameItemMutation();

  // Fetch products for manual addition
  const { data: productsData, isLoading: isLoadingProducts } = useListProductsQuery(
    {
      page: 1,
      pageSize: 100,
      isActive: true,
      sortBy: "code",
      sortOrder: "asc",
    },
    {
      skip: !activeCompanyId || !opname?.warehouseId,
    }
  );

  // Fetch stocks for the warehouse to get expected qty
  const { data: stocksData } = useListStocksQuery(
    {
      warehouseID: opname?.warehouseId || "",
      pageSize: 1000,
      sortBy: "productCode",
      sortOrder: "asc",
    },
    {
      skip: !opname?.warehouseId,
    }
  );

  // Build product options for combobox (exclude already added products)
  const productOptions: ComboboxOption[] = useMemo(() => {
    if (!productsData?.data) return [];

    const addedProductIds = new Set(items.map((item) => item.productId));

    return productsData.data
      .filter((product) => !addedProductIds.has(product.id))
      .map((product) => ({
        value: product.id,
        label: `${product.code} - ${product.name}`,
        searchLabel: `${product.code} ${product.name}`,
        code: product.code,
      }));
  }, [productsData?.data, items]);

  // Populate form when data loads
  useEffect(() => {
    if (opname) {
      // Check if opname can be edited
      if (opname.status !== "draft" && opname.status !== "in_progress") {
        toast({
          title: "Tidak Dapat Diedit",
          description: "Stock opname dengan status ini tidak dapat diedit",
          variant: "destructive",
        });
        router.push(`/inventory/opname/${opnameId}`);
        return;
      }

      setOpnameDate(opname.opnameDate.split("T")[0]);
      setNotes(opname.notes || "");
      setStatus(opname.status as "draft" | "in_progress" | "completed");
      // Deep copy items to avoid mutating RTK Query cached data
      setItems((opname.items || []).map(item => ({ ...item })));
    }
  }, [opname, opnameId, router, toast]);

  const handleActualQtyChange = (index: number, value: string) => {
    const newItems = items.map((item, i) => {
      if (i === index) {
        return {
          ...item,
          actualQty: value,
          difference: (
            parseFloat(value || "0") - parseFloat(item.expectedQty || "0")
          ).toString(),
          isDirty: true,
        };
      }
      return item;
    });
    setItems(newItems);
  };

  const handleNotesChange = (index: number, value: string) => {
    const newItems = items.map((item, i) => {
      if (i === index) {
        return {
          ...item,
          notes: value || undefined,
          isDirty: true,
        };
      }
      return item;
    });
    setItems(newItems);
  };

  // Handle adding product
  const handleAddProduct = async () => {
    if (!selectedProductId) {
      toast({
        title: "Pilih Produk",
        description: "Pilih produk terlebih dahulu",
        variant: "destructive",
      });
      return;
    }

    const product = productsData?.data?.find((p) => p.id === selectedProductId);
    if (!product) return;

    // Get expected qty from stock data if available
    const stockItem = stocksData?.data?.find((s) => s.productID === selectedProductId);
    const expectedQty = stockItem?.quantity || "0";

    try {
      await addItem({
        opnameId,
        productId: product.id,
        expectedQty: expectedQty,
        actualQty: "0",
        notes: "",
      }).unwrap();

      // Refetch to get updated items
      await refetch();
      setSelectedProductId("");

      toast({
        title: "Berhasil",
        description: `${product.name} ditambahkan ke daftar opname`,
      });
    } catch (error) {
      console.error("Failed to add item:", error);
      toast({
        title: "Gagal",
        description: "Gagal menambahkan produk",
        variant: "destructive",
      });
    }
  };

  // Handle delete item confirmation
  const handleDeleteClick = (item: OpnameItemForm) => {
    setItemToDelete(item);
    setShowDeleteConfirm(true);
  };

  // Confirm delete item
  const handleConfirmDelete = async () => {
    if (!itemToDelete) return;

    try {
      await deleteItem({
        opnameId,
        itemId: itemToDelete.id,
      }).unwrap();

      // Remove from local state
      setItems(items.filter((item) => item.id !== itemToDelete.id));
      setShowDeleteConfirm(false);
      setItemToDelete(null);

      toast({
        title: "Berhasil",
        description: `${itemToDelete.productName} dihapus dari daftar opname`,
      });
    } catch (error) {
      console.error("Failed to delete item:", error);
      toast({
        title: "Gagal",
        description: "Gagal menghapus produk",
        variant: "destructive",
      });
    }
  };

  const handleSubmit = async () => {
    try {
      // Update main opname info
      await updateOpname({
        id: opnameId,
        data: {
          opnameDate,
          notes,
          status,
        },
      }).unwrap();

      // Update items that have changed using batch update (creates single audit log)
      const dirtyItems = items.filter((item) => item.isDirty);
      if (dirtyItems.length > 0) {
        await batchUpdateItems({
          opnameId,
          data: {
            items: dirtyItems.map((item) => ({
              itemId: item.id,
              actualQty: item.actualQty,
              notes: item.notes || undefined,
            })),
          },
        }).unwrap();
      }

      toast({
        title: "Berhasil",
        description: "Stock opname berhasil diupdate",
      });

      router.push(`/inventory/opname/${opnameId}`);
    } catch (error) {
      console.error("Failed to update opname:", error);
      toast({
        title: "Gagal",
        description: "Gagal mengupdate stock opname",
        variant: "destructive",
      });
    }
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
            { label: "Edit" },
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
            { label: "Edit" },
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

  const isProcessing = isUpdatingOpname || isUpdatingItems || isAddingItem || isDeletingItem;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Stock Opname", href: "/inventory/opname" },
          { label: "Edit" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Stock Opname
            </h1>
            <p className="text-muted-foreground">
              {opname.opnameNumber} • {opname.warehouseName}
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => router.push(`/inventory/opname/${opnameId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>
        </div>

        {/* Form Card */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Stock Opname</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="status">Status</Label>
                <Select
                  value={status}
                  onValueChange={(value: "draft" | "in_progress" | "completed") =>
                    setStatus(value)
                  }
                >
                  <SelectTrigger id="status" className="bg-background w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="opnameDate">
                  Tanggal Opname <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="opnameDate"
                  type="date"
                  value={opnameDate}
                  onChange={(e) => setOpnameDate(e.target.value)}
                  max={new Date().toISOString().split("T")[0]}
                  className="bg-background"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="notes">Catatan</Label>
              <Textarea
                id="notes"
                placeholder="Catatan untuk stock opname ini..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                className="bg-background"
              />
            </div>

            {/* Manual Product Addition */}
            <div className="space-y-2">
              <Label>Tambah Produk</Label>
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <Combobox
                    options={productOptions}
                    value={selectedProductId}
                    onValueChange={setSelectedProductId}
                    placeholder="Cari dan pilih produk..."
                    searchPlaceholder="Ketik kode atau nama produk..."
                    emptyMessage={
                      isLoadingProducts
                        ? "Memuat produk..."
                        : "Produk tidak ditemukan"
                    }
                    disabled={isLoadingProducts || isAddingItem}
                    renderOption={(option) => (
                      <div className="flex items-center gap-2 w-full">
                        <Package className="h-4 w-4 text-muted-foreground shrink-0" />
                        <div className="flex flex-col min-w-0 flex-1">
                          <span className="font-medium truncate">{option.code}</span>
                          <span className="text-xs text-muted-foreground truncate">
                            {option.label.split(" - ")[1]}
                          </span>
                        </div>
                      </div>
                    )}
                  />
                </div>
                <Button
                  type="button"
                  onClick={handleAddProduct}
                  disabled={!selectedProductId || isAddingItem}
                >
                  {isAddingItem ? (
                    <LoadingSpinner size="sm" className="mr-2" />
                  ) : (
                    <Plus className="mr-2 h-4 w-4" />
                  )}
                  Tambah
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Tambahkan produk yang belum ada dalam daftar opname
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Items Table */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Produk ({items.length})</CardTitle>
              <span className="text-sm text-muted-foreground">
                Isi actual quantity untuk setiap produk
              </span>
            </div>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[100px]">Kode</TableHead>
                    <TableHead>Nama Produk</TableHead>
                    <TableHead className="text-right">Expected Qty</TableHead>
                    <TableHead className="text-right w-[150px]">
                      Actual Qty <span className="text-destructive">*</span>
                    </TableHead>
                    <TableHead className="text-right">Selisih</TableHead>
                    <TableHead className="w-[200px]">Catatan</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item, index) => {
                    const difference = parseFloat(item.difference || "0");
                    return (
                      <TableRow key={item.id}>
                        <TableCell className="font-mono text-sm">
                          {item.productCode}
                        </TableCell>
                        <TableCell>{item.productName}</TableCell>
                        <TableCell className="text-right">
                          {formatNumber(item.expectedQty)}
                        </TableCell>
                        <TableCell className="text-right">
                          <Input
                            type="number"
                            step="0.01"
                            min="0"
                            value={item.actualQty}
                            onChange={(e) =>
                              handleActualQtyChange(index, e.target.value)
                            }
                            className="text-right bg-background"
                          />
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
                            {formatNumber(difference)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Input
                            type="text"
                            value={item.notes || ""}
                            onChange={(e) =>
                              handleNotesChange(index, e.target.value)
                            }
                            placeholder="Catatan..."
                            className="bg-background"
                          />
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDeleteClick(item)}
                            disabled={isDeletingItem}
                            className="text-red-600 hover:text-red-700"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-2">
          <Button
            variant="outline"
            onClick={() => router.push(`/inventory/opname/${opnameId}`)}
            disabled={isProcessing}
          >
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isProcessing}>
            {isProcessing ? (
              <>
                <LoadingSpinner size="sm" className="mr-2" />
                Menyimpan...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Simpan Perubahan
              </>
            )}
          </Button>
        </div>

        {/* Delete Confirmation Dialog */}
        <AlertDialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle className="flex items-center gap-2">
                <AlertTriangle className="h-5 w-5 text-red-500" />
                Konfirmasi Hapus Produk
              </AlertDialogTitle>
              <AlertDialogDescription>
                Apakah Anda yakin ingin menghapus{" "}
                <strong>{itemToDelete?.productName}</strong> dari daftar opname?
                <br /><br />
                Tindakan ini tidak dapat dibatalkan.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setItemToDelete(null)}>
                Batal
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirmDelete}
                className="bg-red-600 hover:bg-red-700"
              >
                {isDeletingItem ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Menghapus...
                  </>
                ) : (
                  "Ya, Hapus"
                )}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  );
}
