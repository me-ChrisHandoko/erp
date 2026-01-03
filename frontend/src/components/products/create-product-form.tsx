/**
 * Create Product Form Component
 *
 * Professional form for creating new products with:
 * - Real-time validation
 * - Currency formatting
 * - Profit margin calculator
 * - Responsive layout
 */

"use client";

import { useState } from "react";
import {
  Package,
  DollarSign,
  Layers,
  Save,
  AlertCircle,
  Info,
  TrendingUp,
  PackageCheck,
  Calendar,
  Bell,
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
import { useCreateProductMutation } from "@/store/services/productApi";
import { toast } from "sonner";
import {
  PRODUCT_CATEGORIES,
  COMMON_UNITS,
  type CreateProductRequest,
} from "@/types/product.types";

interface CreateProductFormProps {
  onSuccess?: (productId: string) => void;
  onCancel?: () => void;
}

export function CreateProductForm({
  onSuccess,
  onCancel,
}: CreateProductFormProps) {
  const [createProduct, { isLoading }] = useCreateProductMutation();

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

  const handleChange = (field: keyof CreateProductRequest, value: any) => {
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

    // Debug log untuk melihat error apa yang terjadi
    if (Object.keys(newErrors).length > 0) {
      console.log("❌ Validation errors:", newErrors);
    }

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
      };

      const result = await createProduct(cleanedData).unwrap();

      toast.success("✓ Produk Berhasil Dibuat", {
        description: `${result.name} telah ditambahkan ke katalog`,
      });

      if (onSuccess) {
        onSuccess(result.id);
      }
    } catch (error: any) {
      toast.error("Gagal Membuat Produk", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          error?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  const profitData = calculateProfit();

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Information */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Code */}
            <div className="space-y-2">
              <Label htmlFor="code" className="text-sm font-medium">
                Kode Produk <span className="text-destructive">*</span>
              </Label>
              <Input
                id="code"
                value={formData.code}
                onChange={(e) =>
                  handleChange("code", e.target.value.toUpperCase())
                }
                onBlur={() => handleBlur("code")}
                placeholder="Contoh: BRS-001"
                className={
                  errors.code && touched.code ? "border-destructive" : ""
                }
              />
              {errors.code && touched.code && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.code}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Kode unik untuk identifikasi produk
              </p>
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

      {/* Pricing & Unit */}
      <Card className="border-2">
        <CardContent>
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
                  type="number"
                  step="0.01"
                  value={formData.baseCost}
                  onChange={(e) => handleChange("baseCost", e.target.value)}
                  onBlur={() => handleBlur("baseCost")}
                  placeholder="0"
                  className={`pl-10 ${
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
                {formatCurrency(formData.baseCost)}
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
                  type="number"
                  step="0.01"
                  value={formData.basePrice}
                  onChange={(e) => handleChange("basePrice", e.target.value)}
                  onBlur={() => handleBlur("basePrice")}
                  placeholder="0"
                  className={`pl-10 ${
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
                {formatCurrency(formData.basePrice)}
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

      {/* Stock & Tracking */}
      <Card className="border-2">
        <CardContent className="pt-6">
          <div className="space-y-6">
            {/* Section Title */}
            <div className="flex items-center gap-2 pb-2 border-b">
              <Layers className="h-5 w-5 text-primary" />
              <h3 className="text-lg font-semibold">Stok & Pelacakan</h3>
            </div>

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
                  Alert otomatis akan muncul ketika stok produk di bawah nilai ini
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

      {/* Form Actions */}
      <div className="flex justify-end gap-3 pt-2">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
            size="lg"
          >
            Batal
          </Button>
        )}
        <Button
          type="submit"
          disabled={isLoading}
          size="lg"
          className="min-w-[150px]"
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
    </form>
  );
}
