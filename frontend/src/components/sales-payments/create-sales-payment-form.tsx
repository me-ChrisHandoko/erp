/**
 * Create Sales Payment Form Component (Enhanced)
 *
 * Form for recording customer payments for invoices with:
 * - Customer autocomplete search
 * - Invoice selection dropdown (unpaid invoices only)
 * - Bank account selection
 * - Auto-fill amount from invoice
 * - Payment method selection
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useForm, Controller } from "react-hook-form";
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
import { Combobox } from "@/components/ui/combobox";
import { useToast } from "@/hooks/use-toast";
import { useCreateSalesPaymentMutation } from "@/store/services/salesPaymentApi";
import { useListCustomersQuery } from "@/store/services/customerApi";
import { useListInvoicesQuery } from "@/store/services/invoiceApi";
import type { CreateSalesPaymentRequest, PaymentMethod } from "@/types/sales-payment.types";
import { PAYMENT_METHOD_LABELS } from "@/types/sales-payment.types";
import { Loader2, Info } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface FormData extends Omit<CreateSalesPaymentRequest, 'customerId' | 'invoiceId'> {
  customerId: string;
  invoiceId: string;
}

export function CreateSalesPaymentForm() {
  const router = useRouter();
  const { toast } = useToast();
  const [createPayment, { isLoading: isSaving }] = useCreateSalesPaymentMutation();
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>("CASH");
  const [selectedCustomerId, setSelectedCustomerId] = useState<string>("");
  const [selectedInvoiceId, setSelectedInvoiceId] = useState<string>("");
  const [selectedBankAccountId, setSelectedBankAccountId] = useState<string>("");

  const {
    register,
    handleSubmit,
    control,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormData>({
    defaultValues: {
      paymentDate: new Date().toISOString().split('T')[0],
      paymentMethod: "CASH",
    }
  });

  // Fetch customers for autocomplete
  const { data: customersData, isLoading: isLoadingCustomers } = useListCustomersQuery({
    page: 1,
    pageSize: 100,
    sortBy: 'name',
    sortOrder: 'asc',
  });

  // Fetch unpaid invoices for selected customer
  const { data: invoicesData, isLoading: isLoadingInvoices } = useListInvoicesQuery(
    {
      customerId: selectedCustomerId,
      status: 'UNPAID', // Only unpaid or partial paid invoices
      page: 1,
      pageSize: 50,
    },
    {
      skip: !selectedCustomerId,
    }
  );

  // Mock bank accounts (in real app, fetch from API)
  const bankAccounts = [
    { id: "bank1", name: "BCA - 1234567890", code: "BCA" },
    { id: "bank2", name: "Mandiri - 0987654321", code: "MANDIRI" },
    { id: "bank3", name: "BRI - 5555666677", code: "BRI" },
  ];

  // Prepare customer options for combobox
  const customerOptions = (customersData?.data || []).map((customer) => ({
    value: customer.id,
    label: `${customer.name} (${customer.code})`,
    search: `${customer.name} ${customer.code}`,
  }));

  // Prepare invoice options
  const invoiceOptions = (invoicesData?.data || []).map((invoice) => ({
    value: invoice.id,
    label: `${invoice.invoiceNumber} - Rp ${Number(invoice.unpaidAmount || 0).toLocaleString('id-ID')}`,
    invoice: invoice,
  }));

  // Auto-fill amount when invoice is selected
  useEffect(() => {
    if (selectedInvoiceId && invoicesData?.data) {
      const selectedInvoice = invoicesData.data.find(inv => inv.id === selectedInvoiceId);
      if (selectedInvoice && selectedInvoice.unpaidAmount) {
        setValue('amount', selectedInvoice.unpaidAmount);
      }
    }
  }, [selectedInvoiceId, invoicesData, setValue]);

  // Get selected invoice details
  const selectedInvoice = invoicesData?.data?.find(inv => inv.id === selectedInvoiceId);

  const onSubmit = async (data: FormData) => {
    if (!selectedCustomerId) {
      toast({
        title: "Pelanggan Belum Dipilih",
        description: "Silakan pilih pelanggan terlebih dahulu",
        variant: "destructive",
      });
      return;
    }

    if (!selectedInvoiceId) {
      toast({
        title: "Invoice Belum Dipilih",
        description: "Silakan pilih invoice terlebih dahulu",
        variant: "destructive",
      });
      return;
    }

    try {
      const payload: CreateSalesPaymentRequest = {
        paymentDate: data.paymentDate,
        customerId: selectedCustomerId,
        invoiceId: selectedInvoiceId,
        amount: data.amount,
        paymentMethod,
        reference: data.reference,
        bankAccountId: selectedBankAccountId || undefined,
        checkNumber: data.checkNumber,
        checkDate: data.checkDate,
        notes: data.notes,
      };

      await createPayment(payload).unwrap();

      toast({
        title: "Pembayaran Berhasil Dicatat",
        description: "Pembayaran telah berhasil disimpan",
      });

      router.push("/sales/payments");
      router.refresh();
    } catch (error: any) {
      toast({
        title: "Gagal Mencatat Pembayaran",
        description: error?.data?.error?.message || "Terjadi kesalahan saat menyimpan pembayaran",
        variant: "destructive",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Info Alert */}
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          Pilih pelanggan terlebih dahulu untuk melihat daftar invoice yang belum dibayar
        </AlertDescription>
      </Alert>

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

        {/* Customer Autocomplete */}
        <div className="space-y-2">
          <Label>
            Pelanggan <span className="text-destructive">*</span>
          </Label>
          <Combobox
            options={customerOptions}
            value={selectedCustomerId}
            onValueChange={(value) => {
              setSelectedCustomerId(value);
              setSelectedInvoiceId(""); // Reset invoice selection
              setValue('amount', ""); // Reset amount
            }}
            placeholder="Cari pelanggan..."
            emptyText="Pelanggan tidak ditemukan"
            searchPlaceholder="Cari nama atau kode..."
            disabled={isLoadingCustomers}
          />
          {!selectedCustomerId && (
            <p className="text-xs text-muted-foreground">
              Ketik untuk mencari pelanggan berdasarkan nama atau kode
            </p>
          )}
        </div>

        {/* Invoice Selection */}
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="invoiceId">
            Invoice <span className="text-destructive">*</span>
          </Label>
          <Select
            value={selectedInvoiceId}
            onValueChange={setSelectedInvoiceId}
            disabled={!selectedCustomerId || isLoadingInvoices}
          >
            <SelectTrigger>
              <SelectValue placeholder={
                !selectedCustomerId
                  ? "Pilih pelanggan terlebih dahulu"
                  : isLoadingInvoices
                  ? "Memuat invoice..."
                  : "Pilih invoice yang belum dibayar"
              } />
            </SelectTrigger>
            <SelectContent>
              {invoiceOptions.length === 0 ? (
                <div className="p-2 text-sm text-muted-foreground text-center">
                  Tidak ada invoice yang belum dibayar
                </div>
              ) : (
                invoiceOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))
              )}
            </SelectContent>
          </Select>
          {selectedInvoice && (
            <Alert className="mt-2">
              <Info className="h-4 w-4" />
              <AlertDescription>
                <div className="grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <span className="text-muted-foreground">Total Invoice:</span>{" "}
                    <span className="font-medium">
                      Rp {Number(selectedInvoice.totalAmount).toLocaleString('id-ID')}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Sudah Dibayar:</span>{" "}
                    <span className="font-medium">
                      Rp {Number(selectedInvoice.paidAmount || 0).toLocaleString('id-ID')}
                    </span>
                  </div>
                  <div className="col-span-2">
                    <span className="text-muted-foreground">Sisa Belum Dibayar:</span>{" "}
                    <span className="font-semibold text-orange-600">
                      Rp {Number(selectedInvoice.unpaidAmount || 0).toLocaleString('id-ID')}
                    </span>
                  </div>
                </div>
              </AlertDescription>
            </Alert>
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
            {...register("amount", {
              required: "Jumlah pembayaran wajib diisi",
              validate: (value) => {
                if (selectedInvoice && Number(value) > Number(selectedInvoice.unpaidAmount)) {
                  return "Jumlah pembayaran melebihi sisa yang belum dibayar";
                }
                return true;
              }
            })}
            placeholder="0.00"
          />
          {errors.amount && (
            <p className="text-sm text-destructive">{errors.amount.message}</p>
          )}
          <p className="text-xs text-muted-foreground">
            Otomatis terisi dari sisa invoice yang belum dibayar
          </p>
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

        {/* Bank Account Selection */}
        {(paymentMethod === "BANK_TRANSFER" || paymentMethod === "CHECK" || paymentMethod === "GIRO") && (
          <div className="space-y-2">
            <Label>Rekening Bank Perusahaan</Label>
            <Select value={selectedBankAccountId} onValueChange={setSelectedBankAccountId}>
              <SelectTrigger>
                <SelectValue placeholder="Pilih rekening bank" />
              </SelectTrigger>
              <SelectContent>
                {bankAccounts.map((bank) => (
                  <SelectItem key={bank.id} value={bank.id}>
                    {bank.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
          disabled={isSaving}
        >
          Batal
        </Button>
        <Button type="submit" disabled={isSaving || !selectedCustomerId || !selectedInvoiceId}>
          {isSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Simpan Pembayaran
        </Button>
      </div>
    </form>
  );
}
