/**
 * Bank Accounts Page - Server Component
 *
 * Server-side rendered page that fetches initial bank accounts data
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
import { BanksClient } from "./banks-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { BankAccountListResponse, BankAccountResponse } from "@/types/company.types";
import type { ApiPaginatedResponse } from "@/types/api";

export default async function BanksPage() {
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
              { label: "Perusahaan", href: "/company" },
              { label: "Rekening Bank" },
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

  // Fetch initial bank accounts data on server
  let initialData: BankAccountListResponse;

  try {
    const response = await apiFetch<ApiPaginatedResponse<BankAccountResponse>>({
      endpoint: '/company/banks',
      params: {
        page: 1,
        page_size: 20,
        sort_by: 'bankName',
        sort_order: 'asc',
      },
      cache: 'no-store', // Always fetch fresh data for now
    });

    // Extract data and pagination from response
    initialData = {
      data: response.data,
      pagination: response.pagination,
    };
  } catch (error) {
    console.error('[Banks Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Perusahaan", href: "/company" },
              { label: "Rekening Bank" },
            ]}
          />
          <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold text-destructive">
                Gagal Memuat Data
              </h3>
              <p className="text-sm text-muted-foreground">
                {error instanceof Error ? error.message : 'Terjadi kesalahan saat memuat data rekening bank'}
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
            { label: "Perusahaan", href: "/company" },
            { label: "Rekening Bank" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <BanksClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
