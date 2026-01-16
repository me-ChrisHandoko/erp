/**
 * Purchase Invoice Detail Page
 *
 * Displays comprehensive invoice information including:
 * - Invoice header (number, dates, supplier)
 * - Invoice items with pricing
 * - Payment information
 * - Status and workflow actions
 * - Financial summary (subtotal, tax, total)
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { FileText, Edit, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { InvoiceDetail } from "@/components/invoices/invoice-detail";
import { useGetPurchaseInvoiceQuery } from "@/store/services/purchaseInvoiceApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function PurchaseInvoiceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const invoiceId = params.id as string;

  const { data, isLoading, error } = useGetPurchaseInvoiceQuery(invoiceId);

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Detail Faktur" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-48" />
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
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Detail Faktur" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data faktur" : "Faktur tidak ditemukan"}
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  const invoice = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Faktur Pembelian", href: "/procurement/invoices" },
          { label: "Detail Faktur" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                Faktur {invoice.invoiceNumber}
              </h1>
            </div>
            <p className="text-muted-foreground">
              Supplier: <span className="font-semibold">{invoice.supplierName}</span>
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/procurement/invoices")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            {invoice.status === "DRAFT" && (
              <Button
                className="shrink-0"
                onClick={() => router.push(`/procurement/invoices/${invoiceId}/edit`)}
              >
                <Edit className="mr-2 h-4 w-4" />
                Edit Faktur
              </Button>
            )}
          </div>
        </div>

        {/* Invoice Detail Component */}
        <InvoiceDetail invoice={invoice} />
      </div>
    </div>
  );
}
