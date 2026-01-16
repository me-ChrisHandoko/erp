/**
 * Edit Product Page
 *
 * Full-page form for editing existing products with:
 * - Pre-filled product data
 * - Basic information editing
 * - Pricing updates
 * - Stock settings
 * - Supplier management (Phase 2)
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Package, ArrowLeft, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { EditProductForm } from "@/components/products/edit-product-form";
import { useGetProductQuery } from "@/store/services/productApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function EditProductPage() {
  const params = useParams();
  const router = useRouter();
  const productId = params.id as string;

  const { data, isLoading, error } = useGetProductQuery(productId);

  const handleSuccess = () => {
    // Navigate back to the product's detail page
    router.push(`/master/products/${productId}`);
  };

  const handleCancel = () => {
    router.push(`/master/products/${productId}`);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Produk", href: "/master/products" },
            { label: "Edit Produk" },
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
            { label: "Produk", href: "/master/products" },
            { label: "Edit Produk" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data produk" : "Produk tidak ditemukan"}
            </AlertDescription>
          </Alert>
          <Button
            variant="outline"
            className="w-fit"
            onClick={() => router.push("/master/products")}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Produk
          </Button>
        </div>
      </div>
    );
  }

  const product = data;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Produk", href: "/master/products" },
          { label: product.name, href: `/master/products/${productId}` },
          { label: "Edit" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Package className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Produk
            </h1>
          </div>
          <p className="text-muted-foreground">
            Perbarui informasi produk{" "}
            <span className="font-mono font-semibold">{product.code}</span> -{" "}
            {product.name}
          </p>
        </div>

        {/* Edit Product Form */}
        <EditProductForm
          product={product}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
