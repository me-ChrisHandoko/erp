/**
 * Goods Receipts Page - Server Component
 *
 * Server-side rendered page that fetches initial goods receipt data
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
import { ReceiptsClient } from "./receipts-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { GoodsReceiptListResponse } from "@/types/goods-receipt.types";

export default async function GoodsReceiptsPage() {
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
              { label: "Pembelian", href: "/procurement" },
              { label: "Penerimaan Barang" },
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

  // Fetch initial goods receipts data on server
  let initialData: GoodsReceiptListResponse;

  try {
    const response = await apiFetch<GoodsReceiptListResponse>({
      endpoint: '/goods-receipts',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'createdAt',
        sort_order: 'desc',
      },
      cache: 'no-store', // Always fetch fresh data for now
    });

    // Response is already GoodsReceiptListResponse
    initialData = response;
  } catch (error) {
    console.error('[Goods Receipts Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Pembelian", href: "/procurement" },
              { label: "Penerimaan Barang" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data penerimaan barang'}
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
            { label: "Pembelian", href: "/procurement" },
            { label: "Penerimaan Barang" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <ReceiptsClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
