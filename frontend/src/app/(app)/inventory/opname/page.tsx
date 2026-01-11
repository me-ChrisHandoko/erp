/**
 * Stock Opname Page - Server Component
 *
 * Server-side rendered page that fetches initial stock opname data
 * and passes to client component for interactivity.
 *
 * Stock opname is the process of physically counting inventory
 * and comparing it with system records for accuracy.
 *
 * Benefits:
 * - Fast initial load (no loading spinner!)
 * - SEO friendly (data in HTML)
 * - Better security (API credentials on server)
 * - Reduced client bundle size
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { OpnameClient } from "./opname-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { StockOpnameListResponse } from "@/types/opname.types";

export default async function StockOpnamePage() {
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
              { label: "Stock Opname" },
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

  // Fetch initial stock opname data on server
  let initialData: StockOpnameListResponse;

  try {
    // Backend returns StockOpnameListResponse directly (not wrapped in ApiSuccessResponse)
    const response = await apiFetch<StockOpnameListResponse>({
      endpoint: "/stock-opnames",
      params: {
        page: 1,
        page_size: 20,
        sort_by: "opnameDate",
        sort_order: "desc",
      },
      cache: "no-store", // Always fetch fresh data for now
    });

    initialData = response;
  } catch (error) {
    console.error("[Stock Opname Page] Failed to fetch initial data:", error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Inventori", href: "/inventory/stock" },
              { label: "Stock Opname" },
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
                  : "Terjadi kesalahan saat memuat data stock opname"}
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
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <OpnameClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
