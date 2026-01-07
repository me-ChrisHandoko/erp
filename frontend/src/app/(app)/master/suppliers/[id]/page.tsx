/**
 * Supplier Detail Page
 *
 * Displays comprehensive supplier information including:
 * - Basic information (code, name, type)
 * - Contact details (email, phone, address)
 * - Business information (tax ID, payment terms, credit limit)
 * - Products supplied
 */

"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import { Building2, Edit, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { SupplierDetail } from "@/components/suppliers/supplier-detail";
import { EditSupplierDialog } from "@/components/suppliers/edit-supplier-dialog";
import { useGetSupplierQuery } from "@/store/services/supplierApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import type { RootState } from "@/store";

export default function SupplierDetailPage() {
  const params = useParams();
  const router = useRouter();
  const supplierId = params.id as string;
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

  // Get activeCompany to ensure company context is ready
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  const { data, isLoading, error } = useGetSupplierQuery(supplierId, {
    skip: !activeCompanyId, // Skip query until company context is available
  });

  // Loading state - Show while company context is initializing or data is loading
  if (!activeCompanyId || isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Supplier", href: "/master/suppliers" },
            { label: "Detail Supplier" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex flex-col items-center justify-center min-h-[400px] gap-4">
            <LoadingSpinner size="lg" />
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold">Memuat Detail Supplier...</h3>
              <p className="text-sm text-muted-foreground">
                {!activeCompanyId
                  ? "Initializing company context..."
                  : "Loading supplier information..."}
              </p>
            </div>
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
            { label: "Supplier", href: "/master/suppliers" },
            { label: "Detail Supplier" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data supplier" : "Supplier tidak ditemukan"}
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  const supplier = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Supplier", href: "/master/suppliers" },
          { label: "Detail Supplier" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Building2 className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                {supplier.name}
              </h1>
            </div>
            <p className="text-muted-foreground">
              Kode: <span className="font-mono font-semibold">{supplier.code}</span>
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/master/suppliers")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            <Button
              className="shrink-0"
              onClick={() => setIsEditDialogOpen(true)}
            >
              <Edit className="mr-2 h-4 w-4" />
              Edit Supplier
            </Button>
          </div>
        </div>

        {/* Supplier Detail Component */}
        <SupplierDetail supplier={supplier} />
      </div>

      {/* Edit Supplier Dialog */}
      <EditSupplierDialog
        supplier={supplier}
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
      />
    </div>
  );
}
