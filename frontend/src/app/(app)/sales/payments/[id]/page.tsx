/**
 * Sales Payment Detail Page - Server Component
 *
 * Displays full payment details with invoice information.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { SalesPaymentDetail } from "@/components/sales-payments/sales-payment-detail";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { SalesPaymentResponse } from "@/types/sales-payment.types";
import type { ApiSuccessResponse } from "@/types/api";
import { notFound } from "next/navigation";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function SalesPaymentDetailPage({ params }: PageProps) {
  const session = await requireAuth();
  const { id } = await params;

  if (!session.activeCompanyId) {
    notFound();
  }

  // Fetch payment data on server
  let payment: SalesPaymentResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<SalesPaymentResponse>>({
      endpoint: `/payments/${id}`,
      cache: 'no-store',
    });

    payment = response.data;
  } catch (error) {
    console.error('[Payment Detail Page] Failed to fetch payment:', error);
    notFound();
  }

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan", href: "/sales/orders" },
            { label: "Pembayaran", href: "/sales/payments" },
            { label: payment.paymentNumber },
          ]}
        />

        <SalesPaymentDetail payment={payment} />
      </div>
    </ErrorBoundary>
  );
}
