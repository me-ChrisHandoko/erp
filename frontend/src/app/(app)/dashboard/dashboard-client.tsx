/**
 * Dashboard Client Component
 *
 * Client-side interactive component for dashboard widgets.
 * Shows real-time data for:
 * - Products with stock alerts
 * - Products without stock setup
 * - Sales overview (placeholder)
 * - Financial overview (placeholder)
 */

"use client";

import { useMemo } from "react";
import { useSelector } from "react-redux";
import Link from "next/link";
import {
  PackageX,
  AlertTriangle,
  Package,
  ArrowRight,
  TrendingUp,
  Wallet,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useListProductsQuery } from "@/store/services/productApi";
import type { RootState } from "@/store";
import type { ProductResponse } from "@/types/product.types";

/**
 * Stock Stats Card - Shows products with stock issues
 */
function StockStatsCard({
  products,
  isLoading,
}: {
  products: ProductResponse[];
  isLoading: boolean;
}) {
  // Calculate stock statistics
  const stats = useMemo(() => {
    if (!products || products.length === 0) {
      return {
        noStockData: 0,
        zeroStock: 0,
        lowStock: 0,
        normalStock: 0,
        total: 0,
        noStockProducts: [] as ProductResponse[],
        zeroStockProducts: [] as ProductResponse[],
        lowStockProducts: [] as ProductResponse[],
      };
    }

    const noStockProducts: ProductResponse[] = [];
    const zeroStockProducts: ProductResponse[] = [];
    const lowStockProducts: ProductResponse[] = [];

    products.forEach((product) => {
      const totalStock = product.currentStock
        ? parseFloat(product.currentStock.totalStock)
        : null;
      const minimumStock = parseFloat(product.minimumStock || "0");

      if (totalStock === null || !product.currentStock) {
        noStockProducts.push(product);
      } else if (totalStock === 0) {
        zeroStockProducts.push(product);
      } else if (totalStock < minimumStock) {
        lowStockProducts.push(product);
      }
    });

    return {
      noStockData: noStockProducts.length,
      zeroStock: zeroStockProducts.length,
      lowStock: lowStockProducts.length,
      normalStock:
        products.length -
        noStockProducts.length -
        zeroStockProducts.length -
        lowStockProducts.length,
      total: products.length,
      noStockProducts: noStockProducts.slice(0, 5),
      zeroStockProducts: zeroStockProducts.slice(0, 5),
      lowStockProducts: lowStockProducts.slice(0, 5),
    };
  }, [products]);

  if (isLoading) {
    return (
      <Card className="col-span-1">
        <CardHeader className="pb-2">
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-20 w-full" />
        </CardContent>
      </Card>
    );
  }

  const hasIssues =
    stats.noStockData > 0 || stats.zeroStock > 0 || stats.lowStock > 0;

  return (
    <Card className="col-span-1">
      <CardHeader className="pb-2">
        <CardTitle className="text-base flex items-center gap-2">
          <Package className="h-4 w-4" />
          Status Stok Produk
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Stats Summary */}
        <div className="grid grid-cols-2 gap-3">
          {/* No Stock Data */}
          <Link
            href="/master/products?stockStatus=no-data"
            className="group block rounded-lg border p-3 hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-center gap-2 text-gray-500">
              <PackageX className="h-4 w-4" />
              <span className="text-sm">Belum Setup</span>
            </div>
            <div className="mt-1 text-2xl font-bold">{stats.noStockData}</div>
          </Link>

          {/* Zero Stock */}
          <Link
            href="/master/products?stockStatus=zero"
            className="group block rounded-lg border p-3 hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-center gap-2 text-red-500">
              <PackageX className="h-4 w-4" />
              <span className="text-sm">Stok Habis</span>
            </div>
            <div className="mt-1 text-2xl font-bold text-red-600">
              {stats.zeroStock}
            </div>
          </Link>

          {/* Low Stock */}
          <Link
            href="/master/products?stockStatus=low"
            className="group block rounded-lg border p-3 hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-center gap-2 text-amber-500">
              <AlertTriangle className="h-4 w-4" />
              <span className="text-sm">Stok Menipis</span>
            </div>
            <div className="mt-1 text-2xl font-bold text-amber-600">
              {stats.lowStock}
            </div>
          </Link>

          {/* Normal Stock */}
          <Link
            href="/master/products?stockStatus=normal"
            className="group block rounded-lg border p-3 hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-center gap-2 text-green-500">
              <Package className="h-4 w-4" />
              <span className="text-sm">Normal</span>
            </div>
            <div className="mt-1 text-2xl font-bold text-green-600">
              {stats.normalStock}
            </div>
          </Link>
        </div>

        {/* Alert Products List */}
        {hasIssues && (
          <div className="border-t pt-3 space-y-3">
            {/* Products without stock data */}
            {stats.noStockProducts.length > 0 && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-gray-600">
                    Perlu Setup Stok
                  </span>
                  {stats.noStockData > 5 && (
                    <Badge variant="secondary" className="text-xs">
                      +{stats.noStockData - 5} lainnya
                    </Badge>
                  )}
                </div>
                <div className="space-y-1">
                  {stats.noStockProducts.map((product) => (
                    <Link
                      key={product.id}
                      href={`/master/products/${product.id}`}
                      className="flex items-center justify-between p-2 rounded hover:bg-muted/50 text-sm"
                    >
                      <span className="font-mono text-xs text-muted-foreground">
                        {product.code}
                      </span>
                      <span className="truncate max-w-[150px]">
                        {product.name}
                      </span>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Zero stock products */}
            {stats.zeroStockProducts.length > 0 && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-red-600">
                    Stok Habis
                  </span>
                  {stats.zeroStock > 5 && (
                    <Badge variant="secondary" className="text-xs">
                      +{stats.zeroStock - 5} lainnya
                    </Badge>
                  )}
                </div>
                <div className="space-y-1">
                  {stats.zeroStockProducts.map((product) => (
                    <Link
                      key={product.id}
                      href={`/master/products/${product.id}`}
                      className="flex items-center justify-between p-2 rounded hover:bg-muted/50 text-sm"
                    >
                      <span className="font-mono text-xs text-muted-foreground">
                        {product.code}
                      </span>
                      <span className="truncate max-w-[150px]">
                        {product.name}
                      </span>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Low stock products */}
            {stats.lowStockProducts.length > 0 && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-amber-600">
                    Stok Menipis
                  </span>
                  {stats.lowStock > 5 && (
                    <Badge variant="secondary" className="text-xs">
                      +{stats.lowStock - 5} lainnya
                    </Badge>
                  )}
                </div>
                <div className="space-y-1">
                  {stats.lowStockProducts.map((product) => (
                    <Link
                      key={product.id}
                      href={`/master/products/${product.id}`}
                      className="flex items-center justify-between p-2 rounded hover:bg-muted/50 text-sm"
                    >
                      <span className="font-mono text-xs text-muted-foreground">
                        {product.code}
                      </span>
                      <span className="truncate max-w-[120px]">
                        {product.name}
                      </span>
                      <Badge
                        variant="outline"
                        className="text-amber-600 border-amber-300 text-xs"
                      >
                        {product.currentStock
                          ? parseFloat(
                              product.currentStock.totalStock
                            ).toLocaleString("id-ID")
                          : 0}{" "}
                        {product.baseUnit}
                      </Badge>
                    </Link>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* View All Link */}
        <div className="pt-2 border-t">
          <Button variant="ghost" size="sm" className="w-full" asChild>
            <Link href="/master/products">
              Lihat Semua Produk
              <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Placeholder Card for Sales
 */
function SalesPlaceholderCard() {
  return (
    <Card className="col-span-1">
      <CardHeader className="pb-2">
        <CardTitle className="text-base flex items-center gap-2">
          <TrendingUp className="h-4 w-4" />
          Total Penjualan
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm">
          Data penjualan akan ditampilkan di sini setelah modul penjualan
          terintegrasi.
        </p>
      </CardContent>
    </Card>
  );
}

/**
 * Placeholder Card for Finance
 */
function FinancePlaceholderCard() {
  return (
    <Card className="col-span-1">
      <CardHeader className="pb-2">
        <CardTitle className="text-base flex items-center gap-2">
          <Wallet className="h-4 w-4" />
          Hutang/Piutang
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm">
          Informasi keuangan akan ditampilkan di sini setelah modul keuangan
          terintegrasi.
        </p>
      </CardContent>
    </Card>
  );
}

export function DashboardClient() {
  // Get activeCompanyId from Redux
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Fetch all products for stock analysis
  // Use larger page size to get more products for dashboard stats
  const { data: productsData, isLoading } = useListProductsQuery(
    { pageSize: 100, isActive: true },
    { skip: !activeCompanyId }
  );

  const products = productsData?.data || [];

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      <h1 className="text-2xl font-bold">Dashboard ERP</h1>

      {/* Main Stats Grid */}
      <div className="grid gap-4 md:grid-cols-3">
        {/* Stock Stats Card */}
        <StockStatsCard products={products} isLoading={isLoading} />

        {/* Sales Placeholder */}
        <SalesPlaceholderCard />

        {/* Finance Placeholder */}
        <FinancePlaceholderCard />
      </div>

      {/* Additional Content Area */}
      <Card className="flex-1">
        <CardContent className="pt-6">
          <p className="text-muted-foreground">
            Konten dashboard tambahan akan ditampilkan di sini. Modul yang akan
            datang termasuk:
          </p>
          <ul className="mt-4 space-y-2 text-sm text-muted-foreground">
            <li>• Grafik penjualan harian/mingguan/bulanan</li>
            <li>• Daftar transaksi terakhir</li>
            <li>• Notifikasi dan pengingat</li>
            <li>• Laporan kinerja bisnis</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
