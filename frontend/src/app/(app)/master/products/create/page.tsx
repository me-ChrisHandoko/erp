/**
 * Create Product Page
 *
 * Full-page form for creating new products with:
 * - Basic information
 * - Pricing
 * - Stock settings
 * - Supplier management (Phase 2)
 */

"use client";

import { useRouter } from "next/navigation";
import { Package, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { CreateProductForm } from "@/components/products/create-product-form";

export default function CreateProductPage() {
  const router = useRouter();

  const handleSuccess = (productId: string) => {
    // Navigate to the newly created product's detail page
    router.push(`/master/products/${productId}`);
  };

  const handleCancel = () => {
    router.push("/master/products");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Produk", href: "/master/products" },
          { label: "Tambah Produk" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Package className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Produk Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat produk baru untuk katalog distribusi Anda
          </p>
        </div>

        {/* Create Product Form */}
        <CreateProductForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
