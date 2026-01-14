/**
 * Adjustments Table Component
 *
 * Displays inventory adjustments in a sortable table with:
 * - Sortable columns (adjustmentNumber, adjustmentDate, status)
 * - Status and type badges with color coding
 * - Context-based action buttons (approve, cancel, edit, delete)
 * - Responsive design
 */

"use client";

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
  CheckCircle,
  XCircle,
  Edit,
  Trash,
  MoreHorizontal,
  ClipboardList,
} from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { AdjustmentStatusBadge } from "./adjustment-status-badge";
import { AdjustmentTypeBadge } from "./adjustment-type-badge";
import {
  ADJUSTMENT_REASON_CONFIG,
  type InventoryAdjustment,
} from "@/types/adjustment.types";
import { format } from "date-fns";
import { id as localeId } from "date-fns/locale";

interface AdjustmentsTableProps {
  adjustments: InventoryAdjustment[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange: (sortBy: string) => void;
  onView: (adjustment: InventoryAdjustment) => void;
  onApprove?: (adjustment: InventoryAdjustment) => void;
  onCancel?: (adjustment: InventoryAdjustment) => void;
  onEdit?: (adjustment: InventoryAdjustment) => void;
  onDelete?: (adjustment: InventoryAdjustment) => void;
  canEdit: boolean;
  canDelete: boolean;
  canApprove: boolean;
}

export function AdjustmentsTable({
  adjustments,
  sortBy = "adjustmentNumber",
  sortOrder = "desc",
  onSortChange,
  onView,
  onApprove,
  onCancel,
  onEdit,
  onDelete,
  canEdit,
  canDelete,
  canApprove,
}: AdjustmentsTableProps) {
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
  const getAvailableActions = (adjustment: InventoryAdjustment) => {
    const actions = [];

    // Always allow view
    actions.push({
      label: "Lihat Detail",
      icon: Eye,
      onClick: () => onView(adjustment),
      variant: "default" as const,
    });

    // Status-based actions
    if (adjustment.status === "DRAFT") {
      // DRAFT: Can edit, delete, approve, or cancel
      if (canEdit && onEdit) {
        actions.push({
          label: "Edit",
          icon: Edit,
          onClick: () => onEdit(adjustment),
          variant: "default" as const,
        });
      }
      if (canDelete && onDelete) {
        actions.push({
          label: "Hapus",
          icon: Trash,
          onClick: () => onDelete(adjustment),
          variant: "destructive" as const,
        });
      }
      if (canApprove && onApprove) {
        actions.push({
          label: "Setujui",
          icon: CheckCircle,
          onClick: () => onApprove(adjustment),
          variant: "default" as const,
        });
      }
      if (canApprove && onCancel) {
        actions.push({
          label: "Batalkan",
          icon: XCircle,
          onClick: () => onCancel(adjustment),
          variant: "destructive" as const,
        });
      }
    }
    // APPROVED and CANCELLED: View only (no additional actions)

    return actions;
  };

  // Format currency
  const formatCurrency = (value: string | number) => {
    const numValue = typeof value === "string" ? parseFloat(value) : value;
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(numValue);
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
                onClick={() => onSortChange("adjustmentNumber")}
              >
                No. Penyesuaian
                <SortIcon column="adjustmentNumber" />
              </Button>
            </TableHead>
            <TableHead>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 px-2"
                onClick={() => onSortChange("adjustmentDate")}
              >
                Tanggal
                <SortIcon column="adjustmentDate" />
              </Button>
            </TableHead>
            <TableHead>Gudang</TableHead>
            <TableHead>Tipe</TableHead>
            <TableHead>Alasan</TableHead>
            <TableHead className="text-center">Item</TableHead>
            <TableHead className="text-right">Nilai</TableHead>
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
          {adjustments.length === 0 ? (
            <TableRow>
              <TableCell colSpan={9}>
                <EmptyState
                  icon={ClipboardList}
                  title="Penyesuaian tidak ditemukan"
                  description="Coba sesuaikan pencarian atau filter Anda"
                />
              </TableCell>
            </TableRow>
          ) : (
            adjustments.map((adjustment) => {
              const actions = getAvailableActions(adjustment);
              const primaryAction = actions[0]; // View is always first
              const secondaryActions = actions.slice(1);
              const reasonConfig = ADJUSTMENT_REASON_CONFIG[adjustment.reason];

              return (
                <TableRow
                  key={adjustment.id}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => onView(adjustment)}
                >
                  {/* Adjustment Number */}
                  <TableCell className="font-mono text-sm font-medium">
                    {adjustment.adjustmentNumber}
                  </TableCell>

                  {/* Adjustment Date */}
                  <TableCell className="text-sm">
                    {format(new Date(adjustment.adjustmentDate), "dd MMM yyyy", {
                      locale: localeId,
                    })}
                  </TableCell>

                  {/* Warehouse */}
                  <TableCell className="text-sm">
                    {adjustment.warehouse?.name || "-"}
                  </TableCell>

                  {/* Type Badge */}
                  <TableCell>
                    <AdjustmentTypeBadge type={adjustment.adjustmentType} />
                  </TableCell>

                  {/* Reason */}
                  <TableCell className="text-sm">
                    {reasonConfig?.label || adjustment.reason}
                  </TableCell>

                  {/* Items Count */}
                  <TableCell className="text-center text-sm">
                    {adjustment.totalItems || adjustment.items?.length || 0}
                  </TableCell>

                  {/* Total Value */}
                  <TableCell className="text-right text-sm font-medium">
                    {formatCurrency(adjustment.totalValue)}
                  </TableCell>

                  {/* Status Badge */}
                  <TableCell>
                    <AdjustmentStatusBadge status={adjustment.status} />
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
