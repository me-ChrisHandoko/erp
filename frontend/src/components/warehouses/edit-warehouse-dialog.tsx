/**
 * Edit Warehouse Dialog Component
 *
 * Modal dialog for editing existing warehouses.
 * Wraps the EditWarehouseForm in a responsive dialog.
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EditWarehouseForm } from "./edit-warehouse-form";
import type { WarehouseResponse } from "@/types/warehouse.types";

interface EditWarehouseDialogProps {
  warehouse: WarehouseResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditWarehouseDialog({
  warehouse,
  open,
  onOpenChange,
}: EditWarehouseDialogProps) {
  if (!warehouse) return null;

  const handleSuccess = () => {
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[90vh] overflow-y-auto p-6">
        <DialogHeader>
          <DialogTitle>Edit Gudang</DialogTitle>
          <DialogDescription>
            Edit informasi gudang: {warehouse.name}
          </DialogDescription>
        </DialogHeader>
        <EditWarehouseForm
          warehouse={warehouse}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </DialogContent>
    </Dialog>
  );
}
