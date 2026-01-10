/**
 * Stock Page - Server Component
 *
 * Server-side rendered page that fetches initial warehouse stock data
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
import { StockClient } from "./stock-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { WarehouseStockListResponse } from "@/types/stock.types";

export default async function StockPage() {
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
              { label: "Stok Gudang" },
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

  // Fetch initial stock data on server
  let initialData: WarehouseStockListResponse;

  try {
    // Backend returns WarehouseStockListResponse directly (not wrapped in ApiSuccessResponse)
    const response = await apiFetch<{
      stocks: any[];
      totalCount: number;
      page: number;
      pageSize: number;
      totalPages: number;
    }>({
      endpoint: '/warehouse-stocks',
      params: {
        page: 1,
        pageSize: 20,
        sortBy: 'productCode',
        sortOrder: 'asc',
      },
      cache: 'no-store', // Always fetch fresh data for now
    });

    // Transform backend response to match our expected structure
    initialData = {
      success: true,
      data: response.stocks || [],
      pagination: {
        page: response.page || 1,
        pageSize: response.pageSize || 20,
        totalItems: response.totalCount || 0,
        totalPages: response.totalPages || 0,
        hasMore: (response.page || 0) < (response.totalPages || 0),
      },
    };
  } catch (error) {
    console.error('[Stock Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Inventori", href: "/inventory/stock" },
              { label: "Stok Gudang" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data stok'}
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
            { label: "Stok Gudang" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <StockClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
