/**
 * Edit Stock Settings Form Component
 *
 * Professional form for updating warehouse stock settings:
 * - Minimum stock threshold
 * - Maximum stock threshold
 * - Storage location (rack/zone)
 *
 * Note: This does NOT update actual stock quantity.
 * Quantity changes are done via inventory movements.
 */

"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import {
  Save,
  AlertCircle,
  Package,
  Warehouse,
  BarChart3,
  Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { useUpdateStockSettingsMutation } from "@/store/services/stockApi";
import { getStockStatus, getStockStatusColor } from "@/types/stock.types";
import type {
  WarehouseStockResponse,
  UpdateWarehouseStockRequest,
} from "@/types/stock.types";

interface EditStockSettingsFormProps {
  stock: WarehouseStockResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

interface FormData {
  minimumStock: string;
  maximumStock: string;
  location: string;
}

export function EditStockSettingsForm({
  stock,
  onSuccess,
  onCancel,
}: EditStockSettingsFormProps) {
  const { toast } = useToast();
  const [updateStockSettings, { isLoading }] = useUpdateStockSettingsMutation();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
  } = useForm<FormData>({
    defaultValues: {
      minimumStock: stock.minimumStock || "0",
      maximumStock: stock.maximumStock || "0",
      location: stock.location || "",
    },
  });

  // Watch minimum stock to validate maximum stock
  const minimumStock = watch("minimumStock");

  // Get stock status and color
  const stockStatus = getStockStatus(
    stock.quantity,
    stock.minimumStock,
    stock.maximumStock
  );
  const stockStatusColor = getStockStatusColor(stockStatus);

  // Extract text color from statusColor className
  const getTextColorClass = (statusColor: string) => {
    if (statusColor.includes("bg-green-"))
      return "text-green-700 dark:text-green-300";
    if (statusColor.includes("bg-yellow-"))
      return "text-yellow-700 dark:text-yellow-300";
    if (statusColor.includes("bg-red-"))
      return "text-red-700 dark:text-red-300";
    if (statusColor.includes("bg-gray-"))
      return "text-gray-700 dark:text-gray-300";
    return "text-green-700 dark:text-green-300"; // default
  };

  const textColorClass = getTextColorClass(stockStatusColor);
  const unitColorClass = textColorClass
    .replace("700", "600")
    .replace("300", "400");

  // Reset form when stock changes
  useEffect(() => {
    reset({
      minimumStock: stock.minimumStock || "0",
      maximumStock: stock.maximumStock || "0",
      location: stock.location || "",
    });
  }, [stock, reset]);

  const onSubmit = async (data: FormData) => {
    try {
      // Validate maximum >= minimum
      if (
        data.maximumStock &&
        Number(data.maximumStock) < Number(data.minimumStock)
      ) {
        toast({
          title: "Validasi Gagal",
          description:
            "Stok maksimum harus lebih besar atau sama dengan stok minimum",
          variant: "destructive",
        });
        return;
      }

      const updateData: UpdateWarehouseStockRequest = {
        minimumStock: data.minimumStock,
        maximumStock: data.maximumStock || undefined,
        location: data.location || undefined,
      };

      await updateStockSettings({
        id: stock.id,
        data: updateData,
      }).unwrap();

      toast({
        title: "Berhasil",
        description: "Pengaturan stok berhasil diperbarui",
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      console.error("Failed to update stock settings:", error);
      toast({
        title: "Gagal Memperbarui",
        description:
          error?.data?.message ||
          "Terjadi kesalahan saat memperbarui pengaturan stok",
        variant: "destructive",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Current Stock Information (Read-only) */}
      <Card className="border-2 overflow-hidden">
        <CardContent>
          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-3">
              {/* Product Info */}
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Package className="h-4 w-4" />
                  Produk
                </div>
                <div className="space-y-2">
                  <Badge variant="outline" className="font-mono text-xs">
                    {stock.productCode}
                  </Badge>
                  <p className="font-semibold text-base leading-tight">
                    {stock.productName}
                  </p>
                  {stock.productCategory && (
                    <p className="text-sm text-muted-foreground">
                      {stock.productCategory}
                    </p>
                  )}
                </div>
              </div>

              {/* Warehouse Info */}
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Warehouse className="h-4 w-4" />
                  Gudang
                </div>
                <div className="space-y-2">
                  {stock.warehouseCode && (
                    <Badge variant="outline" className="font-mono text-xs">
                      {stock.warehouseCode}
                    </Badge>
                  )}
                  <p className="font-semibold text-base">
                    {stock.warehouseName}
                  </p>
                </div>
              </div>

              {/* Current Stock */}
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <BarChart3 className="h-4 w-4" />
                  Stok Saat Ini
                </div>
                <div>
                  <p className={`text-2xl font-bold ${textColorClass}`}>
                    {Number(stock.quantity).toLocaleString("id-ID", {
                      minimumFractionDigits: 0,
                      maximumFractionDigits: 3,
                    })}
                  </p>
                  <p className={`text-sm font-medium ${unitColorClass} mt-1`}>
                    {stock.productUnit}
                  </p>
                </div>
              </div>
            </div>

            {/* Info Note - Full Width */}
            <p className="text-xs text-muted-foreground leading-relaxed">
              Quantity tidak bisa diubah di sini. Gunakan menu Penyesuaian,
              Transfer Gudang, atau Stock Opname untuk mengubah stok.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Stock Settings */}
      <Card className="border-2 overflow-hidden">
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            {/* Minimum Stock */}
            <div className="space-y-2">
              <Label htmlFor="minimumStock" className="text-sm font-medium">
                Stok Minimum <span className="text-destructive">*</span>
              </Label>
              <Input
                id="minimumStock"
                type="number"
                step="0.001"
                min="0"
                {...register("minimumStock", {
                  required: "Stok minimum wajib diisi",
                  min: {
                    value: 0,
                    message: "Stok minimum tidak boleh negatif",
                  },
                })}
                placeholder="0"
                className={errors.minimumStock ? "border-destructive" : ""}
              />
              {errors.minimumStock && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.minimumStock.message}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                <strong>Khusus Gudang Ini:</strong> Threshold untuk alert stok rendah di gudang ini.
                Dapat berbeda dengan default produk.
              </p>
            </div>

            {/* Maximum Stock */}
            <div className="space-y-2">
              <Label htmlFor="maximumStock" className="text-sm font-medium">
                Stok Maksimum
              </Label>
              <Input
                id="maximumStock"
                type="number"
                step="0.001"
                min="0"
                {...register("maximumStock", {
                  min: {
                    value: 0,
                    message: "Stok maksimum tidak boleh negatif",
                  },
                  validate: (value) => {
                    if (
                      value &&
                      minimumStock &&
                      Number(value) < Number(minimumStock)
                    ) {
                      return "Stok maksimum harus >= stok minimum";
                    }
                    return true;
                  },
                })}
                placeholder="0"
                className={errors.maximumStock ? "border-destructive" : ""}
              />
              {errors.maximumStock && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.maximumStock.message}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                <strong>Khusus Gudang Ini:</strong> Threshold untuk alert stok berlebih (opsional)
              </p>
            </div>

            {/* Location */}
            <div className="space-y-2">
              <Label htmlFor="location" className="text-sm font-medium">
                Lokasi Penyimpanan
              </Label>
              <Input
                id="location"
                {...register("location", {
                  maxLength: {
                    value: 100,
                    message: "Lokasi maksimal 100 karakter",
                  },
                })}
                placeholder="Contoh: RAK-A-01, ZONE-B"
                maxLength={100}
                className={errors.location ? "border-destructive" : ""}
              />
              {errors.location && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.location.message}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Rak atau zona penyimpanan di gudang (opsional)
              </p>
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
          className="min-w-37.5"
        >
          {isLoading ? (
            <>
              <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
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
    </form>
  );
}
