/**
 * Edit Stock Settings Dialog
 *
 * Dialog for updating warehouse stock settings:
 * - Minimum stock threshold
 * - Maximum stock threshold
 * - Storage location (rack/zone)
 *
 * Note: This does NOT update actual stock quantity.
 * Quantity changes are done via inventory movements.
 */

"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { Loader2 } from "lucide-react";
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
import { useToast } from "@/hooks/use-toast";
import { useUpdateStockSettingsMutation } from "@/store/services/stockApi";
import type { WarehouseStockResponse, UpdateWarehouseStockRequest } from "@/types/stock.types";

interface EditStockSettingsDialogProps {
  stock: WarehouseStockResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface FormData {
  minimumStock: string;
  maximumStock: string;
  location: string;
}

export function EditStockSettingsDialog({
  stock,
  open,
  onOpenChange,
}: EditStockSettingsDialogProps) {
  const { toast } = useToast();
  const [updateStockSettings, { isLoading }] = useUpdateStockSettingsMutation();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
  } = useForm<FormData>();

  // Watch minimum stock to validate maximum stock
  const minimumStock = watch("minimumStock");

  // Reset form when stock changes
  useEffect(() => {
    if (stock) {
      reset({
        minimumStock: stock.minimumStock || "0",
        maximumStock: stock.maximumStock || "0",
        location: stock.location || "",
      });
    }
  }, [stock, reset]);

  const onSubmit = async (data: FormData) => {
    if (!stock) return;

    try {
      // Validate maximum >= minimum
      if (data.maximumStock && Number(data.maximumStock) < Number(data.minimumStock)) {
        toast({
          title: "Validasi Gagal",
          description: "Stok maksimum harus lebih besar atau sama dengan stok minimum",
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

      onOpenChange(false);
      reset();
    } catch (error: any) {
      console.error("Failed to update stock settings:", error);
      toast({
        title: "Gagal Memperbarui",
        description: error?.data?.message || "Terjadi kesalahan saat memperbarui pengaturan stok",
        variant: "destructive",
      });
    }
  };

  if (!stock) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Edit Pengaturan Stok</DialogTitle>
          <DialogDescription>
            Perbarui pengaturan stok untuk <strong>{stock.productName}</strong> di{" "}
            <strong>{stock.warehouseName}</strong>
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {/* Current Stock (Read-only) */}
          <div className="space-y-2">
            <Label>Stok Saat Ini</Label>
            <div className="p-3 bg-muted rounded-md">
              <p className="text-lg font-semibold">
                {Number(stock.quantity).toLocaleString("id-ID", {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 3,
                })}{" "}
                {stock.productUnit}
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Quantity tidak bisa diubah di sini. Gunakan inventory movements untuk mengubah stok.
              </p>
            </div>
          </div>

          {/* Minimum Stock */}
          <div className="space-y-2">
            <Label htmlFor="minimumStock">
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
            />
            {errors.minimumStock && (
              <p className="text-sm text-destructive">
                {errors.minimumStock.message}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              Threshold untuk alert stok rendah
            </p>
          </div>

          {/* Maximum Stock */}
          <div className="space-y-2">
            <Label htmlFor="maximumStock">Stok Maksimum</Label>
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
                  if (value && minimumStock && Number(value) < Number(minimumStock)) {
                    return "Stok maksimum harus >= stok minimum";
                  }
                  return true;
                },
              })}
              placeholder="0"
            />
            {errors.maximumStock && (
              <p className="text-sm text-destructive">
                {errors.maximumStock.message}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              Threshold untuk alert stok berlebih (opsional)
            </p>
          </div>

          {/* Location */}
          <div className="space-y-2">
            <Label htmlFor="location">Lokasi Penyimpanan</Label>
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
            />
            {errors.location && (
              <p className="text-sm text-destructive">
                {errors.location.message}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              Rak atau zona penyimpanan di gudang (opsional)
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Batal
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Simpan Perubahan
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
