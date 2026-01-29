/**
 * Purchase Invoices Page - Server Component
 *
 * Server-side rendered page that fetches initial invoice data
 * and passes to client component for interactivity.
 *
 * Benefits:
 * - Fast initial load (no loading spinner!)
 * - SEO friendly (data in HTML)
 * - Better security (API credentials on server)
 * - Reduced client bundle size
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { PurchaseInvoicesClient } from "./invoices-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { PurchaseInvoiceListResponse } from "@/types/purchase-invoice.types";

export default async function PurchaseInvoicesPage() {
  // Ensure user is authenticated
  const session = await requireAuth();

  // If no company context, let client component handle initialization
  if (!session.activeCompanyId) {
    // Return minimal shell - CompanyInitializer will handle company selection
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Faktur Pembelian" },
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
  let initialData: PurchaseInvoiceListResponse;

  try {
    // Note: Backend returns { success, data: [...], pagination: {...} } at same level
    // Not nested like ApiSuccessResponse<{ data, pagination }>
    const response = await apiFetch<{ success: boolean; data: any[]; pagination: any }>({
      endpoint: '/purchase-invoices',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'invoiceDate',
        sort_order: 'desc',
      },
      cache: 'no-store', // Always fetch fresh data for now
    });

    // Construct PurchaseInvoiceListResponse from flat response
    initialData = {
      data: response.data,
      pagination: response.pagination,
    };
  } catch (error) {
    console.error('[Purchase Invoices Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Faktur Pembelian" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data faktur pembelian'}
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
            { label: "Faktur Pembelian" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <PurchaseInvoicesClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
