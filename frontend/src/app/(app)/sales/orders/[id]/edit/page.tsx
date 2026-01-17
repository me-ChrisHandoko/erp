/**
 * Edit Sales Order Page
 *
 * Full-page form for editing existing sales orders with:
 * - Pre-filled order data
 * - Customer and warehouse information
 * - Order items management
 * - Pricing and totals
 * - Only DRAFT orders can be edited
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { ShoppingCart, ArrowLeft, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditSalesOrderForm } from "@/components/sales-orders/edit-sales-order-form";
import { useGetSalesOrderQuery } from "@/store/services/salesOrderApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditSalesOrderPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.id as string;

  const { data, isLoading, error } = useGetSalesOrderQuery(orderId);

  const handleSuccess = () => {
    // Navigate back to the order's detail page
    router.push(`/sales/orders/${orderId}`);
  };

  const handleCancel = () => {
    router.push(`/sales/orders/${orderId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pesanan", href: "/sales/orders" },
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
            { label: "Penjualan" },
            { label: "Pesanan", href: "/sales/orders" },
            { label: "Edit Pesanan" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data pesanan" : "Pesanan tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/sales/orders")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Pesanan
          </Button>
        </div>
      </div>
    );
  }

  const salesOrder = data;

  // Check if order can be edited (only DRAFT orders)
  if (salesOrder.status !== "DRAFT") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pesanan", href: "/sales/orders" },
            { label: salesOrder.orderNumber, href: `/sales/orders/${orderId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Hanya pesanan dengan status DRAFT yang dapat diedit. Pesanan ini
              memiliki status {salesOrder.status}.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/sales/orders/${orderId}`)}
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
          { label: "Penjualan" },
          { label: "Pesanan", href: "/sales/orders" },
          { label: salesOrder.orderNumber, href: `/sales/orders/${orderId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <ShoppingCart className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Pesanan Penjualan
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi pesanan{" "}
            <span className="font-mono font-semibold">
              {salesOrder.orderNumber}
            </span>{" "}
            - {salesOrder.customerName}
          </p>
        </div>

        {/* Edit Sales Order Form */}
        <EditSalesOrderForm
          salesOrder={salesOrder}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
