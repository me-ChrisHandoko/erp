/**
 * Edit Supplier Dialog Component
 *
 * Modal dialog for editing existing suppliers
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EditSupplierForm } from "./edit-supplier-form";
import type { SupplierResponse } from "@/types/supplier.types";

interface EditSupplierDialogProps {
  supplier: SupplierResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditSupplierDialog({
  supplier,
  open,
  onOpenChange,
}: EditSupplierDialogProps) {
  const handleSuccess = () => {
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  if (!supplier) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[90vh] overflow-y-auto p-6">
        <DialogHeader>
          <DialogTitle>Edit Supplier</DialogTitle>
          <DialogDescription>
            Perbarui informasi supplier {supplier.code} - {supplier.name}
          </DialogDescription>
        </DialogHeader>
        <EditSupplierForm
          supplier={supplier}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </DialogContent>
    </Dialog>
  );
}
