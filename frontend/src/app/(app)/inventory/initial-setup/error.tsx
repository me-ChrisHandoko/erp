/**
 * Error State for Initial Stock Setup Page
 *
 * Displayed when an error occurs during page load or rendering.
 */

"use client";

import { useEffect } from "react";
import { AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error("[Initial Setup Error]:", error);
  }, [error]);

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      <div className="space-y-1">
        <h1 className="text-3xl font-bold tracking-tight">Setup Stok Awal</h1>
        <p className="text-muted-foreground">
          Setup stok pertama kali untuk produk yang belum pernah memiliki record di gudang
        </p>
      </div>

      <Card className="shadow-sm border-destructive">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-destructive">
            <AlertCircle className="h-5 w-5" />
            Terjadi Kesalahan
          </CardTitle>
          <CardDescription>
            Gagal memuat halaman setup stok awal
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-md bg-destructive/10 p-4">
            <p className="text-sm text-destructive font-mono">
              {error.message || "Unknown error occurred"}
            </p>
            {error.digest && (
              <p className="text-xs text-muted-foreground mt-2">
                Error ID: {error.digest}
              </p>
            )}
          </div>

          <div className="flex gap-2">
            <Button onClick={reset} variant="default">
              Coba Lagi
            </Button>
            <Button
              onClick={() => (window.location.href = "/inventory/stock")}
              variant="outline"
            >
              Kembali ke Stok Barang
            </Button>
          </div>

          <div className="text-xs text-muted-foreground">
            <p>Jika masalah berlanjut, hubungi administrator sistem.</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
