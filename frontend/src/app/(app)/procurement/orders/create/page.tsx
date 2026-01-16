/**
 * Create Purchase Order Page
 *
 * Full-page form for creating new purchase orders with:
 * - Supplier and warehouse selection
 * - Dynamic line items
 * - Real-time calculation
 * - Validation
 */

"use client";

import { useRouter } from "next/navigation";
import { ShoppingCart } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { CreateOrderForm } from "@/components/procurement/create-order-form";

export default function CreateOrderPage() {
  const router = useRouter();

  const handleSuccess = (orderId: string) => {
    // Navigate to the newly created order's detail page
    router.push(`/procurement/orders/${orderId}`);
  };

  const handleCancel = () => {
    router.push("/procurement/orders");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Procurement", href: "/procurement/orders" },
          { label: "Purchase Orders", href: "/procurement/orders" },
          { label: "Buat PO Baru" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <ShoppingCart className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">Buat Purchase Order Baru</h1>
          </div>
          <p className="text-muted-foreground">
            Buat purchase order untuk pembelian barang dari supplier
          </p>
        </div>

        {/* Create Order Form */}
        <CreateOrderForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </div>
    </div>
  );
}
