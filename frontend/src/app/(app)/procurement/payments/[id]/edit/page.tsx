/**
 * Edit Payment Page
 *
 * Full-page form for editing existing supplier payments with:
 * - Pre-filled payment data
 * - Payment information editing
 * - Supplier and amount updates
 * - Cannot edit approved payments (business rule)
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { DollarSign, ArrowLeft, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditPaymentForm } from "@/components/payments/edit-payment-form";
import { useGetPaymentQuery } from "@/store/services/paymentApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditPaymentPage() {
  const params = useParams();
  const router = useRouter();
  const paymentId = params.id as string;

  const { data, isLoading, error } = useGetPaymentQuery(paymentId);

  const handleSuccess = () => {
    // Navigate back to the payment's detail page
    router.push(`/procurement/payments/${paymentId}`);
  };

  const handleCancel = () => {
    router.push(`/procurement/payments/${paymentId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement/orders" },
            { label: "Pembayaran", href: "/procurement/payments" },
            { label: "Edit Pembayaran" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Skeleton className="h-8 w-48" />
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !data) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement/orders" },
            { label: "Pembayaran", href: "/procurement/payments" },
            { label: "Edit Pembayaran" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data pembayaran" : "Pembayaran tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/procurement/payments")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Pembayaran
          </Button>
        </div>
      </div>
    );
  }

  const payment = data;

  // Check if payment is approved (cannot edit)
  if (payment.approvedBy) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement/orders" },
            { label: "Pembayaran", href: "/procurement/payments" },
            { label: payment.paymentNumber, href: `/procurement/payments/${paymentId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Pembayaran yang sudah disetujui tidak dapat diedit. Silakan hubungi administrator jika perlu melakukan perubahan.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/procurement/payments/${paymentId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail Pembayaran
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement/orders" },
          { label: "Pembayaran", href: "/procurement/payments" },
          { label: payment.paymentNumber, href: `/procurement/payments/${paymentId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <DollarSign className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Pembayaran
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi pembayaran{" "}
            <span className="font-mono font-semibold">{payment.paymentNumber}</span>
          </p>
        </div>

        {/* Edit Payment Form */}
        <EditPaymentForm
          payment={payment}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
