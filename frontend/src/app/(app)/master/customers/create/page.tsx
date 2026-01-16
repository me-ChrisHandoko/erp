/**
 * Create Customer Page
 *
 * Full-page form for creating new customers with:
 * - Basic information
 * - Contact details
 * - Address information
 * - Business terms
 */

"use client"

import { useRouter } from "next/navigation"
import { Users, ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { CreateCustomerForm } from "@/components/customers/create-customer-form"

export default function CreateCustomerPage() {
  const router = useRouter()

  const handleSuccess = (customerId: string) => {
    // Navigate to the newly created customer's detail page
    router.push(`/master/customers/${customerId}`)
  }

  const handleCancel = () => {
    router.push("/master/customers")
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pelanggan", href: "/master/customers" },
          { label: "Tambah Pelanggan" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Users className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Pelanggan Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat pelanggan baru untuk sistem distribusi Anda
          </p>
        </div>

        {/* Create Customer Form */}
        <CreateCustomerForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
