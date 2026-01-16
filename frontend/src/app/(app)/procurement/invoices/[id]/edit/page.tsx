/**
 * Edit Purchase Invoice Page
 *
 * Full-page form for editing existing purchase invoices with:
 * - Pre-filled invoice data
 * - Invoice header editing
 * - Line items management
 * - Financial calculations
 * - Notes and references
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { FileText, ArrowLeft, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditInvoiceForm } from "@/components/invoices/edit-invoice-form";
import { useGetPurchaseInvoiceQuery } from "@/store/services/purchaseInvoiceApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditPurchaseInvoicePage() {
  const params = useParams();
  const router = useRouter();
  const invoiceId = params.id as string;

  const { data, isLoading, error } = useGetPurchaseInvoiceQuery(invoiceId);

  const handleSuccess = () => {
    // Navigate back to the invoice's detail page
    router.push(`/procurement/invoices/${invoiceId}`);
  };

  const handleCancel = () => {
    router.push(`/procurement/invoices/${invoiceId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Edit Faktur" },
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
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: "Edit Faktur" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data faktur" : "Faktur tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/procurement/invoices")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Faktur
          </Button>
        </div>
      </div>
    );
  }

  const invoice = data;

  // Only allow editing DRAFT invoices
  if (invoice.status !== "DRAFT") {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Faktur Pembelian", href: "/procurement/invoices" },
            { label: invoice.invoiceNumber, href: `/procurement/invoices/${invoiceId}` },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Hanya faktur dengan status DRAFT yang dapat diedit
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push(`/procurement/invoices/${invoiceId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Detail Faktur
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
          { label: "Faktur Pembelian", href: "/procurement/invoices" },
          { label: invoice.invoiceNumber, href: `/procurement/invoices/${invoiceId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <FileText className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Faktur Pembelian
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi faktur{" "}
            <span className="font-mono font-semibold">{invoice.invoiceNumber}</span>
          </p>
        </div>

        {/* Edit Invoice Form */}
        <EditInvoiceForm
          invoice={invoice}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
