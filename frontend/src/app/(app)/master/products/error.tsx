/**
 * Error UI for Products Page
 *
 * Shown when server-side data fetching fails
 */

"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { AlertCircle } from "lucide-react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("[Products Page Error]", error);
  }, [error]);

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Produk" },
        ]}
      />
      <div className="flex flex-1 flex-col items-center justify-center min-h-[400px] gap-4 p-4">
        <AlertCircle className="h-12 w-12 text-destructive" />
        <div className="text-center space-y-2 max-w-md">
          <h2 className="text-2xl font-bold">Terjadi Kesalahan</h2>
          <p className="text-muted-foreground">
            {error.message || "Gagal memuat data produk. Silakan coba lagi."}
          </p>
          {error.digest && (
            <p className="text-xs text-muted-foreground">
              Error ID: {error.digest}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          <Button onClick={reset}>Coba Lagi</Button>
          <Button variant="outline" onClick={() => window.location.href = "/dashboard"}>
            Kembali ke Dashboard
          </Button>
        </div>
      </div>
    </div>
  );
}
