/**
 * Create Warehouse Form Component
 *
 * Form for creating new warehouses with:
 * - Basic information (code, name, type)
 * - Location details (address, city, province, postal code)
 * - Contact information (phone, email)
 * - Management details (capacity)
 * - Validation and error handling
 */

"use client";

import { useState } from "react";
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
import { useCreateWarehouseMutation } from "@/store/services/warehouseApi";
import type {
  CreateWarehouseRequest,
  WarehouseType,
} from "@/types/warehouse.types";
import { WAREHOUSE_TYPES } from "@/types/warehouse.types";
import { Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface CreateWarehouseFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

export function CreateWarehouseForm({
  onSuccess,
  onCancel,
}: CreateWarehouseFormProps) {
  const { toast } = useToast();
  const [createWarehouse, { isLoading }] = useCreateWarehouseMutation();

  // Form state
  const [formData, setFormData] = useState<CreateWarehouseRequest>({
    code: "",
    name: "",
    type: "MAIN",
  });

  // Error state
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Handle input change
  const handleChange = (
    field: keyof CreateWarehouseRequest,
    value: string | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value || undefined,
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

    // Capacity validation (must be positive number if provided)
    if (formData.capacity && formData.capacity.trim() !== "") {
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
      // Clean up empty fields before submitting
      const cleanedData: CreateWarehouseRequest = {
        code: formData.code.trim(),
        name: formData.name.trim(),
        type: formData.type,
      };

      // Add optional fields only if they have values
      if (formData.address?.trim())
        cleanedData.address = formData.address.trim();
      if (formData.city?.trim()) cleanedData.city = formData.city.trim();
      if (formData.province?.trim())
        cleanedData.province = formData.province.trim();
      if (formData.postalCode?.trim())
        cleanedData.postalCode = formData.postalCode.trim();
      if (formData.phone?.trim()) cleanedData.phone = formData.phone.trim();
      if (formData.email?.trim()) cleanedData.email = formData.email.trim();
      if (formData.managerID?.trim())
        cleanedData.managerID = formData.managerID.trim();
      if (formData.capacity?.trim())
        cleanedData.capacity = formData.capacity.trim();

      await createWarehouse(cleanedData).unwrap();

      toast({
        title: "Berhasil",
        description: "Gudang berhasil ditambahkan",
      });

      onSuccess();
    } catch (error: any) {
      console.error("Error creating warehouse:", error);

      // Handle specific error responses
      let errorMessage = "Gagal menambahkan gudang";

      if (error?.data?.message) {
        errorMessage = error.data.message;
      } else if (error?.status === 400) {
        errorMessage = "Data tidak valid. Periksa kembali form Anda.";
      } else if (error?.status === 409) {
        errorMessage = "Kode gudang sudah digunakan";
      } else if (error?.status === 403) {
        errorMessage = "Anda tidak memiliki izin untuk menambahkan gudang";
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
            value={formData.code}
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
            value={formData.name}
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
      </div>

      {/* Section 2: Location Information */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h3 className="text-lg font-semibold">Informasi Lokasi</h3>
          <p className="text-sm text-muted-foreground">
            Detail alamat dan lokasi gudang (opsional)
          </p>
        </div>

        {/* Address */}
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

        {/* City and Province */}
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

        {/* Postal Code */}
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

        {/* Phone and Email */}
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

        {/* Capacity */}
        <div className="space-y-2">
          <Label htmlFor="capacity">
            Kapasitas (m²)
          </Label>
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

        {/* Manager ID - Placeholder for future implementation */}
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
          {isLoading ? "Menyimpan..." : "Simpan Gudang"}
        </Button>
      </div>
    </form>
  );
}
