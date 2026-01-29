/**
 * Invoice Settings Page
 *
 * Configuration page for purchase invoice settings.
 * Allows management of:
 * - Invoice Control Policy (3-way matching)
 * - Invoice Tolerance Percentage
 */

import { PageHeader } from "@/components/shared/page-header";
import { FileText } from "lucide-react";
import { InvoiceSettingsClient } from "./invoice-settings-client";

export default function InvoiceSettingsPage() {
  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement" },
          { label: "Pengaturan", href: "/procurement/settings" },
          { label: "Faktur Pembelian" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page Title */}
        <div className="flex items-center gap-2">
          <FileText className="h-6 w-6 text-muted-foreground" />
          <h1 className="text-3xl font-bold tracking-tight">
            Pengaturan Faktur Pembelian
          </h1>
        </div>

        {/* Client Component */}
        <InvoiceSettingsClient />
      </div>
    </div>
  );
}
