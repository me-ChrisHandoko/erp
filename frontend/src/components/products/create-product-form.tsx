/**
 * Create Product Form Component
 *
 * Professional form for creating new products with:
 * - Real-time validation
 * - Currency formatting
 * - Profit margin calculator
 * - Responsive layout
 * - Mobile-friendly sticky action bar
 * - Unsaved changes warning
 */

"use client";

import { useState, useEffect, useCallback, useMemo, useRef } from "react";
import {
  DollarSign,
  Layers,
  Save,
  AlertCircle,
  Info,
  TrendingUp,
  PackageCheck,
  Calendar,
  Bell,
  FileText,
} from "lucide-react";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Users, ChevronLeft, ChevronRight, Check, Loader2, CheckCircle2, XCircle } from "lucide-react";
import {
  useCreateProductMutation,
  useLinkSupplierMutation,
  useListProductsQuery,
} from "@/store/services/productApi";
import { toast } from "sonner";
import {
  PRODUCT_CATEGORIES,
  COMMON_UNITS,
  type CreateProductRequest,
} from "@/types/product.types";
import {
  ProductSuppliersSection,
  type SupplierFormData,
} from "./product-suppliers-section";
import {
  ProductUnitsSection,
  type ProductUnitItem,
} from "./product-units-section";

interface CreateProductFormProps {
  onSuccess?: (productId: string) => void;
  onCancel?: () => void;
}

// Tab configuration - defined outside component to prevent recreation on each render
const TABS = ["basic", "pricing", "stock", "suppliers"] as const;
type TabId = typeof TABS[number];

export function CreateProductForm({
  onSuccess,
  onCancel,
}: CreateProductFormProps) {
  const [createProduct, { isLoading: isCreatingProduct }] = useCreateProductMutation();
  const [linkSupplier, { isLoading: isLinkingSupplier }] = useLinkSupplierMutation();

  const isLoading = isCreatingProduct || isLinkingSupplier;

  // Form state
  const [formData, setFormData] = useState<CreateProductRequest>({
    code: "",
    name: "",
    category: "",
    description: "",
    baseUnit: "PCS",
    baseCost: "0",
    basePrice: "0",
    minimumStock: "0",
    barcode: "",
    isBatchTracked: false,
    isPerishable: false,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Real-time code validation - debounced code for API query
  const [debouncedCode, setDebouncedCode] = useState("");
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Debounced code update - all state updates happen in setTimeout callback
  useEffect(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    // Set timer for debounced update (includes reset case)
    debounceTimerRef.current = setTimeout(() => {
      if (!formData.code || formData.code.length < 2) {
        setDebouncedCode("");
      } else {
        setDebouncedCode(formData.code);
      }
    }, formData.code && formData.code.length >= 2 ? 500 : 0);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [formData.code]);

  // Query to check if code exists
  const { data: existingProducts, isFetching: isCheckingCode } = useListProductsQuery(
    { search: debouncedCode, pageSize: 1 },
    { skip: !debouncedCode || debouncedCode.length < 2 }
  );

  // Derive validation status using useMemo
  const codeValidationStatus = useMemo((): "idle" | "checking" | "available" | "taken" => {
    if (!formData.code || formData.code.length < 2) {
      return "idle";
    }

    // If code is different from debounced, we're still waiting for debounce
    if (formData.code !== debouncedCode) {
      return "checking";
    }

    if (isCheckingCode) {
      return "checking";
    }

    if (existingProducts?.data) {
      const codeExists = existingProducts.data.some(
        (p) => p.code.toLowerCase() === debouncedCode.toLowerCase()
      );
      return codeExists ? "taken" : "available";
    }

    return "idle";
  }, [formData.code, debouncedCode, existingProducts, isCheckingCode]);

  // Tab navigation state
  const [activeTab, setActiveTab] = useState<TabId>("basic");

  // Suppliers state
  const [suppliers, setSuppliers] = useState<SupplierFormData[]>([]);

  // Units state - start with base unit
  const [units, setUnits] = useState<ProductUnitItem[]>([]);

  // Derive unsaved changes status using useMemo (avoid setState in useEffect)
  const hasUnsavedChanges = useMemo(() => {
    return (
      formData.code !== "" ||
      formData.name !== "" ||
      formData.category !== "" ||
      formData.description !== "" ||
      formData.barcode !== "" ||
      formData.baseCost !== "0" ||
      formData.basePrice !== "0" ||
      formData.minimumStock !== "0" ||
      formData.isBatchTracked !== false ||
      formData.isPerishable !== false ||
      suppliers.length > 0 ||
      units.length > 0
    );
  }, [formData, suppliers, units]);

  // Tab validation status
  const getTabValidationStatus = useCallback((tab: TabId): "complete" | "error" | "pending" => {
    switch (tab) {
      case "basic":
        const hasBasicData = formData.code.trim() !== "" && formData.name.trim() !== "";
        const hasBasicErrors = !!(errors.code || errors.name);
        if (hasBasicErrors) return "error";
        if (hasBasicData) return "complete";
        return "pending";
      case "pricing":
        const hasPricingData = parseFloat(formData.baseCost) > 0 && parseFloat(formData.basePrice) > 0;
        const hasPricingErrors = !!(errors.baseCost || errors.basePrice || errors.baseUnit);
        if (hasPricingErrors) return "error";
        if (hasPricingData) return "complete";
        return "pending";
      case "stock":
        // Stock section is optional, so it's complete if no errors
        const hasStockErrors = !!errors.minimumStock;
        if (hasStockErrors) return "error";
        return "complete";
      case "suppliers":
        // Suppliers are optional
        return suppliers.length > 0 ? "complete" : "pending";
      default:
        return "pending";
    }
  }, [formData, errors, suppliers]);

  // Tab navigation helpers
  const goToNextTab = useCallback(() => {
    const currentIndex = TABS.indexOf(activeTab);
    if (currentIndex < TABS.length - 1) {
      setActiveTab(TABS[currentIndex + 1]);
    }
  }, [activeTab]);

  const goToPrevTab = useCallback(() => {
    const currentIndex = TABS.indexOf(activeTab);
    if (currentIndex > 0) {
      setActiveTab(TABS[currentIndex - 1]);
    }
  }, [activeTab]);

  // Warn before leaving with unsaved changes
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        e.returnValue = "Anda memiliki perubahan yang belum disimpan. Yakin ingin meninggalkan halaman?";
        return e.returnValue;
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [hasUnsavedChanges]);

  const handleChange = (field: keyof CreateProductRequest, value: string | boolean) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user types
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
  };

  const formatCurrency = (value: string): string => {
    const num = parseFloat(value || "0");
    if (isNaN(num)) return "Rp 0";
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(num);
  };

  // Format number with thousand separators for display in input
  const formatNumberInput = (value: string): string => {
    // Remove non-numeric characters except decimal point
    const cleanValue = value.replace(/[^\d]/g, "");
    if (!cleanValue) return "";
    const num = parseInt(cleanValue, 10);
    if (isNaN(num)) return "";
    return new Intl.NumberFormat("id-ID").format(num);
  };

  // Parse formatted number back to raw value
  const parseFormattedNumber = (value: string): string => {
    const cleanValue = value.replace(/[^\d]/g, "");
    return cleanValue || "0";
  };

  // Handle currency input change
  const handleCurrencyChange = (field: "baseCost" | "basePrice", formattedValue: string) => {
    const rawValue = parseFormattedNumber(formattedValue);
    handleChange(field, rawValue);
  };

  const calculateProfit = () => {
    const cost = parseFloat(formData.baseCost);
    const price = parseFloat(formData.basePrice);
    if (isNaN(cost) || isNaN(price) || cost === 0) return null;
    const margin = ((price - cost) / cost) * 100;
    const profit = price - cost;
    return { margin, profit };
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields
    if (!formData.code.trim()) {
      newErrors.code = "Kode produk wajib diisi";
    } else if (formData.code.length < 2) {
      newErrors.code = "Kode produk minimal 2 karakter";
    } else if (codeValidationStatus === "taken") {
      newErrors.code = "Kode produk ini sudah digunakan";
    }

    if (!formData.name.trim()) {
      newErrors.name = "Nama produk wajib diisi";
    } else if (formData.name.length < 3) {
      newErrors.name = "Nama produk minimal 3 karakter";
    }

    if (!formData.baseUnit.trim()) {
      newErrors.baseUnit = "Satuan dasar wajib diisi";
    }

    // Numeric validations
    const baseCost = parseFloat(formData.baseCost);
    const basePrice = parseFloat(formData.basePrice);
    const minimumStock = parseFloat(formData.minimumStock || "0");

    if (isNaN(baseCost) || baseCost <= 0) {
      newErrors.baseCost = "Harga beli harus lebih dari 0";
    }
    if (isNaN(basePrice) || basePrice <= 0) {
      newErrors.basePrice = "Harga jual harus lebih dari 0";
    }
    if (!isNaN(baseCost) && !isNaN(basePrice) && basePrice < baseCost) {
      newErrors.basePrice = "Harga jual harus lebih besar dari harga beli";
    }
    if (isNaN(minimumStock) || minimumStock < 0) {
      newErrors.minimumStock = "Stok minimum tidak boleh negatif";
    }

    setErrors(newErrors);
    setTouched({
      code: true,
      name: true,
      baseUnit: true,
      baseCost: true,
      basePrice: true,
      minimumStock: true,
    });

    // Validation complete - errors will be displayed in the form

    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      toast.error("Validasi Gagal", {
        description: "Mohon periksa kembali form Anda",
      });
      return;
    }

    try {
      // Clean up empty optional fields before sending
      const cleanedData: CreateProductRequest = {
        code: formData.code,
        name: formData.name,
        baseUnit: formData.baseUnit,
        baseCost: formData.baseCost,
        basePrice: formData.basePrice,
        isBatchTracked: formData.isBatchTracked,
        isPerishable: formData.isPerishable,
        // Only include optional fields if they have values
        ...(formData.category && formData.category.trim() !== "" && { category: formData.category }),
        ...(formData.description && formData.description.trim() !== "" && { description: formData.description }),
        ...(formData.barcode && formData.barcode.trim() !== "" && { barcode: formData.barcode }),
        ...(formData.minimumStock && formData.minimumStock !== "0" && { minimumStock: formData.minimumStock }),
        // Include additional units if any
        ...(units.length > 0 && {
          units: units.map((u) => ({
            unitName: u.unitName,
            conversionRate: u.conversionRate,
            buyPrice: u.buyPrice || undefined,
            sellPrice: u.sellPrice || undefined,
          })),
        }),
      };

      // Create the product first
      const result = await createProduct(cleanedData).unwrap();

      // Link suppliers to the newly created product
      const supplierErrors: string[] = [];
      for (const supplier of suppliers) {
        try {
          await linkSupplier({
            productId: result.id,
            data: {
              supplierId: supplier.supplierId,
              supplierPrice: supplier.supplierPrice,
              leadTimeDays: supplier.leadTimeDays ? parseInt(supplier.leadTimeDays) : undefined,
              minimumOrderQty: supplier.minimumOrderQty || undefined,
              supplierProductCode: supplier.supplierProductCode || undefined,
              supplierProductName: supplier.supplierProductName || undefined,
              isPrimarySupplier: supplier.isPrimarySupplier,
            },
          }).unwrap();
        } catch {
          supplierErrors.push(`Gagal menambahkan supplier ${supplier.supplierName}`);
        }
      }

      // Show success message
      if (supplierErrors.length > 0) {
        toast.warning("Produk Dibuat dengan Peringatan", {
          description: `${result.name} telah ditambahkan, tetapi: ${supplierErrors.join(", ")}`,
        });
      } else {
        toast.success("Produk Berhasil Dibuat", {
          description: suppliers.length > 0
            ? `${result.name} telah ditambahkan dengan ${suppliers.length} supplier`
            : `${result.name} telah ditambahkan ke katalog`,
        });
      }

      if (onSuccess) {
        onSuccess(result.id);
      }
    } catch (error: unknown) {
      const err = error as { data?: { error?: { message?: string }; message?: string }; message?: string };
      toast.error("Gagal Membuat Produk", {
        description:
          err?.data?.error?.message ||
          err?.data?.message ||
          err?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  const profitData = calculateProfit();

  // Tab configuration
  const tabConfig = [
    { id: "basic", label: "Informasi Dasar", icon: FileText },
    { id: "pricing", label: "Harga & Satuan", icon: DollarSign },
    { id: "stock", label: "Stok & Pelacakan", icon: Layers },
    { id: "suppliers", label: "Suppliers", icon: Users },
  ] as const;

  return (
    <form onSubmit={handleSubmit} className="space-y-6 pb-32 md:pb-0">
      {/* Tab Navigation */}
      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as TabId)} className="w-full">
        <TabsList className="grid w-full grid-cols-2 md:grid-cols-4 h-auto gap-1 p-1">
          {tabConfig.map((tab, index) => {
            const status = getTabValidationStatus(tab.id as TabId);
            const Icon = tab.icon;
            return (
              <TabsTrigger
                key={tab.id}
                value={tab.id}
                className="flex items-center gap-2 py-2.5 px-3 text-xs md:text-sm data-[state=active]:bg-primary data-[state=active]:text-primary-foreground relative"
              >
                <div className="flex items-center gap-1.5">
                  <span className="hidden md:flex items-center justify-center w-5 h-5 rounded-full bg-muted text-xs font-medium">
                    {status === "complete" ? (
                      <Check className="h-3 w-3 text-green-600" />
                    ) : status === "error" ? (
                      <AlertCircle className="h-3 w-3 text-destructive" />
                    ) : (
                      index + 1
                    )}
                  </span>
                  <Icon className="h-4 w-4 md:hidden" />
                  <span className="truncate">{tab.label}</span>
                </div>
                {status === "error" && (
                  <Badge variant="destructive" className="absolute -top-1 -right-1 h-4 w-4 p-0 flex items-center justify-center text-[10px] md:hidden">
                    !
                  </Badge>
                )}
              </TabsTrigger>
            );
          })}
        </TabsList>

        {/* Tab 1: Basic Information */}
        <TabsContent value="basic" className="mt-6">
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <FileText className="h-5 w-5 text-primary" />
            Informasi Dasar
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Code */}
            <div className="space-y-2">
              <Label htmlFor="code" className="text-sm font-medium">
                Kode Produk <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <Input
                  id="code"
                  value={formData.code}
                  onChange={(e) =>
                    handleChange("code", e.target.value.toUpperCase())
                  }
                  onBlur={() => handleBlur("code")}
                  placeholder="Contoh: BRS-001"
                  className={`pr-10 ${
                    errors.code && touched.code
                      ? "border-destructive"
                      : codeValidationStatus === "taken"
                      ? "border-destructive"
                      : codeValidationStatus === "available"
                      ? "border-green-500"
                      : ""
                  }`}
                />
                {/* Validation Status Icon */}
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  {codeValidationStatus === "checking" && (
                    <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                  )}
                  {codeValidationStatus === "available" && (
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                  )}
                  {codeValidationStatus === "taken" && (
                    <XCircle className="h-4 w-4 text-destructive" />
                  )}
                </div>
              </div>
              {errors.code && touched.code && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.code}
                </p>
              )}
              {codeValidationStatus === "taken" && !errors.code && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  Kode produk ini sudah digunakan
                </p>
              )}
              {codeValidationStatus === "available" && (
                <p className="flex items-center gap-1 text-sm text-green-600">
                  <CheckCircle2 className="h-3 w-3" />
                  Kode produk tersedia
                </p>
              )}
              {codeValidationStatus === "idle" && (
                <p className="text-xs text-muted-foreground">
                  Kode unik untuk identifikasi produk
                </p>
              )}
            </div>

            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name" className="text-sm font-medium">
                Nama Produk <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => handleChange("name", e.target.value)}
                onBlur={() => handleBlur("name")}
                placeholder="Contoh: Beras Premium 5kg"
                className={
                  errors.name && touched.name ? "border-destructive" : ""
                }
              />
              {errors.name && touched.name && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.name}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Nama lengkap produk yang mudah dikenali
              </p>
            </div>

            {/* Category */}
            <div className="space-y-2">
              <Label htmlFor="category" className="text-sm font-medium">
                Kategori
              </Label>
              <Select
                value={formData.category}
                onValueChange={(value) => handleChange("category", value)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Pilih kategori produk" />
                </SelectTrigger>
                <SelectContent>
                  {PRODUCT_CATEGORIES.map((cat) => (
                    <SelectItem key={cat} value={cat}>
                      {cat}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Kategori untuk memudahkan pencarian
              </p>
            </div>

            {/* Barcode */}
            <div className="space-y-2">
              <Label htmlFor="barcode" className="text-sm font-medium">
                Barcode
              </Label>
              <Input
                id="barcode"
                value={formData.barcode}
                onChange={(e) => handleChange("barcode", e.target.value)}
                placeholder="Contoh: 8991234567890"
              />
              <p className="text-xs text-muted-foreground">
                Barcode untuk scanning (opsional)
              </p>
            </div>

            {/* Description */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="description" className="text-sm font-medium">
                Deskripsi Produk
              </Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => handleChange("description", e.target.value)}
                placeholder="Deskripsi detail tentang produk..."
                rows={3}
                className="resize-none"
              />
              <p className="text-xs text-muted-foreground">
                Informasi tambahan tentang produk
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

          {/* Tab Navigation Buttons */}
          <div className="flex justify-end pt-4">
            <Button
              type="button"
              onClick={goToNextTab}
              className="gap-2"
            >
              Lanjut ke Harga & Satuan
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </TabsContent>

        {/* Tab 2: Pricing & Unit */}
        <TabsContent value="pricing" className="mt-6">
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <DollarSign className="h-5 w-5 text-primary" />
            Harga & Satuan
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid gap-6 sm:grid-cols-3">
            {/* Base Unit */}
            <div className="space-y-2">
              <Label htmlFor="baseUnit" className="text-sm font-medium">
                Satuan Dasar <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.baseUnit}
                onValueChange={(value) => handleChange("baseUnit", value)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {COMMON_UNITS.map((unit) => (
                    <SelectItem key={unit} value={unit}>
                      {unit}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.baseUnit && touched.baseUnit && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.baseUnit}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Satuan terkecil untuk produk ini
              </p>
            </div>

            {/* Base Cost */}
            <div className="space-y-2">
              <Label htmlFor="baseCost" className="text-sm font-medium">
                Harga Beli (HPP) <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  Rp
                </span>
                <Input
                  id="baseCost"
                  type="text"
                  inputMode="numeric"
                  value={formatNumberInput(formData.baseCost)}
                  onChange={(e) => handleCurrencyChange("baseCost", e.target.value)}
                  onBlur={() => handleBlur("baseCost")}
                  placeholder="0"
                  className={`pl-10 text-right font-mono ${
                    errors.baseCost && touched.baseCost
                      ? "border-destructive"
                      : ""
                  }`}
                />
              </div>
              {errors.baseCost && touched.baseCost && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.baseCost}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Harga per {formData.baseUnit || "satuan"}
              </p>
            </div>

            {/* Base Price */}
            <div className="space-y-2">
              <Label htmlFor="basePrice" className="text-sm font-medium">
                Harga Jual <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  Rp
                </span>
                <Input
                  id="basePrice"
                  type="text"
                  inputMode="numeric"
                  value={formatNumberInput(formData.basePrice)}
                  onChange={(e) => handleCurrencyChange("basePrice", e.target.value)}
                  onBlur={() => handleBlur("basePrice")}
                  placeholder="0"
                  className={`pl-10 text-right font-mono ${
                    errors.basePrice && touched.basePrice
                      ? "border-destructive"
                      : ""
                  }`}
                />
              </div>
              {errors.basePrice && touched.basePrice && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.basePrice}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Harga per {formData.baseUnit || "satuan"}
              </p>
            </div>

            {/* Profit Margin Display */}
            {profitData && profitData.margin > 0 && (
              <div className="sm:col-span-3">
                <Alert className="bg-green-50 border-green-200 dark:bg-green-900/10 dark:border-green-900">
                  <TrendingUp className="h-4 w-4 text-green-600" />
                  <AlertDescription className="text-green-800 dark:text-green-200">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">Margin Keuntungan</div>
                        <div className="text-sm text-muted-foreground">
                          Keuntungan per unit
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-2xl font-bold">
                          {profitData.margin.toFixed(1)}%
                        </div>
                        <div className="text-sm">
                          {formatCurrency(profitData.profit.toString())}
                        </div>
                      </div>
                    </div>
                  </AlertDescription>
                </Alert>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

          {/* Unit Conversion Section */}
          <ProductUnitsSection
            units={units}
            onUnitsChange={setUnits}
            baseUnit={formData.baseUnit}
            baseCost={formData.baseCost}
            basePrice={formData.basePrice}
            disabled={isLoading}
          />

          {/* Tab Navigation Buttons */}
          <div className="flex justify-between pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={goToPrevTab}
              className="gap-2"
            >
              <ChevronLeft className="h-4 w-4" />
              Kembali
            </Button>
            <Button
              type="button"
              onClick={goToNextTab}
              className="gap-2"
            >
              Lanjut ke Stok & Pelacakan
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </TabsContent>

        {/* Tab 3: Stock & Tracking */}
        <TabsContent value="stock" className="mt-6">
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Layers className="h-5 w-5 text-primary" />
            Stok & Pelacakan
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="space-y-6">
            {/* Minimum Stock */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Bell className="h-4 w-4 text-muted-foreground" />
                <Label htmlFor="minimumStock" className="text-sm font-medium">
                  Stok Minimum
                </Label>
              </div>
              <Input
                id="minimumStock"
                type="number"
                step="0.01"
                value={formData.minimumStock}
                onChange={(e) => handleChange("minimumStock", e.target.value)}
                onBlur={() => handleBlur("minimumStock")}
                placeholder="Contoh: 10"
                className={
                  errors.minimumStock && touched.minimumStock
                    ? "border-destructive"
                    : ""
                }
              />
              {errors.minimumStock && touched.minimumStock && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.minimumStock}
                </p>
              )}
              <Alert className="bg-blue-50 border-blue-200 dark:bg-blue-900/10 dark:border-blue-900">
                <Info className="h-4 w-4 text-blue-600" />
                <AlertDescription className="text-blue-800 dark:text-blue-200 text-sm">
                  <strong>Default Global:</strong> Nilai ini akan digunakan sebagai default stok minimum saat menambah produk ke gudang baru.
                  Setiap gudang dapat memiliki threshold berbeda sesuai kebutuhannya.
                </AlertDescription>
              </Alert>
            </div>

            <Separator />

            {/* Tracking Options */}
            <div>
              <h4 className="text-sm font-medium mb-4 text-muted-foreground">
                Opsi Pelacakan
              </h4>
              <div className="space-y-3">
                {/* Batch Tracking */}
                <div
                  className={`flex items-start gap-3 rounded-lg border-2 p-4 transition-colors ${
                    formData.isBatchTracked
                      ? "border-primary bg-primary/5"
                      : "border-border hover:border-primary/50 hover:bg-muted/50"
                  }`}
                >
                  <Checkbox
                    id="isBatchTracked"
                    checked={formData.isBatchTracked}
                    onCheckedChange={(checked) => handleChange("isBatchTracked", checked)}
                    className="mt-0.5"
                  />
                  <div
                    className="flex-1 cursor-pointer"
                    onClick={() => handleChange("isBatchTracked", !formData.isBatchTracked)}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <PackageCheck className="h-4 w-4 text-primary" />
                      <Label
                        htmlFor="isBatchTracked"
                        className="cursor-pointer font-semibold text-base"
                      >
                        Pelacakan Batch/Lot
                      </Label>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Aktifkan untuk melacak nomor batch dan tanggal produksi setiap pengadaan
                    </p>
                  </div>
                </div>

                {/* Perishable */}
                <div
                  className={`flex items-start gap-3 rounded-lg border-2 p-4 transition-colors ${
                    formData.isPerishable
                      ? "border-orange-500 bg-orange-50 dark:bg-orange-900/10"
                      : "border-border hover:border-orange-500/50 hover:bg-muted/50"
                  }`}
                >
                  <Checkbox
                    id="isPerishable"
                    checked={formData.isPerishable}
                    onCheckedChange={(checked) => handleChange("isPerishable", checked)}
                    className="mt-0.5"
                  />
                  <div
                    className="flex-1 cursor-pointer"
                    onClick={() => handleChange("isPerishable", !formData.isPerishable)}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <Calendar className="h-4 w-4 text-orange-600" />
                      <Label
                        htmlFor="isPerishable"
                        className="cursor-pointer font-semibold text-base"
                      >
                        Produk Mudah Rusak/Kadaluarsa
                      </Label>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Aktifkan untuk produk yang memiliki tanggal kadaluarsa (makanan, obat, dll)
                    </p>
                  </div>
                </div>

                {formData.isPerishable && (
                  <Alert className="border-yellow-200 bg-yellow-50 dark:border-yellow-900 dark:bg-yellow-900/10">
                    <AlertCircle className="h-4 w-4 text-yellow-600" />
                    <AlertDescription className="text-yellow-800 dark:text-yellow-200">
                      <span className="font-semibold">Metode FEFO Aktif:</span> Produk dengan tanggal kadaluarsa terdekat akan keluar terlebih dahulu (First Expired, First Out)
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

          {/* Tab Navigation Buttons */}
          <div className="flex justify-between pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={goToPrevTab}
              className="gap-2"
            >
              <ChevronLeft className="h-4 w-4" />
              Kembali
            </Button>
            <Button
              type="button"
              onClick={goToNextTab}
              className="gap-2"
            >
              Lanjut ke Suppliers
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </TabsContent>

        {/* Tab 4: Suppliers */}
        <TabsContent value="suppliers" className="mt-6">
          {/* Suppliers Section */}
          <ProductSuppliersSection
            suppliers={suppliers}
            onSuppliersChange={setSuppliers}
            disabled={isLoading}
          />

          {/* Tab Navigation Buttons with Submit */}
          <div className="flex justify-between pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={goToPrevTab}
              className="gap-2"
            >
              <ChevronLeft className="h-4 w-4" />
              Kembali
            </Button>
            <div className="flex gap-3">
              {onCancel && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={onCancel}
                  disabled={isLoading}
                >
                  Batal
                </Button>
              )}
              <Button
                type="submit"
                disabled={isLoading}
                className="gap-2"
              >
                {isLoading ? (
                  <>
                    <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    Menyimpan...
                  </>
                ) : (
                  <>
                    <Save className="h-4 w-4" />
                    Simpan Produk
                  </>
                )}
              </Button>
            </div>
          </div>
        </TabsContent>
      </Tabs>

      {/* Form Actions - Mobile Sticky Footer */}
      <div className="fixed bottom-0 left-0 right-0 p-4 bg-background/95 backdrop-blur-sm border-t shadow-lg md:hidden z-50">
        <div className="flex gap-3 max-w-lg mx-auto">
          {onCancel && (
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={isLoading}
              className="flex-1"
            >
              Batal
            </Button>
          )}
          <Button
            type="submit"
            disabled={isLoading}
            className="flex-1"
          >
            {isLoading ? (
              <>
                <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Menyimpan...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Simpan Produk
              </>
            )}
          </Button>
        </div>
      </div>
    </form>
  );
}
