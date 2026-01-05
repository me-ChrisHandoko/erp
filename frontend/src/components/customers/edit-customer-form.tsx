/**
 * Edit Customer Form Component
 *
 * Professional form for editing existing customers with:
 * - Pre-filled data from existing customer
 * - Real-time validation
 * - Currency formatting for credit limit
 * - Status toggle (isActive)
 * - Responsive design
 */

"use client";

import { useState } from "react";
import {
  Save,
  AlertCircle,
  Mail,
  Calendar,
  ToggleLeft,
  Info,
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useUpdateCustomerMutation } from "@/store/services/customerApi";
import { toast } from "sonner";
import {
  CUSTOMER_TYPES,
  type UpdateCustomerRequest,
  type CustomerResponse,
  type CustomerType,
} from "@/types/customer.types";

interface EditCustomerFormProps {
  customer: CustomerResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditCustomerForm({
  customer,
  onSuccess,
  onCancel,
}: EditCustomerFormProps) {
  const [updateCustomer, { isLoading }] = useUpdateCustomerMutation();

  // Form state - initialize with customer data
  const [formData, setFormData] = useState<UpdateCustomerRequest>({
    code: customer.code,
    name: customer.name,
    customerType: customer.customerType,
    contactPerson: customer.contactPerson || "",
    phone: customer.phone || "",
    email: customer.email || "",
    address: customer.address || "",
    city: customer.city || "",
    province: customer.province || "",
    postalCode: customer.postalCode || "",
    npwp: customer.npwp || "",
    creditLimit: customer.creditLimit || "0",
    creditTermDays: customer.creditTermDays || customer.paymentTerm || 0,
    isActive: customer.isActive,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (field: keyof UpdateCustomerRequest, value: any) => {
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

    // Required fields (optional in UpdateRequest, but validate if provided)
    if (formData.code && formData.code.trim()) {
      if (formData.code.length < 2) {
        newErrors.code = "Kode pelanggan minimal 2 karakter";
      }
    }

    if (formData.name && formData.name.trim()) {
      if (formData.name.length < 3) {
        newErrors.name = "Nama pelanggan minimal 3 karakter";
      }
    }

    // Phone validation (optional but validate if provided)
    if (formData.phone && formData.phone.trim()) {
      const phoneRegex = /^(\+62|62|0)[0-9]{9,13}$/;
      if (!phoneRegex.test(formData.phone.replace(/[\s-]/g, ""))) {
        newErrors.phone = "Format telepon tidak valid (contoh: 08123456789)";
      }
    }

    // Email validation (optional but validate if provided)
    if (formData.email && formData.email.trim()) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email)) {
        newErrors.email = "Format email tidak valid";
      }
    }

    // Numeric validations
    if (formData.creditLimit !== undefined) {
      const creditLimit = parseFloat(formData.creditLimit);
      if (isNaN(creditLimit) || creditLimit < 0) {
        newErrors.creditLimit = "Limit kredit tidak boleh negatif";
      }
    }

    if (formData.creditTermDays !== undefined) {
      const creditTermDays = Number(formData.creditTermDays);
      if (isNaN(creditTermDays) || creditTermDays < 0) {
        newErrors.creditTermDays = "Jangka waktu tidak boleh negatif";
      }
    }

    setErrors(newErrors);

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
      const cleanedData: UpdateCustomerRequest = {
        isActive: formData.isActive,
        // Only include fields that have values
        ...(formData.code &&
          formData.code.trim() !== "" && { code: formData.code }),
        ...(formData.name &&
          formData.name.trim() !== "" && { name: formData.name }),
        ...(formData.customerType && { customerType: formData.customerType }),
        ...(formData.contactPerson &&
          formData.contactPerson.trim() !== "" && {
            contactPerson: formData.contactPerson,
          }),
        ...(formData.phone &&
          formData.phone.trim() !== "" && { phone: formData.phone }),
        ...(formData.email &&
          formData.email.trim() !== "" && { email: formData.email }),
        ...(formData.address &&
          formData.address.trim() !== "" && { address: formData.address }),
        ...(formData.city &&
          formData.city.trim() !== "" && { city: formData.city }),
        ...(formData.province &&
          formData.province.trim() !== "" && { province: formData.province }),
        ...(formData.postalCode &&
          formData.postalCode.trim() !== "" && {
            postalCode: formData.postalCode,
          }),
        ...(formData.npwp &&
          formData.npwp.trim() !== "" && { npwp: formData.npwp }),
        ...(formData.creditLimit &&
          formData.creditLimit !== "0" && {
            creditLimit: formData.creditLimit,
          }),
        ...(formData.creditTermDays &&
          formData.creditTermDays !== 0 && {
            creditTermDays: formData.creditTermDays,
          }),
      };

      const result = await updateCustomer({
        id: customer.id,
        data: cleanedData,
      }).unwrap();

      toast.success("Pelanggan Berhasil Diperbarui", {
        description: `${result.name} telah diperbarui`,
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      toast.error("Gagal Memperbarui Pelanggan", {
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
          <div className="grid gap-6 sm:grid-cols-3">
            {/* Code */}
            <div className="space-y-2">
              <Label htmlFor="code" className="text-sm font-medium">
                Kode Pelanggan <span className="text-destructive">*</span>
              </Label>
              <Input
                id="code"
                value={formData.code}
                onChange={(e) =>
                  handleChange("code", e.target.value.toUpperCase())
                }
                onBlur={() => handleBlur("code")}
                placeholder="Contoh: CUS-001"
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
                Kode unik untuk identifikasi pelanggan
              </p>
              {formData.code !== customer.code && (
                <Alert className="border-amber-500/50 bg-amber-50 dark:bg-amber-950/20">
                  <Info className="h-4 w-4 text-amber-600 dark:text-amber-500" />
                  <AlertDescription className="text-amber-800 dark:text-amber-400">
                    <strong>Perhatian:</strong> Perubahan kode pelanggan dapat
                    mempengaruhi referensi di transaksi. Pastikan tidak ada
                    transaksi aktif yang menggunakan kode lama.
                  </AlertDescription>
                </Alert>
              )}
            </div>

            {/* Name */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="name" className="text-sm font-medium">
                Nama Pelanggan <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => handleChange("name", e.target.value)}
                onBlur={() => handleBlur("name")}
                placeholder="Contoh: Toko Sembako Jaya"
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
                Nama lengkap pelanggan atau nama toko
              </p>
            </div>

            {/* Customer Type */}
            <div className="space-y-2">
              <Label htmlFor="customerType" className="text-sm font-medium">
                Tipe Pelanggan <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.customerType}
                onValueChange={(value) =>
                  handleChange("customerType", value as CustomerType)
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CUSTOMER_TYPES.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Kategori untuk harga dan term pembayaran
              </p>
            </div>

            {/* Status Toggle */}
            <div className="space-y-2 sm:col-span-2">
              <div className="flex items-start gap-3 rounded-lg border-2 p-4 transition-colors">
                <Checkbox
                  id="isActive"
                  checked={formData.isActive}
                  onCheckedChange={(checked) =>
                    handleChange("isActive", checked)
                  }
                  className="mt-0.5"
                />
                <div
                  className="flex-1 cursor-pointer"
                  onClick={() => handleChange("isActive", !formData.isActive)}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <ToggleLeft className="h-4 w-4 text-primary" />
                    <Label
                      htmlFor="isActive"
                      className="cursor-pointer font-semibold"
                    >
                      Status Aktif
                    </Label>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {formData.isActive
                      ? "Pelanggan aktif dan dapat bertransaksi"
                      : "Pelanggan nonaktif dan tidak dapat bertransaksi"}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Contact Information */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Contact Person */}
            <div className="space-y-2">
              <Label htmlFor="contactPerson" className="text-sm font-medium">
                Nama Kontak
              </Label>
              <Input
                id="contactPerson"
                value={formData.contactPerson}
                onChange={(e) => handleChange("contactPerson", e.target.value)}
                placeholder="Contoh: Budi Santoso"
              />
              <p className="text-xs text-muted-foreground">
                Person in charge untuk komunikasi bisnis
              </p>
            </div>

            {/* Phone */}
            <div className="space-y-2">
              <Label htmlFor="phone" className="text-sm font-medium">
                Telepon
              </Label>
              <Input
                id="phone"
                value={formData.phone}
                onChange={(e) => handleChange("phone", e.target.value)}
                onBlur={() => handleBlur("phone")}
                placeholder="Contoh: 08123456789"
                className={
                  errors.phone && touched.phone ? "border-destructive" : ""
                }
              />
              {errors.phone && touched.phone && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.phone}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Nomor telepon untuk konfirmasi pesanan
              </p>
            </div>

            {/* Email */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="email" className="text-sm font-medium">
                Email
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleChange("email", e.target.value)}
                  onBlur={() => handleBlur("email")}
                  placeholder="Contoh: customer@email.com"
                  className={`pl-9 ${
                    errors.email && touched.email ? "border-destructive" : ""
                  }`}
                />
              </div>
              {errors.email && touched.email && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.email}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Email untuk invoice dan komunikasi resmi
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Address */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Address */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="address" className="text-sm font-medium">
                Alamat Lengkap
              </Label>
              <Textarea
                id="address"
                value={formData.address}
                onChange={(e) => handleChange("address", e.target.value)}
                placeholder="Contoh: Jl. Sudirman No. 123"
                rows={3}
                className="resize-none"
              />
            </div>

            {/* City */}
            <div className="space-y-2">
              <Label htmlFor="city" className="text-sm font-medium">
                Kota/Kabupaten
              </Label>
              <Input
                id="city"
                value={formData.city}
                onChange={(e) => handleChange("city", e.target.value)}
                placeholder="Contoh: Jakarta Selatan"
              />
            </div>

            {/* Province */}
            <div className="space-y-2">
              <Label htmlFor="province" className="text-sm font-medium">
                Provinsi
              </Label>
              <Input
                id="province"
                value={formData.province}
                onChange={(e) => handleChange("province", e.target.value)}
                placeholder="Contoh: DKI Jakarta"
              />
            </div>

            {/* Postal Code */}
            <div className="space-y-2">
              <Label htmlFor="postalCode" className="text-sm font-medium">
                Kode Pos
              </Label>
              <Input
                id="postalCode"
                value={formData.postalCode}
                onChange={(e) => handleChange("postalCode", e.target.value)}
                placeholder="Contoh: 12345"
                maxLength={5}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Business Terms */}
      <Card className="border-2">
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* NPWP */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="npwp" className="text-sm font-medium">
                NPWP
              </Label>
              <Input
                id="npwp"
                value={formData.npwp}
                onChange={(e) => handleChange("npwp", e.target.value)}
                placeholder="Contoh: 12.345.678.9-012.345"
                maxLength={20}
              />
              <p className="text-xs text-muted-foreground">
                Nomor Pokok Wajib Pajak untuk PKP
              </p>
            </div>

            {/* Credit Limit */}
            <div className="space-y-2">
              <Label htmlFor="creditLimit" className="text-sm font-medium">
                Limit Kredit
              </Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  Rp
                </span>
                <Input
                  id="creditLimit"
                  type="number"
                  step="0.01"
                  value={formData.creditLimit}
                  onChange={(e) => handleChange("creditLimit", e.target.value)}
                  onBlur={() => handleBlur("creditLimit")}
                  placeholder="0"
                  className={`pl-10 ${
                    errors.creditLimit && touched.creditLimit
                      ? "border-destructive"
                      : ""
                  }`}
                />
              </div>
              {errors.creditLimit && touched.creditLimit && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.creditLimit}
                </p>
              )}
              <p className="text-xs font-medium text-primary">
                {formatCurrency(formData.creditLimit || "0")}
              </p>
              {parseFloat(formData.creditLimit || "0") !==
                parseFloat(customer.creditLimit || "0") && (
                <Alert className="border-blue-500/50 bg-blue-50 dark:bg-blue-950/20">
                  <Info className="h-4 w-4 text-blue-600 dark:text-blue-500" />
                  <AlertDescription className="text-blue-800 dark:text-blue-400">
                    Limit kredit diubah dari{" "}
                    <strong>
                      {formatCurrency(customer.creditLimit || "0")}
                    </strong>{" "}
                    menjadi{" "}
                    <strong>
                      {formatCurrency(formData.creditLimit || "0")}
                    </strong>
                  </AlertDescription>
                </Alert>
              )}
            </div>

            {/* Credit Term Days */}
            <div className="space-y-2">
              <Label htmlFor="creditTermDays" className="text-sm font-medium">
                Jangka Waktu Kredit
              </Label>
              <div className="relative">
                <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="creditTermDays"
                  type="number"
                  value={formData.creditTermDays || ""}
                  onChange={(e) => {
                    const value = e.target.value === "" ? 0 : parseInt(e.target.value);
                    handleChange("creditTermDays", isNaN(value) ? 0 : value);
                  }}
                  onBlur={() => handleBlur("creditTermDays")}
                  placeholder="0"
                  className={`pl-9 ${
                    errors.creditTermDays && touched.creditTermDays
                      ? "border-destructive"
                      : ""
                  }`}
                />
              </div>
              {errors.creditTermDays && touched.creditTermDays && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.creditTermDays}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Jangka waktu pembayaran (hari)
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
