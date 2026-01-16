/**
 * Create Supplier Page
 *
 * Full-page form for creating new suppliers with:
 * - Basic information
 * - Contact details
 * - Address information
 * - Business terms
 */

"use client"

import { useRouter } from "next/navigation"
import { Building2, ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { CreateSupplierForm } from "@/components/suppliers/create-supplier-form"

export default function CreateSupplierPage() {
  const router = useRouter()

  const handleSuccess = (supplierId?: string) => {
    // Navigate to the newly created supplier's detail page or suppliers list
    if (supplierId) {
      router.push(`/master/suppliers/${supplierId}`)
    } else {
      router.push("/master/suppliers")
    }
  }

  const handleCancel = () => {
    router.push("/master/suppliers")
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Supplier", href: "/master/suppliers" },
          { label: "Tambah Supplier" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Building2 className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Supplier Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat supplier baru untuk pengadaan barang
          </p>
        </div>

        {/* Create Supplier Form */}
        <CreateSupplierForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
