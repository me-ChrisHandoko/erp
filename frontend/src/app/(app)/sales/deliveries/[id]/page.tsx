/**
 * Delivery Detail Page
 *
 * Displays comprehensive delivery information including:
 * - Basic information (delivery number, date, status)
 * - Customer and warehouse details
 * - Sales order reference
 * - Delivery items with quantities
 * - Driver/expedition information
 * - Proof of delivery (POD) information
 * - Status tracking timeline
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Truck, Edit, AlertCircle, ArrowLeft, FileDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DeliveryDetail } from "@/components/deliveries/delivery-detail";
import { useGetDeliveryQuery } from "@/store/services/deliveryApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useState } from "react";

export default function DeliveryDetailPage() {
  const params = useParams();
  const router = useRouter();
  const deliveryId = params.id as string;
  const [isDownloading, setIsDownloading] = useState(false);

  const { data, isLoading, error } = useGetDeliveryQuery(deliveryId);

  // Handle PDF download
  const handleDownloadPDF = async () => {
    try {
      setIsDownloading(true);
      const token = localStorage.getItem("token");
      const companyId = localStorage.getItem("company_id");

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/deliveries/${deliveryId}/pdf`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "X-Company-ID": companyId || "",
          },
        }
      );

      if (!response.ok) {
        throw new Error("Failed to download PDF");
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `Surat_Jalan_${data?.deliveryNumber || deliveryId}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error("Error downloading PDF:", error);
      alert("Gagal mengunduh PDF. Silakan coba lagi.");
    } finally {
      setIsDownloading(false);
    }
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
            { label: "Detail Pengiriman" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
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
            { label: "Pengiriman", href: "/sales/deliveries" },
            { label: "Detail Pengiriman" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Gagal memuat detail pengiriman.{" "}
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

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Penjualan" },
          { label: "Pengiriman", href: "/sales/deliveries" },
          { label: data.deliveryNumber },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header with title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Truck className="h-6 w-6 text-primary" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                {data.deliveryNumber}
              </h1>
              <p className="text-sm text-muted-foreground">
                Detail Pengiriman
              </p>
            </div>
          </div>

          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => router.push("/sales/deliveries")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            <Button
              variant="outline"
              onClick={handleDownloadPDF}
              disabled={isDownloading}
            >
              <FileDown className="mr-2 h-4 w-4" />
              {isDownloading ? "Mengunduh..." : "Download PDF"}
            </Button>
            {data.status !== "CANCELLED" && data.status !== "CONFIRMED" && (
              <Button
                onClick={() =>
                  router.push(`/sales/deliveries/${deliveryId}/edit`)
                }
              >
                <Edit className="mr-2 h-4 w-4" />
                Update Status
              </Button>
            )}
          </div>
        </div>

        {/* Delivery detail component */}
        <DeliveryDetail delivery={data} />
      </div>
    </div>
  );
}
