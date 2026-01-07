/**
 * Edit Supplier Form Component
 *
 * Form for editing existing suppliers with:
 * - Pre-filled with current supplier data
 * - Real-time validation matching backend schema
 * - All optional fields (partial update)
 * - Active status toggle
 */

"use client";

import { useState } from "react";
import { Save, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useUpdateSupplierMutation } from "@/store/services/supplierApi";
import { toast } from "sonner";
import {
  SUPPLIER_TYPES,
  INDONESIAN_PROVINCES,
  type SupplierResponse,
  type UpdateSupplierRequest,
} from "@/types/supplier.types";

interface EditSupplierFormProps {
  supplier: SupplierResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditSupplierForm({
  supplier,
  onSuccess,
  onCancel,
}: EditSupplierFormProps) {
  const [updateSupplier, { isLoading }] = useUpdateSupplierMutation();

  // Initialize form with supplier data - matching backend UpdateSupplierRequest
  const [formData, setFormData] = useState<UpdateSupplierRequest>({
    code: supplier.code,
    name: supplier.name,
    type: supplier.type,
    phone: supplier.phone,
    email: supplier.email,
    address: supplier.address,
    city: supplier.city,
    province: supplier.province,
    postalCode: supplier.postalCode,
    npwp: supplier.npwp,
    isPKP: supplier.isPKP,
    contactPerson: supplier.contactPerson,
    contactPhone: supplier.contactPhone,
    paymentTerm: supplier.paymentTerm,
    creditLimit: supplier.creditLimit,
    notes: supplier.notes,
    isActive: supplier.isActive,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (field: keyof UpdateSupplierRequest, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
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

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Code validation (if changed)
    if (formData.code !== undefined) {
      if (!formData.code.trim()) {
        newErrors.code = "Kode supplier wajib diisi";
      } else if (formData.code.length < 1 || formData.code.length > 100) {
        newErrors.code = "Kode supplier harus 1-100 karakter";
      }
    }

    // Name validation (if changed)
    if (formData.name !== undefined) {
      if (!formData.name.trim()) {
        newErrors.name = "Nama supplier wajib diisi";
      } else if (formData.name.length < 2 || formData.name.length > 255) {
        newErrors.name = "Nama supplier harus 2-255 karakter";
      }
    }

    // Email validation (if provided)
    if (formData.email && formData.email.trim()) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email)) {
        newErrors.email = "Format email tidak valid";
      }
    }

    // Phone validation (if provided)
    if (formData.phone && formData.phone.trim()) {
      const phoneRegex = /^(\+62|62|0)[0-9]{9,12}$/;
      if (!phoneRegex.test(formData.phone.replace(/[\s-]/g, ""))) {
        newErrors.phone =
          "Format telepon tidak valid (contoh: 08123456789 atau +628123456789)";
      }
    }

    // Contact phone validation (if provided)
    if (formData.contactPhone && formData.contactPhone.trim()) {
      const phoneRegex = /^(\+62|62|0)[0-9]{9,12}$/;
      if (!phoneRegex.test(formData.contactPhone.replace(/[\s-]/g, ""))) {
        newErrors.contactPhone =
          "Format telepon tidak valid (contoh: 08123456789 atau +628123456789)";
      }
    }

    // NPWP validation (if provided)
    if (formData.npwp && formData.npwp.trim()) {
      const npwpRegex = /^[0-9]{15,16}$/;
      if (!npwpRegex.test(formData.npwp.replace(/[.\-]/g, ""))) {
        newErrors.npwp = "NPWP harus 15-16 digit angka";
      }
    }

    // Payment term validation (if provided)
    if (formData.paymentTerm !== undefined && formData.paymentTerm < 0) {
      newErrors.paymentTerm = "Payment term tidak boleh negatif";
    }

    // Credit limit validation (if provided)
    if (formData.creditLimit) {
      const creditLimit = parseFloat(formData.creditLimit);
      if (isNaN(creditLimit) || creditLimit < 0) {
        newErrors.creditLimit = "Credit limit tidak boleh negatif";
      }
    }

    setErrors(newErrors);
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
      // Build update payload - only include changed/non-empty fields
      const updatePayload: UpdateSupplierRequest = {};

      // Include all fields that have values (partial update)
      if (formData.code !== supplier.code) updatePayload.code = formData.code;
      if (formData.name !== supplier.name) updatePayload.name = formData.name;
      if (formData.type !== supplier.type) updatePayload.type = formData.type;
      if (formData.phone !== supplier.phone)
        updatePayload.phone = formData.phone;
      if (formData.email !== supplier.email)
        updatePayload.email = formData.email;
      if (formData.address !== supplier.address)
        updatePayload.address = formData.address;
      if (formData.city !== supplier.city) updatePayload.city = formData.city;
      if (formData.province !== supplier.province)
        updatePayload.province = formData.province;
      if (formData.postalCode !== supplier.postalCode)
        updatePayload.postalCode = formData.postalCode;
      if (formData.npwp !== supplier.npwp) updatePayload.npwp = formData.npwp;
      if (formData.isPKP !== supplier.isPKP)
        updatePayload.isPKP = formData.isPKP;
      if (formData.contactPerson !== supplier.contactPerson)
        updatePayload.contactPerson = formData.contactPerson;
      if (formData.contactPhone !== supplier.contactPhone)
        updatePayload.contactPhone = formData.contactPhone;
      if (formData.paymentTerm !== supplier.paymentTerm)
        updatePayload.paymentTerm = formData.paymentTerm;
      if (formData.creditLimit !== supplier.creditLimit)
        updatePayload.creditLimit = formData.creditLimit;
      if (formData.notes !== supplier.notes)
        updatePayload.notes = formData.notes;
      if (formData.isActive !== supplier.isActive)
        updatePayload.isActive = formData.isActive;

      console.log("📤 Update payload:", updatePayload);

      // Check if there are any changes
      if (Object.keys(updatePayload).length === 0) {
        toast.info("Tidak Ada Perubahan", {
          description: "Tidak ada data yang diubah",
        });
        if (onSuccess) {
          onSuccess();
        }
        return;
      }

      await updateSupplier({ id: supplier.id, data: updatePayload }).unwrap();

      toast.success("✓ Supplier Berhasil Diperbarui", {
        description: `${formData.name} telah diperbarui`,
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      console.error("❌ Update supplier error:", error);
      console.error("❌ Error details:", {
        status: error?.status,
        data: error?.data,
        message: error?.message,
        originalError: error,
      });

      let errorMessage = "Terjadi kesalahan pada server";

      if (error?.data?.error) {
        errorMessage = error.data.error;
      } else if (error?.data?.message) {
        errorMessage = error.data.message;
      } else if (error?.message) {
        errorMessage = error.message;
      } else if (error?.status === "FETCH_ERROR") {
        errorMessage = "Tidak dapat terhubung ke server";
      } else if (error?.status) {
        errorMessage = `Error ${error.status}: Gagal memperbarui supplier`;
      }

      toast.error("Gagal Memperbarui Supplier", {
        description: errorMessage,
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Informasi Dasar */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Code */}
            <div className="space-y-2">
              <Label htmlFor="code">Kode Supplier</Label>
              <Input
                id="code"
                value={formData.code || ""}
                onChange={(e) =>
                  handleChange("code", e.target.value.toUpperCase())
                }
                onBlur={() => handleBlur("code")}
                placeholder="SUP-001"
                maxLength={100}
                className={
                  errors.code && touched.code ? "border-destructive" : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Kode unik untuk identifikasi supplier (otomatis huruf besar)
              </p>
              {errors.code && touched.code && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.code}
                </p>
              )}
            </div>

            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Nama Supplier</Label>
              <Input
                id="name"
                value={formData.name || ""}
                onChange={(e) => handleChange("name", e.target.value)}
                onBlur={() => handleBlur("name")}
                placeholder="PT Supplier Indonesia"
                maxLength={255}
                className={
                  errors.name && touched.name ? "border-destructive" : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Nama lengkap perusahaan supplier
              </p>
              {errors.name && touched.name && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.name}
                </p>
              )}
            </div>
          </div>

          {/* Type */}
          <div className="space-y-2">
            <Label htmlFor="type">Tipe Supplier</Label>
            <Select
              value={formData.type || ""}
              onValueChange={(value) =>
                handleChange("type", value || undefined)
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Pilih tipe supplier" />
              </SelectTrigger>
              <SelectContent>
                {SUPPLIER_TYPES.map((type) => (
                  <SelectItem key={type} value={type}>
                    {type}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Kategori jenis supplier (Manufacturer, Distributor, dll)
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Kontak */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Email */}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={formData.email || ""}
                onChange={(e) =>
                  handleChange("email", e.target.value || undefined)
                }
                onBlur={() => handleBlur("email")}
                placeholder="supplier@example.com"
                maxLength={255}
                className={
                  errors.email && touched.email ? "border-destructive" : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Email untuk korespondensi dengan supplier
              </p>
              {errors.email && touched.email && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.email}
                </p>
              )}
            </div>

            {/* Phone */}
            <div className="space-y-2">
              <Label htmlFor="phone">Telepon</Label>
              <Input
                id="phone"
                value={formData.phone || ""}
                onChange={(e) =>
                  handleChange("phone", e.target.value || undefined)
                }
                onBlur={() => handleBlur("phone")}
                placeholder="08123456789"
                maxLength={50}
                className={
                  errors.phone && touched.phone ? "border-destructive" : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Nomor telepon supplier (08xxx atau +62xxx)
              </p>
              {errors.phone && touched.phone && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.phone}
                </p>
              )}
            </div>

            {/* Contact Person */}
            <div className="space-y-2">
              <Label htmlFor="contactPerson">Nama PIC</Label>
              <Input
                id="contactPerson"
                value={formData.contactPerson || ""}
                onChange={(e) =>
                  handleChange("contactPerson", e.target.value || undefined)
                }
                placeholder="John Doe"
                maxLength={255}
              />
              <p className="text-xs text-muted-foreground">
                Person In Charge untuk komunikasi
              </p>
            </div>

            {/* Contact Phone */}
            <div className="space-y-2">
              <Label htmlFor="contactPhone">Telepon PIC</Label>
              <Input
                id="contactPhone"
                value={formData.contactPhone || ""}
                onChange={(e) =>
                  handleChange("contactPhone", e.target.value || undefined)
                }
                onBlur={() => handleBlur("contactPhone")}
                placeholder="08123456789"
                maxLength={50}
                className={
                  errors.contactPhone && touched.contactPhone
                    ? "border-destructive"
                    : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Nomor telepon PIC (08xxx atau +62xxx)
              </p>
              {errors.contactPhone && touched.contactPhone && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.contactPhone}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Alamat */}
      <Card className="border-2">
        <CardContent>
          <div className="space-y-6">
            {/* Address */}
            <div className="space-y-2">
              <Label htmlFor="address">Alamat Lengkap</Label>
              <Textarea
                id="address"
                value={formData.address || ""}
                onChange={(e) =>
                  handleChange("address", e.target.value || undefined)
                }
                placeholder="Jl. Contoh No. 123"
                rows={3}
              />
              <p className="text-xs text-muted-foreground">
                Alamat lengkap kantor atau gudang supplier
              </p>
            </div>

            <div className="grid gap-6 sm:grid-cols-3">
              {/* City */}
              <div className="space-y-2">
                <Label htmlFor="city">Kota</Label>
                <Input
                  id="city"
                  value={formData.city || ""}
                  onChange={(e) =>
                    handleChange("city", e.target.value || undefined)
                  }
                  placeholder="Jakarta"
                  maxLength={100}
                />
                <p className="text-xs text-muted-foreground">
                  Nama kota/kabupaten
                </p>
              </div>

              {/* Province */}
              <div className="space-y-2">
                <Label htmlFor="province">Provinsi</Label>
                <Select
                  value={formData.province || ""}
                  onValueChange={(value) =>
                    handleChange("province", value || undefined)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih provinsi" />
                  </SelectTrigger>
                  <SelectContent>
                    {INDONESIAN_PROVINCES.map((province) => (
                      <SelectItem key={province} value={province}>
                        {province}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Provinsi lokasi supplier
                </p>
              </div>

              {/* Postal Code */}
              <div className="space-y-2">
                <Label htmlFor="postalCode">Kode Pos</Label>
                <Input
                  id="postalCode"
                  value={formData.postalCode || ""}
                  onChange={(e) =>
                    handleChange("postalCode", e.target.value || undefined)
                  }
                  placeholder="12345"
                  maxLength={50}
                />
                <p className="text-xs text-muted-foreground">
                  Kode pos area supplier
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Pajak */}
      <Card className="border-2">
        <CardContent>
          <div className="space-y-6">
            {/* NPWP */}
            <div className="space-y-2">
              <Label htmlFor="npwp">NPWP</Label>
              <Input
                id="npwp"
                value={formData.npwp || ""}
                onChange={(e) =>
                  handleChange("npwp", e.target.value || undefined)
                }
                onBlur={() => handleBlur("npwp")}
                placeholder="01.234.567.8-901.000"
                maxLength={50}
                className={
                  errors.npwp && touched.npwp ? "border-destructive" : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Nomor Pokok Wajib Pajak (15-16 digit)
              </p>
              {errors.npwp && touched.npwp && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.npwp}
                </p>
              )}
            </div>

            {/* PKP Status */}
            <div className="flex items-center space-x-2">
              <Checkbox
                id="isPKP"
                checked={formData.isPKP || false}
                onCheckedChange={(checked) => handleChange("isPKP", checked)}
              />
              <Label
                htmlFor="isPKP"
                className="text-sm font-normal cursor-pointer"
              >
                Pengusaha Kena Pajak (PKP)
              </Label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Payment Terms */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Payment Term */}
            <div className="space-y-2">
              <Label htmlFor="paymentTerm">Termin Pembayaran (Hari)</Label>
              <Input
                id="paymentTerm"
                type="number"
                min="0"
                value={formData.paymentTerm ?? ""}
                onChange={(e) =>
                  handleChange(
                    "paymentTerm",
                    e.target.value ? parseInt(e.target.value) : undefined
                  )
                }
                onBlur={() => handleBlur("paymentTerm")}
                placeholder="30"
                className={
                  errors.paymentTerm && touched.paymentTerm
                    ? "border-destructive"
                    : ""
                }
              />
              <p className="text-xs text-muted-foreground">
                Jangka waktu pembayaran (0 = Cash, 30 = NET 30, dst)
              </p>
              {errors.paymentTerm && touched.paymentTerm && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.paymentTerm}
                </p>
              )}
            </div>

            {/* Credit Limit */}
            <div className="space-y-2">
              <Label htmlFor="creditLimit">Limit Kredit</Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  Rp
                </span>
                <Input
                  id="creditLimit"
                  type="number"
                  min="0"
                  step="0.01"
                  value={formData.creditLimit || ""}
                  onChange={(e) =>
                    handleChange("creditLimit", e.target.value || undefined)
                  }
                  onBlur={() => handleBlur("creditLimit")}
                  placeholder="0"
                  className={`pl-10 ${
                    errors.creditLimit && touched.creditLimit
                      ? "border-destructive"
                      : ""
                  }`}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                Maksimal utang yang diperbolehkan
                {formData.creditLimit &&
                  ` (${formatCurrency(formData.creditLimit)})`}
              </p>
              {errors.creditLimit && touched.creditLimit && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.creditLimit}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notes */}
      <Card className="border-2">
        <CardContent>
          <div className="space-y-2">
            <Label htmlFor="notes">Catatan</Label>
            <Textarea
              id="notes"
              value={formData.notes || ""}
              onChange={(e) =>
                handleChange("notes", e.target.value || undefined)
              }
              placeholder="Catatan tambahan tentang supplier..."
              rows={4}
            />
            <p className="text-xs text-muted-foreground">
              Informasi tambahan atau catatan khusus mengenai supplier
            </p>
          </div>
        </CardContent>
      </Card>

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
              onCheckedChange={(checked) => handleChange("isActive", checked)}
              className="mt-0.5"
            />
            <div
              className="flex-1 cursor-pointer"
              onClick={() => handleChange("isActive", !formData.isActive)}
            >
              <div className="flex items-center gap-2 mb-1">
                <Label
                  htmlFor="isActive"
                  className="cursor-pointer font-semibold text-base"
                >
                  Supplier Aktif
                </Label>
              </div>
              <p className="text-sm text-muted-foreground">
                {formData.isActive
                  ? "Supplier dapat digunakan dalam transaksi pembelian"
                  : "Supplier tidak akan muncul dalam daftar pembelian"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Form Actions */}
      <div className="flex justify-end gap-3">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel}>
            Batal
          </Button>
        )}
        <Button type="submit" disabled={isLoading}>
          {isLoading ? (
            <>
              <span className="mr-2">Menyimpan...</span>
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Simpan Perubahan
            </>
          )}
        </Button>
      </div>

      {/* Validation Summary */}
      {Object.keys(errors).length > 0 && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Mohon periksa kembali form. Ada {Object.keys(errors).length} field
            yang perlu diperbaiki.
          </AlertDescription>
        </Alert>
      )}
    </form>
  );
}
