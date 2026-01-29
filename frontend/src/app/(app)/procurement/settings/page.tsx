/**
 * Procurement Settings Page
 *
 * Main settings menu for procurement module.
 * Lists all available settings configurations.
 */

import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Settings, Truck, FileText } from "lucide-react";
import Link from "next/link";

const settingsMenuItems = [
  {
    title: "Toleransi Pengiriman",
    description: "Konfigurasi toleransi penerimaan barang (over/under delivery) berdasarkan produk atau kategori",
    href: "/procurement/settings/delivery-tolerance",
    icon: Truck,
  },
  {
    title: "Pengaturan Faktur Pembelian",
    description: "Konfigurasi kebijakan pembuatan faktur (3-way matching) dan toleransi over-invoice",
    href: "/procurement/settings/invoice",
    icon: FileText,
  },
];

export default function ProcurementSettingsPage() {
  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Pembelian", href: "/procurement" },
          { label: "Pengaturan" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page Title */}
        <div className="flex items-center gap-2">
          <Settings className="h-6 w-6 text-muted-foreground" />
          <h1 className="text-3xl font-bold tracking-tight">
            Pengaturan Pembelian
          </h1>
        </div>

        <p className="text-muted-foreground">
          Kelola pengaturan dan konfigurasi untuk modul pembelian
        </p>

        {/* Settings Menu Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {settingsMenuItems.map((item) => (
            <Link key={item.href} href={item.href}>
              <Card className="h-full transition-colors hover:bg-muted/50 cursor-pointer">
                <CardHeader className="flex flex-row items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                    <item.icon className="h-6 w-6 text-primary" />
                  </div>
                  <div className="space-y-1">
                    <CardTitle className="text-lg">{item.title}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-sm">
                    {item.description}
                  </CardDescription>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
