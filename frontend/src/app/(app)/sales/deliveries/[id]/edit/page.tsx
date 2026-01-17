/**
 * Edit Delivery Page (Status Update)
 *
 * Page for updating delivery status and tracking information.
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Truck, ArrowLeft, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { UpdateDeliveryStatusForm } from "@/components/deliveries/update-delivery-status-form";
import { useGetDeliveryQuery } from "@/store/services/deliveryApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditDeliveryPage() {
  const params = useParams();
  const router = useRouter();
  const deliveryId = params.id as string;

  const { data: delivery, isLoading, error } = useGetDeliveryQuery(deliveryId);

  const handleSuccess = () => {
    // Navigate to delivery detail page on success
    router.push(`/sales/deliveries/${deliveryId}`);
  };

  const handleCancel = () => {
    // Navigate back to delivery detail
    router.push(`/sales/deliveries/${deliveryId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pengiriman", href: "/sales/deliveries" },
            { label: "Update Status" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-12 w-64" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-48 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !delivery) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pengiriman", href: "/sales/deliveries" },
            { label: "Update Status" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Gagal memuat data pengiriman.{" "}
              {error && "message" in error
                ? String(error.message)
                : "Silakan coba lagi."}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push("/sales/deliveries")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Pengiriman
          </Button>
        </div>
      </div>
    );
  }

  // Cannot edit cancelled or confirmed deliveries
  if (delivery.status === "CANCELLED" || delivery.status === "CONFIRMED") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Penjualan" },
            { label: "Pengiriman", href: "/sales/deliveries" },
            { label: delivery.deliveryNumber },
            { label: "Update Status" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Pengiriman dengan status{" "}
              {delivery.status === "CANCELLED" ? "Dibatalkan" : "Dikonfirmasi"}{" "}
              tidak dapat diubah.
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            onClick={() => router.push(`/sales/deliveries/${deliveryId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail
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
          { label: "Pengiriman", href: "/sales/deliveries" },
          { label: delivery.deliveryNumber, href: `/sales/deliveries/${deliveryId}` },
          { label: "Update Status" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Truck className="h-6 w-6 text-primary" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                Update Status Pengiriman
              </h1>
              <p className="text-sm text-muted-foreground">
                {delivery.deliveryNumber}
              </p>
            </div>
          </div>

          <Button variant="outline" onClick={handleCancel}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>
        </div>

        {/* Form */}
        <UpdateDeliveryStatusForm
          delivery={delivery}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
