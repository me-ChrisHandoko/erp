/**
 * Create Supplier Form Component
 *
 * Professional form for creating new suppliers with:
 * - Real-time validation matching backend schema
 * - Currency formatting
 * - Indonesian phone number validation
 * - NPWP (tax ID) validation
 * - PKP status checkbox
 * - Responsive layout
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
import { useCreateSupplierMutation } from "@/store/services/supplierApi";
import { toast } from "sonner";
import {
  SUPPLIER_TYPES,
  INDONESIAN_PROVINCES,
  type CreateSupplierRequest,
} from "@/types/supplier.types";

interface CreateSupplierFormProps {
  onSuccess?: (supplierId?: string) => void;
  onCancel?: () => void;
}

export function CreateSupplierForm({
  onSuccess,
  onCancel,
}: CreateSupplierFormProps) {
  const [createSupplier, { isLoading }] = useCreateSupplierMutation();

  // Form state - matching backend CreateSupplierRequest exactly
  const [formData, setFormData] = useState<CreateSupplierRequest>({
    code: "",
    name: "",
    type: undefined,
    phone: undefined,
    email: undefined,
    address: undefined,
    city: undefined,
    province: undefined,
    postalCode: undefined,
    npwp: undefined,
    isPKP: false,
    contactPerson: undefined,
    contactPhone: undefined,
    paymentTerm: undefined,
    creditLimit: undefined,
    notes: undefined,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (field: keyof CreateSupplierRequest, value: any) => {
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

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields
    if (!formData.code.trim()) {
      newErrors.code = "Kode supplier wajib diisi";
    } else if (formData.code.length < 1 || formData.code.length > 100) {
      newErrors.code = "Kode supplier harus 1-100 karakter";
    }

    if (!formData.name.trim()) {
      newErrors.name = "Nama supplier wajib diisi";
    } else if (formData.name.length < 2 || formData.name.length > 255) {
      newErrors.name = "Nama supplier harus 2-255 karakter";
    }

    // Email validation (if provided)
    if (formData.email && formData.email.trim()) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email)) {
        newErrors.email = "Format email tidak valid";
      }
    }

    // Phone validation (if provided) - Indonesian format
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

    // NPWP validation (if provided) - 15-16 digits
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
    setTouched({
      code: true,
      name: true,
      email: !!formData.email,
      phone: !!formData.phone,
      contactPhone: !!formData.contactPhone,
      npwp: !!formData.npwp,
      paymentTerm: formData.paymentTerm !== undefined,
      creditLimit: !!formData.creditLimit,
    });

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
      const cleanedData: CreateSupplierRequest = {
        code: formData.code,
        name: formData.name,
        ...(formData.type && { type: formData.type }),
        ...(formData.phone && { phone: formData.phone }),
        ...(formData.email && { email: formData.email }),
        ...(formData.address && { address: formData.address }),
        ...(formData.city && { city: formData.city }),
        ...(formData.province && { province: formData.province }),
        ...(formData.postalCode && { postalCode: formData.postalCode }),
        ...(formData.npwp && { npwp: formData.npwp }),
        ...(formData.isPKP !== undefined && { isPKP: formData.isPKP }),
        ...(formData.contactPerson && {
          contactPerson: formData.contactPerson,
        }),
        ...(formData.contactPhone && { contactPhone: formData.contactPhone }),
        ...(formData.paymentTerm !== undefined && {
          paymentTerm: formData.paymentTerm,
        }),
        ...(formData.creditLimit && { creditLimit: formData.creditLimit }),
        ...(formData.notes && { notes: formData.notes }),
      };

      const result = await createSupplier(cleanedData).unwrap();

      toast.success("✓ Supplier Berhasil Dibuat", {
        description: `${formData.name} telah ditambahkan ke daftar supplier`,
      });

      // Always call onSuccess to close the dialog
      if (onSuccess) {
        onSuccess(result?.id || "");
      }
    } catch (error: any) {
      console.error("❌ Create supplier error:", error);
      toast.error("Gagal Membuat Supplier", {
        description:
          error?.data?.error ||
          error?.data?.message ||
          error?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Informasi Dasar */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Code - Required */}
            <div className="space-y-2">
              <Label htmlFor="code">
                Kode Supplier <span className="text-destructive">*</span>
              </Label>
              <Input
                id="code"
                value={formData.code}
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

            {/* Name - Required */}
            <div className="space-y-2">
              <Label htmlFor="name">
                Nama Supplier <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                value={formData.name}
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

          {/* Type - Optional */}
          <div className="space-y-2 mt-6">
            <Label htmlFor="type">Tipe Supplier</Label>
            <Select
              value={formData.type || ""}
              onValueChange={(value) =>
                handleChange("type", value || undefined)
              }
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Pilih tipe supplier" />
              </SelectTrigger>
              <SelectContent>
                {SUPPLIER_TYPES.map((type) => (
                  <SelectItem key={type} value={type}>
                    {type.charAt(0) + type.slice(1).toLowerCase()}
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
              Simpan Supplier
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
