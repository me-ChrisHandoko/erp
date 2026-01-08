/**
 * Customers Page - Server Component
 *
 * Server-side rendered page that fetches initial customer data
 * and passes to client component for interactivity.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { CustomersClient } from "./customers-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { CustomerListResponse } from "@/types/customer.types";
import type { ApiSuccessResponse } from "@/types/api";

export default async function CustomersPage() {
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
              { label: "Pelanggan" },
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

  // Fetch initial customers data on server
  let initialData: CustomerListResponse;

  try {
    const response = await apiFetch<ApiSuccessResponse<CustomerListResponse>>({
      endpoint: '/customers',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'code',
        sort_order: 'asc',
      },
      cache: 'no-store',
    });

    initialData = response.data;
  } catch (error) {
    console.error('[Customers Page] Failed to fetch initial data:', error);

    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Pelanggan" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data pelanggan'}
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
            { label: "Pelanggan" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <CustomersClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
