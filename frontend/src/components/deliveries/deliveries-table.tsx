/**
 * Deliveries Table Component
 *
 * Displays deliveries in a sortable table with:
 * - Sortable columns (deliveryNumber, deliveryDate, status, customer)
 * - Status badges with colors
 * - Type indicators
 * - Action buttons (view, edit)
 * - Responsive design
 * - Delivery tracking information
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
  Truck,
  MoreHorizontal,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import {
  getDeliveryStatusLabel,
  getDeliveryStatusColor,
  getDeliveryTypeLabel,
} from "@/types/delivery.types";
import type { DeliveryResponse } from "@/types/delivery.types";

interface DeliveriesTableProps {
  deliveries: DeliveryResponse[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  canEdit: boolean;
}

export function DeliveriesTable({
  deliveries,
  sortBy = "deliveryDate",
  sortOrder = "desc",
  onSortChange,
  canEdit,
}: DeliveriesTableProps) {
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

  // Format date to Indonesian locale
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }).format(date);
  };

  // Format time to Indonesian locale
  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
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
                  onClick={() => onSortChange("deliveryNumber")}
                >
                  No. Pengiriman
                  <SortIcon column="deliveryNumber" />
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2"
                  onClick={() => onSortChange("deliveryDate")}
                >
                  Tanggal
                  <SortIcon column="deliveryDate" />
                </Button>
              </TableHead>
              <TableHead>Customer</TableHead>
              <TableHead>Gudang</TableHead>
              <TableHead>No. SO</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="text-center">Jenis</TableHead>
              <TableHead>Driver / Ekspedisi</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {deliveries.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9}>
                  <EmptyState
                    icon={Truck}
                    title="Pengiriman tidak ditemukan"
                    description="Coba sesuaikan pencarian atau filter Anda"
                  />
                </TableCell>
              </TableRow>
            ) : (
              deliveries.map((delivery) => (
                <TableRow key={delivery.id}>
                  {/* Delivery Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    {delivery.deliveryNumber}
                  </TableCell>

                  {/* Delivery Date & Time */}
                  <TableCell>
                    <div className="font-medium">
                      {formatDate(delivery.deliveryDate)}
                    </div>
                    {delivery.departureTime && (
                      <div className="text-xs text-muted-foreground">
                        Berangkat: {formatTime(delivery.departureTime)}
                      </div>
                    )}
                    {delivery.arrivalTime && (
                      <div className="text-xs text-green-600">
                        Tiba: {formatTime(delivery.arrivalTime)}
                      </div>
                    )}
                  </TableCell>

                  {/* Customer */}
                  <TableCell>
                    {delivery.customer ? (
                      <div>
                        <div className="font-medium">
                          {delivery.customer.name}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {delivery.customer.code}
                        </div>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Warehouse */}
                  <TableCell>
                    {delivery.warehouse ? (
                      <div>
                        <div className="font-medium">
                          {delivery.warehouse.name}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {delivery.warehouse.code}
                        </div>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Sales Order Number */}
                  <TableCell>
                    {delivery.salesOrder ? (
                      <Link
                        href={`/sales/orders/${delivery.salesOrderId}`}
                        className="font-mono text-sm text-blue-600 hover:underline"
                      >
                        {delivery.salesOrder.soNumber}
                      </Link>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="text-center">
                    <Badge
                      className={`${getDeliveryStatusColor(
                        delivery.status
                      )} text-white hover:opacity-90`}
                    >
                      {getDeliveryStatusLabel(delivery.status)}
                    </Badge>
                    {delivery.receivedBy && (
                      <div className="text-xs text-muted-foreground mt-1">
                        Diterima: {delivery.receivedBy}
                      </div>
                    )}
                  </TableCell>

                  {/* Type */}
                  <TableCell className="text-center">
                    <Badge variant="outline">
                      {getDeliveryTypeLabel(delivery.type)}
                    </Badge>
                  </TableCell>

                  {/* Driver / Expedition */}
                  <TableCell>
                    {delivery.driverName && (
                      <div className="text-sm">
                        <div className="font-medium">
                          {delivery.driverName}
                        </div>
                        {delivery.vehicleNumber && (
                          <div className="text-xs text-muted-foreground">
                            {delivery.vehicleNumber}
                          </div>
                        )}
                      </div>
                    )}
                    {delivery.expeditionService && (
                      <div className="text-sm">
                        <div className="font-medium">
                          {delivery.expeditionService}
                        </div>
                        {delivery.ttnkNumber && (
                          <div className="text-xs text-muted-foreground font-mono">
                            {delivery.ttnkNumber}
                          </div>
                        )}
                      </div>
                    )}
                    {!delivery.driverName && !delivery.expeditionService && (
                      <span className="text-muted-foreground text-sm">-</span>
                    )}
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
                            href={`/sales/deliveries/${delivery.id}`}
                            className="cursor-pointer"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            Lihat Detail
                          </Link>
                        </DropdownMenuItem>
                        {canEdit &&
                          delivery.status !== "CANCELLED" &&
                          delivery.status !== "CONFIRMED" && (
                            <DropdownMenuItem asChild>
                              <Link
                                href={`/sales/deliveries/${delivery.id}/edit`}
                                className="cursor-pointer"
                              >
                                <Pencil className="mr-2 h-4 w-4" />
                                Update Status
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
