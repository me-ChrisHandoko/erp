/**
 * Create Product Dialog Component
 *
 * Modal dialog for creating new products.
 * Wraps the CreateProductForm in a responsive dialog.
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CreateProductForm } from "./create-product-form";

interface CreateProductDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateProductDialog({
  open,
  onOpenChange,
}: CreateProductDialogProps) {
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
          <DialogTitle>Tambah Produk Baru</DialogTitle>
          <DialogDescription>
            Buat produk baru untuk katalog distribusi Anda
          </DialogDescription>
        </DialogHeader>
        <CreateProductForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </DialogContent>
    </Dialog>
  );
}
