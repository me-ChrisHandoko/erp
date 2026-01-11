/**
 * Create Stock Opname Page - Server Component
 *
 * Page for creating new stock opname (physical inventory count).
 * User can select warehouse and import all products for counting.
 */

import { PageHeader } from "@/components/shared/page-header";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { CreateOpnameClient } from "./create-opname-client";
import { requireAuth } from "@/lib/server/auth";

export default async function CreateOpnamePage() {
  // Ensure user is authenticated
  await requireAuth();

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Buat Stock Opname" },
          ]}
        />

        {/* Create form client component */}
        <CreateOpnameClient />
      </div>
    </ErrorBoundary>
  );
}
