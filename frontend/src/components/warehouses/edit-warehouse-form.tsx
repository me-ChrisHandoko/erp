/**
 * Edit Warehouse Form Component
 *
 * Form for editing existing warehouses with:
 * - Pre-populated data from existing warehouse
 * - Same sections as create form
 * - Partial update support
 * - Validation and error handling
 */

"use client";

import { useState, useEffect } from "react";
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
import { useUpdateWarehouseMutation } from "@/store/services/warehouseApi";
import type {
  UpdateWarehouseRequest,
  WarehouseResponse,
  WarehouseType,
} from "@/types/warehouse.types";
import { WAREHOUSE_TYPES } from "@/types/warehouse.types";
import { Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Switch } from "@/components/ui/switch";

interface EditWarehouseFormProps {
  warehouse: WarehouseResponse;
  onSuccess: () => void;
  onCancel: () => void;
}

export function EditWarehouseForm({
  warehouse,
  onSuccess,
  onCancel,
}: EditWarehouseFormProps) {
  const { toast } = useToast();
  const [updateWarehouse, { isLoading }] = useUpdateWarehouseMutation();

  // Form state - Initialize with existing warehouse data
  const [formData, setFormData] = useState<UpdateWarehouseRequest>({
    code: warehouse.code,
    name: warehouse.name,
    type: warehouse.type,
    address: warehouse.address || undefined,
    city: warehouse.city || undefined,
    province: warehouse.province || undefined,
    postalCode: warehouse.postalCode || undefined,
    phone: warehouse.phone || undefined,
    email: warehouse.email || undefined,
    managerID: warehouse.managerID || undefined,
    capacity: warehouse.capacity || undefined,
    isActive: warehouse.isActive,
  });

  // Error state
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Handle input change
  const handleChange = (
    field: keyof UpdateWarehouseRequest,
    value: string | boolean | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value === "" ? undefined : value,
    }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  // Validate form
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields
    if (!formData.code || formData.code.trim() === "") {
      newErrors.code = "Kode gudang wajib diisi";
    }
    if (!formData.name || formData.name.trim() === "") {
      newErrors.name = "Nama gudang wajib diisi";
    }
    if (!formData.type) {
      newErrors.type = "Tipe gudang wajib dipilih";
    }

    // Email validation
    if (formData.email && formData.email.trim() !== "") {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email)) {
        newErrors.email = "Format email tidak valid";
      }
    }

    // Capacity validation
    if (
      formData.capacity &&
      typeof formData.capacity === "string" &&
      formData.capacity.trim() !== ""
    ) {
      const capacityNum = Number(formData.capacity);
      if (isNaN(capacityNum) || capacityNum <= 0) {
        newErrors.capacity = "Kapasitas harus berupa angka positif";
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle submit
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      toast({
        title: "Validasi Gagal",
        description: "Mohon periksa kembali form Anda",
        variant: "destructive",
      });
      return;
    }

    try {
      // Clean up empty fields and prepare update data
      const updateData: UpdateWarehouseRequest = {};

      // Only include fields that have changed
      if (formData.code !== warehouse.code) updateData.code = formData.code?.trim();
      if (formData.name !== warehouse.name) updateData.name = formData.name?.trim();
      if (formData.type !== warehouse.type) updateData.type = formData.type;
      if (formData.address !== warehouse.address)
        updateData.address = formData.address?.trim();
      if (formData.city !== warehouse.city)
        updateData.city = formData.city?.trim();
      if (formData.province !== warehouse.province)
        updateData.province = formData.province?.trim();
      if (formData.postalCode !== warehouse.postalCode)
        updateData.postalCode = formData.postalCode?.trim();
      if (formData.phone !== warehouse.phone)
        updateData.phone = formData.phone?.trim();
      if (formData.email !== warehouse.email)
        updateData.email = formData.email?.trim();
      if (formData.managerID !== warehouse.managerID)
        updateData.managerID = formData.managerID?.trim();
      if (formData.capacity !== warehouse.capacity)
        updateData.capacity = formData.capacity?.trim();
      if (formData.isActive !== warehouse.isActive)
        updateData.isActive = formData.isActive;

      // Check if there are any changes
      if (Object.keys(updateData).length === 0) {
        toast({
          title: "Tidak Ada Perubahan",
          description: "Tidak ada data yang diubah",
        });
        return;
      }

      await updateWarehouse({
        id: warehouse.id,
        data: updateData,
      }).unwrap();

      toast({
        title: "Berhasil",
        description: "Gudang berhasil diperbarui",
      });

      onSuccess();
    } catch (error: any) {
      console.error("Error updating warehouse:", error);

      let errorMessage = "Gagal memperbarui gudang";

      if (error?.data?.message) {
        errorMessage = error.data.message;
      } else if (error?.status === 400) {
        errorMessage = "Data tidak valid. Periksa kembali form Anda.";
      } else if (error?.status === 409) {
        errorMessage = "Kode gudang sudah digunakan";
      } else if (error?.status === 403) {
        errorMessage = "Anda tidak memiliki izin untuk mengubah gudang";
      } else if (error?.status === 404) {
        errorMessage = "Gudang tidak ditemukan";
      }

      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Section 1: Basic Information */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h3 className="text-lg font-semibold">Informasi Dasar</h3>
          <p className="text-sm text-muted-foreground">
            Data dasar gudang (wajib diisi)
          </p>
        </div>

        {/* Code */}
        <div className="space-y-2">
          <Label htmlFor="code">
            Kode Gudang <span className="text-red-500">*</span>
          </Label>
          <Input
            id="code"
            value={formData.code || ""}
            onChange={(e) => handleChange("code", e.target.value)}
            placeholder="Contoh: GDG-001"
            className={errors.code ? "border-red-500" : ""}
          />
          {errors.code && (
            <p className="text-sm text-red-500">{errors.code}</p>
          )}
        </div>

        {/* Name */}
        <div className="space-y-2">
          <Label htmlFor="name">
            Nama Gudang <span className="text-red-500">*</span>
          </Label>
          <Input
            id="name"
            value={formData.name || ""}
            onChange={(e) => handleChange("name", e.target.value)}
            placeholder="Contoh: Gudang Pusat Jakarta"
            className={errors.name ? "border-red-500" : ""}
          />
          {errors.name && (
            <p className="text-sm text-red-500">{errors.name}</p>
          )}
        </div>

        {/* Type */}
        <div className="space-y-2">
          <Label htmlFor="type">
            Tipe Gudang <span className="text-red-500">*</span>
          </Label>
          <Select
            value={formData.type}
            onValueChange={(value) =>
              handleChange("type", value as WarehouseType)
            }
          >
            <SelectTrigger className={errors.type ? "border-red-500" : ""}>
              <SelectValue placeholder="Pilih tipe gudang" />
            </SelectTrigger>
            <SelectContent>
              {WAREHOUSE_TYPES.map((type) => (
                <SelectItem key={type.value} value={type.value}>
                  {type.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.type && (
            <p className="text-sm text-red-500">{errors.type}</p>
          )}
        </div>

        {/* Status */}
        <div className="flex items-center justify-between rounded-lg border p-4">
          <div className="space-y-0.5">
            <Label htmlFor="isActive">Status Gudang</Label>
            <p className="text-sm text-muted-foreground">
              {formData.isActive
                ? "Gudang aktif dan dapat digunakan"
                : "Gudang nonaktif dan tidak dapat digunakan"}
            </p>
          </div>
          <Switch
            id="isActive"
            checked={formData.isActive}
            onCheckedChange={(checked) => handleChange("isActive", checked)}
          />
        </div>
      </div>

      {/* Section 2: Location Information */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h3 className="text-lg font-semibold">Informasi Lokasi</h3>
          <p className="text-sm text-muted-foreground">
            Detail alamat dan lokasi gudang (opsional)
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="address">Alamat</Label>
          <Textarea
            id="address"
            value={formData.address || ""}
            onChange={(e) => handleChange("address", e.target.value)}
            placeholder="Jl. Raya Industri No. 123"
            rows={3}
          />
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="city">Kota</Label>
            <Input
              id="city"
              value={formData.city || ""}
              onChange={(e) => handleChange("city", e.target.value)}
              placeholder="Jakarta"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="province">Provinsi</Label>
            <Input
              id="province"
              value={formData.province || ""}
              onChange={(e) => handleChange("province", e.target.value)}
              placeholder="DKI Jakarta"
            />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="postalCode">Kode Pos</Label>
          <Input
            id="postalCode"
            value={formData.postalCode || ""}
            onChange={(e) => handleChange("postalCode", e.target.value)}
            placeholder="12345"
          />
        </div>
      </div>

      {/* Section 3: Contact Information */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h3 className="text-lg font-semibold">Informasi Kontak</h3>
          <p className="text-sm text-muted-foreground">
            Kontak gudang (opsional)
          </p>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="phone">Telepon</Label>
            <Input
              id="phone"
              value={formData.phone || ""}
              onChange={(e) => handleChange("phone", e.target.value)}
              placeholder="021-1234567"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={formData.email || ""}
              onChange={(e) => handleChange("email", e.target.value)}
              placeholder="gudang@example.com"
              className={errors.email ? "border-red-500" : ""}
            />
            {errors.email && (
              <p className="text-sm text-red-500">{errors.email}</p>
            )}
          </div>
        </div>
      </div>

      {/* Section 4: Management Information */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h3 className="text-lg font-semibold">Informasi Manajemen</h3>
          <p className="text-sm text-muted-foreground">
            Detail manajemen gudang (opsional)
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="capacity">Kapasitas (m²)</Label>
          <Input
            id="capacity"
            type="number"
            step="0.01"
            min="0"
            value={formData.capacity || ""}
            onChange={(e) => handleChange("capacity", e.target.value)}
            placeholder="1000"
            className={errors.capacity ? "border-red-500" : ""}
          />
          {errors.capacity && (
            <p className="text-sm text-red-500">{errors.capacity}</p>
          )}
          <p className="text-xs text-muted-foreground">
            Kapasitas gudang dalam meter persegi
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="managerID">Manager</Label>
          <Input
            id="managerID"
            value={formData.managerID || ""}
            onChange={(e) => handleChange("managerID", e.target.value)}
            placeholder="ID Manager (untuk implementasi masa depan)"
            disabled
            className="bg-muted"
          />
          <p className="text-xs text-muted-foreground">
            Fitur pemilihan manager akan ditambahkan nanti
          </p>
        </div>
      </div>

      {/* Form Actions */}
      <div className="flex gap-3 border-t pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
          className="flex-1"
        >
          Batal
        </Button>
        <Button type="submit" disabled={isLoading} className="flex-1">
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {isLoading ? "Menyimpan..." : "Simpan Perubahan"}
        </Button>
      </div>
    </form>
  );
}
