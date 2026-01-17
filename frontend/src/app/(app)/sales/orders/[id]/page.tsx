/**
 * Sales Order Detail Page - Server Component
 *
 * Server-side rendered page that fetches sales order details
 * and passes to client component for display.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { SalesOrderDetail } from "@/components/sales-orders/sales-order-detail";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import { notFound } from "next/navigation";
import type { SalesOrderResponse } from "@/types/sales-order.types";
import type { ApiSuccessResponse } from "@/types/api";

interface SalesOrderDetailPageProps {
  params: {
    id: string;
  };
}

export default async function SalesOrderDetailPage({
  params,
}: SalesOrderDetailPageProps) {
  // Ensure user is authenticated
  const session = await requireAuth();

  // If no company context, show initialization message
  if (!session.activeCompanyId) {
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Penjualan" },
              { label: "Pesanan", href: "/sales/orders" },
              { label: "Detail" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold">
                Initializing Company Context...
              </h3>
              <p className="text-sm text-muted-foreground">
                Please wait while we set up your company workspace
              </p>
            </div>
          </div>
        </div>
      </ErrorBoundary>
    );
  }

  // Fetch sales order details
  let salesOrder: SalesOrderResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<SalesOrderResponse>>({
      endpoint: `/sales-orders/${params.id}`,
      cache: "no-store", // Always fetch fresh data
    });

    salesOrder = response.data;
  } catch (error) {
    console.error("[Sales Order Detail] Failed to fetch data:", error);

    // If 404, show not found page
    if (error instanceof Error && error.message.includes("404")) {
      notFound();
    }

    // For other errors, show error state
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Penjualan" },
              { label: "Pesanan", href: "/sales/orders" },
              { label: "Detail" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error
                  ? error.message
                  : "Terjadi kesalahan saat memuat detail pesanan"}
              </p>
              <p className="text-xs text-muted-foreground">
                Silakan refresh halaman atau hubungi administrator jika masalah
                berlanjut
              </p>
            </div>
          </div>
        </div>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pesanan", href: "/sales/orders" },
            { label: salesOrder.orderNumber },
          ]}
        />

        {/* Pass fetched data to client component */}
        <SalesOrderDetail salesOrder={salesOrder} />
      </div>
    </ErrorBoundary>
  );
}
