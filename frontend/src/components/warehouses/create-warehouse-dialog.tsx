/**
 * Create Warehouse Dialog Component
 *
 * Modal dialog for creating new warehouses.
 * Wraps the CreateWarehouseForm in a responsive dialog.
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CreateWarehouseForm } from "./create-warehouse-form";

interface CreateWarehouseDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateWarehouseDialog({
  open,
  onOpenChange,
}: CreateWarehouseDialogProps) {
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
          <DialogTitle>Tambah Gudang Baru</DialogTitle>
          <DialogDescription>
            Tambahkan gudang baru untuk manajemen inventori
          </DialogDescription>
        </DialogHeader>
        <CreateWarehouseForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </DialogContent>
    </Dialog>
  );
}
