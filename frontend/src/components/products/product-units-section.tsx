/**
 * Product Units Section Component
 *
 * Manages product unit conversions with:
 * - Visual conversion preview (e.g., 1 KARTON = 24 PCS)
 * - Add/remove additional units
 * - Buy/sell price per unit
 * - Real-time validation
 */

"use client";

import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Plus,
  Trash2,
  Package,
  ArrowRight,
  Calculator,
  AlertCircle,
  Info,
  Layers,
} from "lucide-react";
import { COMMON_UNITS } from "@/types/product.types";
import type { CreateProductUnitRequest } from "@/types/product.types";

// Extended unit type for internal use
export interface ProductUnitItem {
  id: string;
  unitName: string;
  conversionRate: string;
  isBaseUnit: boolean;
  buyPrice?: string;
  sellPrice?: string;
  barcode?: string;
  sku?: string;
  isNew?: boolean; // For tracking new units to be created
  isEdited?: boolean; // For tracking edited units
}

interface ProductUnitsSectionProps {
  units: ProductUnitItem[];
  onUnitsChange: (units: ProductUnitItem[]) => void;
  baseUnit: string;
  baseCost?: string;
  basePrice?: string;
  disabled?: boolean;
}

// Format number with thousand separators
function formatCurrency(value: string): string {
  const num = parseFloat(value);
  if (isNaN(num)) return "Rp 0";
  return `Rp ${num.toLocaleString("id-ID")}`;
}

// Parse currency input
function parseCurrencyInput(value: string): string {
  return value.replace(/[^\d]/g, "");
}

// Format number input with separators
function formatNumberInput(value: string): string {
  const num = parseCurrencyInput(value);
  if (!num) return "";
  return parseInt(num, 10).toLocaleString("id-ID");
}

export function ProductUnitsSection({
  units,
  onUnitsChange,
  baseUnit,
  baseCost,
  basePrice,
  disabled = false,
}: ProductUnitsSectionProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newUnit, setNewUnit] = useState<Partial<CreateProductUnitRequest>>({
    unitName: "",
    conversionRate: "",
    buyPrice: "",
    sellPrice: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Filter out base unit and get additional units
  const additionalUnits = useMemo(() => {
    return units.filter((u) => !u.isBaseUnit);
  }, [units]);

  // Get available units (exclude already added ones)
  const availableUnits = useMemo(() => {
    const usedUnits = units.map((u) => u.unitName);
    return COMMON_UNITS.filter((u) => !usedUnits.includes(u) && u !== baseUnit);
  }, [units, baseUnit]);

  // Calculate suggested prices based on conversion rate
  const suggestedPrices = useMemo(() => {
    if (!newUnit.conversionRate || !baseCost || !basePrice) {
      return { buyPrice: "", sellPrice: "" };
    }
    const rate = parseFloat(newUnit.conversionRate);
    if (isNaN(rate) || rate <= 0) {
      return { buyPrice: "", sellPrice: "" };
    }
    const cost = parseFloat(baseCost);
    const price = parseFloat(basePrice);
    return {
      buyPrice: isNaN(cost) ? "" : Math.round(cost * rate).toString(),
      sellPrice: isNaN(price) ? "" : Math.round(price * rate).toString(),
    };
  }, [newUnit.conversionRate, baseCost, basePrice]);

  // Validate new unit form
  const validateNewUnit = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!newUnit.unitName) {
      newErrors.unitName = "Pilih satuan";
    }

    if (!newUnit.conversionRate) {
      newErrors.conversionRate = "Masukkan nilai konversi";
    } else {
      const rate = parseFloat(newUnit.conversionRate);
      if (isNaN(rate) || rate <= 0) {
        newErrors.conversionRate = "Nilai harus lebih dari 0";
      }
      if (rate < 1) {
        newErrors.conversionRate = "Nilai konversi harus minimal 1";
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle add new unit
  const handleAddUnit = () => {
    if (!validateNewUnit()) return;

    const newUnitItem: ProductUnitItem = {
      id: `new-${Date.now()}`,
      unitName: newUnit.unitName!,
      conversionRate: newUnit.conversionRate!,
      isBaseUnit: false,
      buyPrice: newUnit.buyPrice || suggestedPrices.buyPrice,
      sellPrice: newUnit.sellPrice || suggestedPrices.sellPrice,
      isNew: true,
    };

    onUnitsChange([...units, newUnitItem]);
    setIsDialogOpen(false);
    setNewUnit({
      unitName: "",
      conversionRate: "",
      buyPrice: "",
      sellPrice: "",
    });
    setErrors({});
  };

  // Handle remove unit
  const handleRemoveUnit = (unitId: string) => {
    onUnitsChange(units.filter((u) => u.id !== unitId));
  };

  // Handle currency input change
  const handleCurrencyChange = (field: "buyPrice" | "sellPrice", value: string) => {
    const numericValue = parseCurrencyInput(value);
    setNewUnit((prev) => ({ ...prev, [field]: numericValue }));
  };

  // Apply suggested prices
  const applySuggestedPrices = () => {
    setNewUnit((prev) => ({
      ...prev,
      buyPrice: suggestedPrices.buyPrice,
      sellPrice: suggestedPrices.sellPrice,
    }));
  };

  return (
    <Card className="border-2">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Layers className="h-5 w-5 text-primary" />
            Satuan Konversi
          </CardTitle>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setIsDialogOpen(true)}
            disabled={disabled || availableUnits.length === 0}
            className="gap-2"
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Tambah Satuan</span>
          </Button>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {/* Informative Guide */}
        <Alert className="mb-4 bg-blue-50 border-blue-200 dark:bg-blue-900/10 dark:border-blue-800">
          <Info className="h-4 w-4 text-blue-600" />
          <AlertDescription className="text-blue-800 dark:text-blue-200">
            <div className="space-y-2">
              <p className="font-medium">Apa itu Satuan Konversi?</p>
              <p className="text-sm">
                Satuan konversi memungkinkan Anda menjual produk dalam kemasan berbeda.
                Misalnya, jika satuan dasar adalah <strong>PCS</strong>, Anda bisa menambahkan:
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm mt-2">
                <div className="flex items-center gap-2 p-2 rounded bg-white/50 dark:bg-white/5">
                  <Package className="h-4 w-4 text-blue-600 shrink-0" />
                  <span><strong>KARTON</strong> = 24 PCS</span>
                </div>
                <div className="flex items-center gap-2 p-2 rounded bg-white/50 dark:bg-white/5">
                  <Package className="h-4 w-4 text-blue-600 shrink-0" />
                  <span><strong>LUSIN</strong> = 12 PCS</span>
                </div>
              </div>
              <p className="text-xs text-blue-600 dark:text-blue-300 mt-2">
                💡 Harga per satuan akan dihitung otomatis berdasarkan nilai konversi
              </p>
            </div>
          </AlertDescription>
        </Alert>

        {/* Base Unit Display */}
        <div className="mb-4 p-3 rounded-lg bg-primary/5 border border-primary/20">
          <div className="flex items-center gap-2 mb-1">
            <Package className="h-4 w-4 text-primary" />
            <span className="font-medium">Satuan Dasar</span>
            <Badge variant="secondary" className="ml-auto">
              {baseUnit || "Belum dipilih"}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Satuan terkecil yang digunakan untuk stok dan perhitungan
          </p>
        </div>

        {/* Additional Units List */}
        {additionalUnits.length > 0 ? (
          <div className="space-y-3">
            <Label className="text-sm font-medium text-muted-foreground">
              Satuan Tambahan ({additionalUnits.length})
            </Label>

            {/* Desktop Table */}
            <div className="hidden md:block border rounded-lg overflow-hidden">
              <table className="w-full">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="text-left p-3 text-sm font-medium">Satuan</th>
                    <th className="text-center p-3 text-sm font-medium">Konversi</th>
                    <th className="text-right p-3 text-sm font-medium">Harga Beli</th>
                    <th className="text-right p-3 text-sm font-medium">Harga Jual</th>
                    <th className="text-center p-3 text-sm font-medium w-16">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {additionalUnits.map((unit) => (
                    <tr key={unit.id} className="hover:bg-muted/30 transition-colors">
                      <td className="p-3">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">{unit.unitName}</Badge>
                          {unit.isNew && (
                            <Badge variant="secondary" className="text-xs">Baru</Badge>
                          )}
                        </div>
                      </td>
                      <td className="p-3 text-center">
                        <div className="flex items-center justify-center gap-2 text-sm">
                          <span className="font-medium">1 {unit.unitName}</span>
                          <ArrowRight className="h-3 w-3 text-muted-foreground" />
                          <span className="font-semibold text-primary">
                            {unit.conversionRate} {baseUnit}
                          </span>
                        </div>
                      </td>
                      <td className="p-3 text-right font-mono text-sm">
                        {unit.buyPrice ? formatCurrency(unit.buyPrice) : "-"}
                      </td>
                      <td className="p-3 text-right font-mono text-sm">
                        {unit.sellPrice ? formatCurrency(unit.sellPrice) : "-"}
                      </td>
                      <td className="p-3 text-center">
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveUnit(unit.id)}
                          disabled={disabled}
                          className="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Mobile Cards */}
            <div className="md:hidden space-y-3">
              {additionalUnits.map((unit) => (
                <div
                  key={unit.id}
                  className="p-4 border rounded-lg bg-card hover:bg-muted/30 transition-colors"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-base px-3 py-1">
                        {unit.unitName}
                      </Badge>
                      {unit.isNew && (
                        <Badge variant="secondary" className="text-xs">Baru</Badge>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemoveUnit(unit.id)}
                      disabled={disabled}
                      className="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>

                  {/* Conversion Preview */}
                  <div className="flex items-center gap-2 p-2 rounded-md bg-primary/5 mb-3">
                    <Calculator className="h-4 w-4 text-primary" />
                    <span className="text-sm">
                      <span className="font-medium">1 {unit.unitName}</span>
                      {" = "}
                      <span className="font-semibold text-primary">
                        {unit.conversionRate} {baseUnit}
                      </span>
                    </span>
                  </div>

                  {/* Prices */}
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <span className="text-muted-foreground">Harga Beli:</span>
                      <div className="font-mono font-medium">
                        {unit.buyPrice ? formatCurrency(unit.buyPrice) : "-"}
                      </div>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Harga Jual:</span>
                      <div className="font-mono font-medium">
                        {unit.sellPrice ? formatCurrency(unit.sellPrice) : "-"}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="text-center py-6 border-2 border-dashed rounded-lg">
            <Package className="h-10 w-10 mx-auto text-muted-foreground/50 mb-3" />
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Belum Ada Satuan Konversi
            </p>
            <p className="text-xs text-muted-foreground mb-4 max-w-sm mx-auto">
              Tambahkan satuan konversi jika produk ini dijual dalam kemasan berbeda
              (contoh: KARTON, LUSIN, PACK)
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setIsDialogOpen(true)}
              disabled={disabled || availableUnits.length === 0}
              className="gap-2"
            >
              <Plus className="h-4 w-4" />
              Tambah Satuan Pertama
            </Button>
          </div>
        )}
      </CardContent>

      {/* Add Unit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Plus className="h-5 w-5" />
              Tambah Satuan Konversi
            </DialogTitle>
            <DialogDescription>
              Tambahkan satuan kemasan baru dengan nilai konversi ke satuan dasar
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Unit Selection */}
            <div className="space-y-2">
              <Label htmlFor="unitName">
                Satuan <span className="text-destructive">*</span>
              </Label>
              <Select
                value={newUnit.unitName}
                onValueChange={(value) =>
                  setNewUnit((prev) => ({ ...prev, unitName: value }))
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Pilih satuan" />
                </SelectTrigger>
                <SelectContent>
                  {availableUnits.map((unit) => (
                    <SelectItem key={unit} value={unit}>
                      {unit}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.unitName && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.unitName}
                </p>
              )}
            </div>

            {/* Conversion Rate */}
            <div className="space-y-2">
              <Label htmlFor="conversionRate">
                Nilai Konversi <span className="text-destructive">*</span>
              </Label>
              <div className="flex items-center gap-3 p-3 rounded-lg border bg-muted/30">
                <span className="text-sm font-medium whitespace-nowrap">
                  1 {newUnit.unitName || "SATUAN"}
                </span>
                <span className="text-muted-foreground">=</span>
                <Input
                  id="conversionRate"
                  type="number"
                  min="1"
                  step="1"
                  value={newUnit.conversionRate}
                  onChange={(e) =>
                    setNewUnit((prev) => ({
                      ...prev,
                      conversionRate: e.target.value,
                    }))
                  }
                  placeholder="Contoh: 24"
                  className={`flex-1 text-center font-mono ${errors.conversionRate ? "border-destructive" : ""}`}
                />
                <span className="text-sm font-medium whitespace-nowrap">{baseUnit}</span>
              </div>
              {errors.conversionRate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.conversionRate}
                </p>
              )}
            </div>

            {/* Conversion Preview */}
            {newUnit.unitName && newUnit.conversionRate && parseFloat(newUnit.conversionRate) > 0 && (
              <Alert className="bg-primary/5 border-primary/20">
                <Calculator className="h-4 w-4 text-primary" />
                <AlertDescription className="text-primary">
                  <span className="font-semibold">
                    1 {newUnit.unitName} = {newUnit.conversionRate} {baseUnit}
                  </span>
                </AlertDescription>
              </Alert>
            )}

            <Separator />

            {/* Prices Section */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label className="text-sm font-medium">
                  Harga per {newUnit.unitName || "Satuan"} (Opsional)
                </Label>
                {suggestedPrices.buyPrice && (
                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    onClick={applySuggestedPrices}
                    className="h-auto p-0 text-xs"
                  >
                    Gunakan harga otomatis
                  </Button>
                )}
              </div>

              {suggestedPrices.buyPrice && (
                <p className="text-xs text-muted-foreground">
                  Saran harga berdasarkan konversi: Beli {formatCurrency(suggestedPrices.buyPrice)}, Jual {formatCurrency(suggestedPrices.sellPrice)}
                </p>
              )}

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <Label htmlFor="buyPrice" className="text-xs">
                    Harga Beli
                  </Label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                      Rp
                    </span>
                    <Input
                      id="buyPrice"
                      type="text"
                      inputMode="numeric"
                      value={formatNumberInput(newUnit.buyPrice || "")}
                      onChange={(e) => handleCurrencyChange("buyPrice", e.target.value)}
                      placeholder="0"
                      className="pl-10 text-right font-mono"
                    />
                  </div>
                </div>
                <div className="space-y-1">
                  <Label htmlFor="sellPrice" className="text-xs">
                    Harga Jual
                  </Label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                      Rp
                    </span>
                    <Input
                      id="sellPrice"
                      type="text"
                      inputMode="numeric"
                      value={formatNumberInput(newUnit.sellPrice || "")}
                      onChange={(e) => handleCurrencyChange("sellPrice", e.target.value)}
                      placeholder="0"
                      className="pl-10 text-right font-mono"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsDialogOpen(false);
                setNewUnit({
                  unitName: "",
                  conversionRate: "",
                  buyPrice: "",
                  sellPrice: "",
                });
                setErrors({});
              }}
            >
              Batal
            </Button>
            <Button type="button" onClick={handleAddUnit}>
              Tambah Satuan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
