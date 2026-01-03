/**
 * Products List Page
 *
 * Displays and manages product catalog with:
 * - Product table with sorting & pagination
 * - Search and filter capabilities
 * - Create, edit, delete actions (OWNER/ADMIN only)
 * - Multi-unit and supplier information
 */

"use client";

import { useState, useEffect } from "react";
import { Plus, Search, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { useListProductsQuery } from "@/store/services/productApi";
import { usePermissions } from "@/hooks/use-permissions";
import { ProductsTable } from "@/components/products/products-table";
import { CreateProductDialog } from "@/components/products/create-product-dialog";
import type { ProductFilters } from "@/types/product.types";

export default function ProductsPage() {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string | undefined>(
    undefined
  );
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [filters, setFilters] = useState<ProductFilters>({
    page: 1,
    pageSize: 20,
    sortBy: "code",
    sortOrder: "asc",
  });

  // Check permissions
  const { canEdit } = usePermissions();

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setFilters((prev) => ({ ...prev, page: 1 })); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch products with filters
  const {
    data: productsData,
    isLoading,
    error,
    refetch,
  } = useListProductsQuery({
    ...filters,
    search: debouncedSearch || undefined,
    category: categoryFilter,
    isActive: statusFilter,
  });

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
  };

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => {
      // Toggle sort order if clicking the same column
      if (prev.sortBy === sortBy) {
        return {
          ...prev,
          sortOrder: prev.sortOrder === "asc" ? "desc" : "asc",
        } as ProductFilters;
      }
      // New column, default to ascending
      return {
        ...prev,
        sortBy: sortBy as ProductFilters["sortBy"],
        sortOrder: "asc",
      } as ProductFilters;
    });
  };

  const handleCategoryFilterChange = (category: string | undefined) => {
    setCategoryFilter(category);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  const handleStatusFilterChange = (status: boolean | undefined) => {
    setStatusFilter(status);
    setFilters((prev) => ({ ...prev, page: 1 }));
  };

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Produk" },
          ]}
        />

        {/* Main content */}
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {/* Page title and actions */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-1">
              <h1 className="text-3xl font-bold tracking-tight">
                Daftar Produk
              </h1>
              <p className="text-muted-foreground">
                Kelola produk, unit, dan harga untuk distribusi
              </p>
            </div>
            {canEdit && (
              <Button
                className="shrink-0"
                onClick={() => setIsCreateDialogOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                Tambah Produk
              </Button>
            )}
          </div>

          {/* Products table with search and filters */}
          <Card className="shadow-sm">
            <CardContent>
              {/* Search and Filters Row */}
              <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
                {/* Search */}
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Cari kode atau nama produk..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="pl-9"
                  />
                </div>

                {/* Category Filter */}
                <Select
                  value={categoryFilter || "all"}
                  onValueChange={(value) =>
                    handleCategoryFilterChange(
                      value === "all" ? undefined : value
                    )
                  }
                >
                  <SelectTrigger className="w-full sm:w-[180px]">
                    <SelectValue placeholder="Semua Kategori" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua Kategori</SelectItem>
                    {productsData?.data &&
                      Array.from(
                        new Set(
                          productsData.data
                            .map((p) => p.category)
                            .filter(Boolean)
                        )
                      ).map((category) => (
                        <SelectItem key={category} value={category as string}>
                          {category}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>

                {/* Status Filter */}
                <Select
                  value={
                    statusFilter === undefined
                      ? "all"
                      : statusFilter
                      ? "active"
                      : "inactive"
                  }
                  onValueChange={(value) =>
                    handleStatusFilterChange(
                      value === "all" ? undefined : value === "active"
                    )
                  }
                >
                  <SelectTrigger className="w-full sm:w-[150px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua Status</SelectItem>
                    <SelectItem value="active">Aktif</SelectItem>
                    <SelectItem value="inactive">Nonaktif</SelectItem>
                  </SelectContent>
                </Select>

                {/* Clear Filters Button */}
                {(categoryFilter || statusFilter !== undefined || search) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSearch("");
                      setDebouncedSearch("");
                      setCategoryFilter(undefined);
                      setStatusFilter(undefined);
                      setFilters((prev) => ({ ...prev, page: 1 }));
                    }}
                    className="w-full sm:w-auto"
                  >
                    Reset
                  </Button>
                )}
              </div>

              {isLoading && (
                <div className="py-12">
                  <LoadingSpinner size="lg" text="Memuat data produk..." />
                </div>
              )}

              {error && (
                <ErrorDisplay
                  error={error}
                  title="Gagal memuat data produk"
                  onRetry={refetch}
                />
              )}

              {!isLoading && !error && productsData && (
                <>
                  {productsData.data.length === 0 &&
                  !search &&
                  !categoryFilter &&
                  statusFilter === undefined ? (
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <Package className="mb-4 h-12 w-12 text-muted-foreground" />
                      <h3 className="mb-2 text-lg font-semibold">
                        Belum ada produk
                      </h3>
                      <p className="mb-4 text-sm text-muted-foreground">
                        Mulai dengan menambahkan produk pertama Anda
                      </p>
                      {canEdit && (
                        <Button onClick={() => setIsCreateDialogOpen(true)}>
                          <Plus className="mr-2 h-4 w-4" />
                          Tambah Produk
                        </Button>
                      )}
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {/* Products Table */}
                      <ProductsTable
                        products={productsData.data}
                        sortBy={filters.sortBy}
                        sortOrder={filters.sortOrder}
                        onSortChange={handleSortChange}
                        canEdit={canEdit}
                      />

                      {/* Pagination */}
                      {productsData.pagination.totalPages > 1 && (
                        <div className="flex items-center justify-between border-t pt-4">
                          <div className="text-sm text-muted-foreground">
                            Halaman {productsData.pagination.page} dari{" "}
                            {productsData.pagination.totalPages} (
                            {productsData.pagination.totalItems} produk)
                          </div>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                handlePageChange(
                                  productsData.pagination.page - 1
                                )
                              }
                              disabled={productsData.pagination.page === 1}
                            >
                              Sebelumnya
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                handlePageChange(
                                  productsData.pagination.page + 1
                                )
                              }
                              disabled={
                                productsData.pagination.page >=
                                productsData.pagination.totalPages
                              }
                            >
                              Selanjutnya
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Create Product Dialog */}
      <CreateProductDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </ErrorBoundary>
  );
}
