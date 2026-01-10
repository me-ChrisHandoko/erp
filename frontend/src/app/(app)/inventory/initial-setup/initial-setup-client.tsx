/**
 * Initial Setup Client Component
 *
 * Client-side interactive component for initial stock setup wizard.
 * Phase 2: Full wizard implementation with 5 steps.
 *
 * Steps:
 * 1. Warehouse Selection - Choose target warehouse
 * 2. Input Method - Manual entry or Excel import
 * 3. Data Entry - Manual form or Excel upload
 * 4. Review - Validate and review all entries
 * 5. Success - Confirmation and next actions
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Package,
  Building2,
  FileSpreadsheet,
  Edit3,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Upload,
  Download,
  Plus,
  Trash2,
  AlertCircle,
  DollarSign,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
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
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { useListProductsQuery } from "@/store/services/productApi";
import {
  useSubmitInitialStockMutation,
  useGetWarehouseStockStatusQuery,
} from "@/store/services/initialStockApi";
import { useListStocksQuery } from "@/store/services/stockApi";
import type {
  InitialStockItem,
  ExcelValidationResult,
} from "@/types/initial-stock.types";
import {
  parseExcelFile,
  validateExcelData,
  generateExcelTemplate,
} from "@/lib/excel-validator";

interface InitialSetupClientProps {
  initialWarehouseId?: string;
  context?: string;
  source?: string;
}

type WizardStep = 1 | 2 | 3 | 4 | 5;
type InputMethod = "manual" | "excel";

interface StockItemRow extends InitialStockItem {
  tempId: string;
  productCode?: string;
  productName?: string;
  baseUnit?: string;
}

export function InitialSetupClient({
  initialWarehouseId,
  context,
  source,
}: InitialSetupClientProps) {
  const router = useRouter();

  // Wizard state
  const [currentStep, setCurrentStep] = useState<WizardStep>(1);
  const [selectedWarehouseId, setSelectedWarehouseId] = useState<string>(
    initialWarehouseId || ""
  );
  const [inputMethod, setInputMethod] = useState<InputMethod>("manual");
  const [stockItems, setStockItems] = useState<StockItemRow[]>([]);
  const [notes, setNotes] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Excel upload state
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [isUploadingFile, setIsUploadingFile] = useState(false);
  const [validationResult, setValidationResult] = useState<ExcelValidationResult | null>(null);

  // Fetch data
  const { data: warehousesData, isLoading: isLoadingWarehouses } =
    useListWarehousesQuery({ isActive: true });
  const { data: productsData, isLoading: isLoadingProducts } =
    useListProductsQuery({ isActive: true });
  const { data: statusData } = useGetWarehouseStockStatusQuery();

  // Get existing stocks for selected warehouse (for validation)
  // Use large pageSize to get all stocks for validation (not paginated)
  const {
    data: existingStocksData,
    isLoading: isLoadingExistingStocks,
    isFetching: isFetchingExistingStocks,
    refetch: refetchExistingStocks,
    error: existingStocksError,
  } = useListStocksQuery(
    {
      warehouseID: selectedWarehouseId,
      pageSize: 1000, // Increased to handle warehouses with many products
    },
    {
      skip: !selectedWarehouseId,
      refetchOnMountOrArgChange: true, // Always refetch when component mounts or args change
    }
  );

  // Mutations
  const [submitInitialStock, { isLoading: isSubmitting }] =
    useSubmitInitialStockMutation();

  // Force refetch existing stocks when entering step 3
  useEffect(() => {
    if (currentStep === 3 && selectedWarehouseId) {
      console.log("🔄 Step 3 entered, force refetching stocks for warehouse:", selectedWarehouseId);
      refetchExistingStocks();
    }
  }, [currentStep, selectedWarehouseId, refetchExistingStocks]);

  // Find selected warehouse
  const selectedWarehouse = useMemo(
    () => warehousesData?.data?.find((w) => w.id === selectedWarehouseId),
    [warehousesData, selectedWarehouseId]
  );

  // Check if warehouse already has stock
  const warehouseHasStock = useMemo(
    () => statusData?.find((s) => s.warehouseId === selectedWarehouseId)?.hasInitialStock,
    [statusData, selectedWarehouseId]
  );

  // Check if current step has validation conflicts (for button disable state)
  const hasValidationConflicts = useMemo(() => {
    if (currentStep !== 3 || inputMethod !== "excel" || !validationResult) {
      return false;
    }

    // Check for existing stocks conflicts or duplicates in file
    const hasConflicts = validationResult.existingStocks && validationResult.existingStocks.length > 0;
    const hasDuplicates = validationResult.duplicatesInFile && validationResult.duplicatesInFile.length > 0;

    return hasConflicts || hasDuplicates;
  }, [currentStep, inputMethod, validationResult]);

  // Filter products to show only those without existing stock in selected warehouse
  const availableProducts = useMemo(() => {
    if (!productsData?.data || !selectedWarehouseId) {
      return [];
    }

    // If no existing stocks data yet, show all products
    if (!existingStocksData?.data) {
      console.log("⚠️ No existing stocks data, showing all products");
      return productsData.data;
    }

    // Create a Set of product IDs that already have stock in this warehouse
    const productsWithStock = new Set(
      existingStocksData.data.map((stock) => stock.productID)
    );

    console.log("📊 Available Products Filter Debug:");
    console.log("  - Total products in system:", productsData.data.length);
    console.log("  - Existing stocks loaded:", existingStocksData.data.length);
    console.log("  - Products with stock IDs:", Array.from(productsWithStock));
    console.log("  - Selected warehouse ID:", selectedWarehouseId);

    // Filter out products that already have stock
    const filtered = productsData.data.filter((product) => !productsWithStock.has(product.id));
    console.log("  - Available products after filter:", filtered.length);

    return filtered;
  }, [productsData, existingStocksData, selectedWarehouseId]);

  // Helper: Add new empty row
  const handleAddRow = () => {
    setStockItems([
      ...stockItems,
      {
        tempId: `temp-${Date.now()}`,
        productId: "",
        quantity: "",
        costPerUnit: "",
        location: "",
        minimumStock: "",
        maximumStock: "",
        notes: "",
      },
    ]);
  };

  // Helper: Remove row
  const handleRemoveRow = (tempId: string) => {
    setStockItems(stockItems.filter((item) => item.tempId !== tempId));
  };

  // Helper: Update row field
  const handleUpdateRow = (tempId: string, field: keyof StockItemRow, value: string) => {
    setStockItems(
      stockItems.map((item) => {
        if (item.tempId === tempId) {
          // If productId changes, update related fields
          if (field === "productId") {
            const product = productsData?.data?.find((p) => p.id === value);
            return {
              ...item,
              productId: value,
              productCode: product?.code,
              productName: product?.name,
              baseUnit: product?.baseUnit,
              costPerUnit: product?.baseCost || "",
            };
          }
          return { ...item, [field]: value };
        }
        return item;
      })
    );
  };

  // Excel Upload Handlers
  const handleDownloadTemplate = async () => {
    try {
      const blob = await generateExcelTemplate();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `Template_Stok_Awal_${new Date().toISOString().split("T")[0]}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Error generating template:", error);
      setErrors({ template: "Gagal membuat template Excel" });
    }
  };

  const handleFileUpload = async (file: File) => {
    // Reset state
    setUploadedFile(file);
    setValidationResult(null);
    setErrors({});
    setIsUploadingFile(true);

    try {
      // Check if data is still loading
      if (isLoadingProducts || isLoadingExistingStocks || isFetchingExistingStocks) {
        setErrors({
          upload: "Sedang memuat data produk dan stok. Mohon tunggu sebentar...",
        });
        setIsUploadingFile(false);
        return;
      }

      // 1. Parse Excel file
      const excelRows = await parseExcelFile(file);

      // 2. Validate data with existing stocks
      console.log("🔍 Debug Query State:");
      console.log("  - selectedWarehouseId:", selectedWarehouseId);
      console.log("  - existingStocksData:", existingStocksData);
      console.log("  - existingStocksError:", existingStocksError);
      console.log("  - isLoading:", isLoadingExistingStocks);
      console.log("  - isFetching:", isFetchingExistingStocks);
      console.log("  - Validating with existing stocks:", existingStocksData?.data?.length || 0, "items");

      if (existingStocksError) {
        console.error("❌ Error fetching existing stocks:", existingStocksError);
      }

      const result = validateExcelData(
        excelRows,
        productsData?.data || [],
        existingStocksData?.data || []
      );

      setValidationResult(result);

      // 3. If validation successful, convert to stock items
      if (result.success) {
        const newItems: StockItemRow[] = result.validItems.map((item, index) => {
          const product = productsData?.data?.find((p) => p.id === item.productId);
          return {
            tempId: `excel-${Date.now()}-${index}`,
            productId: item.productId,
            productCode: product?.code,
            productName: product?.name,
            baseUnit: product?.baseUnit,
            quantity: item.quantity,
            costPerUnit: item.costPerUnit,
            location: item.location,
            minimumStock: item.minimumStock,
            maximumStock: item.maximumStock,
            notes: item.notes,
          };
        });
        setStockItems(newItems);
      } else {
        // Show errors
        const errorMessages: Record<string, string> = {};
        if (result.duplicatesInFile.length > 0) {
          errorMessages.duplicates = `${result.duplicatesInFile.length} produk duplikat dalam file`;
        }
        if (result.existingStocks.length > 0) {
          errorMessages.existing = `${result.existingStocks.length} produk sudah memiliki stok`;
        }
        if (result.errors.length > 0) {
          errorMessages.validation = `${result.errors.length} error validasi`;
        }
        setErrors(errorMessages);
      }
    } catch (error) {
      console.error("Error parsing Excel:", error);
      setErrors({
        upload: error instanceof Error ? error.message : "Gagal membaca file Excel",
      });
      setValidationResult(null);
    } finally {
      setIsUploadingFile(false);
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    const validTypes = [
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      "application/vnd.ms-excel",
    ];
    if (!validTypes.includes(file.type)) {
      setErrors({ upload: "Format file harus .xlsx atau .xls" });
      return;
    }

    // Validate file size (max 5MB)
    if (file.size > 5 * 1024 * 1024) {
      setErrors({ upload: "Ukuran file maksimal 5MB" });
      return;
    }

    handleFileUpload(file);
  };

  // Validation
  const validateStep = (step: WizardStep): boolean => {
    const newErrors: Record<string, string> = {};

    if (step === 1 && !selectedWarehouseId) {
      newErrors.warehouse = "Pilih gudang terlebih dahulu";
    }

    if (step === 2 && !inputMethod) {
      newErrors.inputMethod = "Pilih metode input";
    }

    if (step === 3) {
      if (inputMethod === "manual") {
        if (stockItems.length === 0) {
          newErrors.items = "Tambahkan minimal 1 produk";
        } else {
          stockItems.forEach((item, index) => {
            if (!item.productId) {
              newErrors[`product-${index}`] = "Pilih produk";
            }
            if (!item.quantity || parseFloat(item.quantity) <= 0) {
              newErrors[`quantity-${index}`] = "Quantity harus > 0";
            }
            if (!item.costPerUnit || parseFloat(item.costPerUnit) <= 0) {
              newErrors[`cost-${index}`] = "Harga beli harus > 0";
            }
          });
        }
      } else if (inputMethod === "excel") {
        // Excel validation
        if (!uploadedFile) {
          newErrors.upload = "Upload file Excel terlebih dahulu";
        } else if (validationResult && !validationResult.success) {
          newErrors.validation = "File Excel memiliki error validasi";
        } else if (validationResult && validationResult.existingStocks && validationResult.existingStocks.length > 0) {
          // Check for existing stocks conflicts
          newErrors.existingStocks = "Tidak dapat melanjutkan - produk sudah memiliki stok di gudang";
        } else if (validationResult && validationResult.duplicatesInFile && validationResult.duplicatesInFile.length > 0) {
          // Check for duplicates in file
          newErrors.duplicates = "Tidak dapat melanjutkan - ada produk duplikat dalam file";
        } else if (stockItems.length === 0) {
          newErrors.items = "File Excel tidak mengandung data yang valid";
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Navigation
  const handleNext = () => {
    if (validateStep(currentStep)) {
      if (currentStep === 3 && inputMethod === "excel") {
        // Skip to review for Excel (validation happens on upload)
        setCurrentStep(4);
      } else {
        setCurrentStep((currentStep + 1) as WizardStep);
      }
    }
  };

  const handleBack = () => {
    setCurrentStep((currentStep - 1) as WizardStep);
    setErrors({});
  };

  // Submit
  const handleSubmit = async () => {
    if (!selectedWarehouseId || stockItems.length === 0) return;

    try {
      const payload = {
        warehouseId: selectedWarehouseId,
        items: stockItems.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
          costPerUnit: item.costPerUnit,
          location: item.location || undefined,
          minimumStock: item.minimumStock || undefined,
          maximumStock: item.maximumStock || undefined,
          notes: item.notes || undefined,
        })),
        notes: notes || undefined,
      };

      await submitInitialStock(payload).unwrap();
      setCurrentStep(5); // Success step
    } catch (error: any) {
      setErrors({
        submit: error?.data?.message || "Gagal menyimpan data stok awal",
      });
    }
  };

  // Step 1: Warehouse Selection - Enhanced Design
  const renderWarehouseSelection = () => (
    <Card className="border-2">
      <CardHeader className="bg-gradient-to-r from-blue-50 to-blue-100/50 dark:from-blue-950 dark:to-blue-900/50">
        <CardTitle className="flex items-center gap-2 sm:gap-3 text-lg md:text-xl">
          <div className="flex h-8 w-8 sm:h-10 sm:w-10 items-center justify-center rounded-lg bg-blue-600 text-white flex-shrink-0">
            <Building2 className="h-4 w-4 sm:h-5 sm:w-5" />
          </div>
          <span>Pilih Gudang</span>
        </CardTitle>
        <CardDescription className="text-sm md:text-base mt-2">
          Pilih gudang yang akan diisi stok awal. Anda dapat setup stok untuk produk yang belum pernah memiliki record.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 p-6">
        {isLoadingWarehouses ? (
          <LoadingSpinner size="lg" text="Memuat daftar gudang..." />
        ) : warehousesData?.data && warehousesData.data.length > 0 ? (
          <>
            <div className="grid gap-4">
              {warehousesData.data.map((warehouse) => {
                const hasStock = statusData?.find(
                  (s) => s.warehouseId === warehouse.id
                )?.hasInitialStock;

                return (
                  <Card
                    key={warehouse.id}
                    className={`cursor-pointer transition-all duration-300 hover:shadow-lg ${
                      selectedWarehouseId === warehouse.id
                        ? "border-2 border-blue-600 bg-blue-50/50 shadow-md dark:bg-blue-950/30"
                        : "border-2 border-transparent hover:border-blue-200"
                    }`}
                    onClick={() => setSelectedWarehouseId(warehouse.id)}
                  >
                    <CardContent className="p-5">
                      <div className="flex items-start justify-between">
                        <div className="flex items-start gap-4">
                          <div className={`flex h-12 w-12 items-center justify-center rounded-lg transition-colors ${
                            selectedWarehouseId === warehouse.id
                              ? "bg-blue-600 text-white"
                              : "bg-gray-100 text-gray-600"
                          }`}>
                            <Building2 className="h-6 w-6" />
                          </div>
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <p className="font-semibold text-lg">{warehouse.name}</p>
                              {hasStock && (
                                <Badge variant="secondary" className="text-xs bg-amber-100 text-amber-800 border-amber-300">
                                  <AlertCircle className="mr-1 h-3 w-3" />
                                  Ada Stok
                                </Badge>
                              )}
                            </div>
                            <p className="text-sm text-gray-600 font-mono">
                              {warehouse.code}
                            </p>
                            {warehouse.address && (
                              <p className="text-xs text-muted-foreground max-w-md">
                                📍 {warehouse.address}
                              </p>
                            )}
                          </div>
                        </div>
                        {selectedWarehouseId === warehouse.id && (
                          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-600">
                            <CheckCircle2 className="h-5 w-5 text-white" />
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>

            {warehouseHasStock && (
              <Alert className="border-2 border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
                <AlertCircle className="h-5 w-5 text-blue-600" />
                <AlertTitle className="text-blue-900 dark:text-blue-100 font-semibold">
                  💡 Informasi Gudang
                </AlertTitle>
                <AlertDescription className="text-blue-800 dark:text-blue-200">
                  Gudang ini sudah memiliki beberapa produk dengan stok. Setup stok awal hanya dapat dilakukan untuk produk yang <strong>belum pernah memiliki record stok</strong> di gudang ini. Produk yang sudah ada (termasuk yang quantity 0) tidak akan muncul dalam daftar.
                </AlertDescription>
              </Alert>
            )}

            {errors.warehouse && (
              <p className="text-sm text-red-500">{errors.warehouse}</p>
            )}
          </>
        ) : (
          <Alert className="border-2 border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/30">
            <AlertCircle className="h-5 w-5 text-amber-600" />
            <AlertTitle className="text-amber-900 dark:text-amber-100 font-semibold">
              ⚠️ Tidak Ada Gudang
            </AlertTitle>
            <AlertDescription className="text-amber-800 dark:text-amber-200">
              Belum ada gudang aktif. Silakan buat gudang terlebih dahulu di menu
              <strong> Master → Gudang</strong>.
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );

  // Step 2: Input Method Selection - Enhanced
  const renderInputMethodSelection = () => (
    <Card className="border-2">
      <CardHeader className="bg-gradient-to-r from-blue-50 to-blue-100/50 dark:from-blue-950 dark:to-blue-900/50">
        <CardTitle className="flex items-center gap-2 sm:gap-3 text-lg md:text-xl">
          <div className="flex h-8 w-8 sm:h-10 sm:w-10 items-center justify-center rounded-lg bg-blue-600 text-white flex-shrink-0">
            <Edit3 className="h-4 w-4 sm:h-5 sm:w-5" />
          </div>
          <span>Metode Input</span>
        </CardTitle>
        <CardDescription className="text-sm md:text-base mt-2">
          Pilih cara untuk memasukkan data stok awal. Anda bisa input manual atau import dari Excel.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 p-6">
        <div className="grid gap-6 md:grid-cols-2">
          {/* Manual Entry */}
          <Card
            className={`cursor-pointer transition-all duration-300 hover:shadow-lg ${
              inputMethod === "manual"
                ? "border-2 border-blue-600 shadow-md"
                : "border-2 border-transparent hover:border-blue-200"
            }`}
            onClick={() => setInputMethod("manual")}
          >
            <CardContent className="p-8">
              <div className="flex flex-col items-center text-center space-y-4">
                <div className={`flex h-16 w-16 items-center justify-center rounded-2xl transition-colors ${
                  inputMethod === "manual" ? "bg-blue-600" : "bg-blue-100"
                }`}>
                  <Edit3 className={`h-8 w-8 ${
                    inputMethod === "manual" ? "text-white" : "text-blue-600"
                  }`} />
                </div>
                <div>
                  <h3 className="font-bold text-lg mb-2">Input Manual</h3>
                  <p className="text-sm text-muted-foreground">
                    Masukkan data produk satu per satu melalui form interaktif
                  </p>
                </div>
                {inputMethod === "manual" && (
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-600">
                    <CheckCircle2 className="h-5 w-5 text-white" />
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Excel Import */}
          <Card
            className={`cursor-pointer transition-all duration-300 hover:shadow-lg ${
              inputMethod === "excel"
                ? "border-2 border-green-600 shadow-md"
                : "border-2 border-transparent hover:border-green-200"
            }`}
            onClick={() => setInputMethod("excel")}
          >
            <CardContent className="p-8">
              <div className="flex flex-col items-center text-center space-y-4">
                <div className={`flex h-16 w-16 items-center justify-center rounded-2xl transition-colors ${
                  inputMethod === "excel" ? "bg-green-600" : "bg-green-100"
                }`}>
                  <FileSpreadsheet className={`h-8 w-8 ${
                    inputMethod === "excel" ? "text-white" : "text-green-600"
                  }`} />
                </div>
                <div>
                  <h3 className="font-bold text-lg mb-2">Import Excel</h3>
                  <p className="text-sm text-muted-foreground">
                    Upload file Excel dengan format template yang disediakan
                  </p>
                </div>
                {inputMethod === "excel" && (
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-600">
                    <CheckCircle2 className="h-5 w-5 text-white" />
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {errors.inputMethod && (
          <p className="text-sm text-red-500">{errors.inputMethod}</p>
        )}
      </CardContent>
    </Card>
  );

  // Step 3a: Manual Entry - Enhanced
  const renderManualEntry = () => (
    <Card className="border-2">
      <CardHeader className="bg-gradient-to-r from-blue-50 to-blue-100/50 dark:from-blue-950 dark:to-blue-900/50">
        <CardTitle className="flex items-center gap-2 sm:gap-3 text-lg md:text-xl">
          <div className="flex h-8 w-8 sm:h-10 sm:w-10 items-center justify-center rounded-lg bg-blue-600 text-white flex-shrink-0">
            <Edit3 className="h-4 w-4 sm:h-5 sm:w-5" />
          </div>
          <span>Input Manual</span>
        </CardTitle>
        <CardDescription className="text-sm md:text-base">
          Masukkan data stok untuk gudang: <strong>{selectedWarehouse?.name}</strong>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 p-6">
        {/* Warning if pagination might affect results */}
        {existingStocksData?.pagination?.hasMore && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Peringatan: Data Tidak Lengkap</AlertTitle>
            <AlertDescription>
              Gudang ini memiliki lebih dari {existingStocksData.pagination.totalItems} produk dengan stok.
              Beberapa produk mungkin tidak terfilter dengan benar. Silakan hubungi administrator.
            </AlertDescription>
          </Alert>
        )}

        {/* Info about available products - Enhanced */}
        <Alert className="border-2 border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
          <AlertCircle className="h-5 w-5 text-blue-600" />
          <AlertTitle className="text-blue-900 dark:text-blue-100 font-semibold">
            💡 Produk Tersedia: {availableProducts.length} produk
          </AlertTitle>
          <AlertDescription className="text-blue-800 dark:text-blue-200">
            Hanya produk yang <strong>belum pernah memiliki record stok</strong> di <strong>{selectedWarehouse?.name}</strong> yang dapat dipilih untuk setup stok awal.
            {availableProducts.length === 0 && " Semua produk sudah memiliki record stok di gudang ini."}
            {existingStocksData?.data && ` (${existingStocksData.data.length} produk sudah memiliki record stok)`}
          </AlertDescription>
        </Alert>

        {/* Add Button - Enhanced */}
        <div className="flex justify-start">
          <Button
            onClick={handleAddRow}
            size="lg"
            disabled={availableProducts.length === 0}
            className="h-11 px-6 bg-blue-600 hover:bg-blue-700"
          >
            <Plus className="mr-2 h-5 w-5" />
            Tambah Produk
          </Button>
        </div>

        {errors.items && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{errors.items}</AlertDescription>
          </Alert>
        )}

        {/* Stock Items Table */}
        {stockItems.length > 0 && (
          <div className="rounded-md border overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[250px]">Produk *</TableHead>
                  <TableHead className="w-[120px]">Quantity *</TableHead>
                  <TableHead className="w-[150px]">Harga Beli *</TableHead>
                  <TableHead className="w-[150px]">Lokasi</TableHead>
                  <TableHead className="w-[120px]">Min. Stok</TableHead>
                  <TableHead className="w-[120px]">Max. Stok</TableHead>
                  <TableHead className="w-[200px]">Catatan</TableHead>
                  <TableHead className="w-[80px]">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {stockItems.map((item, index) => (
                  <TableRow key={item.tempId}>
                    {/* Product Selection */}
                    <TableCell>
                      <Select
                        value={item.productId}
                        onValueChange={(value) =>
                          handleUpdateRow(item.tempId, "productId", value)
                        }
                      >
                        <SelectTrigger className={errors[`product-${index}`] ? "border-red-500" : ""}>
                          <SelectValue placeholder="Pilih produk..." />
                        </SelectTrigger>
                        <SelectContent>
                          {isLoadingProducts || isLoadingExistingStocks || isFetchingExistingStocks ? (
                            <div className="p-2 text-sm text-muted-foreground">
                              Loading...
                            </div>
                          ) : availableProducts.length === 0 ? (
                            <div className="p-2 text-sm text-muted-foreground">
                              Semua produk sudah memiliki stok di gudang ini
                            </div>
                          ) : (
                            availableProducts.map((product) => (
                              <SelectItem key={product.id} value={product.id}>
                                {product.code} - {product.name}
                              </SelectItem>
                            ))
                          )}
                        </SelectContent>
                      </Select>
                      {errors[`product-${index}`] && (
                        <p className="text-xs text-red-500 mt-1">
                          {errors[`product-${index}`]}
                        </p>
                      )}
                    </TableCell>

                    {/* Quantity */}
                    <TableCell>
                      <div className="space-y-1">
                        <Input
                          type="number"
                          step="0.01"
                          value={item.quantity}
                          onChange={(e) =>
                            handleUpdateRow(item.tempId, "quantity", e.target.value)
                          }
                          placeholder="0"
                          className={errors[`quantity-${index}`] ? "border-red-500" : ""}
                        />
                        {item.baseUnit && (
                          <p className="text-xs text-muted-foreground">
                            {item.baseUnit}
                          </p>
                        )}
                        {errors[`quantity-${index}`] && (
                          <p className="text-xs text-red-500">
                            {errors[`quantity-${index}`]}
                          </p>
                        )}
                      </div>
                    </TableCell>

                    {/* Cost Per Unit */}
                    <TableCell>
                      <Input
                        type="number"
                        step="0.01"
                        value={item.costPerUnit}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "costPerUnit", e.target.value)
                        }
                        placeholder="0"
                        className={errors[`cost-${index}`] ? "border-red-500" : ""}
                      />
                      {errors[`cost-${index}`] && (
                        <p className="text-xs text-red-500 mt-1">
                          {errors[`cost-${index}`]}
                        </p>
                      )}
                    </TableCell>

                    {/* Location */}
                    <TableCell>
                      <Input
                        value={item.location || ""}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "location", e.target.value)
                        }
                        placeholder="Rak A1"
                      />
                    </TableCell>

                    {/* Minimum Stock */}
                    <TableCell>
                      <Input
                        type="number"
                        step="0.01"
                        value={item.minimumStock || ""}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "minimumStock", e.target.value)
                        }
                        placeholder="0"
                      />
                    </TableCell>

                    {/* Maximum Stock */}
                    <TableCell>
                      <Input
                        type="number"
                        step="0.01"
                        value={item.maximumStock || ""}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "maximumStock", e.target.value)
                        }
                        placeholder="0"
                      />
                    </TableCell>

                    {/* Notes */}
                    <TableCell>
                      <Input
                        value={item.notes || ""}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "notes", e.target.value)
                        }
                        placeholder="Catatan..."
                      />
                    </TableCell>

                    {/* Actions */}
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveRow(item.tempId)}
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        {/* General Notes */}
        <div className="space-y-2">
          <Label htmlFor="notes">Catatan Umum (Opsional)</Label>
          <Input
            id="notes"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Catatan untuk setup stok awal ini..."
          />
        </div>
      </CardContent>
    </Card>
  );

  // Step 3b: Excel Import - Enhanced
  const renderExcelImport = () => (
    <Card className="border-2">
      <CardHeader className="bg-gradient-to-r from-green-50 to-green-100/50 dark:from-green-950 dark:to-green-900/50">
        <CardTitle className="flex items-center gap-2 sm:gap-3 text-lg md:text-xl">
          <div className="flex h-8 w-8 sm:h-10 sm:w-10 items-center justify-center rounded-lg bg-green-600 text-white flex-shrink-0">
            <FileSpreadsheet className="h-4 w-4 sm:h-5 sm:w-5" />
          </div>
          <span>Import Excel</span>
        </CardTitle>
        <CardDescription className="text-sm md:text-base">
          Upload file Excel untuk gudang: <strong>{selectedWarehouse?.name}</strong>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6 p-6">
        {/* Download Template - Enhanced */}
        <Alert className="border-2 border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
          <Download className="h-5 w-5 text-blue-600" />
          <AlertTitle className="text-blue-900 dark:text-blue-100 font-semibold">
            📥 Template Excel
          </AlertTitle>
          <AlertDescription className="text-blue-800 dark:text-blue-200">
            Download template Excel terlebih dahulu dan isi sesuai format yang disediakan.
            <Button
              variant="link"
              className="p-0 h-auto ml-2 text-blue-600 hover:text-blue-700 font-semibold"
              onClick={handleDownloadTemplate}
            >
              Download Template →
            </Button>
          </AlertDescription>
        </Alert>

        {/* Upload Area - Enhanced */}
        <div className="border-2 border-dashed border-gray-300 rounded-lg p-12 bg-gradient-to-br from-gray-50 to-gray-100/50 dark:from-gray-900 dark:to-gray-950/50 hover:border-green-400 transition-colors">
          <div className="flex flex-col items-center text-center space-y-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
              <Upload className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold mb-2">Upload File Excel</h3>
              <p className="text-sm text-muted-foreground">
                Klik tombol di bawah untuk memilih file Excel
              </p>
            </div>
            <input
              type="file"
              id="excel-upload"
              accept=".xlsx,.xls"
              className="hidden"
              onChange={handleFileInputChange}
              disabled={isUploadingFile || isLoadingExistingStocks || isFetchingExistingStocks}
            />
            <Button
              onClick={() => document.getElementById("excel-upload")?.click()}
              disabled={isUploadingFile || isLoadingExistingStocks || isFetchingExistingStocks}
              size="lg"
              className="h-11 px-8 bg-green-600 hover:bg-green-700"
            >
              {isUploadingFile ? (
                <>
                  <LoadingSpinner className="mr-2 h-5 w-5" />
                  Memproses...
                </>
              ) : isLoadingExistingStocks || isFetchingExistingStocks ? (
                <>
                  <LoadingSpinner className="mr-2 h-5 w-5" />
                  Memuat data stok...
                </>
              ) : (
                <>
                  <Upload className="mr-2 h-5 w-5" />
                  Pilih File
                </>
              )}
            </Button>
            <p className="text-xs text-muted-foreground font-medium">
              Format: .xlsx, .xls • Maksimal: 5MB
            </p>
          </div>
        </div>

        {/* Uploaded File Info - Enhanced */}
        {uploadedFile && (
          <Alert className="border-2 border-green-200 bg-green-50/50 dark:border-green-800 dark:bg-green-950/30">
            <FileSpreadsheet className="h-5 w-5 text-green-600" />
            <AlertTitle className="text-green-900 dark:text-green-100 font-semibold">
              ✅ File Terpilih
            </AlertTitle>
            <AlertDescription className="text-green-800 dark:text-green-200">
              <span className="font-medium">{uploadedFile.name}</span>
              <span className="text-sm ml-2">
                ({(uploadedFile.size / 1024).toFixed(2)} KB)
              </span>
            </AlertDescription>
          </Alert>
        )}

        {/* Upload Errors - Enhanced */}
        {errors.upload && (
          <Alert variant="destructive" className="border-2">
            <AlertCircle className="h-5 w-5" />
            <AlertTitle className="font-semibold">Error Upload</AlertTitle>
            <AlertDescription className="text-base">{errors.upload}</AlertDescription>
          </Alert>
        )}

        {/* Validation Results */}
        {validationResult && !validationResult.success && (
          <div className="space-y-4">
            {/* Duplicate Conflicts */}
            {validationResult.duplicatesInFile.length > 0 && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>
                  Produk Duplikat dalam File ({validationResult.duplicatesInFile.length})
                </AlertTitle>
                <AlertDescription>
                  <div className="mt-2 space-y-1">
                    {validationResult.duplicatesInFile.slice(0, 5).map((conflict, idx) => (
                      <div key={idx} className="text-sm">
                        • Baris {conflict.row}: {conflict.productCode} - {conflict.productName}
                      </div>
                    ))}
                    {validationResult.duplicatesInFile.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.duplicatesInFile.length - 5} lainnya
                      </div>
                    )}
                  </div>
                </AlertDescription>
              </Alert>
            )}

            {/* Info: Products without stock (valid for initial stock input) */}
            {validationResult.noStockProducts && validationResult.noStockProducts.length > 0 && (
              <Alert className="border-blue-200 bg-blue-50 dark:border-blue-900 dark:bg-blue-950">
                <AlertCircle className="h-4 w-4 text-blue-600" />
                <AlertTitle className="text-blue-900 dark:text-blue-100">
                  Produk Belum Memiliki Stok ({validationResult.noStockProducts.length})
                </AlertTitle>
                <AlertDescription className="text-blue-800 dark:text-blue-200">
                  <p className="mb-2 text-sm">
                    Produk berikut <strong>belum pernah memiliki record stok</strong> di <strong>{selectedWarehouse?.name}</strong> dan siap untuk input stok awal:
                  </p>
                  <div className="mt-2 space-y-1">
                    {validationResult.noStockProducts.slice(0, 5).map((product, idx) => (
                      <div key={idx} className="text-sm">
                        • Baris {product.row}: <strong>{product.productCode}</strong> - {product.productName}
                        <span className="ml-2 text-xs">({product.quantity} unit)</span>
                      </div>
                    ))}
                    {validationResult.noStockProducts.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.noStockProducts.length - 5} produk lainnya
                      </div>
                    )}
                  </div>
                </AlertDescription>
              </Alert>
            )}

            {/* Existing Stock Conflicts */}
            {validationResult.existingStocks.length > 0 && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>
                  Produk Sudah Memiliki Stok ({validationResult.existingStocks.length})
                </AlertTitle>
                <AlertDescription>
                  <div className="mt-2 space-y-1">
                    {validationResult.existingStocks.slice(0, 5).map((conflict, idx) => (
                      <div key={idx} className="text-sm">
                        • Baris {conflict.row}: {conflict.productCode} - {conflict.productName}
                        <br />
                        <span className="ml-4 text-xs">
                          Stok saat ini: {conflict.currentQuantity} | Stok baru: {conflict.newQuantity}
                        </span>
                      </div>
                    ))}
                    {validationResult.existingStocks.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.existingStocks.length - 5} lainnya
                      </div>
                    )}
                  </div>
                </AlertDescription>
              </Alert>
            )}

            {/* Validation Errors */}
            {validationResult.errors.length > 0 && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>
                  Error Validasi ({validationResult.errors.length})
                </AlertTitle>
                <AlertDescription>
                  <div className="mt-2 space-y-1">
                    {validationResult.errors.slice(0, 5).map((error, idx) => (
                      <div key={idx} className="text-sm">
                        • Baris {error.row}, Kolom {error.field}: {error.message}
                      </div>
                    ))}
                    {validationResult.errors.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.errors.length - 5} error lainnya
                      </div>
                    )}
                  </div>
                </AlertDescription>
              </Alert>
            )}
          </div>
        )}

        {/* Success Message */}
        {validationResult && validationResult.success && (
          <Alert>
            <CheckCircle2 className="h-4 w-4 text-green-500" />
            <AlertTitle>Validasi Berhasil</AlertTitle>
            <AlertDescription>
              {validationResult.validItems.length} produk berhasil divalidasi dan siap untuk disimpan.
            </AlertDescription>
          </Alert>
        )}

        {/* Instructions */}
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Panduan Import</AlertTitle>
          <AlertDescription>
            <ul className="list-disc list-inside text-sm space-y-1 mt-2">
              <li>Pastikan format file sesuai dengan template</li>
              <li>Kolom Kode Produk, Quantity, dan Harga Beli wajib diisi</li>
              <li>Gunakan kode produk yang sudah terdaftar di sistem</li>
              <li>Quantity dan harga harus berupa angka positif</li>
              <li>Tidak boleh ada produk duplikat dalam file</li>
              <li>Produk tidak boleh sudah memiliki stok di gudang</li>
            </ul>
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  );

  // Step 4: Review & Validation - Enhanced
  const renderReview = () => {
    const totalItems = stockItems.length;
    const totalQuantity = stockItems.reduce(
      (sum, item) => sum + parseFloat(item.quantity || "0"),
      0
    );
    const totalValue = stockItems.reduce(
      (sum, item) =>
        sum +
        parseFloat(item.quantity || "0") * parseFloat(item.costPerUnit || "0"),
      0
    );

    return (
      <Card className="border-2">
        <CardHeader className="bg-gradient-to-r from-green-50 to-green-100/50 dark:from-green-950 dark:to-green-900/50">
          <CardTitle className="flex items-center gap-2 sm:gap-3 text-lg md:text-xl">
            <div className="flex h-8 w-8 sm:h-10 sm:w-10 items-center justify-center rounded-lg bg-green-600 text-white flex-shrink-0">
              <CheckCircle2 className="h-4 w-4 sm:h-5 sm:w-5" />
            </div>
            <span>Review & Validasi</span>
          </CardTitle>
          <CardDescription className="text-sm md:text-base">
            Periksa kembali data sebelum menyimpan ke sistem
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6 p-6">
          {/* Summary Cards - Enhanced */}
          <div className="grid gap-4 md:grid-cols-3">
            <Card className="border-2 border-blue-100 bg-gradient-to-br from-blue-50 to-blue-100/30 dark:border-blue-900 dark:from-blue-950 dark:to-blue-900/30">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600 text-white">
                    <Package className="h-6 w-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Item</p>
                    <p className="text-3xl font-bold text-blue-600">{totalItems}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card className="border-2 border-purple-100 bg-gradient-to-br from-purple-50 to-purple-100/30 dark:border-purple-900 dark:from-purple-950 dark:to-purple-900/30">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600 text-white">
                    <Package className="h-6 w-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Quantity</p>
                    <p className="text-3xl font-bold text-purple-600">
                      {totalQuantity.toLocaleString("id-ID")}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card className="border-2 border-green-100 bg-gradient-to-br from-green-50 to-green-100/30 dark:border-green-900 dark:from-green-950 dark:to-green-900/30">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600 text-white">
                    <DollarSign className="h-6 w-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Nilai</p>
                    <p className="text-2xl font-bold text-green-600">
                      Rp {totalValue.toLocaleString("id-ID")}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Warehouse Info - Enhanced */}
          <Card className="border-2 border-blue-100 bg-blue-50/30 dark:border-blue-900 dark:bg-blue-950/30">
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600 text-white">
                  <Building2 className="h-6 w-6" />
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Gudang</p>
                  <p className="text-lg font-bold">
                    {selectedWarehouse?.name}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Kode: {selectedWarehouse?.code}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Separator className="my-6" />

          {/* Items List - Enhanced */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Package className="h-5 w-5 text-blue-600" />
              <Label className="text-base font-semibold">Daftar Produk ({stockItems.length} item)</Label>
            </div>
            <div className="rounded-lg border-2 overflow-hidden shadow-sm">
              <Table>
                <TableHeader className="bg-gray-50 dark:bg-gray-900">
                  <TableRow>
                    <TableHead className="font-semibold">Produk</TableHead>
                    <TableHead className="text-right font-semibold">Quantity</TableHead>
                    <TableHead className="text-right font-semibold">Harga Beli</TableHead>
                    <TableHead className="text-right font-semibold">Subtotal</TableHead>
                    <TableHead className="font-semibold">Lokasi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stockItems.map((item, index) => {
                    const subtotal =
                      parseFloat(item.quantity || "0") *
                      parseFloat(item.costPerUnit || "0");
                    return (
                      <TableRow
                        key={item.tempId}
                        className={index % 2 === 0 ? "bg-white dark:bg-gray-950" : "bg-gray-50/50 dark:bg-gray-900/50"}
                      >
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900">
                              <Package className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                            </div>
                            <div>
                              <p className="font-semibold">{item.productName}</p>
                              <p className="text-sm text-muted-foreground">
                                {item.productCode}
                              </p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <span className="font-medium">
                            {parseFloat(item.quantity).toLocaleString("id-ID")}
                          </span>
                          <span className="text-sm text-muted-foreground ml-1">
                            {item.baseUnit}
                          </span>
                        </TableCell>
                        <TableCell className="text-right font-medium">
                          Rp {parseFloat(item.costPerUnit).toLocaleString("id-ID")}
                        </TableCell>
                        <TableCell className="text-right">
                          <span className="font-bold text-green-600">
                            Rp {subtotal.toLocaleString("id-ID")}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="font-normal">
                            {item.location || "Tidak ada"}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </div>

          {/* Notes - Enhanced */}
          {notes && (
            <>
              <Separator className="my-6" />
              <Card className="border-2 border-amber-100 bg-amber-50/30 dark:border-amber-900 dark:bg-amber-950/30">
                <CardContent className="p-6">
                  <div className="flex gap-3">
                    <AlertCircle className="h-5 w-5 text-amber-600 flex-shrink-0 mt-0.5" />
                    <div className="space-y-1">
                      <Label className="text-base font-semibold">Catatan</Label>
                      <p className="text-sm text-muted-foreground">{notes}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </>
          )}

          {/* Submit Error - Enhanced */}
          {errors.submit && (
            <Alert variant="destructive" className="border-2">
              <AlertCircle className="h-5 w-5" />
              <AlertTitle className="font-semibold">Terjadi Kesalahan</AlertTitle>
              <AlertDescription className="text-base">{errors.submit}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    );
  };

  // Step 5: Success - Enhanced Celebration
  const renderSuccess = () => (
    <Card className="border-2 border-green-200 shadow-xl">
      <CardContent className="p-12">
        <div className="flex flex-col items-center text-center space-y-8">
          {/* Success Animation */}
          <div className="relative">
            <div className="absolute inset-0 rounded-full bg-green-400 blur-2xl opacity-30 animate-pulse"></div>
            <div className="relative rounded-full bg-gradient-to-br from-green-500 to-green-600 p-8 shadow-2xl">
              <CheckCircle2 className="h-20 w-20 text-white" />
            </div>
          </div>

          {/* Success Message */}
          <div className="space-y-3">
            <h2 className="text-3xl font-bold bg-gradient-to-r from-green-600 to-green-700 bg-clip-text text-transparent">
              🎉 Setup Stok Awal Berhasil!
            </h2>
            <p className="text-muted-foreground max-w-md text-lg">
              Data stok awal untuk gudang <strong className="text-green-700">{selectedWarehouse?.name}</strong>{" "}
              telah berhasil disimpan ke sistem.
            </p>
          </div>

          {/* Stats Summary */}
          <div className="grid grid-cols-2 gap-6 w-full max-w-md">
            <div className="rounded-lg border-2 border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950/30">
              <p className="text-2xl font-bold text-green-700">
                {stockItems.length}
              </p>
              <p className="text-sm text-muted-foreground">Produk Ditambahkan</p>
            </div>
            <div className="rounded-lg border-2 border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-950/30">
              <p className="text-2xl font-bold text-blue-700">
                {selectedWarehouse?.name.split(' ')[0]}
              </p>
              <p className="text-sm text-muted-foreground">Gudang</p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-4 pt-4">
            <Button
              onClick={() => router.push("/inventory/stock")}
              className="h-12 px-8 text-base font-medium bg-green-600 hover:bg-green-700"
              size="lg"
            >
              <Package className="mr-2 h-5 w-5" />
              Lihat Stok Barang
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setCurrentStep(1);
                setSelectedWarehouseId("");
                setInputMethod("manual");
                setStockItems([]);
                setNotes("");
                setErrors({});
              }}
              className="h-12 px-6 text-base font-medium"
              size="lg"
            >
              <Building2 className="mr-2 h-5 w-5" />
              Setup Gudang Lain
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );

  // Render current step content
  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return renderWarehouseSelection();
      case 2:
        return renderInputMethodSelection();
      case 3:
        return inputMethod === "manual"
          ? renderManualEntry()
          : renderExcelImport();
      case 4:
        return renderReview();
      case 5:
        return renderSuccess();
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col gap-6 p-4 pt-0 shrink-0">
      {/* Page Header with Gradient Background */}
      <div className="relative overflow-hidden rounded-lg bg-gradient-to-r from-blue-600 to-blue-700 p-4 sm:p-6 md:p-8 shadow-lg">
        <div className="relative z-10 space-y-2">
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="flex h-10 w-10 sm:h-12 sm:w-12 items-center justify-center rounded-lg bg-white/20 backdrop-blur-sm flex-shrink-0">
              <Package className="h-5 w-5 sm:h-6 sm:w-6 text-white" />
            </div>
            <div className="min-w-0">
              <h1 className="text-xl sm:text-2xl md:text-3xl font-bold tracking-tight text-white">Setup Stok Awal</h1>
              <p className="text-blue-100 mt-1 text-xs sm:text-sm">
                Setup stok pertama kali untuk produk yang belum pernah memiliki record di gudang
              </p>
            </div>
          </div>
        </div>
        {/* Decorative circles - hidden on mobile */}
        <div className="absolute -right-10 -top-10 h-24 w-24 sm:h-32 sm:w-32 md:h-40 md:w-40 rounded-full bg-white/10"></div>
        <div className="absolute -bottom-10 -left-10 h-24 w-24 sm:h-32 sm:w-32 md:h-40 md:w-40 rounded-full bg-white/10"></div>
      </div>

      {/* Wizard Steps Indicator - Responsive */}
      {currentStep !== 5 && (
        <Card className="border-2 shadow-md">
          <CardContent className="p-4 md:p-6">
            {/* Mobile: Compact Horizontal Steps */}
            <div className="md:hidden">
              <div className="flex items-center justify-center mb-3">
                {[
                  { step: 1, label: "Pilih Gudang", icon: Building2 },
                  { step: 2, label: "Metode Input", icon: Edit3 },
                  { step: 3, label: "Input Data", icon: FileSpreadsheet },
                  { step: 4, label: "Review", icon: CheckCircle2 },
                ].map((item, index) => (
                  <div key={item.step} className="flex items-center">
                    <div className="flex flex-col items-center gap-1">
                      <div
                        className={`flex h-8 w-8 items-center justify-center rounded-full border-2 font-semibold transition-all duration-300 ${
                          currentStep >= item.step
                            ? "border-blue-600 bg-blue-600 text-white shadow-md"
                            : "border-gray-300 bg-white text-gray-400"
                        }`}
                      >
                        {currentStep > item.step ? (
                          <CheckCircle2 className="h-4 w-4" />
                        ) : (
                          <item.icon className="h-4 w-4" />
                        )}
                      </div>
                      <p className={`text-[10px] font-semibold text-center ${
                        currentStep >= item.step ? "text-blue-600" : "text-gray-500"
                      }`}>
                        {item.label}
                      </p>
                    </div>
                    {index < 3 && (
                      <ChevronRight
                        className={`mx-2 h-4 w-4 flex-shrink-0 transition-colors duration-300 ${
                          currentStep > item.step
                            ? "text-blue-600"
                            : "text-gray-300"
                        }`}
                      />
                    )}
                  </div>
                ))}
              </div>
              <div className="text-center">
                <p className="text-xs text-muted-foreground">
                  Langkah {currentStep} dari 4
                </p>
              </div>
            </div>

            {/* Desktop: Full Steps with Labels */}
            <div className="hidden md:flex items-center justify-center">
              {[
                { step: 1, label: "Pilih Gudang", icon: Building2 },
                { step: 2, label: "Metode Input", icon: Edit3 },
                { step: 3, label: "Input Data", icon: FileSpreadsheet },
                { step: 4, label: "Review", icon: CheckCircle2 },
              ].map((item, index) => (
                <div key={item.step} className="flex items-center">
                  <div className="flex flex-col items-center gap-2">
                    <div
                      className={`flex h-12 w-12 items-center justify-center rounded-full border-2 font-semibold transition-all duration-300 ${
                        currentStep >= item.step
                          ? "border-blue-600 bg-blue-600 text-white shadow-lg shadow-blue-200"
                          : "border-gray-300 bg-white text-gray-400"
                      }`}
                    >
                      {currentStep > item.step ? (
                        <CheckCircle2 className="h-6 w-6" />
                      ) : (
                        <item.icon className="h-5 w-5" />
                      )}
                    </div>
                    <div className="flex flex-col items-center">
                      <p className={`text-xs font-semibold ${
                        currentStep >= item.step ? "text-blue-600" : "text-gray-500"
                      }`}>
                        Step {item.step}
                      </p>
                      <p className={`text-sm font-medium ${
                        currentStep >= item.step ? "text-gray-900" : "text-gray-500"
                      }`}>
                        {item.label}
                      </p>
                    </div>
                  </div>
                  {index < 3 && (
                    <ChevronRight
                      className={`mx-4 md:mx-6 lg:mx-8 h-6 w-6 md:h-7 md:w-7 lg:h-8 lg:w-8 flex-shrink-0 transition-colors duration-300 ${
                        currentStep > item.step
                          ? "text-blue-600"
                          : "text-gray-300"
                      }`}
                    />
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step Content */}
      {renderStepContent()}

      {/* Navigation Buttons - Enhanced */}
      {currentStep !== 5 && (
        <Card className="border-2 shadow-md bg-gradient-to-r from-gray-50 to-white dark:from-gray-900 dark:to-gray-950">
          <CardContent className="p-4 sm:p-6">
            <div className="flex flex-col sm:flex-row justify-between items-center gap-3 sm:gap-0">
              <Button
                variant="outline"
                onClick={handleBack}
                disabled={currentStep === 1 || isSubmitting}
                className="h-10 sm:h-12 px-4 sm:px-6 text-sm sm:text-base font-medium hover:bg-gray-100 w-full sm:w-auto"
              >
                <ChevronLeft className="mr-2 h-4 w-4 sm:h-5 sm:w-5" />
                Kembali
              </Button>

              <div className="text-xs sm:text-sm text-muted-foreground font-medium order-first sm:order-none">
                Langkah {currentStep} dari 4
              </div>

              {currentStep < 4 ? (
                <Button
                  onClick={handleNext}
                  disabled={isSubmitting || hasValidationConflicts}
                  className={`h-10 sm:h-12 px-6 sm:px-8 text-sm sm:text-base font-medium bg-blue-600 hover:bg-blue-700 w-full sm:w-auto ${
                    hasValidationConflicts ? "opacity-50 cursor-not-allowed" : ""
                  }`}
                >
                  Selanjutnya
                  <ChevronRight className="ml-2 h-4 w-4 sm:h-5 sm:w-5" />
                </Button>
              ) : (
                <Button
                  onClick={handleSubmit}
                  disabled={isSubmitting}
                  className="h-10 sm:h-12 px-6 sm:px-8 text-sm sm:text-base font-medium bg-green-600 hover:bg-green-700 w-full sm:w-auto"
                >
                  {isSubmitting ? (
                    <>
                      <LoadingSpinner size="sm" className="mr-2" />
                      Menyimpan...
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="mr-2 h-4 w-4 sm:h-5 sm:w-5" />
                      Simpan Setup Stok Awal
                    </>
                  )}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
