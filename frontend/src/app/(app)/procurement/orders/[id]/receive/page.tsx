/**
 * Create Goods Receipt Page
 *
 * Page for creating a goods receipt from a confirmed purchase order.
 * Shows the form with required field indicators based on product tracking settings.
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { PackageCheck, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CreateGoodsReceiptForm } from "@/components/goods-receipts/create-goods-receipt-form";
import { useGetPurchaseOrderQuery } from "@/store/services/purchaseOrderApi";

export default function CreateGoodsReceiptPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.id as string;

  const { data: order, isLoading, error } = useGetPurchaseOrderQuery(orderId);

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement" },
            { label: "Purchase Orders", href: "/procurement/orders" },
            { label: "Detail PO", href: `/procurement/orders/${orderId}` },
            { label: "Terima Barang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    );
  }

  // Error state
  if (error || !order) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement" },
            { label: "Purchase Orders", href: "/procurement/orders" },
            { label: "Terima Barang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error
                ? "Gagal memuat data purchase order"
                : "Purchase order tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push("/procurement/orders")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar PO
          </Button>
        </div>
      </div>
    );
  }

  // Check if PO is in CONFIRMED status
  if (order.status !== "CONFIRMED") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Pembelian", href: "/procurement" },
            { label: "Purchase Orders", href: "/procurement/orders" },
            { label: order.poNumber, href: `/procurement/orders/${orderId}` },
            { label: "Terima Barang" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Penerimaan barang hanya bisa dilakukan untuk PO dengan status
              DIKONFIRMASI. Status PO saat ini:{" "}
              <span className="font-semibold">{order.status}</span>
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push(`/procurement/orders/${orderId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail PO
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
          { label: "Pembelian", href: "/procurement" },
          { label: "Purchase Orders", href: "/procurement/orders" },
          { label: order.poNumber, href: `/procurement/orders/${orderId}` },
          { label: "Terima Barang" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <PackageCheck className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                Terima Barang
              </h1>
            </div>
            <p className="text-muted-foreground">
              Buat penerimaan barang untuk {order.poNumber} &bull;{" "}
              {order.supplier?.name}
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => router.push(`/procurement/orders/${orderId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>
        </div>

        {/* Form */}
        <CreateGoodsReceiptForm
          purchaseOrder={order}
          onSuccess={() => router.push("/procurement/receipts")}
          onCancel={() => router.push(`/procurement/orders/${orderId}`)}
        />
      </div>
    </div>
  );
}
