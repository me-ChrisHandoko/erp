/**
 * Adjustments Page - Server Component
 *
 * Server-side rendered page that fetches initial inventory adjustment data
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
import { AdjustmentsClient } from "./adjustments-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { AdjustmentListResponse } from "@/types/adjustment.types";
import type { ApiSuccessResponse } from "@/types/api";

export default async function AdjustmentsPage() {
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
              { label: "Inventori", href: "/inventory/stock" },
              { label: "Penyesuaian" },
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

  // Fetch initial adjustments data on server
  let initialData: AdjustmentListResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<AdjustmentListResponse>>({
      endpoint: '/inventory-adjustments',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'adjustmentNumber',
        sort_order: 'desc', // Latest adjustments first
      },
      cache: 'no-store', // Always fetch fresh data for now
    });

    // Extract data from success response envelope
    initialData = response.data;
  } catch (error) {
    console.error('[Adjustments Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Inventori", href: "/inventory/stock" },
              { label: "Penyesuaian" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data penyesuaian stok'}
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
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <AdjustmentsClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
