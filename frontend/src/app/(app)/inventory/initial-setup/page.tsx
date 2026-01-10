/**
 * Initial Stock Setup Page - Server Component
 *
 * Server-side rendered page for one-time initial stock setup.
 * Handles authentication and passes context to client component.
 *
 * Context-aware via URL params:
 * - ?warehouse={id} - Pre-select warehouse from warehouse detail page
 * - ?context=onboarding - Coming from onboarding flow
 * - ?source=dashboard - Coming from dashboard widget
 */

import { requireAuth } from "@/lib/server/auth";
import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { InitialSetupClient } from "./initial-setup-client";

interface InitialSetupPageProps {
  searchParams: Promise<{
    warehouse?: string;
    context?: string;
    source?: string;
  }>;
}

export default async function InitialSetupPage({
  searchParams,
}: InitialSetupPageProps) {
  // Ensure user is authenticated
  const session = await requireAuth();

  // If no company context, let CompanyInitializer handle company selection
  if (!session.activeCompanyId) {
    return (
      <ErrorBoundary>
        <div className="flex flex-col">
          <PageHeader
            breadcrumbs={[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Persediaan", href: "/inventory/stock" },
              { label: "Setup Stok Awal" },
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

  // Next.js 15+: searchParams is now a Promise that must be awaited
  const params = await searchParams;

  // Parse context from URL params
  const warehouseId = params.warehouse;
  const context = params.context; // 'onboarding' | undefined
  const source = params.source; // 'dashboard' | undefined

  return (
    <ErrorBoundary>
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Persediaan", href: "/inventory/stock" },
          { label: "Setup Stok Awal" },
        ]}
      />
      <InitialSetupClient
        initialWarehouseId={warehouseId}
        context={context}
        source={source}
      />
    </ErrorBoundary>
  );
}
