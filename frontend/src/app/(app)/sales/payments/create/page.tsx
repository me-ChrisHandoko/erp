/**
 * Create Sales Payment Page
 *
 * Server component that renders the create payment form.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CreateSalesPaymentForm } from "@/components/sales-payments/create-sales-payment-form";
import { requireAuth } from "@/lib/server/auth";
import { redirect } from "next/navigation";

export default async function CreateSalesPaymentPage() {
  const session = await requireAuth();

  if (!session.activeCompanyId) {
    redirect("/sales/payments");
  }

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan", href: "/sales/orders" },
            { label: "Pembayaran", href: "/sales/payments" },
            { label: "Catat Pembayaran" },
          ]}
        />

        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Catat Pembayaran Pelanggan
            </h1>
            <p className="text-muted-foreground">
              Catat pembayaran yang diterima dari pelanggan untuk invoice
            </p>
          </div>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle>Informasi Pembayaran</CardTitle>
            </CardHeader>
            <CardContent>
              <CreateSalesPaymentForm />
            </CardContent>
          </Card>
        </div>
      </div>
    </ErrorBoundary>
  );
}
