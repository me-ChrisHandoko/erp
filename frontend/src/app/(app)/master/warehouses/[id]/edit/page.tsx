/**
 * Edit Warehouse Page
 *
 * Full-page form for editing existing warehouses with:
 * - Pre-filled warehouse data
 * - Basic information editing
 * - Location updates
 * - Capacity adjustments
 * - Type changes
 */

"use client"

import { useParams, useRouter } from "next/navigation"
import { Warehouse, ArrowLeft, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { EditWarehouseForm } from "@/components/warehouses/edit-warehouse-form"
import { useGetWarehouseQuery } from "@/store/services/warehouseApi"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"

export default function EditWarehousePage() {
  const params = useParams()
  const router = useRouter()
  const warehouseId = params.id as string

  const { data, isLoading, error } = useGetWarehouseQuery(warehouseId)

  const handleSuccess = () => {
    // Navigate back to the warehouse's detail page
    router.push(`/master/warehouses/${warehouseId}`)
  }

  const handleCancel = () => {
    router.push(`/master/warehouses/${warehouseId}`)
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Gudang", href: "/master/warehouses" },
            { label: "Edit Gudang" },
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
            { label: "Gudang", href: "/master/warehouses" },
            { label: "Edit Gudang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data gudang" : "Gudang tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/master/warehouses")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Gudang
          </Button>
        </div>
      </div>
    )
  }

  const warehouse = data

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Gudang", href: "/master/warehouses" },
          { label: warehouse.name, href: `/master/warehouses/${warehouseId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Warehouse className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Gudang
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi gudang{" "}
            <span className="font-mono font-semibold">{warehouse.code}</span> -{" "}
            {warehouse.name}
          </p>
        </div>

        {/* Edit Warehouse Form */}
        <EditWarehouseForm
          warehouse={warehouse}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
