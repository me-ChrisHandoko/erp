/**
 * Loading UI for Products Page
 *
 * Shown while server is fetching initial product data
 */

import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { PageHeader } from "@/components/shared/page-header";

export default function Loading() {
  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Produk" },
        ]}
      />
      <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
        <LoadingSpinner size="lg" />
        <div className="text-center space-y-2">
          <h3 className="text-lg font-semibold">Memuat Data Produk</h3>
          <p className="text-sm text-muted-foreground">
            Mohon tunggu sebentar...
          </p>
        </div>
      </div>
    </div>
  );
}
