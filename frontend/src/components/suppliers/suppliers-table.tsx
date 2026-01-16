/**
 * Suppliers Table Component
 *
 * Displays suppliers in a sortable table with:
 * - Sortable columns (code, name, type, city)
 * - Status badges (active/inactive)
 * - Type badges (category)
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
  Building2,
  MoreHorizontal,
  Mail,
  Phone,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { SupplierResponse } from "@/types/supplier.types";

interface SuppliersTableProps {
  suppliers: SupplierResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function SuppliersTable({
  suppliers,
  sortBy = "code",
  sortOrder = "asc",
  onSortChange,
  canEdit,
}: SuppliersTableProps) {
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
                  Nama Supplier
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
              <TableHead>Kontak</TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("city")}
                >
                  Kota
                  <SortIcon column="city" />
                </Button>
              </TableHead>
              <TableHead className="text-center">Payment Terms</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {suppliers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8}>
                  <EmptyState
                    icon={Building2}
                    title="Supplier tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              suppliers.map((supplier) => (
                <TableRow key={supplier.id}>
                  {/* Code */}
                  <TableCell className="font-mono text-sm font-medium">
                    {supplier.code}
                  </TableCell>

                  {/* Name */}
                  <TableCell>
                    <div className="font-medium">{supplier.name}</div>
                    {supplier.contactPerson && (
                      <div className="text-sm text-muted-foreground">
                        PIC: {supplier.contactPerson}
                      </div>
                    )}
                  </TableCell>

                  {/* Type */}
                  <TableCell>
                    {supplier.type && (
                      <Badge className="bg-blue-500 text-white hover:bg-blue-600">
                        {supplier.type.charAt(0) + supplier.type.slice(1).toLowerCase()}
                      </Badge>
                    )}
                  </TableCell>

                  {/* Contact Info */}
                  <TableCell>
                    <div className="space-y-1 text-sm">
                      {supplier.email && (
                        <div className="flex items-center gap-1 text-muted-foreground">
                          <Mail className="h-3 w-3" />
                          <span className="truncate max-w-[200px]">
                            {supplier.email}
                          </span>
                        </div>
                      )}
                      {supplier.phone && (
                        <div className="flex items-center gap-1 text-muted-foreground">
                          <Phone className="h-3 w-3" />
                          <span>{supplier.phone}</span>
                        </div>
                      )}
                    </div>
                  </TableCell>

                  {/* City */}
                  <TableCell>
                    {supplier.city && (
                      <div className="text-sm">
                        <div className="font-medium">{supplier.city}</div>
                        {supplier.province && (
                          <div className="text-xs text-muted-foreground">
                            {supplier.province}
                          </div>
                        )}
                      </div>
                    )}
                  </TableCell>

                  {/* Payment Terms */}
                  <TableCell className="text-center">
                    {supplier.paymentTerm > 0 ? (
                      <Badge variant="outline">{supplier.paymentTerm} hari</Badge>
                    ) : (
                      <Badge variant="secondary">Tunai</Badge>
                    )}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge
                      className={
                        supplier.isActive
                          ? "bg-green-500 text-white hover:bg-green-600"
                          : "bg-red-500 text-white hover:bg-red-600"
                      }
                    >
                      {supplier.isActive ? "Aktif" : "Nonaktif"}
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
                            href={`/master/suppliers/${supplier.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/master/suppliers/${supplier.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Supplier
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
