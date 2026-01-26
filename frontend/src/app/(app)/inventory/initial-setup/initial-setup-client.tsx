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
  ArrowRight,
  ArrowLeft,
  Upload,
  Download,
  Plus,
  Trash2,
  AlertCircle,
  DollarSign,
  HelpCircle,
  Check,
  ClipboardCheck,
} from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
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
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  context: _context, // TODO: Use for onboarding flow customization
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  source: _source, // TODO: Use for analytics/UX tracking
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
  const [validationResult, setValidationResult] =
    useState<ExcelValidationResult | null>(null);

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
    () =>
      statusData?.find((s) => s.warehouseId === selectedWarehouseId)
        ?.hasInitialStock,
    [statusData, selectedWarehouseId]
  );

  // Check if current step has validation conflicts (for button disable state)
  const hasValidationConflicts = useMemo(() => {
    if (currentStep !== 3 || inputMethod !== "excel" || !validationResult) {
      return false;
    }

    // Check for existing stocks conflicts or duplicates in file
    const hasConflicts =
      validationResult.existingStocks &&
      validationResult.existingStocks.length > 0;
    const hasDuplicates =
      validationResult.duplicatesInFile &&
      validationResult.duplicatesInFile.length > 0;

    return hasConflicts || hasDuplicates;
  }, [currentStep, inputMethod, validationResult]);

  // Filter products to show only those without existing stock in selected warehouse
  const availableProducts = useMemo(() => {
    if (!productsData?.data || !selectedWarehouseId) {
      return [];
    }

    // If no existing stocks data yet, show all products
    if (!existingStocksData?.data) {
      return productsData.data;
    }

    // Create a Set of product IDs that already have stock in this warehouse
    const productsWithStock = new Set(
      existingStocksData.data.map((stock) => stock.productID)
    );

    // Filter out products that already have stock
    const filtered = productsData.data.filter(
      (product) => !productsWithStock.has(product.id)
    );

    return filtered;
  }, [productsData, existingStocksData, selectedWarehouseId]);

  // Calculate remaining available products (not yet selected in manual input)
  const remainingAvailableProducts = useMemo(() => {
    const selectedProductIds = new Set(
      stockItems.map((item) => item.productId).filter(Boolean)
    );
    return availableProducts.filter(
      (product) => !selectedProductIds.has(product.id)
    );
  }, [availableProducts, stockItems]);

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
  const handleUpdateRow = (
    tempId: string,
    field: keyof StockItemRow,
    value: string
  ) => {
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
              // Auto-fill from Product.minimumStock as default
              minimumStock: product?.minimumStock || "",
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
      // Pass available products to generate template with actual product codes
      const templateProducts = availableProducts.map((p) => ({
        code: p.code,
        name: p.name,
      }));
      const blob = await generateExcelTemplate(templateProducts);
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `Template_Stok_Awal_${selectedWarehouse?.name || "Gudang"}_${
        new Date().toISOString().split("T")[0]
      }.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch {
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
      if (
        isLoadingProducts ||
        isLoadingExistingStocks ||
        isFetchingExistingStocks
      ) {
        setErrors({
          upload:
            "Sedang memuat data produk dan stok. Mohon tunggu sebentar...",
        });
        setIsUploadingFile(false);
        return;
      }

      // 1. Parse Excel file
      const excelRows = await parseExcelFile(file);

      // 2. Validate data with existing stocks
      const result = validateExcelData(
        excelRows,
        productsData?.data || [],
        existingStocksData?.data || []
      );

      setValidationResult(result);

      // 3. If validation successful, convert to stock items
      if (result.success) {
        const newItems: StockItemRow[] = result.validItems.map(
          (item, index) => {
            const product = productsData?.data?.find(
              (p) => p.id === item.productId
            );
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
          }
        );
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
      setErrors({
        upload:
          error instanceof Error ? error.message : "Gagal membaca file Excel",
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
        } else if (
          validationResult &&
          validationResult.existingStocks &&
          validationResult.existingStocks.length > 0
        ) {
          // Check for existing stocks conflicts
          newErrors.existingStocks =
            "Tidak dapat melanjutkan - produk sudah memiliki stok di gudang";
        } else if (
          validationResult &&
          validationResult.duplicatesInFile &&
          validationResult.duplicatesInFile.length > 0
        ) {
          // Check for duplicates in file
          newErrors.duplicates =
            "Tidak dapat melanjutkan - ada produk duplikat dalam file";
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
    } catch (error: unknown) {
      const errorMessage =
        error &&
        typeof error === "object" &&
        "data" in error &&
        error.data &&
        typeof error.data === "object" &&
        "message" in error.data &&
        typeof error.data.message === "string"
          ? error.data.message
          : "Gagal menyimpan data stok awal";
      setErrors({
        submit: errorMessage,
      });
    }
  };

  // Step 1: Warehouse Selection - Enhanced Design
  const renderWarehouseSelection = () => (
    <Card>
      <CardHeader>
        <CardTitle>Pilih Gudang</CardTitle>
        <CardDescription>
          Pilih gudang untuk setup stok awal. Setiap gudang memiliki inventori terpisah.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="warehouse">
            Gudang <span className="text-destructive">*</span>
          </Label>
          {isLoadingWarehouses ? (
            <div className="flex items-center gap-2 h-10 px-3 border rounded-md bg-muted">
              <LoadingSpinner size="sm" />
              <span className="text-sm text-muted-foreground">Memuat gudang...</span>
            </div>
          ) : warehousesData?.data && warehousesData.data.length > 0 ? (
            <Select
              value={selectedWarehouseId}
              onValueChange={setSelectedWarehouseId}
            >
              <SelectTrigger
                id="warehouse"
                className={cn(
                  "w-full bg-background",
                  errors.warehouse && "border-destructive"
                )}
              >
                <SelectValue placeholder="Pilih gudang" />
              </SelectTrigger>
              <SelectContent>
                {warehousesData.data.map((warehouse) => {
                  const hasStock = statusData?.find(
                    (s) => s.warehouseId === warehouse.id
                  )?.hasInitialStock;

                  return (
                    <SelectItem key={warehouse.id} value={warehouse.id}>
                      <div className="flex items-center gap-2">
                        <Building2 className="h-4 w-4 text-muted-foreground" />
                        <span>{warehouse.name}</span>
                        {warehouse.code && (
                          <span className="text-muted-foreground">
                            ({warehouse.code})
                          </span>
                        )}
                        {hasStock && (
                          <Badge
                            variant="secondary"
                            className="text-xs bg-amber-100 text-amber-800 border-amber-300 ml-1"
                          >
                            Ada Stok
                          </Badge>
                        )}
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          ) : (
            <Alert className="border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/30">
              <AlertCircle className="h-4 w-4 text-amber-600" />
              <AlertDescription className="text-amber-800 dark:text-amber-200">
                Belum ada gudang aktif. Silakan buat gudang terlebih dahulu di menu{" "}
                <strong>Master → Gudang</strong>.
              </AlertDescription>
            </Alert>
          )}
          {errors.warehouse && (
            <p className="text-sm text-destructive">{errors.warehouse}</p>
          )}
        </div>

        {warehouseHasStock && selectedWarehouseId && (
          <Alert className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
            <AlertCircle className="h-4 w-4 text-blue-600" />
            <AlertDescription className="text-blue-800 dark:text-blue-200">
              Gudang ini sudah memiliki beberapa produk dengan stok. Setup stok
              awal hanya dapat dilakukan untuk produk yang{" "}
              <strong>belum pernah memiliki record stok</strong> di gudang ini.
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );

  // Step 2: Input Method Selection
  const renderInputMethodSelection = () => (
    <Card>
      <CardHeader>
        <CardTitle>Metode Input</CardTitle>
        <CardDescription>
          Pilih cara untuk memasukkan data stok awal. Anda bisa input manual atau import dari Excel.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          {/* Manual Entry */}
          <div
            className={`cursor-pointer rounded-lg border p-4 transition-all hover:shadow-md ${
              inputMethod === "manual"
                ? "border-primary bg-primary/5"
                : "border-border hover:border-primary/50"
            }`}
            onClick={() => setInputMethod("manual")}
          >
            <div className="flex items-center gap-3">
              <div
                className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors ${
                  inputMethod === "manual"
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted"
                }`}
              >
                <Edit3 className="h-5 w-5" />
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold">Input Manual</h3>
                <p className="text-sm text-muted-foreground">
                  Masukkan data produk satu per satu
                </p>
              </div>
              {inputMethod === "manual" && (
                <Check className="h-5 w-5 text-primary shrink-0" />
              )}
            </div>
          </div>

          {/* Excel Import */}
          <div
            className={`cursor-pointer rounded-lg border p-4 transition-all hover:shadow-md ${
              inputMethod === "excel"
                ? "border-primary bg-primary/5"
                : "border-border hover:border-primary/50"
            }`}
            onClick={() => setInputMethod("excel")}
          >
            <div className="flex items-center gap-3">
              <div
                className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors ${
                  inputMethod === "excel"
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted"
                }`}
              >
                <FileSpreadsheet className="h-5 w-5" />
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold">Import Excel</h3>
                <p className="text-sm text-muted-foreground">
                  Upload file Excel dengan template
                </p>
              </div>
              {inputMethod === "excel" && (
                <Check className="h-5 w-5 text-primary shrink-0" />
              )}
            </div>
          </div>
        </div>

        {errors.inputMethod && (
          <p className="text-sm text-red-500">{errors.inputMethod}</p>
        )}
      </CardContent>
    </Card>
  );

  // Step 3a: Manual Entry - Enhanced
  const renderManualEntry = () => (
    <Card>
      <CardHeader>
        <CardTitle>Input Manual</CardTitle>
        <CardDescription>
          Masukkan data stok untuk gudang: <strong>{selectedWarehouse?.name}</strong>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Warning if pagination might affect results */}
        {existingStocksData?.pagination?.hasMore && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Peringatan: Data Tidak Lengkap</AlertTitle>
            <AlertDescription>
              Gudang ini memiliki lebih dari{" "}
              {existingStocksData.pagination.totalItems} produk dengan stok.
              Beberapa produk mungkin tidak terfilter dengan benar. Silakan
              hubungi administrator.
            </AlertDescription>
          </Alert>
        )}

        {/* Info about available products */}
        <Alert className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
          <AlertCircle className="h-4 w-4 text-blue-600" />
          <AlertDescription className="text-blue-800 dark:text-blue-200">
            <strong>Produk Tersedia: {availableProducts.length} produk.</strong>{" "}
            Hanya produk yang belum pernah memiliki record stok di {selectedWarehouse?.name} yang dapat dipilih untuk setup stok awal.
            {existingStocksData?.data && existingStocksData.data.length > 0 &&
              ` (${existingStocksData.data.length} produk sudah memiliki record stok)`}
          </AlertDescription>
        </Alert>

        {/* Add Button - Enhanced */}
        <div className="flex justify-start">
          <Button
            onClick={handleAddRow}
            disabled={remainingAvailableProducts.length === 0}
          >
            <Plus className="mr-2 h-4 w-4" />
            Tambah Produk
          </Button>
          {remainingAvailableProducts.length === 0 && stockItems.length > 0 && (
            <span className="ml-3 text-sm text-muted-foreground self-center">
              Semua produk sudah dipilih
            </span>
          )}
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
                  <TableHead className="w-62.5">Produk *</TableHead>
                  <TableHead className="w-30">Quantity *</TableHead>
                  <TableHead className="w-37.5">Harga Beli *</TableHead>
                  <TableHead className="w-37.5">Lokasi</TableHead>
                  <TableHead className="w-30">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div className="flex items-center gap-1 cursor-help">
                            Min. Stok
                            <HelpCircle className="h-3 w-3 text-muted-foreground" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="top" className="max-w-xs">
                          <p className="text-xs">
                            <strong>Stok Minimum Gudang</strong>: Threshold untuk alert stok rendah di gudang ini.
                            Otomatis terisi dari default produk, dapat diubah sesuai kebutuhan gudang.
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableHead>
                  <TableHead className="w-30">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div className="flex items-center gap-1 cursor-help">
                            Max. Stok
                            <HelpCircle className="h-3 w-3 text-muted-foreground" />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="top" className="max-w-xs">
                          <p className="text-xs">
                            <strong>Stok Maksimum Gudang</strong>: Threshold untuk alert stok berlebih di gudang ini.
                            Opsional, dapat dikosongkan jika tidak diperlukan.
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableHead>
                  <TableHead className="w-50">Catatan</TableHead>
                  <TableHead className="w-20">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {stockItems.map((item, index) => (
                  <TableRow key={item.tempId}>
                    {/* Product Selection */}
                    <TableCell className="align-top">
                      <Select
                        value={item.productId}
                        onValueChange={(value) =>
                          handleUpdateRow(item.tempId, "productId", value)
                        }
                      >
                        <SelectTrigger
                          className={cn(
                            "w-full",
                            errors[`product-${index}`] && "border-red-500"
                          )}
                        >
                          <SelectValue placeholder="Pilih produk..." />
                        </SelectTrigger>
                        <SelectContent>
                          {isLoadingProducts ||
                          isLoadingExistingStocks ||
                          isFetchingExistingStocks ? (
                            <div className="p-2 text-sm text-muted-foreground">
                              Loading...
                            </div>
                          ) : availableProducts.length === 0 ? (
                            <div className="p-2 text-sm text-muted-foreground">
                              Semua produk sudah memiliki stok di gudang ini
                            </div>
                          ) : (
                            (() => {
                              // Filter out products already selected in other rows
                              // But keep the current row's selected product
                              const filteredProducts = availableProducts.filter(
                                (product) => {
                                  const isSelectedInOtherRow = stockItems.some(
                                    (otherItem) =>
                                      otherItem.tempId !== item.tempId &&
                                      otherItem.productId === product.id
                                  );
                                  return !isSelectedInOtherRow;
                                }
                              );

                              if (filteredProducts.length === 0) {
                                return (
                                  <div className="p-2 text-sm text-muted-foreground">
                                    Semua produk sudah dipilih
                                  </div>
                                );
                              }

                              return filteredProducts.map((product) => (
                                <SelectItem key={product.id} value={product.id}>
                                  {product.code} - {product.name}
                                </SelectItem>
                              ));
                            })()
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
                    <TableCell className="align-top">
                      <Input
                        type="number"
                        step="0.01"
                        value={item.quantity}
                        onChange={(e) =>
                          handleUpdateRow(
                            item.tempId,
                            "quantity",
                            e.target.value
                          )
                        }
                        placeholder="0"
                        className={
                          errors[`quantity-${index}`] ? "border-red-500" : ""
                        }
                      />
                      {item.baseUnit && (
                        <p className="text-xs text-muted-foreground mt-1">
                          {item.baseUnit}
                        </p>
                      )}
                      {errors[`quantity-${index}`] && (
                        <p className="text-xs text-red-500 mt-1">
                          {errors[`quantity-${index}`]}
                        </p>
                      )}
                    </TableCell>

                    {/* Cost Per Unit */}
                    <TableCell className="align-top">
                      <Input
                        type="number"
                        step="0.01"
                        value={item.costPerUnit}
                        onChange={(e) =>
                          handleUpdateRow(
                            item.tempId,
                            "costPerUnit",
                            e.target.value
                          )
                        }
                        placeholder="0"
                        className={
                          errors[`cost-${index}`] ? "border-red-500" : ""
                        }
                      />
                      {errors[`cost-${index}`] && (
                        <p className="text-xs text-red-500 mt-1">
                          {errors[`cost-${index}`]}
                        </p>
                      )}
                    </TableCell>

                    {/* Location */}
                    <TableCell className="align-top">
                      <Input
                        value={item.location || ""}
                        onChange={(e) =>
                          handleUpdateRow(
                            item.tempId,
                            "location",
                            e.target.value
                          )
                        }
                        placeholder="Rak A1"
                      />
                    </TableCell>

                    {/* Minimum Stock */}
                    <TableCell className="align-top">
                      <Input
                        type="number"
                        step="0.01"
                        value={item.minimumStock || ""}
                        onChange={(e) =>
                          handleUpdateRow(
                            item.tempId,
                            "minimumStock",
                            e.target.value
                          )
                        }
                        placeholder="0"
                      />
                    </TableCell>

                    {/* Maximum Stock */}
                    <TableCell className="align-top">
                      <Input
                        type="number"
                        step="0.01"
                        value={item.maximumStock || ""}
                        onChange={(e) =>
                          handleUpdateRow(
                            item.tempId,
                            "maximumStock",
                            e.target.value
                          )
                        }
                        placeholder="0"
                      />
                    </TableCell>

                    {/* Notes */}
                    <TableCell className="align-top">
                      <Input
                        value={item.notes || ""}
                        onChange={(e) =>
                          handleUpdateRow(item.tempId, "notes", e.target.value)
                        }
                        placeholder="Catatan..."
                      />
                    </TableCell>

                    {/* Actions */}
                    <TableCell className="align-top">
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
    <Card>
      <CardHeader>
        <CardTitle>Import Excel</CardTitle>
        <CardDescription>
          Upload file Excel untuk gudang: <strong>{selectedWarehouse?.name}</strong>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Download Template */}
        <Alert className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
          <Download className="h-4 w-4 text-blue-600" />
          <AlertDescription className="text-blue-800 dark:text-blue-200">
            Download template Excel terlebih dahulu dan isi sesuai format yang disediakan.{" "}
            <Button
              variant="link"
              className="p-0 h-auto text-blue-600 hover:text-blue-700 font-medium"
              onClick={handleDownloadTemplate}
            >
              Download Template
            </Button>
          </AlertDescription>
        </Alert>

        {/* Upload Area */}
        <div className="border border-dashed border-muted-foreground/25 rounded-lg p-6 hover:border-primary/50 transition-colors">
          <div className="flex items-center gap-4">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-muted">
              <Upload className="h-6 w-6 text-muted-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold">Upload File Excel</h3>
              <p className="text-sm text-muted-foreground">
                Format: .xlsx, .xls • Maksimal: 5MB
              </p>
            </div>
            <input
              type="file"
              id="excel-upload"
              accept=".xlsx,.xls"
              className="hidden"
              onChange={handleFileInputChange}
              disabled={
                isUploadingFile ||
                isLoadingExistingStocks ||
                isFetchingExistingStocks
              }
            />
            <Button
              onClick={() => document.getElementById("excel-upload")?.click()}
              disabled={
                isUploadingFile ||
                isLoadingExistingStocks ||
                isFetchingExistingStocks
              }
            >
              {isUploadingFile ? (
                <>
                  <LoadingSpinner className="mr-2 h-4 w-4" />
                  Memproses...
                </>
              ) : isLoadingExistingStocks || isFetchingExistingStocks ? (
                <>
                  <LoadingSpinner className="mr-2 h-4 w-4" />
                  Memuat...
                </>
              ) : (
                <>
                  <Upload className="mr-2 h-4 w-4" />
                  Pilih File
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Uploaded File Info */}
        {uploadedFile && (
          <Alert className="border-green-200 bg-green-50/50 dark:border-green-800 dark:bg-green-950/30">
            <FileSpreadsheet className="h-4 w-4 text-green-600" />
            <AlertDescription className="text-green-800 dark:text-green-200">
              <strong>File terpilih:</strong> {uploadedFile.name} ({(uploadedFile.size / 1024).toFixed(2)} KB)
            </AlertDescription>
          </Alert>
        )}

        {/* Upload Errors - Enhanced */}
        {errors.upload && (
          <Alert variant="destructive" className="border-2">
            <AlertCircle className="h-5 w-5" />
            <AlertTitle className="font-semibold">Error Upload</AlertTitle>
            <AlertDescription className="text-base">
              {errors.upload}
            </AlertDescription>
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
                  Produk Duplikat dalam File (
                  {validationResult.duplicatesInFile.length})
                </AlertTitle>
                <AlertDescription>
                  <div className="mt-2 space-y-1">
                    {validationResult.duplicatesInFile
                      .slice(0, 5)
                      .map((conflict, idx) => (
                        <div key={idx} className="text-sm">
                          • Baris {conflict.row}: {conflict.productCode} -{" "}
                          {conflict.productName}
                        </div>
                      ))}
                    {validationResult.duplicatesInFile.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.duplicatesInFile.length - 5}{" "}
                        lainnya
                      </div>
                    )}
                  </div>
                </AlertDescription>
              </Alert>
            )}

            {/* Info: Products without stock (valid for initial stock input) */}
            {validationResult.noStockProducts &&
              validationResult.noStockProducts.length > 0 && (
                <Alert className="border-blue-200 bg-blue-50 dark:border-blue-900 dark:bg-blue-950">
                  <AlertCircle className="h-4 w-4 text-blue-600" />
                  <AlertTitle className="text-blue-900 dark:text-blue-100">
                    Produk Belum Memiliki Stok (
                    {validationResult.noStockProducts.length})
                  </AlertTitle>
                  <AlertDescription className="text-blue-800 dark:text-blue-200">
                    <p className="mb-2 text-sm">
                      Produk berikut{" "}
                      <strong>belum pernah memiliki record stok</strong> di{" "}
                      <strong>{selectedWarehouse?.name}</strong> dan siap untuk
                      input stok awal:
                    </p>
                    <div className="mt-2 space-y-1">
                      {validationResult.noStockProducts
                        .slice(0, 5)
                        .map((product, idx) => (
                          <div key={idx} className="text-sm">
                            • Baris {product.row}:{" "}
                            <strong>{product.productCode}</strong> -{" "}
                            {product.productName}
                            <span className="ml-2 text-xs">
                              ({product.quantity} unit)
                            </span>
                          </div>
                        ))}
                      {validationResult.noStockProducts.length > 5 && (
                        <div className="text-sm font-semibold">
                          ... dan {validationResult.noStockProducts.length - 5}{" "}
                          produk lainnya
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
                  Produk Sudah Memiliki Stok (
                  {validationResult.existingStocks.length})
                </AlertTitle>
                <AlertDescription>
                  <div className="mt-2 space-y-1">
                    {validationResult.existingStocks
                      .slice(0, 5)
                      .map((conflict, idx) => (
                        <div key={idx} className="text-sm">
                          • Baris {conflict.row}: {conflict.productCode} -{" "}
                          {conflict.productName}
                          <br />
                          <span className="ml-4 text-xs">
                            Stok saat ini: {conflict.currentQuantity} | Stok
                            baru: {conflict.newQuantity}
                          </span>
                        </div>
                      ))}
                    {validationResult.existingStocks.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.existingStocks.length - 5}{" "}
                        lainnya
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
                        • Baris {error.row}, Kolom {error.field}:{" "}
                        {error.message}
                      </div>
                    ))}
                    {validationResult.errors.length > 5 && (
                      <div className="text-sm font-semibold">
                        ... dan {validationResult.errors.length - 5} error
                        lainnya
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
              {validationResult.validItems.length} produk berhasil divalidasi
              dan siap untuk disimpan.
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
      <Card>
        <CardHeader>
          <CardTitle>Review & Validasi</CardTitle>
          <CardDescription>
            Periksa kembali data sebelum menyimpan ke sistem
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Summary Info - 2 Columns */}
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div className="flex items-start gap-2">
              <Building2 className="h-4 w-4 text-muted-foreground mt-0.5" />
              <div>
                <span className="text-muted-foreground">Gudang</span>
                <p className="font-medium">{selectedWarehouse?.name} ({selectedWarehouse?.code})</p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              {inputMethod === "manual" ? (
                <Edit3 className="h-4 w-4 text-muted-foreground mt-0.5" />
              ) : (
                <FileSpreadsheet className="h-4 w-4 text-muted-foreground mt-0.5" />
              )}
              <div>
                <span className="text-muted-foreground">Metode</span>
                <p className="font-medium">
                  {inputMethod === "manual" ? "Input Manual" : "Import Excel"}
                  {inputMethod === "excel" && uploadedFile && (
                    <span className="text-muted-foreground font-normal"> ({uploadedFile.name})</span>
                  )}
                </p>
              </div>
            </div>
          </div>

          {/* Items List */}
          <div className="space-y-3">
            <Label className="text-base font-semibold">
              Daftar Produk
            </Label>
            <div className="rounded-md border overflow-hidden">
              {/* Scrollable table container for many items */}
              <div className={cn(
                "overflow-x-auto",
                stockItems.length > 10 && "max-h-[500px] overflow-y-auto"
              )}>
                <Table>
                  <TableHeader className="sticky top-0 z-10 bg-background">
                    <TableRow>
                      <TableHead className="w-12 text-center">No.</TableHead>
                      <TableHead>Produk</TableHead>
                      <TableHead className="text-right">Quantity</TableHead>
                      <TableHead className="text-right">Harga Beli</TableHead>
                      <TableHead className="text-right">Subtotal</TableHead>
                      <TableHead>Lokasi</TableHead>
                      <TableHead className="text-center">Min</TableHead>
                      <TableHead className="text-center">Max</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {stockItems.map((item, index) => {
                      const subtotal =
                        parseFloat(item.quantity || "0") *
                        parseFloat(item.costPerUnit || "0");
                      return (
                        <TableRow key={item.tempId}>
                          {/* Row Number */}
                          <TableCell className="text-center font-medium text-muted-foreground">
                            {index + 1}
                          </TableCell>
                          {/* Product */}
                          <TableCell>
                            <div>
                              <p className="font-medium">{item.productName}</p>
                              <p className="text-xs text-muted-foreground">{item.productCode}</p>
                            </div>
                          </TableCell>
                          {/* Quantity */}
                          <TableCell className="text-right">
                            {parseFloat(item.quantity).toLocaleString("id-ID")}
                            {item.baseUnit && (
                              <span className="text-xs text-muted-foreground ml-1">
                                {item.baseUnit}
                              </span>
                            )}
                          </TableCell>
                          {/* Cost */}
                          <TableCell className="text-right">
                            Rp {parseFloat(item.costPerUnit).toLocaleString("id-ID")}
                          </TableCell>
                          {/* Subtotal */}
                          <TableCell className="text-right font-medium">
                            Rp {subtotal.toLocaleString("id-ID")}
                          </TableCell>
                          {/* Location */}
                          <TableCell className="text-muted-foreground">
                            {item.location || "-"}
                          </TableCell>
                          {/* Min Stock */}
                          <TableCell className="text-center text-muted-foreground">
                            {item.minimumStock
                              ? parseFloat(item.minimumStock).toLocaleString("id-ID")
                              : "-"}
                          </TableCell>
                          {/* Max Stock */}
                          <TableCell className="text-center text-muted-foreground">
                            {item.maximumStock
                              ? parseFloat(item.maximumStock).toLocaleString("id-ID")
                              : "-"}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                  {/* Table Footer with Totals */}
                  <tfoot className="bg-muted/50 border-t sticky bottom-0">
                    <tr>
                      <td colSpan={2} className="px-4 py-2 font-semibold text-right text-sm">
                        TOTAL
                      </td>
                      <td className="px-4 py-2 text-right text-sm font-semibold">
                        {totalQuantity.toLocaleString("id-ID")}
                      </td>
                      <td className="px-4 py-2 text-right text-muted-foreground text-sm">
                        -
                      </td>
                      <td className="px-4 py-2 text-right text-sm font-semibold text-green-600">
                        Rp {totalValue.toLocaleString("id-ID")}
                      </td>
                      <td colSpan={3} className="px-4 py-2"></td>
                    </tr>
                  </tfoot>
                </Table>
              </div>
            </div>
          </div>

          {/* Notes */}
          {notes && (
            <div className="space-y-2">
              <Label className="text-sm font-medium">Catatan</Label>
              <p className="text-sm text-muted-foreground bg-muted p-3 rounded-md">{notes}</p>
            </div>
          )}

          {/* Submit Error */}
          {errors.submit && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Terjadi Kesalahan</AlertTitle>
              <AlertDescription>{errors.submit}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    );
  };

  // Step 5: Success - Simple Celebration
  const renderSuccess = () => (
    <Card className="border-green-200 shadow-lg">
      <CardContent className="p-8">
        <div className="flex flex-col items-center text-center space-y-6">
          {/* Success Icon */}
          <div className="rounded-full bg-green-100 p-4 dark:bg-green-900/30">
            <CheckCircle2 className="h-12 w-12 text-green-600" />
          </div>

          {/* Success Message */}
          <div className="space-y-2">
            <h2 className="text-2xl font-bold text-green-700">
              Setup Stok Awal Berhasil!
            </h2>
            <p className="text-muted-foreground">
              <strong>{stockItems.length} produk</strong> berhasil ditambahkan ke gudang{" "}
              <strong>{selectedWarehouse?.name}</strong>
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 pt-2">
            <Button
              onClick={() => router.push("/inventory/stock")}
              className="bg-green-600 hover:bg-green-700"
            >
              <Package className="mr-2 h-4 w-4" />
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
            >
              <Building2 className="mr-2 h-4 w-4" />
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

  // Step definitions for the wizard
  const steps = [
    { number: 1, icon: Building2, label: "Gudang" },
    { number: 2, icon: Edit3, label: "Metode" },
    { number: 3, icon: FileSpreadsheet, label: "Input" },
    { number: 4, icon: ClipboardCheck, label: "Review" },
  ];

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page Header - Simple Style */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Setup Stok Awal</h1>
          <p className="text-muted-foreground">
            Setup stok pertama kali untuk produk yang belum memiliki record di gudang
          </p>
        </div>
      </div>

      {/* Step Indicator - Line Connector Style */}
      {currentStep !== 5 && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-center gap-2">
              {steps.map((step) => {
                const StepIcon = step.icon;
                return (
                  <div key={step.number} className="flex items-center gap-2">
                    <div className="flex flex-col items-center gap-1.5">
                      <div
                        className={cn(
                          "flex h-10 w-10 items-center justify-center rounded-full transition-colors",
                          step.number === currentStep
                            ? "bg-primary text-primary-foreground"
                            : step.number < currentStep
                            ? "bg-green-500 text-white"
                            : "bg-muted text-muted-foreground"
                        )}
                      >
                        {step.number < currentStep ? (
                          <Check className="h-5 w-5" />
                        ) : (
                          <StepIcon className="h-5 w-5" />
                        )}
                      </div>
                      <span
                        className={cn(
                          "text-xs font-medium transition-colors",
                          step.number === currentStep
                            ? "text-primary"
                            : step.number < currentStep
                            ? "text-green-600 dark:text-green-500"
                            : "text-muted-foreground"
                        )}
                      >
                        {step.label}
                      </span>
                    </div>
                    {step.number < 4 && (
                      <div
                        className={cn(
                          "h-0.5 w-8 sm:w-12 transition-colors mb-4",
                          step.number < currentStep ? "bg-green-500" : "bg-muted"
                        )}
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step Content */}
      {renderStepContent()}

      {/* Navigation Buttons */}
      {currentStep !== 5 && (
        <div className="flex justify-between items-center">
          <Button
            variant="outline"
            onClick={handleBack}
            disabled={currentStep === 1 || isSubmitting}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>

          {currentStep < 4 ? (
            <Button
              onClick={handleNext}
              disabled={isSubmitting || hasValidationConflicts}
            >
              Selanjutnya
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          ) : (
            <Button
              onClick={handleSubmit}
              disabled={isSubmitting}
              className="bg-green-600 hover:bg-green-700"
            >
              {isSubmitting ? (
                <>
                  <LoadingSpinner size="sm" className="mr-2" />
                  Menyimpan...
                </>
              ) : (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  Simpan Setup Stok Awal
                </>
              )}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
