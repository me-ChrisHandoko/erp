/**
 * Create Payment Page
 *
 * Full-page form for creating new supplier payments with:
 * - Payment information
 * - Supplier selection
 * - Amount and payment method
 * - Optional purchase order reference
 * - Bank account selection
 */

"use client";

import { useRouter } from "next/navigation";
import { DollarSign, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreatePaymentForm } from "@/components/payments/create-payment-form";

export default function CreatePaymentPage() {
  const router = useRouter();

  const handleSuccess = (paymentId: string) => {
    // Navigate to the newly created payment's detail page
    router.push(`/procurement/payments/${paymentId}`);
  };

  const handleCancel = () => {
    router.push("/procurement/payments");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement/orders" },
          { label: "Pembayaran", href: "/procurement/payments" },
          { label: "Tambah Pembayaran" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <DollarSign className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Pembayaran Supplier
            </h1>
          </div>
          <p className="text-muted-foreground">
            Catat pembayaran baru untuk supplier Anda
          </p>
        </div>

        {/* Create Payment Form */}
        <CreatePaymentForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
