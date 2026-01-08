/**
 * Edit Warehouse Form Component
 *
 * Professional form for editing existing warehouses with:
 * - Pre-populated data from existing warehouse
 * - Real-time validation with touched state
 * - Organized sections with visual indicators
 * - Partial update support
 * - Responsive layout
 */

"use client";

import { useState } from "react";
import {
  Warehouse,
  MapPin,
  Phone,
  Settings,
  AlertCircle,
  Save,
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
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { useUpdateWarehouseMutation } from "@/store/services/warehouseApi";
import type {
  UpdateWarehouseRequest,
  WarehouseResponse,
  WarehouseType,
} from "@/types/warehouse.types";
import { WAREHOUSE_TYPES } from "@/types/warehouse.types";
import { toast } from "sonner";

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

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (
    field: keyof UpdateWarehouseRequest,
    value: string | boolean | undefined
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value === "" ? undefined : value,
    }));
    // Clear error when user types
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields
    if (!formData.code || formData.code.trim() === "") {
      newErrors.code = "Kode gudang wajib diisi";
    } else if (formData.code.length < 2) {
      newErrors.code = "Kode gudang minimal 2 karakter";
    }

    if (!formData.name || formData.name.trim() === "") {
      newErrors.name = "Nama gudang wajib diisi";
    } else if (formData.name.length < 3) {
      newErrors.name = "Nama gudang minimal 3 karakter";
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
    setTouched({
      code: true,
      name: true,
      type: true,
      email: true,
      capacity: true,
    });

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
      // Clean up empty fields and prepare update data
      const updateData: UpdateWarehouseRequest = {};

      // Only include fields that have changed
      if (formData.code !== warehouse.code)
        updateData.code = formData.code?.trim();
      if (formData.name !== warehouse.name)
        updateData.name = formData.name?.trim();
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
        toast("Tidak Ada Perubahan", {
          description: "Tidak ada data yang diubah",
        });
        return;
      }

      await updateWarehouse({
        id: warehouse.id,
        data: updateData,
      }).unwrap();

      toast.success("✓ Gudang Berhasil Diperbarui", {
        description: `${formData.name} telah diperbarui`,
      });

      onSuccess();
    } catch (error: any) {
      toast.error("Gagal Memperbarui Gudang", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          error?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Information */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Code */}
            <div className="space-y-2">
              <Label htmlFor="code" className="text-sm font-medium">
                Kode Gudang <span className="text-destructive">*</span>
              </Label>
              <Input
                id="code"
                value={formData.code || ""}
                onChange={(e) =>
                  handleChange("code", e.target.value.toUpperCase())
                }
                onBlur={() => handleBlur("code")}
                placeholder="Contoh: GDG-001"
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
                Kode unik untuk identifikasi gudang
              </p>
            </div>

            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name" className="text-sm font-medium">
                Nama Gudang <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                value={formData.name || ""}
                onChange={(e) => handleChange("name", e.target.value)}
                onBlur={() => handleBlur("name")}
                placeholder="Contoh: Gudang Pusat Jakarta"
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
                Nama lengkap gudang yang mudah dikenali
              </p>
            </div>

            {/* Type */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="type" className="text-sm font-medium">
                Tipe Gudang <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.type}
                onValueChange={(value) =>
                  handleChange("type", value as WarehouseType)
                }
              >
                <SelectTrigger
                  className={
                    errors.type && touched.type ? "w-full border-destructive" : "w-full"
                  }
                >
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
              {errors.type && touched.type && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.type}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Pilih kategori yang sesuai dengan fungsi gudang dalam sistem
                distribusi Anda (Gudang Utama, Cabang, atau Transit)
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Location Information */}
      <Card className="border-2">
        <CardContent>
          <div className="space-y-4">
            {/* Address */}
            <div className="space-y-2">
              <Label htmlFor="address" className="text-sm font-medium">
                Alamat
              </Label>
              <Textarea
                id="address"
                value={formData.address || ""}
                onChange={(e) => handleChange("address", e.target.value)}
                placeholder="Jl. Raya Industri No. 123"
                rows={3}
              />
              <p className="text-xs text-muted-foreground">
                Alamat lengkap lokasi gudang
              </p>
            </div>

            {/* City and Province */}
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="city" className="text-sm font-medium">
                  Kota
                </Label>
                <Input
                  id="city"
                  value={formData.city || ""}
                  onChange={(e) => handleChange("city", e.target.value)}
                  placeholder="Jakarta"
                />
                <p className="text-xs text-muted-foreground">
                  Kota lokasi gudang
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="province" className="text-sm font-medium">
                  Provinsi
                </Label>
                <Input
                  id="province"
                  value={formData.province || ""}
                  onChange={(e) => handleChange("province", e.target.value)}
                  placeholder="DKI Jakarta"
                />
                <p className="text-xs text-muted-foreground">
                  Provinsi lokasi gudang
                </p>
              </div>
            </div>

            {/* Postal Code */}
            <div className="space-y-2">
              <Label htmlFor="postalCode" className="text-sm font-medium">
                Kode Pos
              </Label>
              <Input
                id="postalCode"
                value={formData.postalCode || ""}
                onChange={(e) => handleChange("postalCode", e.target.value)}
                placeholder="12345"
              />
              <p className="text-xs text-muted-foreground">
                Kode pos area gudang
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Contact & Management */}
      <div className="grid gap-6 sm:grid-cols-2">
        {/* Contact Information */}
        <Card className="border-2">
          <CardContent>
            <div className="space-y-4">
              {/* Phone */}
              <div className="space-y-2">
                <Label htmlFor="phone" className="text-sm font-medium">
                  Telepon
                </Label>
                <Input
                  id="phone"
                  value={formData.phone || ""}
                  onChange={(e) => handleChange("phone", e.target.value)}
                  placeholder="021-1234567"
                />
                <p className="text-xs text-muted-foreground">
                  Nomor telepon gudang
                </p>
              </div>

              {/* Email */}
              <div className="space-y-2">
                <Label htmlFor="email" className="text-sm font-medium">
                  Email
                </Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email || ""}
                  onChange={(e) => handleChange("email", e.target.value)}
                  onBlur={() => handleBlur("email")}
                  placeholder="gudang@example.com"
                  className={
                    errors.email && touched.email ? "border-destructive" : ""
                  }
                />
                {errors.email && touched.email && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.email}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  Email kontak gudang
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Management Information */}
        <Card className="border-2">
          <CardContent>
            <div className="space-y-4">
              {/* Capacity */}
              <div className="space-y-2">
                <Label htmlFor="capacity" className="text-sm font-medium">
                  Kapasitas (m²)
                </Label>
                <Input
                  id="capacity"
                  type="number"
                  step="0.01"
                  min="0"
                  value={formData.capacity || ""}
                  onChange={(e) => handleChange("capacity", e.target.value)}
                  onBlur={() => handleBlur("capacity")}
                  placeholder="1000"
                  className={
                    errors.capacity && touched.capacity
                      ? "border-destructive"
                      : ""
                  }
                />
                {errors.capacity && touched.capacity && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.capacity}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  Kapasitas total dalam meter persegi
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Status & Options */}
      <Card className="border-2">
        <CardContent>
          {/* Active Status */}
          <div
            className={`flex items-start gap-3 rounded-lg border-2 p-4 transition-colors ${
              formData.isActive
                ? "border-green-500 bg-green-50 dark:bg-green-900/10"
                : "border-red-500 bg-red-50 dark:bg-red-900/10"
            }`}
          >
            <Checkbox
              id="isActive"
              checked={formData.isActive}
              onCheckedChange={(checked) =>
                handleChange("isActive", checked as boolean)
              }
              className="mt-0.5"
            />
            <div
              className="flex-1 cursor-pointer"
              onClick={() => handleChange("isActive", !formData.isActive)}
            >
              <div className="flex items-center gap-2 mb-1">
                <Warehouse
                  className={`h-4 w-4 ${
                    formData.isActive ? "text-green-600" : "text-red-600"
                  }`}
                />
                <Label
                  htmlFor="isActive"
                  className="cursor-pointer font-semibold text-base"
                >
                  Gudang Aktif
                </Label>
              </div>
              <p className="text-sm text-muted-foreground">
                {formData.isActive
                  ? "Gudang dapat digunakan untuk transaksi dan operasional"
                  : "Gudang tidak akan muncul dalam daftar operasional"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Form Actions */}
      <div className="flex justify-end gap-3 pt-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
          size="lg"
        >
          Batal
        </Button>
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
              Simpan Perubahan
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
