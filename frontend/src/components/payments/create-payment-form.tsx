/**
 * Create Payment Form Component
 *
 * Professional form for creating new supplier payments with:
 * - Real-time validation
 * - Currency formatting
 * - Supplier and PO selection
 * - Payment method selection
 */

"use client";

import { useState, useEffect } from "react";
import {
  DollarSign,
  Save,
  AlertCircle,
  Building2,
  Calendar,
  CreditCard,
  FileText,
  Landmark,
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useCreatePaymentMutation } from "@/store/services/paymentApi";
import { useListSuppliersQuery } from "@/store/services/supplierApi";
import { useListPurchaseOrdersQuery } from "@/store/services/purchaseOrderApi";
import { useGetBankAccountsQuery } from "@/store/services/companyApi";
import { toast } from "sonner";
import {
  PAYMENT_METHOD,
  PAYMENT_METHOD_LABELS,
  type CreatePaymentRequest,
} from "@/types/payment.types";

interface CreatePaymentFormProps {
  onSuccess?: (paymentId: string) => void;
  onCancel?: () => void;
}

export function CreatePaymentForm({
  onSuccess,
  onCancel,
}: CreatePaymentFormProps) {
  const [createPayment, { isLoading }] = useCreatePaymentMutation();

  // Fetch suppliers
  const { data: suppliersData } = useListSuppliersQuery({
    page: 1,
    pageSize: 100,
    sortBy: "name",
    sortOrder: "asc",
  });

  // Fetch purchase orders
  const { data: purchaseOrdersData } = useListPurchaseOrdersQuery({
    page: 1,
    pageSize: 100,
    sortBy: "poDate",
    sortOrder: "desc",
  });

  // Fetch bank accounts
  const { data: bankAccountsData } = useGetBankAccountsQuery();

  // Form state
  const [formData, setFormData] = useState<CreatePaymentRequest>({
    paymentDate: new Date().toISOString().split("T")[0],
    supplierId: "",
    amount: "",
    paymentMethod: PAYMENT_METHOD.BANK_TRANSFER,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Filter POs by selected supplier
  const [filteredPOs, setFilteredPOs] = useState<any[]>([]);

  useEffect(() => {
    if (formData.supplierId && purchaseOrdersData?.data) {
      const filtered = purchaseOrdersData.data.filter(
        (po: any) => po.supplierId === formData.supplierId
      );
      setFilteredPOs(filtered);
    } else {
      setFilteredPOs([]);
    }
  }, [formData.supplierId, purchaseOrdersData]);

  const handleChange = (field: keyof CreatePaymentRequest, value: any) => {
    setFormData((prev) => {
      const updated = { ...prev, [field]: value };

      // Clear PO selection when supplier changes
      if (field === "supplierId" && value !== prev.supplierId) {
        updated.purchaseOrderId = undefined;
      }

      return updated;
    });

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

    // Payment date
    if (!formData.paymentDate) {
      newErrors.paymentDate = "Tanggal pembayaran harus diisi";
    }

    // Supplier
    if (!formData.supplierId) {
      newErrors.supplierId = "Supplier harus dipilih";
    }

    // Amount
    if (!formData.amount || parseFloat(formData.amount) <= 0) {
      newErrors.amount = "Jumlah pembayaran harus lebih dari 0";
    }

    // Payment method
    if (!formData.paymentMethod) {
      newErrors.paymentMethod = "Metode pembayaran harus dipilih";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Mark all fields as touched
    const allTouched = Object.keys(formData).reduce(
      (acc, key) => ({ ...acc, [key]: true }),
      {}
    );
    setTouched(allTouched);

    // Validate
    if (!validate()) {
      toast.error("Mohon lengkapi semua field yang wajib diisi");
      return;
    }

    try {
      const result = await createPayment(formData).unwrap();
      toast.success("Pembayaran berhasil dicatat");
      onSuccess?.(result.id);
    } catch (error: any) {
      console.error("Failed to create payment:", error);
      toast.error(
        error?.data?.error || "Gagal mencatat pembayaran. Silakan coba lagi."
      );
    }
  };

  const suppliers = suppliersData?.data || [];
  const bankAccounts = bankAccountsData?.data || [];

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Informasi Pembayaran
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Row 1: Supplier & Payment Date - 2 columns */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Supplier */}
            <div className="space-y-2 w-full">
              <Label htmlFor="supplierId">
                Supplier <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.supplierId}
                onValueChange={(value) => handleChange("supplierId", value)}
                disabled={isLoading}
              >
                <SelectTrigger id="supplierId" className="w-full">
                  <Building2 className="mr-2 h-4 w-4 text-muted-foreground" />
                  <SelectValue placeholder="Pilih supplier" />
                </SelectTrigger>
                <SelectContent>
                  {suppliers.map((supplier: any) => (
                    <SelectItem key={supplier.id} value={supplier.id}>
                      {supplier.name} {supplier.code && `(${supplier.code})`}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {touched.supplierId && errors.supplierId && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.supplierId}
                </p>
              )}
            </div>

            {/* Payment Date */}
            <div className="space-y-2 w-full">
              <Label htmlFor="paymentDate">
                Tanggal Pembayaran <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <Calendar className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="paymentDate"
                  type="date"
                  value={formData.paymentDate}
                  onChange={(e) => handleChange("paymentDate", e.target.value)}
                  onBlur={() => handleBlur("paymentDate")}
                  className="pl-9 w-full"
                  disabled={isLoading}
                />
              </div>
              {touched.paymentDate && errors.paymentDate && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.paymentDate}
                </p>
              )}
            </div>
          </div>

          {/* Row 2: Amount & Payment Method - 2 columns */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Amount */}
            <div className="space-y-2 w-full">
              <Label htmlFor="amount">
                Jumlah Pembayaran <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <DollarSign className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="amount"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="0.00"
                  value={formData.amount}
                  onChange={(e) => handleChange("amount", e.target.value)}
                  onBlur={() => handleBlur("amount")}
                  className="pl-9 w-full"
                  disabled={isLoading}
                />
              </div>
              {touched.amount && errors.amount && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.amount}
                </p>
              )}
            </div>

            {/* Payment Method */}
            <div className="space-y-2 w-full">
              <Label htmlFor="paymentMethod">
                Metode Pembayaran <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.paymentMethod}
                onValueChange={(value) => handleChange("paymentMethod", value)}
                disabled={isLoading}
              >
                <SelectTrigger id="paymentMethod" className="w-full">
                  <CreditCard className="mr-2 h-4 w-4 text-muted-foreground" />
                  <SelectValue placeholder="Pilih metode" />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(PAYMENT_METHOD_LABELS).map(([key, label]) => (
                    <SelectItem key={key} value={key}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {touched.paymentMethod && errors.paymentMethod && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  {errors.paymentMethod}
                </p>
              )}
            </div>
          </div>

          {/* Row 3: Purchase Order, Bank Account & Reference - 3 columns */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Purchase Order (Optional) */}
            <div className="space-y-2 w-full">
              <Label htmlFor="purchaseOrderId">Purchase Order (Opsional)</Label>
              <Select
                value={formData.purchaseOrderId || ""}
                onValueChange={(value) =>
                  handleChange("purchaseOrderId", value || undefined)
                }
                disabled={isLoading || !formData.supplierId || filteredPOs.length === 0}
              >
                <SelectTrigger id="purchaseOrderId" className="w-full">
                  <FileText className="mr-2 h-4 w-4 text-muted-foreground" />
                  <SelectValue placeholder={
                    !formData.supplierId
                      ? "Pilih supplier terlebih dahulu"
                      : filteredPOs.length === 0
                      ? "Tidak ada PO untuk supplier ini"
                      : "Pilih PO (opsional)"
                  } />
                </SelectTrigger>
                <SelectContent>
                  {filteredPOs.map((po: any) => (
                    <SelectItem key={po.id} value={po.id}>
                      {po.poNumber} - Rp {Number(po.totalAmount).toLocaleString("id-ID")}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Bank Account (Optional) */}
            <div className="space-y-2 w-full">
              <Label htmlFor="bankAccountId">Rekening Bank (Opsional)</Label>
              <Select
                value={formData.bankAccountId || ""}
                onValueChange={(value) =>
                  handleChange("bankAccountId", value || undefined)
                }
                disabled={isLoading}
              >
                <SelectTrigger id="bankAccountId" className="w-full">
                  <Landmark className="mr-2 h-4 w-4 text-muted-foreground" />
                  <SelectValue placeholder="Pilih rekening (opsional)" />
                </SelectTrigger>
                <SelectContent>
                  {bankAccounts.map((bank: any) => (
                    <SelectItem key={bank.id} value={bank.id}>
                      {bank.bankName} - {bank.accountNumber}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Reference */}
            <div className="space-y-2 w-full">
              <Label htmlFor="reference">Referensi (Opsional)</Label>
              <Input
                id="reference"
                placeholder="Nomor referensi transaksi"
                value={formData.reference || ""}
                onChange={(e) => handleChange("reference", e.target.value)}
                disabled={isLoading}
                maxLength={100}
                className="w-full"
              />
            </div>
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Catatan (Opsional)</Label>
            <Textarea
              id="notes"
              placeholder="Tambahkan catatan untuk pembayaran ini"
              value={formData.notes || ""}
              onChange={(e) => handleChange("notes", e.target.value)}
              disabled={isLoading}
              rows={4}
            />
          </div>
        </CardContent>
      </Card>

      {/* Form Actions */}
      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          Batal
        </Button>
        <Button type="submit" disabled={isLoading}>
          <Save className="mr-2 h-4 w-4" />
          {isLoading ? "Menyimpan..." : "Simpan Pembayaran"}
        </Button>
      </div>
    </form>
  );
}
