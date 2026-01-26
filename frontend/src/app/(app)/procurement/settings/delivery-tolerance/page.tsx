/**
 * Delivery Tolerance Settings Page
 *
 * Configuration page for delivery tolerance settings (SAP Model).
 * Allows management of hierarchical tolerance rules for:
 * - Company-level defaults
 * - Category-level overrides
 * - Product-level specific tolerances
 */

import { PageHeader } from "@/components/shared/page-header";
import { Settings } from "lucide-react";
import { DeliveryToleranceClient } from "./delivery-tolerance-client";

export default function DeliveryToleranceSettingsPage() {
  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement" },
          { label: "Pengaturan" },
          { label: "Toleransi Pengiriman" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page Title */}
        <div className="flex items-center gap-2">
          <Settings className="h-6 w-6 text-muted-foreground" />
          <h1 className="text-3xl font-bold tracking-tight">
            Pengaturan Toleransi Pengiriman
          </h1>
        </div>

        {/* Client Component */}
        <DeliveryToleranceClient />
      </div>
    </div>
  );
}
