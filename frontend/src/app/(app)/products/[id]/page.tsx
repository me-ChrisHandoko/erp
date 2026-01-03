/**
 * Product Detail Page
 *
 * Displays comprehensive product information including:
 * - Basic information (code, name, category)
 * - Pricing details (cost, price, margin)
 * - Units and conversions
 * - Stock information
 * - Suppliers
 * - Product attributes
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { Package, Edit, AlertCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ProductDetail } from "@/components/products/product-detail";
import { useGetProductQuery } from "@/store/services/productApi";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function ProductDetailPage() {
  const params = useParams();
  const router = useRouter();
  const productId = params.id as string;

  const { data, isLoading, error } = useGetProductQuery(productId);

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Produk", href: "/products" },
            { label: "Detail Produk" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
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
            { label: "Produk", href: "/products" },
            { label: "Detail Produk" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error ? "Gagal memuat data produk" : "Produk tidak ditemukan"}
            </AlertDescription>
          </Alert>
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
          { label: "Produk", href: "/products" },
          { label: "Detail Produk" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Package className="h-6 w-6 text-muted-foreground" />
              <h1 className="text-3xl font-bold tracking-tight">
                {product.name}
              </h1>
            </div>
            <p className="text-muted-foreground">
              Kode: <span className="font-mono font-semibold">{product.code}</span>
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => router.push("/products")}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            <Button
              className="shrink-0"
              onClick={() => router.push(`/products/edit/${productId}`)}
            >
              <Edit className="mr-2 h-4 w-4" />
              Edit Produk
            </Button>
          </div>
        </div>

        {/* Product Detail Component */}
        <ProductDetail product={product} />
      </div>
    </div>
  );
}
