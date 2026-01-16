/**
 * Edit Adjustment Page
 *
 * Full-page form for editing DRAFT inventory adjustments with:
 * - Pre-filled adjustment data
 * - Multi-step wizard
 * - Warehouse and product validation
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Package, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditAdjustmentForm } from "@/components/adjustments/edit-adjustment-form";
import { useGetAdjustmentQuery } from "@/store/services/adjustmentApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditAdjustmentPage() {
  const params = useParams();
  const router = useRouter();
  const adjustmentId = params.id as string;

  const { data, isLoading, error } = useGetAdjustmentQuery(adjustmentId);

  const handleSuccess = () => {
    // Navigate back to the adjustment's detail page
    router.push(`/inventory/adjustments/${adjustmentId}`);
  };

  const handleCancel = () => {
    router.push(`/inventory/adjustments/${adjustmentId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
            { label: "Edit Penyesuaian" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Skeleton className="h-8 w-48" />
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-48 w-full" />
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
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
            { label: "Edit Penyesuaian" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data penyesuaian" : "Penyesuaian tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/inventory/adjustments")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Penyesuaian
          </Button>
        </div>
      </div>
    );
  }

  const adjustment = data;

  // Check if adjustment is editable (only DRAFT can be edited)
  if (adjustment.status !== "DRAFT") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
            { label: adjustment.adjustmentNumber, href: `/inventory/adjustments/${adjustmentId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Penyesuaian dengan status {adjustment.status} tidak dapat diedit. Hanya penyesuaian dengan status DRAFT yang dapat diedit.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/inventory/adjustments/${adjustmentId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail Penyesuaian
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
          { label: adjustment.adjustmentNumber, href: `/inventory/adjustments/${adjustmentId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Package className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Penyesuaian Stok
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui penyesuaian{" "}
            <span className="font-mono font-semibold">{adjustment.adjustmentNumber}</span>
          </p>
        </div>

        {/* Edit Adjustment Form */}
        <EditAdjustmentForm
          adjustment={adjustment}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
