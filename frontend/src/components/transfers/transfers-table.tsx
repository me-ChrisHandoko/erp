/**
 * Transfers Table Component
 *
 * Displays stock transfers in a sortable table with:
 * - Sortable columns (transferNumber, transferDate, status)
 * - Status badges with color coding
 * - Context-based action buttons (ship, receive, cancel, edit, delete)
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
  Eye,
  Send,
  CheckCircle,
  XCircle,
  Edit,
  Trash,
  MoreHorizontal,
  PackageOpen,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { TransferStatusBadge } from "./transfer-status-badge";
import type { StockTransfer } from "@/types/transfer.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";

interface TransfersTableProps {
  transfers: StockTransfer[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  onView: (transfer: StockTransfer) => void;
  onShip?: (transfer: StockTransfer) => void;
  onReceive?: (transfer: StockTransfer) => void;
  onCancel?: (transfer: StockTransfer) => void;
  onEdit?: (transfer: StockTransfer) => void;
  onDelete?: (transfer: StockTransfer) => void;
  canEdit: boolean;
  canDelete: boolean;
  canApprove: boolean;
}

export function TransfersTable({
  transfers,
  sortBy = "transferNumber",
  sortOrder = "desc",
  onSortChange,
  onView,
  onShip,
  onReceive,
  onCancel,
  onEdit,
  onDelete,
  canEdit,
  canDelete,
  canApprove,
}: TransfersTableProps) {
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

  // Determine available actions based on status and permissions
  const getAvailableActions = (transfer: StockTransfer) => {
    const actions = [];

    // Always allow view
    actions.push({
      label: "Lihat Detail",
      icon: Eye,
      onClick: () => onView(transfer),
      variant: "default" as const,
    });

    // Status-based actions
    if (transfer.status === "DRAFT") {
      // DRAFT: Can edit, delete, and ship
      if (canEdit && onEdit) {
        actions.push({
          label: "Edit",
          icon: Edit,
          onClick: () => onEdit(transfer),
          variant: "default" as const,
        });
      }
      if (canDelete && onDelete) {
        actions.push({
          label: "Hapus",
          icon: Trash,
          onClick: () => onDelete(transfer),
          variant: "destructive" as const,
        });
      }
      if ((canEdit || canApprove) && onShip) {
        actions.push({
          label: "Kirim Transfer",
          icon: Send,
          onClick: () => onShip(transfer),
          variant: "default" as const,
        });
      }
    } else if (transfer.status === "SHIPPED") {
      // SHIPPED: Can receive or cancel
      if (canApprove && onReceive) {
        actions.push({
          label: "Terima Transfer",
          icon: CheckCircle,
          onClick: () => onReceive(transfer),
          variant: "default" as const,
        });
      }
      if (canApprove && onCancel) {
        actions.push({
          label: "Batalkan",
          icon: XCircle,
          onClick: () => onCancel(transfer),
          variant: "destructive" as const,
        });
      }
    }
    // RECEIVED and CANCELLED: View only (no additional actions)

    return actions;
  };

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("transferNumber")}
              >
                No. Transfer
                <SortIcon column="transferNumber" />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("transferDate")}
              >
                Tanggal
                <SortIcon column="transferDate" />
              </Button>
            </TableHead>
            <TableHead>Dari Gudang</TableHead>
            <TableHead>Ke Gudang</TableHead>
            <TableHead className="text-center">Jumlah Item</TableHead>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("status")}
              >
                Status
                <SortIcon column="status" />
              </Button>
            </TableHead>
            <TableHead className="w-[70px]">
              <span className="sr-only">Aksi</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {transfers.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7}>
                <EmptyState
                  icon={PackageOpen}
                  title="Transfer tidak ditemukan"
                  description="Coba sesuaikan pencarian atau filter Anda"
                />
              </TableCell>
            </TableRow>
          ) : (
            transfers.map((transfer) => {
              const actions = getAvailableActions(transfer);
              const primaryAction = actions[0]; // View is always first
              const secondaryActions = actions.slice(1);

              return (
                <TableRow
                  key={transfer.id}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => onView(transfer)}
                >
                  {/* Transfer Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    {transfer.transferNumber}
                  </TableCell>

                  {/* Transfer Date */}
                  <TableCell className="text-sm">
                    {format(new Date(transfer.transferDate), "dd MMM yyyy", {
                      locale: localeId,
                    })}
                  </TableCell>

                  {/* Source Warehouse */}
                  <TableCell className="text-sm">
                    {transfer.sourceWarehouse?.name || "-"}
                  </TableCell>

                  {/* Destination Warehouse */}
                  <TableCell className="text-sm">
                    {transfer.destWarehouse?.name || "-"}
                  </TableCell>

                  {/* Items Count */}
                  <TableCell className="text-center text-sm">
                    {transfer.items?.length || 0}
                  </TableCell>

                  {/* Status Badge */}
                  <TableCell>
                    <TransferStatusBadge status={transfer.status} />
                  </TableCell>

                  {/* Actions */}
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    {secondaryActions.length > 0 ? (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <span className="sr-only">Buka menu</span>
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                          <DropdownMenuSeparator />
                          {actions.map((action, index) => {
                            const Icon = action.icon;
                            return (
                              <DropdownMenuItem
                                key={index}
                                onClick={action.onClick}
                                className={
                                  action.variant === "destructive"
                                    ? "text-destructive focus:text-destructive"
                                    : ""
                                }
                              >
                                <Icon className="mr-2 h-4 w-4" />
                                {action.label}
                              </DropdownMenuItem>
                            );
                          })}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    ) : (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => primaryAction.onClick()}
                        className="h-8"
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>
    </div>
  );
}
