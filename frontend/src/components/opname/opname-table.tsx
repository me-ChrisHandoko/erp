/**
 * Opname Table Component
 *
 * Displays stock opname records in a table format with:
 * - Sortable columns
 * - Status badges with different colors
 * - Action buttons (View, Edit, Delete, Approve)
 * - Formatted dates and numbers
 */

"use client";

import { useRouter } from "next/navigation";
import { Eye, Edit, Trash2, CheckCircle, MoreHorizontal, ChevronUp, ChevronDown, ChevronsUpDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useDeleteOpnameMutation, useApproveOpnameMutation } from "@/store/services/opnameApi";
import { useToast } from "@/hooks/use-toast";
import type { StockOpname } from "@/types/opname.types";
import { OPNAME_STATUS_CONFIG } from "@/types/opname.types";
import { useState } from "react";
import { cn } from "@/lib/utils";

interface OpnameTableProps {
  opnames: StockOpname[];
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSortChange?: (sortBy: string) => void;
  canEdit?: boolean;
  canDelete?: boolean;
  canApprove?: boolean;
}

export function OpnameTable({
  opnames,
  sortBy,
  sortOrder,
  onSortChange,
  canEdit = false,
  canDelete = false,
  canApprove = false,
}: OpnameTableProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [deleteOpname] = useDeleteOpnameMutation();
  const [approveOpname] = useApproveOpnameMutation();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [approveDialogOpen, setApproveDialogOpen] = useState(false);
  const [selectedOpnameId, setSelectedOpnameId] = useState<string | null>(null);

  const getStatusBadge = (status: StockOpname["status"]) => {
    const config = OPNAME_STATUS_CONFIG[status];
    return (
      <Badge variant={config.variant} className={config.className}>
        {config.label}
      </Badge>
    );
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  };

  const formatNumber = (value: string | number) => {
    const num = typeof value === "string" ? parseFloat(value) : value;
    return num.toLocaleString("id-ID", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    });
  };

  const handleDelete = async () => {
    if (!selectedOpnameId) return;

    try {
      await deleteOpname(selectedOpnameId).unwrap();
      toast({
        title: "Berhasil",
        description: "Stock opname berhasil dihapus",
      });
      setDeleteDialogOpen(false);
      setSelectedOpnameId(null);
    } catch (error) {
      toast({
        title: "Gagal",
        description: "Gagal menghapus stock opname",
        variant: "destructive",
      });
    }
  };

  const handleApprove = async () => {
    if (!selectedOpnameId) return;

    try {
      await approveOpname({ id: selectedOpnameId }).unwrap();
      toast({
        title: "Berhasil",
        description: "Stock opname berhasil disetujui dan penyesuaian stok telah diterapkan",
      });
      setApproveDialogOpen(false);
      setSelectedOpnameId(null);
    } catch (error) {
      toast({
        title: "Gagal",
        description: "Gagal menyetujui stock opname",
        variant: "destructive",
      });
    }
  };

  const openDeleteDialog = (opnameId: string) => {
    setSelectedOpnameId(opnameId);
    setDeleteDialogOpen(true);
  };

  const openApproveDialog = (opnameId: string) => {
    setSelectedOpnameId(opnameId);
    setApproveDialogOpen(true);
  };

  // Render sortable header with indicator
  const SortableHeader = ({
    column,
    label,
    className,
  }: {
    column: string;
    label: string;
    className?: string;
  }) => {
    const isSorted = sortBy === column;
    const isAsc = sortOrder === "asc";

    return (
      <TableHead className={cn("cursor-pointer select-none", className)}>
        <button
          type="button"
          onClick={() => onSortChange?.(column)}
          className="flex items-center gap-1 hover:text-foreground transition-colors w-full"
        >
          <span>{label}</span>
          {isSorted ? (
            isAsc ? (
              <ChevronUp className="h-4 w-4 text-primary" />
            ) : (
              <ChevronDown className="h-4 w-4 text-primary" />
            )
          ) : (
            <ChevronsUpDown className="h-4 w-4 text-muted-foreground/50" />
          )}
        </button>
      </TableHead>
    );
  };

  return (
    <>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <SortableHeader column="opnameNumber" label="Nomor Opname" />
              <SortableHeader column="opnameDate" label="Tanggal" />
              <SortableHeader column="warehouse" label="Gudang" />
              <TableHead className="text-right">Total Item</TableHead>
              <TableHead className="text-right">Expected Qty</TableHead>
              <TableHead className="text-right">Actual Qty</TableHead>
              <TableHead className="text-right">Selisih</TableHead>
              <SortableHeader column="status" label="Status" />
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {opnames.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={9}
                  className="h-24 text-center text-muted-foreground"
                >
                  Tidak ada data stock opname
                </TableCell>
              </TableRow>
            ) : (
              opnames.map((opname) => (
                <TableRow key={opname.id}>
                  <TableCell className="font-mono font-medium">
                    {opname.opnameNumber}
                  </TableCell>
                  <TableCell>{formatDate(opname.opnameDate)}</TableCell>
                  <TableCell>
                    {opname.warehouseName || opname.warehouseId}
                  </TableCell>
                  <TableCell className="text-right">
                    {opname.totalItems}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatNumber(opname.totalExpectedQty)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatNumber(opname.totalActualQty)}
                  </TableCell>
                  <TableCell className="text-right">
                    <span
                      className={
                        parseFloat(opname.totalDifference) > 0
                          ? "text-green-600 font-semibold"
                          : parseFloat(opname.totalDifference) < 0
                          ? "text-red-600 font-semibold"
                          : "text-muted-foreground"
                      }
                    >
                      {parseFloat(opname.totalDifference) > 0 && "+"}
                      {formatNumber(opname.totalDifference)}
                    </span>
                  </TableCell>
                  <TableCell>{getStatusBadge(opname.status)}</TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                        <DropdownMenuSeparator />

                        {/* View Detail */}
                        <DropdownMenuItem
                          onClick={() => router.push(`/inventory/opname/${opname.id}`)}
                          className="cursor-pointer"
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          Lihat Detail
                        </DropdownMenuItem>

                        {/* Edit - Only for draft and in_progress */}
                        {canEdit &&
                          (opname.status === "draft" ||
                            opname.status === "in_progress") && (
                            <DropdownMenuItem
                              onClick={() =>
                                router.push(`/inventory/opname/edit/${opname.id}`)
                              }
                              className="cursor-pointer"
                            >
                              <Edit className="mr-2 h-4 w-4" />
                              Edit Stock Opname
                            </DropdownMenuItem>
                          )}

                        {/* Approve - Only for completed */}
                        {canApprove && opname.status === "completed" && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => openApproveDialog(opname.id)}
                              className="cursor-pointer text-green-600 focus:text-green-600"
                            >
                              <CheckCircle className="mr-2 h-4 w-4" />
                              Approve
                            </DropdownMenuItem>
                          </>
                        )}

                        {/* Delete - Only for draft */}
                        {canDelete && opname.status === "draft" && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => openDeleteDialog(opname.id)}
                              className="cursor-pointer text-red-600 focus:text-red-600"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Hapus
                            </DropdownMenuItem>
                          </>
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

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Stock Opname?</AlertDialogTitle>
            <AlertDialogDescription>
              Tindakan ini tidak dapat dibatalkan. Stock opname akan dihapus
              secara permanen dari sistem.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Hapus
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Approve Confirmation Dialog */}
      <AlertDialog open={approveDialogOpen} onOpenChange={setApproveDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Approve Stock Opname?</AlertDialogTitle>
            <AlertDialogDescription>
              Dengan menyetujui stock opname ini, penyesuaian stok akan
              diterapkan ke sistem. Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleApprove}
              className="bg-green-600 text-white hover:bg-green-700"
            >
              Approve
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
