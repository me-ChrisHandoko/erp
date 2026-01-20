/**
 * Products Table Component
 *
 * Displays products in a sortable table with:
 * - Sortable columns (code, name, price, stock)
 * - Status badges (active/inactive)
 * - Category filtering
 * - Action buttons (view, edit)
 * - Responsive design
 */

"use client";

import Link from "next/link";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  Pencil,
  Package,
  MoreHorizontal,
  PackageX,
  AlertTriangle,
  CheckCircle2,
} from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { EmptyState } from "@/components/shared/empty-state";
import type { ProductResponse } from "@/types/product.types";

/**
 * Stock Status Indicator Component
 * Shows stock level with icon and tooltip
 */
function StockStatusIndicator({ product }: { product: ProductResponse }) {
  const totalStock = product.currentStock
    ? parseFloat(product.currentStock.totalStock)
    : null;
  const minimumStock = parseFloat(product.minimumStock || "0");

  // No stock data available (no warehouse_stock records)
  if (totalStock === null || !product.currentStock) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center justify-center">
              <Badge
                variant="outline"
                className="border-gray-300 text-gray-500 gap-1"
              >
                <PackageX className="h-3 w-3" />
                N/A
              </Badge>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>Belum ada data stok</p>
            <p className="text-xs text-muted-foreground">
              Setup stok awal di menu Inventory
            </p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Zero stock
  if (totalStock === 0) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center justify-center">
              <Badge className="bg-red-500 text-white hover:bg-red-600 gap-1">
                <PackageX className="h-3 w-3" />
                0
              </Badge>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>Stok habis</p>
            {product.currentStock.warehouses.length > 0 && (
              <div className="text-xs mt-1">
                {product.currentStock.warehouses.map((wh) => (
                  <div key={wh.warehouseId}>
                    {wh.warehouseName}: {wh.quantity} {product.baseUnit}
                  </div>
                ))}
              </div>
            )}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Low stock (below minimum)
  if (totalStock > 0 && totalStock < minimumStock) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center justify-center">
              <Badge className="bg-amber-500 text-white hover:bg-amber-600 gap-1">
                <AlertTriangle className="h-3 w-3" />
                {totalStock.toLocaleString("id-ID")}
              </Badge>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>Stok menipis (min: {minimumStock})</p>
            {product.currentStock.warehouses.length > 0 && (
              <div className="text-xs mt-1">
                {product.currentStock.warehouses.map((wh) => (
                  <div key={wh.warehouseId}>
                    {wh.warehouseName}: {parseFloat(wh.quantity).toLocaleString("id-ID")} {product.baseUnit}
                  </div>
                ))}
              </div>
            )}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Normal stock (above minimum)
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="flex items-center justify-center">
            <Badge className="bg-green-500 text-white hover:bg-green-600 gap-1">
              <CheckCircle2 className="h-3 w-3" />
              {totalStock.toLocaleString("id-ID")}
            </Badge>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>Stok tersedia</p>
          {product.currentStock.warehouses.length > 0 && (
            <div className="text-xs mt-1">
              {product.currentStock.warehouses.map((wh) => (
                <div key={wh.warehouseId}>
                  {wh.warehouseName}: {parseFloat(wh.quantity).toLocaleString("id-ID")} {product.baseUnit}
                </div>
              ))}
            </div>
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface ProductsTableProps {
  products: ProductResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function ProductsTable({
  products,
  sortBy = "code",
  sortOrder = "asc",
  onSortChange,
  canEdit,
}: ProductsTableProps) {
  // Sort icon component
  const SortIcon = ({ column }: { column: string }) => {
    if (sortBy !== column) {
      return <ArrowUpDown className="ml-2 h-4 w-4 text-muted-foreground" />;
    }
    return sortOrder === "asc" ? (
      <ArrowUp className="ml-2 h-4 w-4" />
    ) : (
      <ArrowDown className="ml-2 h-4 w-4" />
    );
  };

  return (
    <>
      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("code")}
                >
                  Kode
                  <SortIcon column="code" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("name")}
                >
                  Nama Produk
                  <SortIcon column="name" />
                </Button>
              </TableHead>
              <TableHead>Kategori</TableHead>
              <TableHead className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("basePrice")}
                >
                  Harga Jual
                  <SortIcon column="basePrice" />
                </Button>
              </TableHead>
              <TableHead className="text-right">Harga Beli</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="text-center">Unit</TableHead>
              <TableHead className="text-center">Stok</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {products.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9}>
                  <EmptyState
                    icon={Package}
                    title="Produk tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              products.map((product) => (
                <TableRow key={product.id}>
                  {/* Code */}
                  <TableCell className="font-mono text-sm font-medium">
                    {product.code}
                  </TableCell>

                  {/* Name */}
                  <TableCell>
                    <div className="font-medium">{product.name}</div>
                    {product.description && (
                      <div className="text-sm text-muted-foreground line-clamp-1">
                        {product.description}
                      </div>
                    )}
                  </TableCell>

                  {/* Category */}
                  <TableCell>
                    {product.category && (
                      <Badge className="bg-blue-500 text-white hover:bg-blue-600">
                        {product.category}
                      </Badge>
                    )}
                  </TableCell>

                  {/* Base Price */}
                  <TableCell className="text-right font-medium">
                    <div>
                      Rp {Number(product.basePrice).toLocaleString("id-ID")}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      /{product.baseUnit}
                    </div>
                  </TableCell>

                  {/* Base Cost */}
                  <TableCell className="text-right text-muted-foreground">
                    <div>
                      Rp {Number(product.baseCost).toLocaleString("id-ID")}
                    </div>
                    <div className="text-xs">/{product.baseUnit}</div>
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge
                      className={
                        product.isActive
                          ? "bg-green-500 text-white hover:bg-green-600"
                          : "bg-red-500 text-white hover:bg-red-600"
                      }
                    >
                      {product.isActive ? "Aktif" : "Nonaktif"}
                    </Badge>
                  </TableCell>

                  {/* Units Count */}
                  <TableCell className="text-center">
                    {product.units && product.units.length > 1 ? (
                      <Badge variant="secondary">
                        {product.units.length} unit
                      </Badge>
                    ) : (
                      <span className="text-sm text-muted-foreground">
                        1 unit
                      </span>
                    )}
                  </TableCell>

                  {/* Stock Status */}
                  <TableCell className="text-center">
                    <StockStatusIndicator product={product} />
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                        >
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/master/products/${product.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/master/products/${product.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Produk
                            </Link>
                          </DropdownMenuItem>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </>
  );
}
