/**
 * Edit Sales Payment Form Component
 *
 * Form for editing existing customer payment records.
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
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
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { useUpdateSalesPaymentMutation } from "@/store/services/salesPaymentApi";
import type { SalesPaymentResponse, UpdateSalesPaymentRequest, PaymentMethod } from "@/types/sales-payment.types";
import { PAYMENT_METHOD_LABELS } from "@/types/sales-payment.types";
import { Loader2 } from "lucide-react";

interface EditSalesPaymentFormProps {
  payment: SalesPaymentResponse;
}

export function EditSalesPaymentForm({ payment }: EditSalesPaymentFormProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [updatePayment, { isLoading }] = useUpdateSalesPaymentMutation();
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>(payment.paymentMethod);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<UpdateSalesPaymentRequest>();

  // Set default values
  useEffect(() => {
    setValue("paymentDate", payment.paymentDate.split('T')[0]);
    setValue("customerId", payment.customerId);
    setValue("invoiceId", payment.invoiceId);
    setValue("amount", payment.amount);
    setValue("reference", payment.reference || "");
    setValue("bankAccountId", payment.bankAccountId || "");
    setValue("checkNumber", payment.checkNumber || "");
    setValue("checkDate", payment.checkDate?.split('T')[0] || "");
    setValue("notes", payment.notes || "");
  }, [payment, setValue]);

  const onSubmit = async (data: UpdateSalesPaymentRequest) => {
    try {
      const payload: UpdateSalesPaymentRequest = {
        ...data,
        paymentMethod,
      };

      await updatePayment({ id: payment.id, data: payload }).unwrap();

      toast({
        title: "Pembayaran Berhasil Diupdate",
        description: "Perubahan pembayaran telah disimpan",
      });

      router.push(`/sales/payments/${payment.id}`);
      router.refresh();
    } catch (error: any) {
      toast({
        title: "Gagal Mengupdate Pembayaran",
        description: error?.data?.error?.message || "Terjadi kesalahan saat menyimpan perubahan",
        variant: "destructive",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2">
        {/* Payment Date */}
        <div className="space-y-2">
          <Label htmlFor="paymentDate">
            Tanggal Pembayaran <span className="text-destructive">*</span>
          </Label>
          <Input
            id="paymentDate"
            type="date"
            {...register("paymentDate", { required: "Tanggal pembayaran wajib diisi" })}
          />
          {errors.paymentDate && (
            <p className="text-sm text-destructive">{errors.paymentDate.message}</p>
          )}
        </div>

        {/* Customer ID */}
        <div className="space-y-2">
          <Label htmlFor="customerId">
            ID Pelanggan <span className="text-destructive">*</span>
          </Label>
          <Input
            id="customerId"
            {...register("customerId", { required: "ID pelanggan wajib diisi" })}
            disabled
            className="bg-muted"
          />
          {errors.customerId && (
            <p className="text-sm text-destructive">{errors.customerId.message}</p>
          )}
        </div>

        {/* Invoice ID */}
        <div className="space-y-2">
          <Label htmlFor="invoiceId">
            ID Invoice <span className="text-destructive">*</span>
          </Label>
          <Input
            id="invoiceId"
            {...register("invoiceId", { required: "ID invoice wajib diisi" })}
            disabled
            className="bg-muted"
          />
          {errors.invoiceId && (
            <p className="text-sm text-destructive">{errors.invoiceId.message}</p>
          )}
        </div>

        {/* Amount */}
        <div className="space-y-2">
          <Label htmlFor="amount">
            Jumlah Pembayaran <span className="text-destructive">*</span>
          </Label>
          <Input
            id="amount"
            type="number"
            step="0.01"
            {...register("amount", { required: "Jumlah pembayaran wajib diisi" })}
          />
          {errors.amount && (
            <p className="text-sm text-destructive">{errors.amount.message}</p>
          )}
        </div>

        {/* Payment Method */}
        <div className="space-y-2">
          <Label htmlFor="paymentMethod">
            Metode Pembayaran <span className="text-destructive">*</span>
          </Label>
          <Select value={paymentMethod} onValueChange={(value) => setPaymentMethod(value as PaymentMethod)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(PAYMENT_METHOD_LABELS).map(([key, label]) => (
                <SelectItem key={key} value={key}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Reference */}
        <div className="space-y-2">
          <Label htmlFor="reference">Referensi / No. Transaksi</Label>
          <Input
            id="reference"
            {...register("reference")}
            placeholder="Nomor referensi transfer, cek, dll"
          />
        </div>

        {/* Bank Account ID */}
        {(paymentMethod === "BANK_TRANSFER" || paymentMethod === "CHECK" || paymentMethod === "GIRO") && (
          <div className="space-y-2">
            <Label htmlFor="bankAccountId">ID Rekening Bank</Label>
            <Input
              id="bankAccountId"
              {...register("bankAccountId")}
              placeholder="ID rekening bank perusahaan"
            />
          </div>
        )}

        {/* Check Number */}
        {(paymentMethod === "CHECK" || paymentMethod === "GIRO") && (
          <div className="space-y-2">
            <Label htmlFor="checkNumber">Nomor Cek/Giro</Label>
            <Input
              id="checkNumber"
              {...register("checkNumber")}
              placeholder="Nomor cek atau giro"
            />
          </div>
        )}

        {/* Check Date */}
        {(paymentMethod === "CHECK" || paymentMethod === "GIRO") && (
          <div className="space-y-2">
            <Label htmlFor="checkDate">Tanggal Jatuh Tempo Cek/Giro</Label>
            <Input
              id="checkDate"
              type="date"
              {...register("checkDate")}
            />
          </div>
        )}
      </div>

      {/* Notes */}
      <div className="space-y-2">
        <Label htmlFor="notes">Catatan</Label>
        <Textarea
          id="notes"
          {...register("notes")}
          placeholder="Catatan tambahan (opsional)"
          rows={3}
        />
      </div>

      {/* Action Buttons */}
      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={() => router.back()}
          disabled={isLoading}
        >
          Batal
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Simpan Perubahan
        </Button>
      </div>
    </form>
  );
}
