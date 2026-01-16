/**
 * Create Adjustment Page
 *
 * Full-page form for creating new inventory adjustments with:
 * - Multi-step wizard (Info → Products → Review)
 * - Warehouse selection
 * - Adjustment type and reason selection
 */

"use client";

import { useRouter } from "next/navigation";
import { Package } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { CreateAdjustmentForm } from "@/components/adjustments/create-adjustment-form";

export default function CreateAdjustmentPage() {
  const router = useRouter();

  const handleSuccess = (adjustmentId: string) => {
    // Navigate to the newly created adjustment's detail page
    router.push(`/inventory/adjustments/${adjustmentId}`);
  };

  const handleCancel = () => {
    router.push("/inventory/adjustments");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Penyesuaian Stok", href: "/inventory/adjustments" },
          { label: "Buat Penyesuaian" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Package className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Buat Penyesuaian Stok Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Kelola penyesuaian stok untuk koreksi, barang rusak, kadaluarsa, dan lainnya
          </p>
        </div>

        {/* Create Adjustment Form */}
        <CreateAdjustmentForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
