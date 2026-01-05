/**
 * Create Customer Dialog Component
 *
 * Modal dialog for creating new customers.
 * Wraps the CreateCustomerForm in a responsive dialog.
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CreateCustomerForm } from "./create-customer-form";

interface CreateCustomerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateCustomerDialog({
  open,
  onOpenChange,
}: CreateCustomerDialogProps) {
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
          <DialogTitle>Tambah Pelanggan Baru</DialogTitle>
          <DialogDescription>
            Buat pelanggan baru untuk manajemen distribusi Anda
          </DialogDescription>
        </DialogHeader>
        <CreateCustomerForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </DialogContent>
    </Dialog>
  );
}
