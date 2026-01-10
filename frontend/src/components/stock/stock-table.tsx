/**
 * Stock Table Component
 *
 * Displays warehouse stocks in a sortable table with:
 * - Sortable columns (product code, name, quantity)
 * - Stock status badges (normal, low, critical, out of stock)
 * - Warehouse information
 * - Location details
 * - Edit stock settings action
 * - Responsive design
 */

"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
  Package,
  MapPin,
  MoreHorizontal,
  Settings,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { EditStockSettingsDialog } from "./edit-stock-settings-dialog";
import {
  getStockStatus,
  getStockStatusColor,
  getStockStatusLabel,
  formatStockQuantity,
} from "@/types/stock.types";
import type { WarehouseStockResponse } from "@/types/stock.types";

interface StockTableProps {
  stocks: WarehouseStockResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit?: boolean;
}

export function StockTable({
  stocks,
  sortBy = "productCode",
  sortOrder = "asc",
  onSortChange,
  canEdit = true,
}: StockTableProps) {
  // Edit dialog state
  const [editStock, setEditStock] = useState<WarehouseStockResponse | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

  const handleEditClick = (stock: WarehouseStockResponse) => {
    setEditStock(stock);
    setIsEditDialogOpen(true);
  };

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
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("productCode")}
                >
                  Kode Produk
                  <SortIcon column="productCode" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("productName")}
                >
                  Nama Produk
                  <SortIcon column="productName" />
                </Button>
              </TableHead>
              <TableHead>Gudang</TableHead>
              <TableHead className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("quantity")}
                >
                  Jumlah Stok
                  <SortIcon column="quantity" />
                </Button>
              </TableHead>
              <TableHead className="text-right">Stok Minimum</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead>Lokasi</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {stocks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8}>
                  <EmptyState
                    icon={Package}
                    title="Stok tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
            stocks.map((stock) => {
              const status = getStockStatus(
                stock.quantity,
                stock.minimumStock,
                stock.maximumStock
              );
              const statusColor = getStockStatusColor(status);
              const statusLabel = getStockStatusLabel(status);

              return (
                <TableRow key={stock.id}>
                  {/* Product Code */}
                  <TableCell className="font-mono text-sm font-medium">
                    {stock.productCode}
                  </TableCell>

                  {/* Product Name */}
                  <TableCell>
                    <div className="font-medium">{stock.productName}</div>
                    {stock.productCategory && (
                      <div className="text-sm text-muted-foreground">
                        {stock.productCategory}
                      </div>
                    )}
                  </TableCell>

                  {/* Warehouse */}
                  <TableCell>
                    <div className="font-medium">{stock.warehouseName}</div>
                    {stock.warehouseCode && (
                      <div className="text-xs text-muted-foreground font-mono">
                        {stock.warehouseCode}
                      </div>
                    )}
                  </TableCell>

                  {/* Quantity */}
                  <TableCell className="text-right font-medium">
                    <div>
                      {formatStockQuantity(stock.quantity, stock.productUnit)}
                    </div>
                    {stock.lastCountDate && (
                      <div className="text-xs text-muted-foreground">
                        Hitung:{" "}
                        {new Date(stock.lastCountDate).toLocaleDateString(
                          "id-ID"
                        )}
                      </div>
                    )}
                  </TableCell>

                  {/* Minimum Stock */}
                  <TableCell className="text-right text-muted-foreground">
                    {formatStockQuantity(
                      stock.minimumStock,
                      stock.productUnit
                    )}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge className={statusColor}>{statusLabel}</Badge>
                  </TableCell>

                  {/* Location */}
                  <TableCell>
                    {stock.location ? (
                      <div className="flex items-center gap-1 text-sm">
                        <MapPin className="h-3 w-3 text-muted-foreground" />
                        <span>{stock.location}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="text-right">
                    {canEdit && (
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
                          <DropdownMenuItem
                            onClick={() => handleEditClick(stock)}
                            className="cursor-pointer"
                          >
                            <Settings className="mr-2 h-4 w-4" />
                            Edit Pengaturan
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </TableCell>
                </TableRow>
              );
            })
          )}
          </TableBody>
        </Table>
      </div>

      {/* Edit Stock Settings Dialog */}
      <EditStockSettingsDialog
        stock={editStock}
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
      />
    </>
  );
}
