/**
 * Edit Sales Payment Page - Server Component
 *
 * Fetches payment data and renders edit form.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EditSalesPaymentForm } from "@/components/sales-payments/edit-sales-payment-form";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { SalesPaymentResponse } from "@/types/sales-payment.types";
import type { ApiSuccessResponse } from "@/types/api";
import { notFound } from "next/navigation";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function EditSalesPaymentPage({ params }: PageProps) {
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
    console.error('[Edit Payment Page] Failed to fetch payment:', error);
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
            { label: payment.paymentNumber, href: `/sales/payments/${payment.id}` },
            { label: "Edit" },
          ]}
        />

        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Pembayaran
            </h1>
            <p className="text-muted-foreground">
              Edit informasi pembayaran pelanggan
            </p>
          </div>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle>Informasi Pembayaran</CardTitle>
            </CardHeader>
            <CardContent>
              <EditSalesPaymentForm payment={payment} />
            </CardContent>
          </Card>
        </div>
      </div>
    </ErrorBoundary>
  );
}
