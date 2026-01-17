/**
 * Create Sales Invoice Page
 *
 * In a distribution ERP, sales invoices are typically generated from:
 * 1. Sales Orders (SO) - Invoice before or during delivery
 * 2. Deliveries (DO) - Invoice after goods are delivered
 *
 * This page will be expanded to support direct invoice creation in the future.
 */

"use client";

import { useRouter } from "next/navigation";
import { FileText, Package, Truck, ArrowLeft, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function CreateInvoicePage() {
  const router = useRouter();

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Penjualan" },
          { label: "Faktur", href: "/sales/invoices" },
          { label: "Buat Faktur" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <FileText className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Buat Faktur Penjualan
            </h1>
          </div>
          <p className="text-muted-foreground">
            Pilih metode pembuatan faktur penjualan
          </p>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            Dalam sistem distribusi, faktur penjualan biasanya dibuat dari Pesanan Penjualan atau Surat Jalan.
            Fitur pembuatan faktur manual akan segera tersedia.
          </AlertDescription>
        </Alert>

        {/* Invoice Creation Methods */}
        <div className="grid gap-4 md:grid-cols-2">
          {/* From Sales Order */}
          <Card className="cursor-pointer hover:border-primary transition-colors">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Package className="h-5 w-5 text-primary" />
                <CardTitle>Dari Pesanan Penjualan</CardTitle>
              </div>
              <CardDescription>
                Buat faktur dari pesanan penjualan yang sudah ada
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Pilih pesanan penjualan dan sistem akan otomatis membuat faktur dengan detail produk dan harga.
              </p>
              <Button
                onClick={() => router.push("/sales/orders")}
                className="w-full"
              >
                Lihat Pesanan Penjualan
              </Button>
            </CardContent>
          </Card>

          {/* From Delivery */}
          <Card className="cursor-pointer hover:border-primary transition-colors">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Truck className="h-5 w-5 text-primary" />
                <CardTitle>Dari Surat Jalan</CardTitle>
              </div>
              <CardDescription>
                Buat faktur dari surat jalan yang sudah dikirim
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                Pilih surat jalan yang sudah dikirim dan buat faktur berdasarkan barang yang sudah diterima pelanggan.
              </p>
              <Button
                onClick={() => router.push("/sales/deliveries")}
                className="w-full"
              >
                Lihat Surat Jalan
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Coming Soon - Manual Creation */}
        <Card className="border-dashed">
          <CardHeader>
            <CardTitle className="text-muted-foreground">Pembuatan Manual (Segera Hadir)</CardTitle>
            <CardDescription>
              Fitur untuk membuat faktur penjualan langsung tanpa pesanan atau surat jalan
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Fitur ini akan memungkinkan Anda membuat faktur penjualan langsung dengan memilih pelanggan dan produk.
            </p>
          </CardContent>
        </Card>

        {/* Back Button */}
        <div className="flex justify-end">
          <Button
            variant="outline"
            onClick={() => router.push("/sales/invoices")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Faktur
          </Button>
        </div>
      </div>
    </div>
  );
}
