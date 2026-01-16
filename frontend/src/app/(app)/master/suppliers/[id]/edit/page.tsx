/**
 * Edit Supplier Page
 *
 * Full-page form for editing existing suppliers with:
 * - Pre-filled supplier data
 * - Basic information editing
 * - Contact details updates
 * - Address information
 * - Business terms
 */

"use client"

import { useParams, useRouter } from "next/navigation"
import { Building2, ArrowLeft, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { EditSupplierForm } from "@/components/suppliers/edit-supplier-form"
import { useGetSupplierQuery } from "@/store/services/supplierApi"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"

export default function EditSupplierPage() {
  const params = useParams()
  const router = useRouter()
  const supplierId = params.id as string

  const { data, isLoading, error } = useGetSupplierQuery(supplierId)

  const handleSuccess = () => {
    // Navigate back to the supplier's detail page
    router.push(`/master/suppliers/${supplierId}`)
  }

  const handleCancel = () => {
    router.push(`/master/suppliers/${supplierId}`)
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Supplier", href: "/master/suppliers" },
            { label: "Edit Supplier" },
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
    )
  }

  // Error state
  if (error || !data) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Supplier", href: "/master/suppliers" },
            { label: "Edit Supplier" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data supplier" : "Supplier tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/master/suppliers")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Supplier
          </Button>
        </div>
      </div>
    )
  }

  const supplier = data

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Supplier", href: "/master/suppliers" },
          { label: supplier.name, href: `/master/suppliers/${supplierId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Building2 className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Supplier
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi supplier{" "}
            <span className="font-mono font-semibold">{supplier.code}</span> -{" "}
            {supplier.name}
          </p>
        </div>

        {/* Edit Supplier Form */}
        <EditSupplierForm
          supplier={supplier}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
