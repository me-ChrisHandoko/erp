/**
 * Create Supplier Dialog Component
 *
 * Modal dialog for creating new suppliers.
 * Wraps the CreateSupplierForm in a responsive dialog.
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CreateSupplierForm } from "./create-supplier-form";

interface CreateSupplierDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateSupplierDialog({
  open,
  onOpenChange,
}: CreateSupplierDialogProps) {
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
          <DialogTitle>Tambah Supplier Baru</DialogTitle>
          <DialogDescription>
            Buat supplier baru untuk pengadaan barang Anda
          </DialogDescription>
        </DialogHeader>
        <CreateSupplierForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </DialogContent>
    </Dialog>
  );
}
