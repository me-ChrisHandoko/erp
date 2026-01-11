/**
 * Create Transfer Dialog Component
 *
 * Multi-step wizard for creating stock transfers:
 * - Step 1: Select warehouses (source ≠ destination) and date
 * - Step 2: Select products and quantities with stock validation
 * - Step 3: Review and submit
 */

"use client";

import React, { useState } from "react";
import { useSelector } from "react-redux";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import {
  ArrowRight,
  ArrowLeft,
  ArrowUpDown,
  Calendar as CalendarIcon,
  Check,
  Warehouse,
  Package,
  Trash2,
  AlertCircle,
  Loader2,
  Building2,
  ShoppingCart,
  ClipboardCheck,
} from "lucide-react";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { useListProductsQuery } from "@/store/services/productApi";
import { useListStocksQuery } from "@/store/services/stockApi";
import {
  useCreateTransferMutation,
  useUpdateTransferMutation,
} from "@/store/services/transferApi";
import type { RootState } from "@/store";
import type { CreateTransferRequest, StockTransfer } from "@/types/transfer.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface CreateTransferDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  transferToEdit?: StockTransfer | null; // Optional: Transfer to edit (DRAFT only)
}

interface TransferItem {
  productId: string;
  productCode: string;
  productName: string;
  quantity: string;
  batchId?: string;
  notes?: string;
  availableStock?: number;
}

export function CreateTransferDialog({
  open,
  onOpenChange,
  onSuccess,
  transferToEdit,
}: CreateTransferDialogProps) {
  // Determine if we're in edit mode
  const isEditMode = !!transferToEdit;

  // Current step (1-3)
  const [currentStep, setCurrentStep] = useState(1);

  // Step 1: Warehouse Selection
  const [sourceWarehouseId, setSourceWarehouseId] = useState<string>("");
  const [destWarehouseId, setDestWarehouseId] = useState<string>("");
  const [transferDate, setTransferDate] = useState<Date>(new Date());
  const [notes, setNotes] = useState<string>("");

  // Step 2: Product Selection
  const [items, setItems] = useState<TransferItem[]>([]);
  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [quantity, setQuantity] = useState<string>("");
  const [itemNotes, setItemNotes] = useState<string>("");

  // Validation errors
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Track if currently swapping warehouses to prevent unwanted resets
  const isSwappingRef = React.useRef(false);

  // Track if currently in pre-fill mode to prevent unwanted resets
  const isPrefillingRef = React.useRef(false);

  // Get active company
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Fetch warehouses
  const { data: warehousesData, isLoading: loadingWarehouses } =
    useListWarehousesQuery(
      { page: 1, pageSize: 100 },
      { skip: !activeCompanyId }
    );

  // Fetch stocks from source warehouse to get available products
  // Note: Don't send zeroStock=false because backend doesn't filter when false
  // We'll filter zero stock on frontend instead
  const { data: stocksData, isLoading: loadingStocks } = useListStocksQuery(
    {
      page: 1,
      pageSize: 100, // Max allowed by backend validation
      warehouseID: sourceWarehouseId,
      // Don't send zeroStock parameter - we'll filter client-side
    },
    { skip: !activeCompanyId || currentStep !== 2 || !sourceWarehouseId }
  );

  // Extract productIDs that have stock in source warehouse (exclude zero stock)
  const availableProductIds = React.useMemo(() => {
    if (!stocksData?.data) {
      console.log("🔍 [Stock Filter] No stocks data available");
      return new Set<string>();
    }

    console.log("🔍 [Stock Filter] Raw stocks data:", stocksData.data.length, "items");

    // Filter out products with zero stock
    const filtered = stocksData.data.filter((stock) => {
      const qty = parseFloat(stock.quantity);
      return qty > 0;
    });

    console.log("✅ [Stock Filter] Products with stock > 0:", filtered.length);

    return new Set(filtered.map((stock) => stock.productID));
  }, [stocksData]);

  // Fetch all products for details
  const { data: allProductsData, isLoading: loadingProducts } =
    useListProductsQuery(
      {
        page: 1,
        pageSize: 100, // Max allowed by backend validation
        isActive: true,
      },
      { skip: !activeCompanyId || currentStep !== 2 }
    );

  // Filter products that have stock in source warehouse
  const productsData = React.useMemo(() => {
    // If no products data yet, return undefined
    if (!allProductsData) {
      console.log("🔍 [Product Filter] No products data available");
      return undefined;
    }

    console.log("🔍 [Product Filter] All products:", allProductsData.data.length);

    // If source warehouse not selected yet, return all products
    if (!sourceWarehouseId) {
      console.log("⚠️ [Product Filter] No source warehouse selected, showing all products");
      return allProductsData;
    }

    // If no stocks data yet or no available products, return empty data
    if (availableProductIds.size === 0) {
      console.log("⚠️ [Product Filter] No available products with stock, returning empty");
      return {
        ...allProductsData,
        data: [],
      };
    }

    // Filter products that have stock in source warehouse
    const filtered = allProductsData.data.filter((product) =>
      availableProductIds.has(product.id)
    );

    console.log("✅ [Product Filter] Filtered products:", filtered.length);
    console.log("✅ [Product Filter] Product codes:", filtered.map(p => p.code).join(", "));

    return {
      ...allProductsData,
      data: filtered,
    };
  }, [allProductsData, availableProductIds, sourceWarehouseId]);

  // Create/Update transfer mutations
  const [createTransfer, { isLoading: isCreating }] =
    useCreateTransferMutation();
  const [updateTransfer, { isLoading: isUpdating }] =
    useUpdateTransferMutation();

  // Reset form
  const resetForm = () => {
    setCurrentStep(1);
    setSourceWarehouseId("");
    setDestWarehouseId("");
    setTransferDate(new Date());
    setNotes("");
    setItems([]);
    setSelectedProductId("");
    setQuantity("");
    setItemNotes("");
    setErrors({});
  };

  // Reset form when modal is closed OR pre-fill when in edit mode
  React.useEffect(() => {
    if (!open) {
      // Reset form when modal closes
      resetForm();
      isPrefillingRef.current = false;
    } else if (open && transferToEdit) {
      // Pre-fill form when in edit mode
      console.log("📝 [EditMode] Pre-filling form with transfer:", transferToEdit);

      // Set pre-filling flag to prevent warehouse change effect from resetting
      isPrefillingRef.current = true;

      // Step 1: Warehouse selection
      setSourceWarehouseId(transferToEdit.sourceWarehouseId || "");
      setDestWarehouseId(transferToEdit.destWarehouseId || "");
      setTransferDate(new Date(transferToEdit.transferDate));
      setNotes(transferToEdit.notes || "");

      // Step 2: Items
      const editItems: TransferItem[] = transferToEdit.items.map((item) => ({
        productId: item.productId,
        productCode: item.product?.code || "",
        productName: item.product?.name || "",
        quantity: item.quantity,
        batchId: item.batchId,
        notes: item.notes,
      }));
      setItems(editItems);

      console.log("✅ [EditMode] Form pre-filled with items:", editItems);

      // Reset pre-filling flag after a short delay to allow all state updates to complete
      setTimeout(() => {
        isPrefillingRef.current = false;
      }, 100);
    } else if (open && !transferToEdit) {
      // Ensure form is reset when opening in create mode
      console.log("📝 [CreateMode] Resetting form for new transfer");
      resetForm();
    }
  }, [open, transferToEdit]);

  // Reset destination warehouse when source warehouse changes
  React.useEffect(() => {
    // Skip reset if currently swapping warehouses
    if (isSwappingRef.current) {
      isSwappingRef.current = false;
      return;
    }

    // Skip reset if currently pre-filling form in edit mode
    if (isPrefillingRef.current) {
      return;
    }

    // Clear destination warehouse and errors when source warehouse changes
    if (destWarehouseId) {
      setDestWarehouseId("");
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors.destWarehouse;
        return newErrors;
      });
    }
  }, [sourceWarehouseId]);

  // Validate Step 1
  const validateStep1 = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!sourceWarehouseId) {
      newErrors.sourceWarehouse = "Pilih gudang asal";
    }
    if (!destWarehouseId) {
      newErrors.destWarehouse = "Pilih gudang tujuan";
    }
    if (sourceWarehouseId && destWarehouseId && sourceWarehouseId === destWarehouseId) {
      newErrors.destWarehouse = "Gudang tujuan harus berbeda dengan gudang asal";
    }
    if (!transferDate) {
      newErrors.transferDate = "Pilih tanggal transfer";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Validate Step 2
  const validateStep2 = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (items.length === 0) {
      newErrors.items = "Tambahkan minimal 1 produk";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Add product to items
  const handleAddProduct = () => {
    const newErrors: Record<string, string> = {};

    if (!selectedProductId) {
      newErrors.product = "Pilih produk";
    }
    if (!quantity || parseFloat(quantity) <= 0) {
      newErrors.quantity = "Masukkan jumlah yang valid";
    }

    // Find product details
    const product = productsData?.data.find((p) => p.id === selectedProductId);
    if (!product && selectedProductId) {
      newErrors.product = "Produk tidak ditemukan";
    }

    // Get available stock
    const stock = stocksData?.data.find((s) => s.productID === selectedProductId);
    if (selectedProductId && !stock) {
      newErrors.product = "Stok produk tidak ditemukan";
    }

    // Validate quantity against available stock
    if (stock && quantity && parseFloat(quantity) > parseFloat(stock.quantity)) {
      newErrors.quantity = `Jumlah melebihi stok tersedia (${parseFloat(stock.quantity).toLocaleString("id-ID", {
        minimumFractionDigits: 0,
        maximumFractionDigits: 3,
      })})`;
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    if (!product) return;

    // Check if product already in list
    if (items.some((item) => item.productId === selectedProductId)) {
      setErrors({ product: "Produk sudah ditambahkan" });
      return;
    }

    // Add to items with available stock info
    const newItem: TransferItem = {
      productId: selectedProductId,
      productCode: product.code,
      productName: product.name,
      quantity,
      notes: itemNotes || undefined,
      availableStock: stock ? parseFloat(stock.quantity) : undefined,
    };

    setItems([...items, newItem]);

    // Clear form
    setSelectedProductId("");
    setQuantity("");
    setItemNotes("");
    setErrors({});
  };

  // Remove product from items
  const handleRemoveProduct = (index: number) => {
    setItems(items.filter((_, i) => i !== index));
  };

  // Navigate steps
  const handleNext = () => {
    if (currentStep === 1 && validateStep1()) {
      setCurrentStep(2);
    } else if (currentStep === 2 && validateStep2()) {
      setCurrentStep(3);
    }
  };

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
      setErrors({});
    }
  };

  // Swap warehouses
  const handleSwapWarehouses = () => {
    if (sourceWarehouseId && destWarehouseId) {
      // Set flag to prevent useEffect from resetting destination warehouse
      isSwappingRef.current = true;

      // Swap the warehouse IDs
      const temp = sourceWarehouseId;
      setSourceWarehouseId(destWarehouseId);
      setDestWarehouseId(temp);

      // Clear any related errors
      setErrors({});
    }
  };

  // Submit transfer
  const handleSubmit = async () => {
    try {
      const requestData: CreateTransferRequest = {
        transferDate: format(transferDate, "yyyy-MM-dd"),
        sourceWarehouseId,
        destWarehouseId,
        notes: notes || undefined,
        items: items.map((item) => ({
          productId: item.productId,
          batchId: item.batchId,
          quantity: item.quantity,
          notes: item.notes,
        })),
      };

      console.log(`📤 [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Request data:`, JSON.stringify(requestData, null, 2));
      console.log(`📤 [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Items detail:`, items);
      console.log(`📤 [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Source/Dest:`, {
        source: sourceWarehouseId,
        dest: destWarehouseId,
        date: transferDate,
        notes: notes
      });

      let result;
      if (isEditMode && transferToEdit) {
        // Update existing transfer
        result = await updateTransfer({
          id: transferToEdit.id,
          data: requestData,
        }).unwrap();
        console.log("✅ [UpdateTransfer] Success");
      } else {
        // Create new transfer
        result = await createTransfer(requestData).unwrap();
        console.log("✅ [CreateTransfer] Success");
      }

      // Show success toast
      toast.success(isEditMode ? "✓ Transfer Berhasil Diperbarui" : "✓ Transfer Berhasil Dibuat", {
        description: isEditMode
          ? `Transfer ${result.transferNumber} telah diperbarui.`
          : `Transfer ${result.transferNumber} telah dibuat dan siap untuk diproses.`,
      });

      // Success - call onSuccess which will close modal and refetch
      // The resetForm will be automatically handled by useEffect when modal closes
      if (onSuccess) {
        onSuccess();
      } else {
        onOpenChange(false);
      }
    } catch (error: any) {
      console.error(`❌ [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Error:`, error);
      console.error(`❌ [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Error data:`, error?.data);
      console.error(`❌ [${isEditMode ? "UpdateTransfer" : "CreateTransfer"}] Error status:`, error?.status);

      // Show error toast
      toast.error(isEditMode ? "Gagal Memperbarui Transfer" : "Gagal Membuat Transfer", {
        description: error?.data?.message || (isEditMode
          ? "Terjadi kesalahan saat memperbarui transfer. Silakan coba lagi."
          : "Terjadi kesalahan saat membuat transfer. Silakan coba lagi."),
      });

      setErrors({
        submit: error?.data?.message || (isEditMode ? "Gagal memperbarui transfer" : "Gagal membuat transfer"),
      });
    }
  };

  // Get warehouse name
  const getWarehouseName = (id: string) => {
    return warehousesData?.data.find((w) => w.id === id)?.name || "-";
  };

  // Calculate total quantity
  const totalQuantity = items.reduce(
    (sum, item) => sum + parseFloat(item.quantity || "0"),
    0
  );

  // Convert products to ComboboxOption[]
  const productOptions: ComboboxOption[] = React.useMemo(() => {
    if (!productsData?.data) return [];
    return productsData.data.map((product) => ({
      value: product.id,
      label: product.name,
      searchLabel: `${product.code} ${product.name}`, // For better search
    }));
  }, [productsData]);

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen && document.activeElement instanceof HTMLElement) {
          // Blur any focused element to prevent aria-hidden warning
          document.activeElement.blur();
        }
        onOpenChange(isOpen);
      }}
    >
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            {isEditMode ? "Edit Transfer Gudang" : "Buat Transfer Gudang Baru"}
          </DialogTitle>
          <DialogDescription>
            Langkah {currentStep} dari 3: {
              currentStep === 1 ? "Pilih Gudang & Tanggal" :
              currentStep === 2 ? "Pilih Produk & Jumlah" :
              "Review & Konfirmasi"
            }
          </DialogDescription>
        </DialogHeader>

        {/* Step Indicator */}
        <div className="flex items-center justify-center gap-2 py-4">
          {[
            { number: 1, icon: Building2, label: "Gudang" },
            { number: 2, icon: ShoppingCart, label: "Produk" },
            { number: 3, icon: ClipboardCheck, label: "Review" },
          ].map((step) => {
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
                {step.number < 3 && (
                  <div
                    className={cn(
                      "h-0.5 w-12 transition-colors mb-4",
                      step.number < currentStep ? "bg-green-500" : "bg-muted"
                    )}
                  />
                )}
              </div>
            );
          })}
        </div>

        <Separator />

        {/* Step 1: Warehouse Selection */}
        {currentStep === 1 && (
          <div className="space-y-4">
            {/* Source Warehouse */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="sourceWarehouse">
                  Dari Gudang <span className="text-destructive">*</span>
                </Label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleSwapWarehouses}
                  disabled={!sourceWarehouseId || !destWarehouseId}
                  className="h-8 gap-1.5 text-xs text-muted-foreground hover:text-foreground"
                  title="Tukar gudang asal dan tujuan"
                >
                  <ArrowUpDown className="h-3.5 w-3.5" />
                  Tukar
                </Button>
              </div>
              <Select
                value={sourceWarehouseId}
                onValueChange={(value) => {
                  setSourceWarehouseId(value);
                  setErrors({});
                }}
              >
                <SelectTrigger
                  id="sourceWarehouse"
                  className={cn("w-full", errors.sourceWarehouse && "border-destructive")}
                >
                  <SelectValue placeholder="Pilih gudang asal" />
                </SelectTrigger>
                <SelectContent position="popper" sideOffset={4}>
                  {loadingWarehouses ? (
                    <SelectItem value="loading" disabled>
                      Memuat...
                    </SelectItem>
                  ) : (
                    warehousesData?.data.map((warehouse) => (
                      <SelectItem key={warehouse.id} value={warehouse.id}>
                        <div className="flex items-center gap-2">
                          <Warehouse className="h-4 w-4" />
                          {warehouse.name}
                          {warehouse.code && (
                            <span className="text-xs text-muted-foreground">
                              ({warehouse.code})
                            </span>
                          )}
                        </div>
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
              {errors.sourceWarehouse && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.sourceWarehouse}
                </p>
              )}
            </div>

            {/* Destination Warehouse */}
            <div className="space-y-2">
              <Label htmlFor="destWarehouse">
                Ke Gudang <span className="text-destructive">*</span>
              </Label>
              <Select
                value={destWarehouseId}
                onValueChange={(value) => {
                  setDestWarehouseId(value);
                  setErrors({});
                }}
              >
                <SelectTrigger
                  id="destWarehouse"
                  className={cn("w-full", errors.destWarehouse && "border-destructive")}
                >
                  <SelectValue placeholder="Pilih gudang tujuan" />
                </SelectTrigger>
                <SelectContent position="popper" sideOffset={4}>
                  {loadingWarehouses ? (
                    <SelectItem value="loading" disabled>
                      Memuat...
                    </SelectItem>
                  ) : (
                    warehousesData?.data
                      .filter((w) => w.id !== sourceWarehouseId) // Exclude source warehouse
                      .map((warehouse) => (
                        <SelectItem key={warehouse.id} value={warehouse.id}>
                          <div className="flex items-center gap-2">
                            <Warehouse className="h-4 w-4" />
                            {warehouse.name}
                            {warehouse.code && (
                              <span className="text-xs text-muted-foreground">
                                ({warehouse.code})
                              </span>
                            )}
                          </div>
                        </SelectItem>
                      ))
                  )}
                </SelectContent>
              </Select>
              {errors.destWarehouse && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.destWarehouse}
                </p>
              )}
            </div>

            {/* Transfer Date */}
            <div className="space-y-2">
              <Label>
                Tanggal Transfer <span className="text-destructive">*</span>
              </Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !transferDate && "text-muted-foreground",
                      errors.transferDate && "border-destructive"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {transferDate ? (
                      format(transferDate, "dd MMMM yyyy", { locale: localeId })
                    ) : (
                      <span>Pilih tanggal</span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={transferDate}
                    onSelect={(date) => {
                      if (date) {
                        setTransferDate(date);
                        setErrors({});
                      }
                    }}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
              {errors.transferDate && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.transferDate}
                </p>
              )}
            </div>

            {/* Notes */}
            <div className="space-y-2">
              <Label htmlFor="notes">Catatan (Opsional)</Label>
              <Textarea
                id="notes"
                placeholder="Tambahkan catatan untuk transfer ini..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
              />
            </div>
          </div>
        )}

        {/* Step 2: Product Selection */}
        {currentStep === 2 && (
          <div className="space-y-4">
            {/* Product Selection Form */}
            <div className="rounded-lg border p-4 space-y-4 bg-muted/30">
              <h3 className="font-semibold text-sm">Tambah Produk</h3>

              {/* Product Combobox */}
              <div className="space-y-2">
                <Label htmlFor="product">
                  Produk <span className="text-destructive">*</span>
                </Label>
                <Combobox
                  options={productOptions}
                  value={selectedProductId}
                  onValueChange={(value) => {
                    setSelectedProductId(value);
                    setErrors({});
                  }}
                  placeholder="Pilih produk..."
                  searchPlaceholder="Cari produk..."
                  emptyMessage={
                    loadingProducts || loadingStocks
                      ? "Memuat..."
                      : !sourceWarehouseId
                      ? "Pilih gudang asal terlebih dahulu"
                      : availableProductIds.size === 0
                      ? "Tidak ada produk dengan stok di gudang ini"
                      : "Produk tidak ditemukan"
                  }
                  disabled={loadingProducts || loadingStocks || !sourceWarehouseId}
                  className={cn(errors.product && "border-destructive")}
                  renderOption={(option, selected) => {
                    const product = productsData?.data.find(
                      (p) => p.id === option.value
                    );
                    if (!product) return null;

                    // Get stock quantity for this product
                    const stock = stocksData?.data.find(
                      (s) => s.productID === product.id
                    );

                    return (
                      <>
                        <Check
                          className={cn(
                            "mr-2 h-4 w-4",
                            selected ? "opacity-100" : "opacity-0"
                          )}
                        />
                        <div className="flex items-center justify-between flex-1 gap-2">
                          <div className="flex items-center gap-2">
                            <Package className="h-4 w-4 text-muted-foreground" />
                            <span className="font-mono text-xs">
                              {product.code}
                            </span>
                            {" - "}
                            <span>{product.name}</span>
                          </div>
                          {stock && (
                            <span className="text-xs text-muted-foreground ml-auto">
                              Stok: {parseFloat(stock.quantity).toLocaleString("id-ID", {
                                minimumFractionDigits: 0,
                                maximumFractionDigits: 3,
                              })}
                            </span>
                          )}
                        </div>
                      </>
                    );
                  }}
                  renderTrigger={(selectedOption) => {
                    if (!selectedOption) return "Pilih produk...";

                    const product = productsData?.data.find(
                      (p) => p.id === selectedOption.value
                    );
                    if (!product) return "Pilih produk...";

                    const stock = stocksData?.data.find(
                      (s) => s.productID === product.id
                    );

                    return (
                      <div className="flex items-center justify-between flex-1 gap-2">
                        <div className="flex items-center gap-2">
                          <Package className="h-4 w-4" />
                          <span className="font-mono text-xs">
                            {product.code}
                          </span>
                          {" - "}
                          <span>{product.name}</span>
                        </div>
                        {stock && (
                          <span className="text-xs text-muted-foreground">
                            Stok: {parseFloat(stock.quantity).toLocaleString("id-ID", {
                              minimumFractionDigits: 0,
                              maximumFractionDigits: 3,
                            })}
                          </span>
                        )}
                      </div>
                    );
                  }}
                />
                {errors.product && (
                  <p className="text-sm text-destructive flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.product}
                  </p>
                )}
              </div>

              {/* Quantity */}
              <div className="space-y-2">
                <Label htmlFor="quantity">
                  Jumlah <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="quantity"
                  type="number"
                  step="0.001"
                  min="0"
                  placeholder="0.000"
                  value={quantity}
                  onChange={(e) => {
                    setQuantity(e.target.value);
                    setErrors({});
                  }}
                  className={cn("bg-background", errors.quantity && "border-destructive")}
                />
                {selectedProductId && (() => {
                  const stock = stocksData?.data.find((s) => s.productID === selectedProductId);
                  return stock ? (
                    <p className="text-xs text-muted-foreground">
                      Stok tersedia: {parseFloat(stock.quantity).toLocaleString("id-ID", {
                        minimumFractionDigits: 0,
                        maximumFractionDigits: 3,
                      })}
                    </p>
                  ) : null;
                })()}
                {errors.quantity && (
                  <p className="text-sm text-destructive flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.quantity}
                  </p>
                )}
              </div>

              {/* Item Notes */}
              <div className="space-y-2">
                <Label htmlFor="itemNotes">Catatan Item (Opsional)</Label>
                <Input
                  id="itemNotes"
                  placeholder="Catatan untuk item ini..."
                  value={itemNotes}
                  onChange={(e) => setItemNotes(e.target.value)}
                  className="bg-background"
                />
              </div>

              {/* Add Button */}
              <Button
                type="button"
                onClick={handleAddProduct}
                className="w-full"
                variant="secondary"
              >
                <Package className="mr-2 h-4 w-4" />
                Tambah Produk
              </Button>
            </div>

            {/* Items List */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-sm">
                  Daftar Produk ({items.length})
                </h3>
                {items.length > 0 && (
                  <Badge variant="secondary">
                    Total: {totalQuantity.toLocaleString("id-ID", {
                      minimumFractionDigits: 0,
                      maximumFractionDigits: 3,
                    })}
                  </Badge>
                )}
              </div>

              {items.length === 0 ? (
                <div className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
                  <Package className="mx-auto h-8 w-8 mb-2 opacity-50" />
                  Belum ada produk ditambahkan
                </div>
              ) : (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Kode</TableHead>
                        <TableHead>Nama Produk</TableHead>
                        <TableHead className="text-right">Jumlah</TableHead>
                        <TableHead>Catatan</TableHead>
                        <TableHead className="w-[50px]"></TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {items.map((item, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-mono text-sm">
                            {item.productCode}
                          </TableCell>
                          <TableCell className="font-medium">
                            {item.productName}
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
                          <TableCell>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => handleRemoveProduct(index)}
                              className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}

              {errors.items && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.items}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Step 3: Review & Submit */}
        {currentStep === 3 && (
          <div className="space-y-6">
            {/* Transfer Summary */}
            <div className="rounded-lg border p-4 space-y-4 bg-muted/30">
              <h3 className="font-semibold">Ringkasan Transfer</h3>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* From Warehouse */}
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Dari Gudang</p>
                  <div className="flex items-center gap-2">
                    <Warehouse className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">{getWarehouseName(sourceWarehouseId)}</p>
                  </div>
                </div>

                {/* To Warehouse */}
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Ke Gudang</p>
                  <div className="flex items-center gap-2">
                    <Warehouse className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">{getWarehouseName(destWarehouseId)}</p>
                  </div>
                </div>

                {/* Transfer Date */}
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Tanggal Transfer</p>
                  <div className="flex items-center gap-2">
                    <CalendarIcon className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">
                      {format(transferDate, "dd MMMM yyyy", { locale: localeId })}
                    </p>
                  </div>
                </div>

                {/* Total Items */}
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Total Produk</p>
                  <div className="flex items-center gap-2">
                    <Package className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">{items.length} produk</p>
                  </div>
                </div>
              </div>

              {/* Notes */}
              {notes && (
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Catatan</p>
                  <p className="text-sm bg-background p-3 rounded-md border">
                    {notes}
                  </p>
                </div>
              )}
            </div>

            {/* Items Review */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold">Daftar Produk ({items.length})</h3>
                <Badge>
                  Total Qty: {totalQuantity.toLocaleString("id-ID", {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 3,
                  })}
                </Badge>
              </div>

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
                    {items.map((item, index) => (
                      <TableRow key={index}>
                        <TableCell className="font-mono text-sm">
                          {item.productCode}
                        </TableCell>
                        <TableCell className="font-medium">
                          {item.productName}
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
            </div>

            {/* Confirmation Note */}
            <div className="rounded-md bg-blue-50 dark:bg-blue-950 p-4 text-sm">
              <div className="flex items-start gap-3">
                <AlertCircle className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                <div className="space-y-1">
                  <p className="font-medium text-blue-900 dark:text-blue-100">
                    Konfirmasi Transfer
                  </p>
                  <p className="text-blue-700 dark:text-blue-300">
                    Transfer akan dibuat dengan status <strong>DRAFT</strong>. Anda
                    dapat mengedit atau menghapus transfer sebelum dikirim.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Error Display */}
        {errors.submit && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-start gap-2">
            <AlertCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
            <span>{errors.submit}</span>
          </div>
        )}

        {/* Footer Buttons */}
        <DialogFooter className={cn("gap-2", currentStep === 2 && "sm:gap-2")}>
          {/* Cancel Button - Always visible */}
          <Button
            type="button"
            onClick={() => onOpenChange(false)}
            disabled={isCreating}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Batal
          </Button>

          {/* Back Button - Show from step 2 onwards */}
          {currentStep > 1 && (
            <Button
              type="button"
              variant="outline"
              onClick={handleBack}
              disabled={isCreating}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
          )}

          {/* Next/Submit Button */}
          {currentStep < 3 ? (
            <Button type="button" onClick={handleNext}>
              Lanjut
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          ) : (
            <Button
              type="button"
              onClick={handleSubmit}
              disabled={isCreating || isUpdating}
            >
              {isCreating || isUpdating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {isEditMode ? "Memperbarui Transfer..." : "Membuat Transfer..."}
                </>
              ) : (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  {isEditMode ? "Perbarui Transfer" : "Buat Transfer"}
                </>
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
