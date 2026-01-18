/**
 * Product Suppliers Section Component
 *
 * Reusable component for managing product-supplier relationships
 * within create and edit product forms.
 *
 * Features:
 * - Add/Edit/Remove suppliers
 * - Supplier-specific pricing
 * - Lead time and MOQ configuration
 * - Primary supplier setting
 */

"use client";

import { useState } from "react";
import {
  Building2,
  Plus,
  Trash2,
  Edit2,
  Star,
  AlertCircle,
  X,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { useListSuppliersQuery } from "@/store/services/supplierApi";

// Local supplier data structure for form state
export interface SupplierFormData {
  id?: string; // Only present for existing suppliers (edit mode)
  supplierId: string;
  supplierCode?: string;
  supplierName?: string;
  supplierPrice: string;
  leadTimeDays: string;
  minimumOrderQty: string;
  supplierProductCode: string;
  supplierProductName: string;
  isPrimarySupplier: boolean;
  isNew?: boolean; // Flag to indicate newly added supplier
  isEdited?: boolean; // Flag to indicate edited supplier
  isDeleted?: boolean; // Flag to indicate deleted supplier
}

interface ProductSuppliersSectionProps {
  suppliers: SupplierFormData[];
  onSuppliersChange: (suppliers: SupplierFormData[]) => void;
  existingSupplierIds?: string[]; // IDs of suppliers already linked (to exclude from dropdown)
  disabled?: boolean;
}

const formatCurrency = (value: string | number): string => {
  const num = typeof value === "string" ? parseFloat(value || "0") : value;
  if (isNaN(num)) return "Rp 0";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num);
};

export function ProductSuppliersSection({
  suppliers,
  onSuppliersChange,
  existingSupplierIds = [],
  disabled = false,
}: ProductSuppliersSectionProps) {
  const [isAddingSupplier, setIsAddingSupplier] = useState(false);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [formData, setFormData] = useState<SupplierFormData>({
    supplierId: "",
    supplierPrice: "",
    leadTimeDays: "7",
    minimumOrderQty: "",
    supplierProductCode: "",
    supplierProductName: "",
    isPrimarySupplier: false,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Fetch available suppliers
  const { data: suppliersData } = useListSuppliersQuery({
    isActive: true,
    pageSize: 100,
  });

  // Get list of suppliers excluding already linked ones
  const linkedSupplierIds = [
    ...existingSupplierIds,
    ...suppliers.filter((s) => !s.isDeleted).map((s) => s.supplierId),
  ];

  const availableSuppliers = (suppliersData?.data || []).filter(
    (s) =>
      !linkedSupplierIds.includes(s.id) ||
      (editingIndex !== null && suppliers[editingIndex]?.supplierId === s.id)
  );

  const resetForm = () => {
    setFormData({
      supplierId: "",
      supplierPrice: "",
      leadTimeDays: "7",
      minimumOrderQty: "",
      supplierProductCode: "",
      supplierProductName: "",
      isPrimarySupplier: suppliers.filter((s) => !s.isDeleted).length === 0,
    });
    setErrors({});
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.supplierId) {
      newErrors.supplierId = "Supplier wajib dipilih";
    }

    if (!formData.supplierPrice || parseFloat(formData.supplierPrice) <= 0) {
      newErrors.supplierPrice = "Harga supplier wajib diisi";
    }

    if (formData.leadTimeDays && parseInt(formData.leadTimeDays) < 0) {
      newErrors.leadTimeDays = "Lead time tidak boleh negatif";
    }

    if (formData.minimumOrderQty && parseFloat(formData.minimumOrderQty) < 0) {
      newErrors.minimumOrderQty = "MOQ tidak boleh negatif";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleAddSupplier = () => {
    if (!validateForm()) return;

    const selectedSupplier = suppliersData?.data?.find(
      (s) => s.id === formData.supplierId
    );

    const newSupplier: SupplierFormData = {
      ...formData,
      supplierCode: selectedSupplier?.code,
      supplierName: selectedSupplier?.name,
      isNew: true,
    };

    // If this is primary, unset other primaries
    let updatedSuppliers = [...suppliers];
    if (newSupplier.isPrimarySupplier) {
      updatedSuppliers = updatedSuppliers.map((s) => ({
        ...s,
        isPrimarySupplier: false,
        isEdited: s.id ? true : s.isEdited,
      }));
    }

    onSuppliersChange([...updatedSuppliers, newSupplier]);
    setIsAddingSupplier(false);
    resetForm();
  };

  const handleEditSupplier = () => {
    if (editingIndex === null || !validateForm()) return;

    const selectedSupplier = suppliersData?.data?.find(
      (s) => s.id === formData.supplierId
    );

    const updatedSupplier: SupplierFormData = {
      ...suppliers[editingIndex],
      ...formData,
      supplierCode: selectedSupplier?.code,
      supplierName: selectedSupplier?.name,
      isEdited: suppliers[editingIndex].id
        ? true
        : suppliers[editingIndex].isEdited,
    };

    // If this is primary, unset other primaries
    let updatedSuppliers = [...suppliers];
    if (updatedSupplier.isPrimarySupplier) {
      updatedSuppliers = updatedSuppliers.map((s, i) => ({
        ...s,
        isPrimarySupplier: i === editingIndex ? true : false,
        isEdited: i !== editingIndex && s.id ? true : s.isEdited,
      }));
    }

    updatedSuppliers[editingIndex] = updatedSupplier;
    onSuppliersChange(updatedSuppliers);
    setEditingIndex(null);
    resetForm();
  };

  const handleRemoveSupplier = (index: number) => {
    const supplier = suppliers[index];

    if (supplier.id) {
      // Existing supplier - mark as deleted
      const updatedSuppliers = [...suppliers];
      updatedSuppliers[index] = { ...supplier, isDeleted: true };
      onSuppliersChange(updatedSuppliers);
    } else {
      // New supplier - just remove from array
      const updatedSuppliers = suppliers.filter((_, i) => i !== index);
      onSuppliersChange(updatedSuppliers);
    }
  };

  const handleSetPrimary = (index: number) => {
    const updatedSuppliers = suppliers.map((s, i) => ({
      ...s,
      isPrimarySupplier: i === index,
      isEdited: s.id ? true : s.isEdited,
    }));
    onSuppliersChange(updatedSuppliers);
  };

  const startEdit = (index: number) => {
    const supplier = suppliers[index];
    setFormData({
      supplierId: supplier.supplierId,
      supplierPrice: supplier.supplierPrice,
      leadTimeDays: supplier.leadTimeDays || "7",
      minimumOrderQty: supplier.minimumOrderQty || "",
      supplierProductCode: supplier.supplierProductCode || "",
      supplierProductName: supplier.supplierProductName || "",
      isPrimarySupplier: supplier.isPrimarySupplier,
    });
    setEditingIndex(index);
    setIsAddingSupplier(false);
  };

  const cancelEdit = () => {
    setEditingIndex(null);
    setIsAddingSupplier(false);
    resetForm();
  };

  const startAdd = () => {
    resetForm();
    setEditingIndex(null);
    setIsAddingSupplier(true);
  };

  // Filter out deleted suppliers for display
  const visibleSuppliers = suppliers.filter((s) => !s.isDeleted);

  return (
    <Card className="border-2">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Building2 className="h-5 w-5" />
            Suppliers
          </CardTitle>
          {!isAddingSupplier && editingIndex === null && !disabled && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={startAdd}
              disabled={availableSuppliers.length === 0}
            >
              <Plus className="mr-2 h-4 w-4" />
              Tambah Supplier
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Add/Edit Form */}
        {(isAddingSupplier || editingIndex !== null) && (
          <div className="border rounded-lg p-4 bg-muted/30 space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">
                {editingIndex !== null
                  ? "Edit Supplier"
                  : "Tambah Supplier Baru"}
              </h4>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={cancelEdit}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              {/* Supplier Selection */}
              <div className="space-y-2 sm:col-span-2">
                <Label>
                  Supplier <span className="text-destructive">*</span>
                </Label>
                <Select
                  value={formData.supplierId}
                  onValueChange={(value) =>
                    setFormData((prev) => ({ ...prev, supplierId: value }))
                  }
                  disabled={editingIndex !== null}
                >
                  <SelectTrigger
                    className={`w-full ${
                      errors.supplierId ? "border-destructive" : ""
                    }`}
                  >
                    <SelectValue placeholder="Pilih supplier..." />
                  </SelectTrigger>
                  <SelectContent>
                    {availableSuppliers.length === 0 ? (
                      <div className="p-2 text-sm text-muted-foreground text-center">
                        Semua supplier sudah terhubung
                      </div>
                    ) : (
                      availableSuppliers.map((supplier) => (
                        <SelectItem key={supplier.id} value={supplier.id}>
                          {supplier.code} - {supplier.name}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
                {errors.supplierId && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.supplierId}
                  </p>
                )}
              </div>

              {/* Supplier Price */}
              <div className="space-y-2">
                <Label>
                  Harga dari Supplier{" "}
                  <span className="text-destructive">*</span>
                </Label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                    Rp
                  </span>
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    value={formData.supplierPrice}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        supplierPrice: e.target.value,
                      }))
                    }
                    className={`pl-10 ${
                      errors.supplierPrice ? "border-destructive" : ""
                    }`}
                    placeholder="0"
                  />
                </div>
                {errors.supplierPrice && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.supplierPrice}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  {formatCurrency(formData.supplierPrice)}
                </p>
              </div>

              {/* Lead Time */}
              <div className="space-y-2">
                <Label>Lead Time (hari)</Label>
                <Input
                  type="number"
                  min="0"
                  value={formData.leadTimeDays}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      leadTimeDays: e.target.value,
                    }))
                  }
                  className={errors.leadTimeDays ? "border-destructive" : ""}
                  placeholder="7"
                />
                {errors.leadTimeDays && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.leadTimeDays}
                  </p>
                )}
              </div>

              {/* Min Order Qty */}
              <div className="space-y-2">
                <Label>Min. Order Qty</Label>
                <Input
                  type="number"
                  step="0.001"
                  min="0"
                  value={formData.minimumOrderQty}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      minimumOrderQty: e.target.value,
                    }))
                  }
                  className={errors.minimumOrderQty ? "border-destructive" : ""}
                  placeholder="1"
                />
                {errors.minimumOrderQty && (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <AlertCircle className="h-3 w-3" />
                    {errors.minimumOrderQty}
                  </p>
                )}
              </div>

              {/* Supplier Product Code */}
              <div className="space-y-2">
                <Label>Kode Produk di Supplier</Label>
                <Input
                  value={formData.supplierProductCode}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      supplierProductCode: e.target.value,
                    }))
                  }
                  placeholder="Opsional"
                />
              </div>

              {/* Supplier Product Name */}
              <div className="space-y-2">
                <Label>Nama Produk di Supplier</Label>
                <Input
                  value={formData.supplierProductName}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      supplierProductName: e.target.value,
                    }))
                  }
                  placeholder="Opsional"
                />
              </div>

              {/* Primary Supplier Checkbox */}
              <div className="sm:col-span-2 flex items-center space-x-2">
                <Checkbox
                  id="isPrimary"
                  checked={formData.isPrimarySupplier}
                  onCheckedChange={(checked) =>
                    setFormData((prev) => ({
                      ...prev,
                      isPrimarySupplier: checked === true,
                    }))
                  }
                />
                <Label
                  htmlFor="isPrimary"
                  className="text-sm font-normal cursor-pointer"
                >
                  Jadikan sebagai supplier utama
                </Label>
              </div>
            </div>

            {/* Form Actions */}
            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={cancelEdit}>
                Batal
              </Button>
              <Button
                type="button"
                onClick={
                  editingIndex !== null ? handleEditSupplier : handleAddSupplier
                }
              >
                <Check className="mr-2 h-4 w-4" />
                {editingIndex !== null ? "Simpan Perubahan" : "Tambah Supplier"}
              </Button>
            </div>
          </div>
        )}

        {/* Suppliers List */}
        {visibleSuppliers.length > 0 ? (
          <>
            {/* Desktop Table View */}
            <div className="hidden md:block border rounded-lg overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Supplier</TableHead>
                    <TableHead className="text-right">Harga</TableHead>
                    <TableHead className="text-center">Lead Time</TableHead>
                    <TableHead className="text-center">MOQ</TableHead>
                    <TableHead className="text-center">Status</TableHead>
                    {!disabled && <TableHead className="w-25"></TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visibleSuppliers.map((supplier, index) => {
                    const actualIndex = suppliers.findIndex(
                      (s) => s === supplier
                    );
                    return (
                      <TableRow key={supplier.supplierId + index}>
                        <TableCell>
                          <div>
                            <div className="font-medium">
                              {supplier.supplierCode} - {supplier.supplierName}
                            </div>
                            {supplier.supplierProductCode && (
                              <div className="text-xs text-muted-foreground">
                                Kode: {supplier.supplierProductCode}
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-right font-mono">
                          {formatCurrency(supplier.supplierPrice)}
                        </TableCell>
                        <TableCell className="text-center">
                          {supplier.leadTimeDays || "-"} hari
                        </TableCell>
                        <TableCell className="text-center">
                          {supplier.minimumOrderQty || "-"}
                        </TableCell>
                        <TableCell className="text-center">
                          {supplier.isPrimarySupplier ? (
                            <Badge variant="default" className="gap-1">
                              <Star className="h-3 w-3" />
                              Utama
                            </Badge>
                          ) : (
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => handleSetPrimary(actualIndex)}
                              disabled={disabled}
                              className="text-xs"
                            >
                              Set Utama
                            </Button>
                          )}
                        </TableCell>
                        {!disabled && (
                          <TableCell>
                            <div className="flex items-center gap-1">
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => startEdit(actualIndex)}
                                disabled={
                                  editingIndex !== null || isAddingSupplier
                                }
                              >
                                <Edit2 className="h-4 w-4" />
                              </Button>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemoveSupplier(actualIndex)}
                                className="text-destructive hover:text-destructive"
                                disabled={
                                  editingIndex !== null || isAddingSupplier
                                }
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </div>
                          </TableCell>
                        )}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>

            {/* Mobile Card View */}
            <div className="md:hidden space-y-3">
              {visibleSuppliers.map((supplier, index) => {
                const actualIndex = suppliers.findIndex((s) => s === supplier);
                return (
                  <div
                    key={supplier.supplierId + index}
                    className="border rounded-lg p-4 space-y-3"
                  >
                    {/* Header with name and primary badge */}
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">
                          {supplier.supplierCode} - {supplier.supplierName}
                        </div>
                        {supplier.supplierProductCode && (
                          <div className="text-xs text-muted-foreground">
                            Kode: {supplier.supplierProductCode}
                          </div>
                        )}
                      </div>
                      {supplier.isPrimarySupplier && (
                        <Badge variant="default" className="gap-1 shrink-0">
                          <Star className="h-3 w-3" />
                          Utama
                        </Badge>
                      )}
                    </div>

                    {/* Details Grid */}
                    <div className="grid grid-cols-3 gap-2 text-sm">
                      <div>
                        <div className="text-muted-foreground text-xs">
                          Harga
                        </div>
                        <div className="font-mono font-medium">
                          {formatCurrency(supplier.supplierPrice)}
                        </div>
                      </div>
                      <div>
                        <div className="text-muted-foreground text-xs">
                          Lead Time
                        </div>
                        <div>{supplier.leadTimeDays || "-"} hari</div>
                      </div>
                      <div>
                        <div className="text-muted-foreground text-xs">MOQ</div>
                        <div>{supplier.minimumOrderQty || "-"}</div>
                      </div>
                    </div>

                    {/* Actions */}
                    {!disabled && (
                      <div className="flex items-center gap-2 pt-2 border-t">
                        {!supplier.isPrimarySupplier && (
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleSetPrimary(actualIndex)}
                            className="text-xs flex-1"
                          >
                            <Star className="mr-1 h-3 w-3" />
                            Set Utama
                          </Button>
                        )}
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => startEdit(actualIndex)}
                          disabled={editingIndex !== null || isAddingSupplier}
                          className="flex-1"
                        >
                          <Edit2 className="mr-1 h-4 w-4" />
                          Edit
                        </Button>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => handleRemoveSupplier(actualIndex)}
                          className="text-destructive hover:text-destructive flex-1"
                          disabled={editingIndex !== null || isAddingSupplier}
                        >
                          <Trash2 className="mr-1 h-4 w-4" />
                          Hapus
                        </Button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </>
        ) : !isAddingSupplier && editingIndex === null ? (
          <div className="text-center py-8 text-muted-foreground border rounded-lg">
            <Building2 className="mx-auto h-12 w-12 mb-4 opacity-50" />
            <p>Belum ada supplier terhubung</p>
            <p className="text-sm">
              {disabled
                ? "Supplier dapat ditambahkan setelah produk dibuat"
                : "Klik 'Tambah Supplier' untuk menghubungkan supplier"}
            </p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
