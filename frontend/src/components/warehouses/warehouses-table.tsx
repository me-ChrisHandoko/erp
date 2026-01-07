/**
 * Warehouses Table Component
 *
 * Displays warehouses in a sortable table with:
 * - Sortable columns (code, name, type)
 * - Status badges (active/inactive)
 * - Type badges (color-coded by warehouse type)
 * - Location display (city + province)
 * - Action buttons (view, edit)
 * - Responsive design
 */

"use client";

import { useState } from "react";
import Link from "next/link";
import { EditWarehouseDialog } from "./edit-warehouse-dialog";
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
  Warehouse,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { WarehouseResponse } from "@/types/warehouse.types";
import {
  getWarehouseTypeLabel,
  getWarehouseTypeBadgeColor,
  formatWarehouseLocation,
  formatCapacity,
} from "@/types/warehouse.types";

interface WarehousesTableProps {
  warehouses: WarehouseResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function WarehousesTable({
  warehouses,
  sortBy = "code",
  sortOrder = "asc",
  onSortChange,
  canEdit,
}: WarehousesTableProps) {
  // Edit dialog state
  const [editWarehouse, setEditWarehouse] =
    useState<WarehouseResponse | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

  const handleEditClick = (warehouse: WarehouseResponse) => {
    setEditWarehouse(warehouse);
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
                  Nama Gudang
                  <SortIcon column="name" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("type")}
                >
                  Tipe
                  <SortIcon column="type" />
                </Button>
              </TableHead>
              <TableHead>Lokasi</TableHead>
              <TableHead className="text-right">Kapasitas</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {warehouses.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState
                    icon={Warehouse}
                    title="Gudang tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              warehouses.map((warehouse) => (
                <TableRow key={warehouse.id}>
                  {/* Code */}
                  <TableCell className="font-mono text-sm font-medium">
                    {warehouse.code}
                  </TableCell>

                  {/* Name */}
                  <TableCell>
                    <div className="font-medium">{warehouse.name}</div>
                    {warehouse.address && (
                      <div className="text-sm text-muted-foreground line-clamp-1">
                        {warehouse.address}
                      </div>
                    )}
                  </TableCell>

                  {/* Type */}
                  <TableCell>
                    <Badge
                      className={getWarehouseTypeBadgeColor(warehouse.type)}
                    >
                      {getWarehouseTypeLabel(warehouse.type)}
                    </Badge>
                  </TableCell>

                  {/* Location */}
                  <TableCell>
                    <div className="text-sm">
                      {formatWarehouseLocation(warehouse)}
                    </div>
                  </TableCell>

                  {/* Capacity */}
                  <TableCell className="text-right font-medium">
                    {formatCapacity(warehouse.capacity)}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge
                      className={
                        warehouse.isActive
                          ? "bg-green-500 text-white hover:bg-green-600"
                          : "bg-red-500 text-white hover:bg-red-600"
                      }
                    >
                      {warehouse.isActive ? "Aktif" : "Nonaktif"}
                    </Badge>
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
                            href={`/master/warehouses/${warehouse.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem
                            onClick={() => handleEditClick(warehouse)}
                            className="cursor-pointer"
                          >
                            <Pencil className="mr-2 h-4 w-4" />
                            Edit Gudang
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

      {/* Edit Warehouse Dialog */}
      <EditWarehouseDialog
        warehouse={editWarehouse}
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
      />
    </>
  );
}
