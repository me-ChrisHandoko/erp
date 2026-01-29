/**
 * Invoice Settings Client Component
 *
 * Manages purchase invoice settings:
 * - Invoice Control Policy (ORDERED vs RECEIVED - 3-way matching)
 * - Invoice Tolerance Percentage (over-invoice tolerance)
 */

"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  FileText,
  Info,
  Save,
  Loader2,
  ShieldCheck,
  Percent,
  Package,
  ClipboardCheck,
  Truck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import {
  useGetCompanyQuery,
  useUpdateCompanyMutation,
} from "@/store/services/companyApi";
import { usePermissions } from "@/hooks/use-permissions";
import { toast } from "sonner";
import {
  INVOICE_CONTROL_POLICIES,
  INVOICE_CONTROL_POLICY_LABELS,
  type InvoiceControlPolicy,
} from "@/types/company.types";

// Form schema
const invoiceSettingsSchema = z.object({
  invoiceControlPolicy: z.enum(INVOICE_CONTROL_POLICIES, {
    message: "Pilih kebijakan kontrol faktur",
  }),
  invoiceTolerancePct: z
    .number()
    .min(0, "Toleransi minimal 0%")
    .max(100, "Toleransi maksimal 100%"),
});

type InvoiceSettingsFormData = z.infer<typeof invoiceSettingsSchema>;

export function InvoiceSettingsClient() {
  const permissions = usePermissions();
  const canEdit = permissions.canEdit("system-config");

  // Fetch company data
  const {
    data: company,
    isLoading,
    error,
    refetch,
  } = useGetCompanyQuery();

  // Update mutation
  const [updateCompany, { isLoading: isSaving }] = useUpdateCompanyMutation();

  // Form
  const form = useForm<InvoiceSettingsFormData>({
    resolver: zodResolver(invoiceSettingsSchema),
    defaultValues: {
      invoiceControlPolicy: "ORDERED",
      invoiceTolerancePct: 0,
    },
  });

  // Update form when company data loads
  useEffect(() => {
    if (company) {
      form.reset({
        invoiceControlPolicy: company.invoiceControlPolicy || "ORDERED",
        invoiceTolerancePct: company.invoiceTolerancePct || 0,
      });
    }
  }, [company, form]);

  const onSubmit = async (data: InvoiceSettingsFormData) => {
    try {
      await updateCompany({
        invoiceControlPolicy: data.invoiceControlPolicy,
        invoiceTolerancePct: data.invoiceTolerancePct,
      }).unwrap();
      toast.success("Pengaturan Disimpan", {
        description: "Pengaturan faktur pembelian berhasil diperbarui",
      });
    } catch (error: any) {
      toast.error("Gagal Menyimpan", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const watchPolicy = form.watch("invoiceControlPolicy");

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <ErrorDisplay
        error="Tidak dapat memuat pengaturan faktur pembelian"
        title="Gagal Memuat Data"
        onRetry={refetch}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Info Alert - Invoice Control Policy */}
      <Alert className="border-blue-200 bg-blue-50/50">
        <Info className="h-4 w-4 text-blue-600" />
        <AlertTitle className="text-blue-900">
          Kebijakan Kontrol Faktur (3-Way Matching)
        </AlertTitle>
        <AlertDescription className="text-blue-800">
          <p className="mb-3">
            Tentukan kapan faktur pembelian dapat dibuat berdasarkan referensi
            Purchase Order (PO):
          </p>
          <div className="grid gap-3 md:grid-cols-2">
            <div className="flex items-start gap-2 p-3 bg-white/50 rounded-lg border border-blue-200">
              <Package className="h-5 w-5 text-blue-600 mt-0.5" />
              <div>
                <p className="font-medium text-blue-900">ORDERED (Default)</p>
                <p className="text-sm text-blue-700">
                  Faktur dapat dibuat berdasarkan qty yang dipesan di PO, tanpa
                  harus menunggu penerimaan barang (GRN).
                </p>
              </div>
            </div>
            <div className="flex items-start gap-2 p-3 bg-white/50 rounded-lg border border-blue-200">
              <Truck className="h-5 w-5 text-green-600 mt-0.5" />
              <div>
                <p className="font-medium text-blue-900">
                  RECEIVED (3-Way Matching)
                </p>
                <p className="text-sm text-blue-700">
                  Faktur hanya dapat dibuat untuk qty yang sudah diterima (GRN).
                  Seperti SAP dan Odoo.
                </p>
              </div>
            </div>
          </div>
        </AlertDescription>
      </Alert>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          {/* Invoice Control Policy Card */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle className="text-lg">
                    Kebijakan Kontrol Faktur
                  </CardTitle>
                  <CardDescription>
                    Tentukan dasar pembuatan faktur pembelian
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <FormField
                control={form.control}
                name="invoiceControlPolicy"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Kebijakan Invoice</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      value={field.value}
                      disabled={!canEdit}
                    >
                      <FormControl>
                        <SelectTrigger className="w-full md:w-[400px]">
                          <SelectValue placeholder="Pilih kebijakan" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {INVOICE_CONTROL_POLICIES.map((policy) => (
                          <SelectItem key={policy} value={policy}>
                            <div className="flex items-center gap-2">
                              {policy === "ORDERED" ? (
                                <Package className="h-4 w-4 text-blue-600" />
                              ) : (
                                <Truck className="h-4 w-4 text-green-600" />
                              )}
                              <span>{INVOICE_CONTROL_POLICY_LABELS[policy]}</span>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {watchPolicy === "ORDERED" ? (
                        <span className="flex items-center gap-1">
                          <Badge variant="outline" className="text-blue-600">
                            ORDERED
                          </Badge>
                          Invoice dapat dibuat maksimal sebesar qty PO yang
                          belum di-invoice
                        </span>
                      ) : (
                        <span className="flex items-center gap-1">
                          <Badge variant="outline" className="text-green-600">
                            RECEIVED
                          </Badge>
                          Invoice hanya dapat dibuat maksimal sebesar qty yang
                          sudah diterima (GRN) dan belum di-invoice
                        </span>
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Visual Flow Diagram */}
              <div className="mt-6 p-4 bg-muted/50 rounded-lg">
                <p className="text-sm font-medium mb-3">Alur Pembuatan Faktur:</p>
                {watchPolicy === "ORDERED" ? (
                  <div className="flex flex-wrap items-center gap-2 text-sm">
                    <Badge className="bg-blue-100 text-blue-800">
                      <ClipboardCheck className="h-3 w-3 mr-1" />
                      PO (100 pcs)
                    </Badge>
                    <span className="text-muted-foreground">→</span>
                    <Badge className="bg-purple-100 text-purple-800">
                      <FileText className="h-3 w-3 mr-1" />
                      Invoice (max 100 pcs)
                    </Badge>
                    <span className="text-xs text-muted-foreground ml-2">
                      (Langsung bisa invoice tanpa GRN)
                    </span>
                  </div>
                ) : (
                  <div className="flex flex-wrap items-center gap-2 text-sm">
                    <Badge className="bg-blue-100 text-blue-800">
                      <ClipboardCheck className="h-3 w-3 mr-1" />
                      PO (100 pcs)
                    </Badge>
                    <span className="text-muted-foreground">→</span>
                    <Badge className="bg-green-100 text-green-800">
                      <Truck className="h-3 w-3 mr-1" />
                      GRN (80 pcs)
                    </Badge>
                    <span className="text-muted-foreground">→</span>
                    <Badge className="bg-purple-100 text-purple-800">
                      <FileText className="h-3 w-3 mr-1" />
                      Invoice (max 80 pcs)
                    </Badge>
                    <span className="text-xs text-muted-foreground ml-2">
                      (Harus GRN dulu)
                    </span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <Separator />

          {/* Invoice Tolerance Card */}
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Percent className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle className="text-lg">Toleransi Over-Invoice</CardTitle>
                  <CardDescription>
                    Izinkan invoice melebihi qty dengan batas toleransi tertentu
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <FormField
                control={form.control}
                name="invoiceTolerancePct"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Persentase Toleransi (%)</FormLabel>
                    <FormControl>
                      <div className="flex items-center gap-2 w-full md:w-[200px]">
                        <Input
                          type="number"
                          step="0.01"
                          min="0"
                          max="100"
                          placeholder="0"
                          {...field}
                          onChange={(e) =>
                            field.onChange(parseFloat(e.target.value) || 0)
                          }
                          disabled={!canEdit}
                          className="font-mono"
                        />
                        <span className="text-muted-foreground">%</span>
                      </div>
                    </FormControl>
                    <FormDescription>
                      Contoh: Jika toleransi 5% dan sisa qty = 20, maka max
                      invoice = 20 × 1.05 = 21 pcs
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Tolerance Example */}
              <div className="mt-4 p-4 bg-muted/50 rounded-lg">
                <p className="text-sm font-medium mb-2">Contoh Perhitungan:</p>
                <div className="text-sm text-muted-foreground space-y-1">
                  <p>• Sisa qty yang bisa di-invoice: <span className="font-mono">20 pcs</span></p>
                  <p>• Toleransi: <span className="font-mono">{form.watch("invoiceTolerancePct") || 0}%</span></p>
                  <p>
                    • Max invoice:{" "}
                    <span className="font-mono font-medium text-foreground">
                      {(20 * (1 + (form.watch("invoiceTolerancePct") || 0) / 100)).toFixed(2)} pcs
                    </span>
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Save Button */}
          {canEdit && (
            <div className="flex justify-end">
              <Button type="submit" disabled={isSaving}>
                {isSaving ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Save className="mr-2 h-4 w-4" />
                )}
                Simpan Pengaturan
              </Button>
            </div>
          )}
        </form>
      </Form>
    </div>
  );
}
