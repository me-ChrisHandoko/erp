/**
 * Loading State for Initial Stock Setup Page
 *
 * Displayed while the page is loading.
 */

import { LoadingSpinner } from "@/components/shared/loading-spinner";

export default function Loading() {
  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      <div className="space-y-1">
        <div className="h-9 w-64 bg-muted animate-pulse rounded" />
        <div className="h-5 w-96 bg-muted animate-pulse rounded" />
      </div>
      <div className="flex flex-1 items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" text="Memuat setup stok awal..." />
      </div>
    </div>
  );
}
