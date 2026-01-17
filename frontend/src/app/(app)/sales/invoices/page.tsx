/**
 * Sales Invoices Page - Server Component
 *
 * Server-side rendered page that fetches initial invoice data
 * and passes to client component for interactivity.
 */

import { Metadata } from "next";
import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { InvoicesClient } from "./invoices-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { InvoiceListResponse } from "@/types/invoice.types";
import type { ApiSuccessResponse } from "@/types/api";

export const metadata: Metadata = {
  title: "Faktur Penjualan | Sales",
  description: "Kelola faktur penjualan dan tagihan pelanggan",
};

export default async function InvoicesPage() {
  // Ensure user is authenticated
  const session = await requireAuth();

  // If no company context, let client component handle initialization
  if (!session.activeCompanyId) {
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Faktur Penjualan" },
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

  // Fetch initial invoices data on server
  let initialData: InvoiceListResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<InvoiceListResponse>>({
      endpoint: '/invoices',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'invoiceDate',
        sort_order: 'desc',
      },
      cache: 'no-store',
    });

    // Extract data from success response envelope
    initialData = response.data;
  } catch (error) {
    console.error('[Invoices Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Faktur Penjualan" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data faktur'}
              </p>
              <p className="text-xs text-muted-foreground">
                Silakan refresh halaman atau hubungi administrator jika masalah berlanjut
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
            { label: "Faktur Penjualan" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <InvoicesClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
