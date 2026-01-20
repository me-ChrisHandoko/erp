/**
 * Create Stock Opname Client Component
 *
 * Form for creating new stock opname with warehouse selection
 * and product import functionality.
 */

"use client";

import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import { ArrowLeft, Save, Download, Plus, Trash2, Package, AlertTriangle } from "lucide-react";
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
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { useToast } from "@/hooks/use-toast";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { useListStocksQuery } from "@/store/services/stockApi";
import { useListProductsQuery } from "@/store/services/productApi";
import { useCreateOpnameMutation } from "@/store/services/opnameApi";
import type { RootState } from "@/store";
import type { CreateStockOpnameItemRequest } from "@/types/opname.types";

interface OpnameItemForm {
  productId: string;
  productCode: string;
  productName: string;
  expectedQty: string;
  actualQty: string;
  notes?: string;
}

export function CreateOpnameClient() {
  const router = useRouter();
  const { toast } = useToast();
  const [warehouseId, setWarehouseId] = useState<string>("");
  const [opnameDate, setOpnameDate] = useState<string>(
    new Date().toISOString().split("T")[0]
  );
  const [notes, setNotes] = useState<string>("");
  const [items, setItems] = useState<OpnameItemForm[]>([]);
  const [isImporting, setIsImporting] = useState(false);
  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [showImportConfirm, setShowImportConfirm] = useState(false);

  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Fetch warehouses
  const { data: warehousesData, isLoading: isLoadingWarehouses } =
    useListWarehousesQuery(
      {
        isActive: true,
        pageSize: 100,
        sortBy: "name",
        sortOrder: "asc",
      },
      {
        skip: !activeCompanyId,
      }
    );

  // Fetch stocks for selected warehouse
  const { data: stocksData, isLoading: isLoadingStocks } = useListStocksQuery(
    {
      warehouseID: warehouseId,
      pageSize: 1000, // Get all products
      sortBy: "productCode",
      sortOrder: "asc",
    },
    {
      skip: !warehouseId,
    }
  );

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
      skip: !activeCompanyId || !warehouseId,
    }
  );

  const [createOpname, { isLoading: isCreating }] = useCreateOpnameMutation();

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

  // Handle adding product manually
  const handleAddProduct = () => {
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

    const newItem: OpnameItemForm = {
      productId: product.id,
      productCode: product.code,
      productName: product.name,
      expectedQty: expectedQty,
      actualQty: "",
      notes: "",
    };

    setItems([...items, newItem]);
    setSelectedProductId("");

    toast({
      title: "Berhasil",
      description: `${product.name} ditambahkan ke daftar opname`,
    });
  };

  const handleImportProducts = () => {
    if (!stocksData?.data) {
      toast({
        title: "Gagal",
        description: "Tidak ada data stok untuk diimport",
        variant: "destructive",
      });
      return;
    }

    // Show confirmation if there are existing items
    if (items.length > 0) {
      setShowImportConfirm(true);
      return;
    }

    doImportProducts();
  };

  const doImportProducts = () => {
    if (!stocksData?.data) return;

    setIsImporting(true);
    const importedItems: OpnameItemForm[] = stocksData.data.map((stock) => ({
      productId: stock.productID,
      productCode: stock.productCode || "",
      productName: stock.productName || "",
      expectedQty: stock.quantity || "0",
      actualQty: "", // User will fill this
      notes: "",
    }));

    setItems(importedItems);
    setIsImporting(false);
    setShowImportConfirm(false);

    toast({
      title: "Berhasil",
      description: `${importedItems.length} produk berhasil diimport`,
    });
  };

  const handleActualQtyChange = (index: number, value: string) => {
    const newItems = [...items];
    newItems[index].actualQty = value;
    setItems(newItems);
  };

  const handleNotesChange = (index: number, value: string) => {
    const newItems = [...items];
    newItems[index].notes = value;
    setItems(newItems);
  };

  const handleRemoveItem = (index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    setItems(newItems);
  };

  const handleSubmit = async () => {
    // Validation
    if (!warehouseId) {
      toast({
        title: "Validasi Gagal",
        description: "Pilih gudang terlebih dahulu",
        variant: "destructive",
      });
      return;
    }

    if (!opnameDate) {
      toast({
        title: "Validasi Gagal",
        description: "Tanggal opname wajib diisi",
        variant: "destructive",
      });
      return;
    }

    if (items.length === 0) {
      toast({
        title: "Validasi Gagal",
        description: "Minimal harus ada 1 produk untuk stock opname",
        variant: "destructive",
      });
      return;
    }

    // Check if all items have actual qty
    const itemsWithoutActualQty = items.filter(
      (item) => !item.actualQty || item.actualQty === ""
    );
    if (itemsWithoutActualQty.length > 0) {
      toast({
        title: "Validasi Gagal",
        description: "Semua produk harus memiliki actual quantity",
        variant: "destructive",
      });
      return;
    }

    try {
      const opnameItems: CreateStockOpnameItemRequest[] = items.map((item) => ({
        productId: item.productId,
        expectedQty: item.expectedQty,
        actualQty: item.actualQty,
        notes: item.notes,
      }));

      await createOpname({
        warehouseId,
        opnameDate,
        notes,
        items: opnameItems,
      }).unwrap();

      toast({
        title: "Berhasil",
        description: "Stock opname berhasil dibuat",
      });

      router.push("/inventory/opname");
    } catch (error) {
      console.error("Failed to create opname:", error);
      toast({
        title: "Gagal",
        description: "Gagal membuat stock opname",
        variant: "destructive",
      });
    }
  };

  const calculateDifference = (expected: string, actual: string): number => {
    const exp = parseFloat(expected) || 0;
    const act = parseFloat(actual) || 0;
    return act - exp;
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">
            Buat Stock Opname
          </h1>
          <p className="text-muted-foreground">
            Buat penghitungan fisik inventory untuk gudang
          </p>
        </div>
        <Button
          variant="outline"
          onClick={() => router.push("/inventory/opname")}
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
          {/* Warehouse Selection */}
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="warehouse">
                Gudang <span className="text-destructive">*</span>
              </Label>
              <Select value={warehouseId} onValueChange={setWarehouseId}>
                <SelectTrigger id="warehouse" className="bg-background w-full">
                  <SelectValue placeholder="Pilih gudang" />
                </SelectTrigger>
                <SelectContent>
                  {isLoadingWarehouses ? (
                    <SelectItem value="loading" disabled>
                      Memuat gudang...
                    </SelectItem>
                  ) : (
                    warehousesData?.data?.map((warehouse) => (
                      <SelectItem key={warehouse.id} value={warehouse.id}>
                        {warehouse.name}
                      </SelectItem>
                    ))
                  )}
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

          {/* Notes */}
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

          {/* Import Button */}
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleImportProducts}
              disabled={!warehouseId || isLoadingStocks || isImporting}
            >
              <Download className="mr-2 h-4 w-4" />
              Import Produk dari Gudang
            </Button>
            {isLoadingStocks && (
              <span className="text-sm text-muted-foreground">
                Memuat produk...
              </span>
            )}
          </div>

          {/* Manual Product Addition */}
          <div className="space-y-2">
            <Label>Tambah Produk Manual</Label>
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
                  disabled={!warehouseId || isLoadingProducts}
                  renderOption={(option, selected) => (
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
                disabled={!selectedProductId || !warehouseId}
              >
                <Plus className="mr-2 h-4 w-4" />
                Tambah
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Pilih produk satu per satu atau gunakan tombol &quot;Import Produk dari Gudang&quot; untuk mengimport semua produk sekaligus
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Items Table */}
      {items.length > 0 && (
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
                    const difference = calculateDifference(
                      item.expectedQty,
                      item.actualQty
                    );
                    return (
                      <TableRow key={index}>
                        <TableCell className="font-mono text-sm">
                          {item.productCode}
                        </TableCell>
                        <TableCell>{item.productName}</TableCell>
                        <TableCell className="text-right">
                          {parseFloat(item.expectedQty).toLocaleString("id-ID")}
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
                            placeholder="0"
                          />
                        </TableCell>
                        <TableCell className="text-right">
                          {item.actualQty && (
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
                              {difference.toLocaleString("id-ID")}
                            </span>
                          )}
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
                            onClick={() => handleRemoveItem(index)}
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
      )}

      {/* Action Buttons */}
      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={() => router.push("/inventory/opname")}
          disabled={isCreating}
        >
          Batal
        </Button>
        <Button onClick={handleSubmit} disabled={isCreating || items.length === 0}>
          {isCreating ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Menyimpan...
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Simpan Stock Opname
            </>
          )}
        </Button>
      </div>

      {/* Import Confirmation Dialog */}
      <AlertDialog open={showImportConfirm} onOpenChange={setShowImportConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
              Konfirmasi Import
            </AlertDialogTitle>
            <AlertDialogDescription>
              Anda sudah memiliki <strong>{items.length} produk</strong> dalam daftar.
              Import dari gudang akan <strong>menimpa semua data</strong> yang sudah ada.
              <br /><br />
              Apakah Anda yakin ingin melanjutkan?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={doImportProducts}
              className="bg-amber-600 hover:bg-amber-700"
            >
              Ya, Timpa Data
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
