/**
 * Customers Table Component
 *
 * Displays customers in a sortable table with:
 * - Sortable columns (code, name, creditLimit)
 * - Status badges (active/inactive)
 * - Customer type filtering
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
  Users,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import type { CustomerResponse } from "@/types/customer.types";

interface CustomersTableProps {
  customers: CustomerResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function CustomersTable({
  customers,
  sortBy = "code",
  sortOrder = "asc",
  onSortChange,
  canEdit,
}: CustomersTableProps) {
  // Format currency helper
  const formatCurrency = (value: string | null | undefined): string => {
    if (!value) return "Rp 0";
    const num = parseFloat(value);
    if (isNaN(num)) return "Rp 0";
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(num);
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
                  Nama
                  <SortIcon column="name" />
                </Button>
              </TableHead>
              <TableHead>Tipe</TableHead>
              <TableHead>Telepon</TableHead>
              <TableHead>Kota</TableHead>
              <TableHead className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("creditLimit")}
                >
                  Limit Kredit
                  <SortIcon column="creditLimit" />
                </Button>
              </TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {customers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8}>
                  <EmptyState
                    icon={Users}
                    title="Pelanggan tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              customers.map((customer) => (
                <TableRow key={customer.id}>
                  {/* Code */}
                  <TableCell className="font-mono text-sm font-medium">
                    {customer.code}
                  </TableCell>

                  {/* Name */}
                  <TableCell>
                    <div className="font-medium">{customer.name}</div>
                    {customer.contactPerson && (
                      <div className="text-sm text-muted-foreground">
                        {customer.contactPerson}
                      </div>
                    )}
                  </TableCell>

                  {/* Customer Type */}
                  <TableCell>
                    <Badge className="bg-blue-500 text-white hover:bg-blue-600">
                      {customer.customerType}
                    </Badge>
                  </TableCell>

                  {/* Phone */}
                  <TableCell className="text-sm">
                    {customer.phone || (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* City */}
                  <TableCell className="text-sm">
                    {customer.city || (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Credit Limit */}
                  <TableCell className="text-right font-medium">
                    {formatCurrency(customer.creditLimit)}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge
                      className={
                        customer.isActive
                          ? "bg-green-500 text-white hover:bg-green-600"
                          : "bg-red-500 text-white hover:bg-red-600"
                      }
                    >
                      {customer.isActive ? "Aktif" : "Nonaktif"}
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
                            href={`/master/customers/${customer.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit && (
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/master/customers/${customer.id}/edit`}
                              className="cursor-pointer"
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit Pelanggan
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
