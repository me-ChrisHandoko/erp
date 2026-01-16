/**
 * Create Transfer Page
 *
 * Full-page form for creating new stock transfers with:
 * - Multi-step wizard (Warehouse → Products → Review)
 * - Stock validation
 * - Warehouse selection
 */

"use client";

import { useRouter } from "next/navigation";
import { Package, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreateTransferForm } from "@/components/transfers/create-transfer-form";

export default function CreateTransferPage() {
  const router = useRouter();

  const handleSuccess = (transferId: string) => {
    // Navigate to the newly created transfer's detail page
    router.push(`/inventory/transfers/${transferId}`);
  };

  const handleCancel = () => {
    router.push("/inventory/transfers");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Transfer Gudang", href: "/inventory/transfers" },
          { label: "Buat Transfer" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Package className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Buat Transfer Gudang Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Transfer stok antar gudang dengan validasi real-time
          </p>
        </div>

        {/* Create Transfer Form */}
        <CreateTransferForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
