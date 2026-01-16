/**
 * Create Adjustment Form Component
 *
 * Multi-step wizard for creating inventory adjustments:
 * - Step 1: Select warehouse, type, reason, and date
 * - Step 2: Select products and adjustment quantities
 * - Step 3: Review and submit
 */

"use client";

import React, { useState } from "react";
import { useSelector } from "react-redux";
import { useRouter } from "next/navigation";
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
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import {
  ArrowRight,
  ArrowLeft,
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
  ArrowUp,
  ArrowDown,
} from "lucide-react";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { useListStocksQuery } from "@/store/services/stockApi";
import { useListProductsQuery } from "@/store/services/productApi";
import { useCreateAdjustmentMutation } from "@/store/services/adjustmentApi";
import type { RootState } from "@/store";
import {
  ADJUSTMENT_REASON_CONFIG,
  ADJUSTMENT_TYPE_CONFIG,
  type CreateAdjustmentRequest,
  type AdjustmentType,
  type AdjustmentReason,
} from "@/types/adjustment.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface CreateAdjustmentFormProps {
  onSuccess?: (adjustmentId: string) => void;
  onCancel?: () => void;
}

interface AdjustmentItem {
  productId: string;
  productCode: string;
  productName: string;
  quantityAdjusted: string;
  unitCost: string;
  batchId?: string;
  notes?: string;
  currentStock?: number;
}

export function CreateAdjustmentForm({
  onSuccess,
  onCancel,
}: CreateAdjustmentFormProps) {
  const router = useRouter();

  // Current step (1-3)
  const [currentStep, setCurrentStep] = useState(1);

  // Step 1: Basic Info
  const [warehouseId, setWarehouseId] = useState<string>("");
  const [adjustmentType, setAdjustmentType] = useState<AdjustmentType>("DECREASE");
  const [reason, setReason] = useState<AdjustmentReason>("CORRECTION");
  const [adjustmentDate, setAdjustmentDate] = useState<Date>(new Date());
  const [notes, setNotes] = useState<string>("");

  // Step 2: Product Selection
  const [items, setItems] = useState<AdjustmentItem[]>([]);
  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [quantityAdjusted, setQuantityAdjusted] = useState<string>("");
  const [itemNotes, setItemNotes] = useState<string>("");

  // Validation errors
  const [errors, setErrors] = useState<Record<string, string>>({});

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

  // Fetch stocks from warehouse
  const { data: stocksData, isLoading: loadingStocks } = useListStocksQuery(
    {
      page: 1,
      pageSize: 100,
      warehouseID: warehouseId,
    },
    { skip: !activeCompanyId || currentStep !== 2 || !warehouseId }
  );

  // Fetch products to get baseCost for auto-fill
  const skipProductsQuery = !activeCompanyId || currentStep !== 2;
  const { data: productsData, isLoading: loadingProducts } = useListProductsQuery(
    { page: 1, pageSize: 100, isActive: true },
    { skip: skipProductsQuery }
  );

  // Create mutation
  const [createAdjustment, { isLoading: isCreating }] =
    useCreateAdjustmentMutation();

  // Validate Step 1
  const validateStep1 = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!warehouseId) {
      newErrors.warehouse = "Pilih gudang";
    }
    if (!adjustmentType) {
      newErrors.adjustmentType = "Pilih tipe penyesuaian";
    }
    if (!reason) {
      newErrors.reason = "Pilih alasan penyesuaian";
    }
    if (!adjustmentDate) {
      newErrors.adjustmentDate = "Pilih tanggal";
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
    if (!quantityAdjusted || parseFloat(quantityAdjusted) === 0) {
      newErrors.quantityAdjusted = "Masukkan jumlah yang valid";
    }
    if (loadingProducts) {
      newErrors.product = "Menunggu data produk dimuat...";
    }

    // Find stock entry
    const stock = stocksData?.data.find((s) => s.productID === selectedProductId);
    if (!stock && selectedProductId) {
      newErrors.product = "Produk tidak ditemukan di gudang ini";
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    if (!stock) return;

    // Check if product already in list
    if (items.some((item) => item.productId === selectedProductId)) {
      setErrors({ product: "Produk sudah ditambahkan" });
      return;
    }

    // Get cost from products data (auto-fill price)
    const product = productsData?.data.find((p) => p.id === selectedProductId);
    const productBaseCost = parseFloat(product?.baseCost || "0");
    const productBasePrice = parseFloat(product?.basePrice || "0");
    const autoUnitCost = productBaseCost > 0 ? productBaseCost.toString() :
                         productBasePrice > 0 ? productBasePrice.toString() : "0";

    // Add to items
    const newItem: AdjustmentItem = {
      productId: selectedProductId,
      productCode: stock.productCode,
      productName: stock.productName,
      quantityAdjusted: `${Math.abs(parseFloat(quantityAdjusted))}`,
      unitCost: autoUnitCost,
      notes: itemNotes || undefined,
      currentStock: parseFloat(stock.quantity),
    };

    setItems([...items, newItem]);

    // Clear form
    setSelectedProductId("");
    setQuantityAdjusted("");
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

  // Submit adjustment
  const handleSubmit = async () => {
    try {
      const requestData: CreateAdjustmentRequest = {
        adjustmentDate: format(adjustmentDate, "yyyy-MM-dd"),
        warehouseId,
        adjustmentType,
        reason,
        notes: notes || undefined,
        items: items.map((item) => ({
          productId: item.productId,
          batchId: item.batchId,
          quantityAdjusted: Math.abs(parseFloat(item.quantityAdjusted)).toString(),
          unitCost: item.unitCost,
          notes: item.notes,
        })),
      };

      const result = await createAdjustment(requestData).unwrap();

      toast.success("Penyesuaian Berhasil Dibuat", {
        description: `Penyesuaian ${result.adjustmentNumber} telah dibuat dan siap untuk disetujui.`,
      });

      if (onSuccess) {
        onSuccess(result.id);
      } else {
        router.push(`/inventory/adjustments/${result.id}`);
      }
    } catch (error: any) {
      toast.error("Gagal Membuat Penyesuaian", {
        description:
          error?.data?.message ||
          "Terjadi kesalahan. Silakan coba lagi.",
      });

      setErrors({
        submit:
          error?.data?.message || "Gagal membuat penyesuaian",
      });
    }
  };

  // Get warehouse name
  const getWarehouseName = (id: string) => {
    return warehousesData?.data.find((w) => w.id === id)?.name || "-";
  };

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

  // Calculate totals
  const totalValue = items.reduce((sum, item) => {
    const qty = Math.abs(parseFloat(item.quantityAdjusted || "0"));
    const cost = parseFloat(item.unitCost || "0");
    return sum + qty * cost;
  }, 0);

  // Convert warehouse stocks to ComboboxOption[]
  const productOptions: ComboboxOption[] = React.useMemo(() => {
    if (!stocksData?.data) return [];
    return stocksData.data
      .filter((stock) => stock.productCode && stock.productName)
      .map((stock) => ({
        value: stock.productID,
        label: stock.productName,
        searchLabel: `${stock.productCode} ${stock.productName}`,
      }));
  }, [stocksData]);

  const typeConfig = ADJUSTMENT_TYPE_CONFIG[adjustmentType];
  const reasonConfig = ADJUSTMENT_REASON_CONFIG[reason];

  return (
    <div className="space-y-6">
      {/* Step Indicator */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center gap-2">
            {[
              { number: 1, icon: Building2, label: "Info" },
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
        </CardContent>
      </Card>

      {/* Step 1: Basic Info */}
      {currentStep === 1 && (
        <Card>
          <CardHeader>
            <CardTitle>Informasi Dasar</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Warehouse */}
            <div className="space-y-2">
              <Label htmlFor="warehouse">
                Gudang <span className="text-destructive">*</span>
              </Label>
              <Select
                value={warehouseId}
                onValueChange={(value) => {
                  setWarehouseId(value);
                  if (value !== warehouseId) {
                    setItems([]);
                    setSelectedProductId("");
                  }
                  setErrors({});
                }}
              >
                <SelectTrigger
                  id="warehouse"
                  className={cn("w-full", errors.warehouse && "border-destructive")}
                >
                  <SelectValue placeholder="Pilih gudang" />
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
                        </div>
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
              {errors.warehouse && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.warehouse}
                </p>
              )}
            </div>

            {/* Adjustment Type */}
            <div className="space-y-2">
              <Label htmlFor="adjustmentType">
                Tipe Penyesuaian <span className="text-destructive">*</span>
              </Label>
              <Select
                value={adjustmentType}
                onValueChange={(value) => {
                  setAdjustmentType(value as AdjustmentType);
                  setErrors({});
                }}
              >
                <SelectTrigger
                  id="adjustmentType"
                  className={cn(
                    "w-full",
                    errors.adjustmentType && "border-destructive"
                  )}
                >
                  <SelectValue placeholder="Pilih tipe" />
                </SelectTrigger>
                <SelectContent position="popper" sideOffset={4}>
                  <SelectItem value="INCREASE">
                    <div className="flex items-center gap-2">
                      <ArrowUp className="h-4 w-4 text-green-600" />
                      Penambahan Stok
                    </div>
                  </SelectItem>
                  <SelectItem value="DECREASE">
                    <div className="flex items-center gap-2">
                      <ArrowDown className="h-4 w-4 text-red-600" />
                      Pengurangan Stok
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
              {errors.adjustmentType && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.adjustmentType}
                </p>
              )}
            </div>

            {/* Reason */}
            <div className="space-y-2">
              <Label htmlFor="reason">
                Alasan <span className="text-destructive">*</span>
              </Label>
              <Select
                value={reason}
                onValueChange={(value) => {
                  setReason(value as AdjustmentReason);
                  setErrors({});
                }}
              >
                <SelectTrigger
                  id="reason"
                  className={cn("w-full", errors.reason && "border-destructive")}
                >
                  <SelectValue placeholder="Pilih alasan" />
                </SelectTrigger>
                <SelectContent position="popper" sideOffset={4}>
                  {Object.entries(ADJUSTMENT_REASON_CONFIG).map(([key, config]) => (
                    <SelectItem key={key} value={key}>
                      <div className="flex flex-col">
                        <span>{config.label}</span>
                        <span className="text-xs text-muted-foreground">
                          {config.description}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.reason && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.reason}
                </p>
              )}
            </div>

            {/* Date */}
            <div className="space-y-2">
              <Label>
                Tanggal <span className="text-destructive">*</span>
              </Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !adjustmentDate && "text-muted-foreground",
                      errors.adjustmentDate && "border-destructive"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {adjustmentDate ? (
                      format(adjustmentDate, "dd MMMM yyyy", { locale: localeId })
                    ) : (
                      <span>Pilih tanggal</span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={adjustmentDate}
                    onSelect={(date) => {
                      if (date) {
                        setAdjustmentDate(date);
                        setErrors({});
                      }
                    }}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
              {errors.adjustmentDate && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.adjustmentDate}
                </p>
              )}
            </div>

            {/* Notes */}
            <div className="space-y-2">
              <Label htmlFor="notes">Catatan (Opsional)</Label>
              <Textarea
                id="notes"
                placeholder="Tambahkan catatan untuk penyesuaian ini..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step 2: Product Selection */}
      {currentStep === 2 && (
        <div className="space-y-4">
          {/* Product Selection Form */}
          <Card>
            <CardHeader>
              <CardTitle>Tambah Produk</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
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
                    loadingStocks
                      ? "Memuat..."
                      : !warehouseId
                      ? "Pilih gudang terlebih dahulu"
                      : "Tidak ada produk di gudang ini"
                  }
                  disabled={loadingStocks || !warehouseId}
                  className={cn(errors.product && "border-destructive")}
                  renderOption={(option, selected) => {
                    const stock = stocksData?.data.find(
                      (s) => s.productID === option.value
                    );
                    if (!stock) return null;

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
                              {stock.productCode}
                            </span>
                            {" - "}
                            <span>{stock.productName}</span>
                          </div>
                          <span className="text-xs text-muted-foreground ml-auto">
                            Stok: {parseFloat(stock.quantity).toLocaleString("id-ID")}
                          </span>
                        </div>
                      </>
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
                <Label htmlFor="quantityAdjusted">
                  Jumlah {adjustmentType === "INCREASE" ? "Penambahan" : "Pengurangan"}{" "}
                  <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="quantityAdjusted"
                  type="number"
                  step="0.001"
                  min="0"
                  placeholder="0.000"
                  value={quantityAdjusted}
                  onChange={(e) => {
                    setQuantityAdjusted(e.target.value);
                    setErrors({});
                  }}
                  className={cn(
                    "bg-background",
                    errors.quantityAdjusted && "border-destructive"
                  )}
                />
                {errors.quantityAdjusted && (
                  <p className="text-sm text-destructive flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.quantityAdjusted}
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
            </CardContent>
          </Card>

          {/* Items List */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Daftar Produk ({items.length})</CardTitle>
                {items.length > 0 && (
                  <Badge variant="secondary">
                    Total: {formatCurrency(totalValue)}
                  </Badge>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {items.length === 0 ? (
                <div className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
                  <Package className="mx-auto h-8 w-8 mb-2 opacity-50" />
                  Belum ada produk ditambahkan
                </div>
              ) : (
                <div className="space-y-2">
                  {items.map((item, index) => {
                    const qty = Math.abs(parseFloat(item.quantityAdjusted || "0"));
                    const cost = parseFloat(item.unitCost || "0");
                    const itemTotal = qty * cost;

                    return (
                      <div
                        key={index}
                        className="flex items-center justify-between gap-2 p-3 rounded-lg border bg-background"
                      >
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-mono text-xs text-muted-foreground">
                              {item.productCode}
                            </span>
                          </div>
                          <div className="font-medium text-sm truncate" title={item.productName}>
                            {item.productName}
                          </div>
                          <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                            <span>
                              Qty:{" "}
                              <span
                                className={cn(
                                  "font-medium",
                                  adjustmentType === "INCREASE"
                                    ? "text-green-600"
                                    : "text-red-600"
                                )}
                              >
                                {adjustmentType === "INCREASE" ? "+" : "-"}
                                {Math.abs(parseFloat(item.quantityAdjusted)).toLocaleString("id-ID")}
                              </span>
                            </span>
                            <span>×</span>
                            <span>{formatCurrency(cost)}</span>
                          </div>
                          {item.notes && (
                            <div className="text-xs text-muted-foreground mt-1 italic">
                              Catatan: {item.notes}
                            </div>
                          )}
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="text-right">
                            <div className="font-semibold text-sm">
                              {formatCurrency(itemTotal)}
                            </div>
                          </div>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemoveProduct(index)}
                            className="h-8 w-8 p-0 text-destructive hover:text-destructive flex-shrink-0"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}

              {errors.items && (
                <p className="text-sm text-destructive flex items-center gap-1 mt-2">
                  <AlertCircle className="h-3 w-3" />
                  {errors.items}
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Step 3: Review & Submit */}
      {currentStep === 3 && (
        <div className="space-y-4">
          {/* Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Ringkasan Penyesuaian</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Gudang</p>
                  <div className="flex items-center gap-2">
                    <Warehouse className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">{getWarehouseName(warehouseId)}</p>
                  </div>
                </div>

                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Tanggal</p>
                  <div className="flex items-center gap-2">
                    <CalendarIcon className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">
                      {format(adjustmentDate, "dd MMMM yyyy", { locale: localeId })}
                    </p>
                  </div>
                </div>

                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Tipe</p>
                  <Badge className={typeConfig.className}>{typeConfig.label}</Badge>
                </div>

                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Alasan</p>
                  <p className="font-medium">{reasonConfig.label}</p>
                </div>
              </div>

              {notes && (
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">Catatan</p>
                  <p className="text-sm bg-muted p-3 rounded-md border">{notes}</p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Items Review */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Daftar Produk ({items.length})</CardTitle>
                <Badge>{formatCurrency(totalValue)}</Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {items.map((item, index) => {
                  const qty = Math.abs(parseFloat(item.quantityAdjusted || "0"));
                  const cost = parseFloat(item.unitCost || "0");
                  const itemTotal = qty * cost;

                  return (
                    <div
                      key={index}
                      className="flex items-center justify-between gap-2 p-3 rounded-lg border bg-background"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-xs text-muted-foreground">
                            {item.productCode}
                          </span>
                        </div>
                        <div className="font-medium text-sm truncate" title={item.productName}>
                          {item.productName}
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                          <span>
                            Qty:{" "}
                            <span
                              className={cn(
                                "font-medium",
                                adjustmentType === "INCREASE"
                                  ? "text-green-600"
                                  : "text-red-600"
                              )}
                            >
                              {adjustmentType === "INCREASE" ? "+" : "-"}
                              {Math.abs(parseFloat(item.quantityAdjusted)).toLocaleString("id-ID")}
                            </span>
                          </span>
                          <span>×</span>
                          <span>{formatCurrency(cost)}</span>
                        </div>
                        {item.notes && (
                          <div className="text-xs text-muted-foreground mt-1 italic">
                            Catatan: {item.notes}
                          </div>
                        )}
                      </div>
                      <div className="text-right flex-shrink-0">
                        <div className="font-semibold text-sm">
                          {formatCurrency(itemTotal)}
                        </div>
                      </div>
                    </div>
                  );
                })}
                {/* Total Row */}
                <div className="flex items-center justify-between p-3 rounded-lg bg-muted/50 font-semibold">
                  <span>Total Nilai</span>
                  <span>{formatCurrency(totalValue)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Confirmation Note */}
          <div className="rounded-md bg-blue-50 dark:bg-blue-950 p-4 text-sm">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
              <div className="space-y-1">
                <p className="font-medium text-blue-900 dark:text-blue-100">
                  Konfirmasi Penyesuaian
                </p>
                <p className="text-blue-700 dark:text-blue-300">
                  Penyesuaian akan dibuat dengan status <strong>DRAFT</strong>.
                  Anda dapat mengedit atau menghapus sebelum disetujui. Setelah
                  disetujui, perubahan stok akan langsung diterapkan.
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
      <div className="flex items-center justify-between gap-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel || (() => router.push("/inventory/adjustments"))}
          disabled={isCreating}
        >
          Batal
        </Button>

        <div className="flex items-center gap-2">
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

          {currentStep < 3 ? (
            <Button type="button" onClick={handleNext}>
              Lanjut
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          ) : (
            <Button
              type="button"
              onClick={handleSubmit}
              disabled={isCreating}
            >
              {isCreating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Membuat...
                </>
              ) : (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  Buat Penyesuaian
                </>
              )}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
