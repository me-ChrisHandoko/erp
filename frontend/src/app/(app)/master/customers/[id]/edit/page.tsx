/**
 * Edit Customer Page
 *
 * Full-page form for editing existing customers with:
 * - Pre-filled customer data
 * - Basic information editing
 * - Contact details updates
 * - Address information
 * - Business terms
 */

"use client"

import { useParams, useRouter } from "next/navigation"
import { Users, ArrowLeft, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/shared/page-header"
import { EditCustomerForm } from "@/components/customers/edit-customer-form"
import { useGetCustomerQuery } from "@/store/services/customerApi"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"

export default function EditCustomerPage() {
  const params = useParams()
  const router = useRouter()
  const customerId = params.id as string

  const { data, isLoading, error } = useGetCustomerQuery(customerId)

  const handleSuccess = () => {
    // Navigate back to the customer's detail page
    router.push(`/master/customers/${customerId}`)
  }

  const handleCancel = () => {
    router.push(`/master/customers/${customerId}`)
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pelanggan", href: "/master/customers" },
            { label: "Edit Pelanggan" },
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
            { label: "Pelanggan", href: "/master/customers" },
            { label: "Edit Pelanggan" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data pelanggan" : "Pelanggan tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/master/customers")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Pelanggan
          </Button>
        </div>
      </div>
    )
  }

  const customer = data

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pelanggan", href: "/master/customers" },
          { label: customer.name, href: `/master/customers/${customerId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Users className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Pelanggan
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi pelanggan{" "}
            <span className="font-mono font-semibold">{customer.code}</span> -{" "}
            {customer.name}
          </p>
        </div>

        {/* Edit Customer Form */}
        <EditCustomerForm
          customer={customer}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
