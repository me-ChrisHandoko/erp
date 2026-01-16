/**
 * Payment Detail Page - Server Component
 *
 * Displays detailed information about a specific payment including:
 * - Payment information
 * - Supplier details
 * - Payment method and reference
 * - Related purchase order (if any)
 * - Timestamps and approval info
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { PaymentDetail } from "@/components/payments/payment-detail";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { PaymentResponse } from "@/types/payment.types";
import type { ApiSuccessResponse } from "@/types/api";
import { notFound } from "next/navigation";

interface PaymentPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function PaymentPage({ params }: PaymentPageProps) {
  // Ensure user is authenticated
  await requireAuth();

  // Await params
  const { id } = await params;

  // Fetch payment data
  let payment: PaymentResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<PaymentResponse>>({
      endpoint: `/supplier-payments/${id}`,
      cache: 'no-store',
    });

    payment = response.data;
  } catch (error) {
    console.error('[Payment Detail Page] Failed to fetch payment:', error);
    // Return 404 if payment not found
    notFound();
  }

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement/orders" },
            { label: "Pembayaran", href: "/procurement/payments" },
            { label: payment.paymentNumber },
          ]}
        />

        <PaymentDetail payment={payment} />
      </div>
    </ErrorBoundary>
  );
}
