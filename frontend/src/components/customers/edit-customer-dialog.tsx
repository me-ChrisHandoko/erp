/**
 * Edit Customer Dialog Component
 *
 * Modal dialog for editing existing customers
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EditCustomerForm } from "./edit-customer-form";
import type { CustomerResponse } from "@/types/customer.types";

interface EditCustomerDialogProps {
  customer: CustomerResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditCustomerDialog({
  customer,
  open,
  onOpenChange,
}: EditCustomerDialogProps) {
  const handleSuccess = () => {
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  if (!customer) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[90vh] overflow-y-auto p-6">
        <DialogHeader>
          <DialogTitle>Edit Pelanggan</DialogTitle>
          <DialogDescription>
            Perbarui informasi pelanggan {customer.code} - {customer.name}
          </DialogDescription>
        </DialogHeader>
        <EditCustomerForm
          customer={customer}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </DialogContent>
    </Dialog>
  );
}
