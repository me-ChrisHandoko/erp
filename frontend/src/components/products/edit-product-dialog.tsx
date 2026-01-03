/**
 * Edit Product Dialog Component
 *
 * Modal dialog for editing existing products
 */

"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EditProductForm } from "./edit-product-form";
import type { ProductResponse } from "@/types/product.types";

interface EditProductDialogProps {
  product: ProductResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditProductDialog({
  product,
  open,
  onOpenChange,
}: EditProductDialogProps) {
  const handleSuccess = () => {
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  if (!product) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[90vh] overflow-y-auto p-6">
        <DialogHeader>
          <DialogTitle>Edit Produk</DialogTitle>
          <DialogDescription>
            Perbarui informasi produk {product.code} - {product.name}
          </DialogDescription>
        </DialogHeader>
        <EditProductForm
          product={product}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </DialogContent>
    </Dialog>
  );
}
