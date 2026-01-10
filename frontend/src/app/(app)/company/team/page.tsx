/**
 * Team Management Page - Server Component
 *
 * Server-side rendered page that fetches initial team data
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
import { TeamClient } from "./team-client";
import { apiFetch } from "@/lib/server/api-fetch";
import { requireAuth } from "@/lib/server/auth";
import type { Tenant, TenantUser } from "@/types/tenant.types";
import type { ApiSuccessResponse } from "@/types/api";

interface TeamInitialData {
  tenant: Tenant;
  users: TenantUser[];
}

export default async function TeamPage() {
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
              { label: "Tim" },
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

  // Fetch initial team data on server
  let initialData: TeamInitialData;

  try {
    // Fetch tenant and users in parallel for better performance
    const [tenantResponse, usersResponse] = await Promise.all([
      apiFetch<ApiSuccessResponse<Tenant>>({
        endpoint: '/tenant',
        cache: 'no-store', // Always fetch fresh data for now
      }),
      apiFetch<ApiSuccessResponse<TenantUser[]>>({
        endpoint: '/company/users',
        cache: 'no-store', // Always fetch fresh data for now
      }),
    ]);

    // Extract data from success response envelope
    initialData = {
      tenant: tenantResponse.data,
      users: usersResponse.data,
    };
  } catch (error) {
    console.error('[Team Page] Failed to fetch initial data:', error);

    // Return error state - let client handle retry
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Perusahaan", href: "/company" },
              { label: "Tim" },
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
                  : 'Terjadi kesalahan saat memuat data tim'}
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
            { label: "Tim" },
          ]}
        />

        {/* Pass server-fetched data to client component */}
        <TeamClient initialData={initialData} />
      </div>
    </ErrorBoundary>
  );
}
