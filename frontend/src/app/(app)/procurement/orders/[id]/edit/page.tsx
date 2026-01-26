/**
 * Edit Purchase Order Page
 *
 * Full-page form for editing DRAFT purchase orders with:
 * - Pre-filled order data
 * - Supplier and warehouse selection
 * - Dynamic line items
 * - Validation
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { ShoppingCart, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditOrderForm } from "@/components/procurement/edit-order-form";
import { useGetPurchaseOrderQuery } from "@/store/services/purchaseOrderApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditOrderPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.id as string;

  const { data, isLoading, error } = useGetPurchaseOrderQuery(orderId);

  const handleSuccess = () => {
    // Navigate back to the order's detail page
    router.push(`/procurement/orders/${orderId}`);
  };

  const handleCancel = () => {
    router.push(`/procurement/orders/${orderId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian" },
            { label: "Pesanan Pembelian", href: "/procurement/orders" },
            { label: "Edit Pesanan" },
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
            { label: "Pembelian" },
            { label: "Pesanan Pembelian", href: "/procurement/orders" },
            { label: "Edit Pesanan" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data pesanan pembelian" : "Pesanan pembelian tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/procurement/orders")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Pesanan
          </Button>
        </div>
      </div>
    );
  }

  const order = data;

  // Check if order is editable (only DRAFT can be edited)
  if (order.status !== "DRAFT") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian" },
            { label: "Pesanan Pembelian", href: "/procurement/orders" },
            { label: order.poNumber, href: `/procurement/orders/${orderId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Pesanan pembelian dengan status {order.status} tidak dapat diedit. Hanya pesanan dengan
              status DRAFT yang dapat diedit.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/procurement/orders/${orderId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail Pesanan
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
          { label: "Pembelian" },
          { label: "Pesanan Pembelian", href: "/procurement/orders" },
          { label: order.poNumber, href: `/procurement/orders/${orderId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <ShoppingCart className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">Edit Pesanan Pembelian</h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui pesanan pembelian{" "}
            <span className="font-mono font-semibold">{order.poNumber}</span>
          </p>
        </div>

        {/* Edit Order Form */}
        <EditOrderForm order={order} onSuccess={handleSuccess} onCancel={handleCancel} />
      </div>
    </div>
  );
}
