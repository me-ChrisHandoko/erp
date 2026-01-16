/**
 * Create Warehouse Page
 *
 * Full-page form for creating new warehouses with:
 * - Basic information
 * - Location details
 * - Capacity settings
 * - Type selection
 */

"use client"

import { useRouter } from "next/navigation"
import { Warehouse, ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { CreateWarehouseForm } from "@/components/warehouses/create-warehouse-form"

export default function CreateWarehousePage() {
  const router = useRouter()

  const handleSuccess = () => {
    // Navigate to the warehouses list page
    router.push("/master/warehouses")
  }

  const handleCancel = () => {
    router.push("/master/warehouses")
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Gudang", href: "/master/warehouses" },
          { label: "Tambah Gudang" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Warehouse className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Gudang Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat gudang baru untuk manajemen inventori
          </p>
        </div>

        {/* Create Warehouse Form */}
        <CreateWarehouseForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
