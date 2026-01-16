/**
 * Warehouse Detail Page
 *
 * Displays comprehensive warehouse information including:
 * - Basic information (code, name, type, status)
 * - Location details (address, city, province, postal code)
 * - Contact information (phone, email)
 * - Management details (manager, capacity)
 * - Timestamps
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Warehouse, Edit, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { WarehouseDetail } from "@/components/warehouses/warehouse-detail";
import { useGetWarehouseQuery } from "@/store/services/warehouseApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function WarehouseDetailPage() {
  const params = useParams();
  const router = useRouter();
  const warehouseId = params.id as string;

  const { data, isLoading, error } = useGetWarehouseQuery(warehouseId);

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Gudang", href: "/master/warehouses" },
            { label: "Detail Gudang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !data) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Gudang", href: "/master/warehouses" },
            { label: "Detail Gudang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data gudang" : "Gudang tidak ditemukan"}
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  const warehouse = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Gudang", href: "/master/warehouses" },
          { label: "Detail Gudang" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Warehouse className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                {warehouse.name}
              </h1>
            </div>
            <p className="text-muted-foreground">
              Kode:{" "}
              <span className="font-mono font-semibold">{warehouse.code}</span>
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/master/warehouses")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            <Button
              className="shrink-0"
              onClick={() =>
                router.push(`/master/warehouses/${warehouseId}/edit`)
              }
            >
              <Edit className="mr-2 h-4 w-4" />
              Edit Gudang
            </Button>
          </div>
        </div>

        {/* Warehouse Detail Component */}
        <WarehouseDetail warehouse={warehouse} />
      </div>
    </div>
  );
}
