/**
 * Global Loading State for App Routes
 *
 * Ditampilkan otomatis saat navigasi ke halaman dengan async data fetching.
 * Next.js App Router secara otomatis wrap halaman dalam Suspense boundary,
 * sehingga loading.tsx akan muncul selama Server Component sedang fetch data.
 *
 * Catatan:
 * - Sidebar dan header tetap visible (ada di layout.tsx)
 * - Hanya area konten yang diganti dengan loading state
 * - Memberikan feedback visual instan saat navigasi
 */

import { LoadingSpinner } from "@/components/shared/loading-spinner";

export default function Loading() {
  return (
    <div className="flex flex-1 items-center justify-center min-h-[400px]">
      <LoadingSpinner size="lg" text="Memuat halaman..." />
    </div>
  );
}
