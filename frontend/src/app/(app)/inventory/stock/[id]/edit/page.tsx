/**
 * Edit Stock Settings Page
 *
 * Full-page form for editing warehouse stock settings:
 * - Minimum stock threshold
 * - Maximum stock threshold
 * - Storage location
 */

"use client"

import { useParams, useRouter } from "next/navigation"
import { useState, useEffect } from "react"
import { Settings, AlertCircle } from "lucide-react"
import { PageHeader } from "@/components/shared/page-header"
import { EditStockSettingsForm } from "@/components/stock/edit-stock-settings-form"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"
import type { WarehouseStockResponse } from "@/types/stock.types"

export default function EditStockSettingsPage() {
  const params = useParams()
  const router = useRouter()
  const stockId = params.id as string

  const [stock, setStock] = useState<WarehouseStockResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<boolean>(false)

  // Load stock data from sessionStorage
  useEffect(() => {
    try {
      const storedStock = sessionStorage.getItem(`stock_${stockId}`)
      if (storedStock) {
        setStock(JSON.parse(storedStock))
        setIsLoading(false)
      } else {
        setError(true)
        setIsLoading(false)
      }
    } catch (err) {
      console.error("Failed to load stock data:", err)
      setError(true)
      setIsLoading(false)
    }
  }, [stockId])

  const handleSuccess = () => {
    // Clean up sessionStorage and navigate back
    sessionStorage.removeItem(`stock_${stockId}`)
    router.push("/inventory/stock")
  }

  const handleCancel = () => {
    // Clean up sessionStorage and navigate back
    sessionStorage.removeItem(`stock_${stockId}`)
    router.push("/inventory/stock")
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Stok Gudang", href: "/inventory/stock" },
            { label: "Edit Pengaturan" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Skeleton className="h-8 w-48" />
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </div>
    )
  }

  // Error state
  if (error || !stock) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Stok Gudang", href: "/inventory/stock" },
            { label: "Edit Pengaturan" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Data stok tidak ditemukan. Silakan kembali ke halaman daftar stok.
            </AlertDescription>
          </Alert>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Stok Gudang", href: "/inventory/stock" },
          { label: "Edit Pengaturan" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Settings className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Pengaturan Stok
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui pengaturan stok untuk{" "}
            <span className="font-semibold">{stock.productName}</span> di{" "}
            <span className="font-semibold">{stock.warehouseName}</span>
          </p>
        </div>

        {/* Edit Stock Settings Form */}
        <EditStockSettingsForm
          stock={stock}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  )
}
