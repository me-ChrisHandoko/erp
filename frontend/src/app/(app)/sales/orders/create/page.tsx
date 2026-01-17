/**
 * Create Sales Order Page
 *
 * Full-page form for creating new sales orders with:
 * - Customer selection
 * - Order items management
 * - Pricing and totals calculation
 * - Order submission
 */

"use client";

import { useRouter } from "next/navigation";
import { ShoppingCart, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreateSalesOrderForm } from "@/components/sales-orders/create-sales-order-form";

export default function CreateSalesOrderPage() {
  const router = useRouter();

  const handleSuccess = (orderId: string) => {
    // Navigate to the newly created order's detail page
    router.push(`/sales/orders/${orderId}`);
  };

  const handleCancel = () => {
    router.push("/sales/orders");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Penjualan" },
          { label: "Pesanan", href: "/sales/orders" },
          { label: "Buat Pesanan" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <ShoppingCart className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Buat Pesanan Penjualan
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat pesanan penjualan baru dari pelanggan
          </p>
        </div>

        {/* Create Sales Order Form */}
        <CreateSalesOrderForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
