/**
 * Edit Transfer Page
 *
 * Full-page form for editing DRAFT stock transfers with:
 * - Pre-filled transfer data
 * - Multi-step wizard
 * - Stock validation
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Package, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditTransferForm } from "@/components/transfers/edit-transfer-form";
import { useGetTransferQuery } from "@/store/services/transferApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ArrowLeft } from "lucide-react";

export default function EditTransferPage() {
  const params = useParams();
  const router = useRouter();
  const transferId = params.id as string;

  const { data, isLoading, error } = useGetTransferQuery(transferId);

  const handleSuccess = () => {
    // Navigate back to the transfer's detail page
    router.push(`/inventory/transfers/${transferId}`);
  };

  const handleCancel = () => {
    router.push(`/inventory/transfers/${transferId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Transfer Gudang", href: "/inventory/transfers" },
            { label: "Edit Transfer" },
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
            { label: "Transfer Gudang", href: "/inventory/transfers" },
            { label: "Edit Transfer" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data transfer" : "Transfer tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/inventory/transfers")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Transfer
          </Button>
        </div>
      </div>
    );
  }

  const transfer = data;

  // Check if transfer is editable (only DRAFT can be edited)
  if (transfer.status !== "DRAFT") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Transfer Gudang", href: "/inventory/transfers" },
            { label: transfer.transferNumber, href: `/inventory/transfers/${transferId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Transfer dengan status {transfer.status} tidak dapat diedit. Hanya transfer dengan status DRAFT yang dapat diedit.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/inventory/transfers/${transferId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail Transfer
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
          { label: "Transfer Gudang", href: "/inventory/transfers" },
          { label: transfer.transferNumber, href: `/inventory/transfers/${transferId}` },
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
              Edit Transfer Gudang
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui transfer{" "}
            <span className="font-mono font-semibold">{transfer.transferNumber}</span>
          </p>
        </div>

        {/* Edit Transfer Form */}
        <EditTransferForm
          transfer={transfer}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
