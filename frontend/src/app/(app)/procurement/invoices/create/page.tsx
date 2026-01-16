/**
 * Create Purchase Invoice Page
 *
 * Full-page form for creating new purchase invoices with:
 * - Supplier selection
 * - Invoice header (number, dates)
 * - Product line items
 * - Financial calculations
 * - Notes and references
 */

"use client";

import { useRouter } from "next/navigation";
import { FileText, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreateInvoiceForm } from "@/components/invoices/create-invoice-form";

export default function CreatePurchaseInvoicePage() {
  const router = useRouter();

  const handleSuccess = (invoiceId: string) => {
    // Navigate to the newly created invoice's detail page
    router.push(`/procurement/invoices/${invoiceId}`);
  };

  const handleCancel = () => {
    router.push("/procurement/invoices");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Faktur Pembelian", href: "/procurement/invoices" },
          { label: "Tambah Faktur" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <FileText className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Faktur Pembelian Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat faktur pembelian baru dari supplier
          </p>
        </div>

        {/* Create Invoice Form */}
        <CreateInvoiceForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
