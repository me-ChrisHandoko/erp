/**
 * Create Delivery Page
 *
 * Page for creating new deliveries with form and validation.
 */

"use client";

import { useRouter } from "next/navigation";
import { Truck, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreateDeliveryForm } from "@/components/deliveries/create-delivery-form";

export default function CreateDeliveryPage() {
  const router = useRouter();

  const handleSuccess = (deliveryId: string) => {
    // Navigate to delivery detail page on success
    router.push(`/sales/deliveries/${deliveryId}`);
  };

  const handleCancel = () => {
    // Navigate back to deliveries list
    router.push("/sales/deliveries");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Penjualan" },
          { label: "Pengiriman", href: "/sales/deliveries" },
          { label: "Buat Pengiriman" },
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
                Buat Pengiriman Baru
              </h1>
              <p className="text-sm text-muted-foreground">
                Buat pengiriman barang ke customer
              </p>
            </div>
          </div>

          <Button variant="outline" onClick={handleCancel}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>
        </div>

        {/* Form */}
        <CreateDeliveryForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </div>
    </div>
  );
}
