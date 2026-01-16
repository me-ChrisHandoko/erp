/**
 * Edit Invoice Form Component
 *
 * Form for editing existing purchase invoices with:
 * - Pre-filled invoice data
 * - Invoice header editing
 * - Line items management
 * - Financial calculations
 * - Real-time validation
 */

"use client";

import { useState, useEffect } from "react";
import {
  FileText,
  Building2,
  Calendar,
  Package,
  Save,
  AlertCircle,
  Plus,
  Trash2,
  DollarSign,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useUpdatePurchaseInvoiceMutation } from "@/store/services/purchaseInvoiceApi";
import { useListPurchaseOrdersQuery } from "@/store/services/purchaseOrderApi";
import { toast } from "sonner";
import type {
  PurchaseInvoiceResponse,
  UpdatePurchaseInvoiceRequest,
} from "@/types/purchase-invoice.types";

interface EditInvoiceFormProps {
  invoice: PurchaseInvoiceResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

interface InvoiceItem {
  id: string;
  productId: string;
  productCode: string;
  productName: string;
  quantity: string;
  unitName: string;
  unitPrice: string;
  discountPercent: string;
  discountAmount: string;
  taxPercent: string;
  taxAmount: string;
  totalAmount: string;
}

export function EditInvoiceForm({
  invoice,
  onSuccess,
  onCancel,
}: EditInvoiceFormProps) {
  const [updateInvoice, { isLoading }] = useUpdatePurchaseInvoiceMutation();

  // Initialize form data from invoice
  const [formData, setFormData] = useState({
    supplierId: invoice.supplierId,
    supplierName: invoice.supplierName,
    invoiceNumber: invoice.invoiceNumber,
    invoiceDate: invoice.invoiceDate.split("T")[0], // Convert to YYYY-MM-DD
    dueDate: invoice.dueDate.split("T")[0],
    purchaseOrderId: invoice.purchaseOrderId || "",
    purchaseOrderNumber: invoice.poNumber || "",
    notes: invoice.notes || "",
    taxPercentage: invoice.taxRate || "11",
  });

  // Fetch POs filtered by selected supplier
  const { data: purchaseOrdersData, isFetching: isLoadingPOs } = useListPurchaseOrdersQuery(
    {
      supplierId: formData.supplierId,
      status: "CONFIRMED", // Only show confirmed POs
      pageSize: 100
    },
    {
      skip: !formData.supplierId // Only fetch when supplier is selected
    }
  );
  const purchaseOrders = purchaseOrdersData?.data || [];

  // Initialize items from invoice
  const [items, setItems] = useState<InvoiceItem[]>(
    invoice.items?.map((item) => ({
      id: `item-${item.productId}`,
      productId: item.productId,
      productCode: item.productCode || "",
      productName: item.productName,
      quantity: item.quantity,
      unitId: item.unitId,
      unitName: item.unitName,
      unitPrice: item.unitPrice,
      discountPercent: item.discountPct || "0",
      discountAmount: item.discountAmount || "0",
      taxPercent: formData.taxPercentage, // Use invoice-level tax rate
      taxAmount: item.taxAmount || "0",
      totalAmount: item.lineTotal,
    })) || []
  );

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (field: string, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
  };

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

  const handleAddItem = () => {
    const newItem: InvoiceItem = {
      id: `item-${Date.now()}`,
      productId: "",
      productCode: "",
      productName: "",
      quantity: "1",
      unitName: "PCS",
      unitPrice: "0",
      discountPercent: "0",
      discountAmount: "0",
      taxPercent: formData.taxPercentage,
      taxAmount: "0",
      totalAmount: "0",
    };
    setItems([...items, newItem]);
  };

  const handleRemoveItem = (itemId: string) => {
    setItems(items.filter((item) => item.id !== itemId));
  };

  const handleItemChange = (itemId: string, field: keyof InvoiceItem, value: string) => {
    setItems((prevItems) =>
      prevItems.map((item) => {
        if (item.id !== itemId) return item;

        const updated = { ...item, [field]: value };

        if (["quantity", "unitPrice", "discountPercent", "taxPercent"].includes(field)) {
          const qty = parseFloat(updated.quantity || "0");
          const price = parseFloat(updated.unitPrice || "0");
          const discPct = parseFloat(updated.discountPercent || "0");
          const taxPct = parseFloat(updated.taxPercent || "0");

          const subtotal = qty * price;
          const discount = (subtotal * discPct) / 100;
          const taxableAmount = subtotal - discount;
          const tax = (taxableAmount * taxPct) / 100;
          const total = taxableAmount + tax;

          updated.discountAmount = discount.toFixed(2);
          updated.taxAmount = tax.toFixed(2);
          updated.totalAmount = total.toFixed(2);
        }

        return updated;
      })
    );
  };

  const calculateTotals = () => {
    const subtotal = items.reduce((sum, item) => {
      const qty = parseFloat(item.quantity || "0");
      const price = parseFloat(item.unitPrice || "0");
      return sum + qty * price;
    }, 0);

    const totalDiscount = items.reduce((sum, item) => {
      return sum + parseFloat(item.discountAmount || "0");
    }, 0);

    const totalTax = items.reduce((sum, item) => {
      return sum + parseFloat(item.taxAmount || "0");
    }, 0);

    const total = subtotal - totalDiscount + totalTax;

    return { subtotal, totalDiscount, totalTax, total };
  };

  const totals = calculateTotals();

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // invoiceNumber is read-only (already generated), no validation needed
    if (!formData.invoiceDate) {
      newErrors.invoiceDate = "Tanggal faktur wajib diisi";
    }
    if (!formData.dueDate) {
      newErrors.dueDate = "Jatuh tempo wajib diisi";
    }

    if (formData.invoiceDate && formData.dueDate) {
      const invoiceDate = new Date(formData.invoiceDate);
      const dueDate = new Date(formData.dueDate);
      if (dueDate < invoiceDate) {
        newErrors.dueDate = "Jatuh tempo tidak boleh lebih awal dari tanggal faktur";
      }
    }

    if (items.length === 0) {
      newErrors.items = "Minimal harus ada 1 item";
    } else {
      items.forEach((item, index) => {
        if (!item.productName.trim()) {
          newErrors[`item-${index}-product`] = "Produk wajib diisi";
        }
        const qty = parseFloat(item.quantity || "0");
        const price = parseFloat(item.unitPrice || "0");
        if (qty <= 0) {
          newErrors[`item-${index}-quantity`] = "Kuantitas harus > 0";
        }
        if (price <= 0) {
          newErrors[`item-${index}-price`] = "Harga harus > 0";
        }
      });
    }

    setErrors(newErrors);
    setTouched({
      invoiceDate: true,
      dueDate: true,
    });

    if (Object.keys(newErrors).length > 0) {
      console.log("❌ Validation errors:", newErrors);
    }

    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      toast.error("Validasi Gagal", {
        description: "Mohon periksa kembali form Anda",
      });
      return;
    }

    try {
      const updateData: UpdatePurchaseInvoiceRequest = {
        invoiceDate: formData.invoiceDate,
        dueDate: formData.dueDate,
        supplierId: formData.supplierId,
        discountAmount: totals.totalDiscount > 0 ? totals.totalDiscount.toString() : undefined,
        taxRate: formData.taxPercentage,
        notes: formData.notes || undefined,
      };

      await updateInvoice({
        invoiceId: invoice.id,
        data: updateData,
      }).unwrap();

      toast.success("Faktur Berhasil Diperbarui", {
        description: `Faktur ${formData.invoiceNumber} telah diperbarui`,
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      toast.error("Gagal Memperbarui Faktur", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          error?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Invoice Header */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Informasi Faktur
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Supplier (Read-only) */}
            <div className="space-y-2 sm:col-span-2">
              <Label className="text-sm font-medium">Supplier</Label>
              <div className="flex items-center gap-2 rounded-md border bg-muted px-3 py-2">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">{formData.supplierName}</span>
              </div>
              <p className="text-xs text-muted-foreground">
                Supplier tidak dapat diubah setelah faktur dibuat
              </p>
            </div>

            {/* Invoice Number - Read-only */}
            <div className="space-y-2">
              <Label htmlFor="invoiceNumber" className="text-sm font-medium">
                Nomor Faktur
              </Label>
              <Input
                id="invoiceNumber"
                value={formData.invoiceNumber}
                disabled
                className="bg-muted cursor-not-allowed"
              />
              <p className="text-xs text-muted-foreground">
                Nomor faktur tidak dapat diubah setelah dibuat
              </p>
            </div>

            {/* Purchase Order Reference */}
            <div className="space-y-2">
              <Label htmlFor="purchaseOrder" className="text-sm font-medium">
                Referensi PO
              </Label>
              <Select
                value={formData.purchaseOrderId}
                onValueChange={(value) => {
                  const selectedPO = purchaseOrders.find(po => po.id === value);
                  handleChange("purchaseOrderId", value);
                  if (selectedPO) {
                    handleChange("purchaseOrderNumber", selectedPO.poNumber);
                  }
                }}
                disabled={!formData.supplierId}
              >
                <SelectTrigger className="w-full">
                  <SelectValue
                    placeholder={
                      !formData.supplierId
                        ? "Pilih supplier terlebih dahulu"
                        : isLoadingPOs
                        ? "Memuat PO..."
                        : purchaseOrders.length === 0
                        ? "Tidak ada PO untuk supplier ini"
                        : "Pilih PO (opsional)..."
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {purchaseOrders.map((po) => (
                    <SelectItem key={po.id} value={po.id}>
                      {po.poNumber} - {new Date(po.poDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                PO akan difilter berdasarkan supplier yang dipilih
              </p>
            </div>

            {/* Invoice Date */}
            <div className="space-y-2">
              <Label htmlFor="invoiceDate" className="text-sm font-medium">
                Tanggal Faktur <span className="text-destructive">*</span>
              </Label>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  id="invoiceDate"
                  type="date"
                  value={formData.invoiceDate}
                  onChange={(e) => handleChange("invoiceDate", e.target.value)}
                  onBlur={() => handleBlur("invoiceDate")}
                  className={
                    errors.invoiceDate && touched.invoiceDate
                      ? "border-destructive"
                      : ""
                  }
                />
              </div>
              {errors.invoiceDate && touched.invoiceDate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.invoiceDate}
                </p>
              )}
            </div>

            {/* Due Date */}
            <div className="space-y-2">
              <Label htmlFor="dueDate" className="text-sm font-medium">
                Jatuh Tempo <span className="text-destructive">*</span>
              </Label>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  id="dueDate"
                  type="date"
                  value={formData.dueDate}
                  onChange={(e) => handleChange("dueDate", e.target.value)}
                  onBlur={() => handleBlur("dueDate")}
                  className={
                    errors.dueDate && touched.dueDate
                      ? "border-destructive"
                      : ""
                  }
                />
              </div>
              {errors.dueDate && touched.dueDate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.dueDate}
                </p>
              )}
            </div>

            {/* Notes */}
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="notes" className="text-sm font-medium">
                Catatan
              </Label>
              <Textarea
                id="notes"
                value={formData.notes}
                onChange={(e) => handleChange("notes", e.target.value)}
                placeholder="Catatan tambahan untuk faktur ini..."
                rows={3}
                className="resize-none"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Invoice Items */}
      <Card className="border-2">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Package className="h-5 w-5" />
              Item Faktur
            </CardTitle>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddItem}
              disabled={isLoading}
            >
              <Plus className="mr-2 h-4 w-4" />
              Tambah Item
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {errors.items && (
            <Alert variant="destructive" className="mb-4">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{errors.items}</AlertDescription>
            </Alert>
          )}

          {items.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Package className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>Belum ada item dalam faktur</p>
              <p className="text-sm">Klik "Tambah Item" untuk menambahkan produk</p>
            </div>
          ) : (
            <div className="rounded-md border overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[250px]">Produk</TableHead>
                    <TableHead className="w-[100px]">Qty</TableHead>
                    <TableHead className="w-[80px]">Unit</TableHead>
                    <TableHead className="w-[120px] text-right">Harga</TableHead>
                    <TableHead className="w-[80px] text-right">Diskon %</TableHead>
                    <TableHead className="w-[80px] text-right">PPN %</TableHead>
                    <TableHead className="w-[120px] text-right">Total</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item, index) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <Input
                          value={item.productName}
                          onChange={(e) =>
                            handleItemChange(item.id, "productName", e.target.value)
                          }
                          placeholder="Nama produk"
                          className="h-8"
                        />
                        {errors[`item-${index}-product`] && (
                          <p className="text-xs text-destructive mt-1">
                            {errors[`item-${index}-product`]}
                          </p>
                        )}
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          value={item.quantity}
                          onChange={(e) =>
                            handleItemChange(item.id, "quantity", e.target.value)
                          }
                          className="h-8"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          value={item.unitName}
                          onChange={(e) =>
                            handleItemChange(item.id, "unitName", e.target.value)
                          }
                          className="h-8"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          value={item.unitPrice}
                          onChange={(e) =>
                            handleItemChange(item.id, "unitPrice", e.target.value)
                          }
                          className="h-8 text-right"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          value={item.discountPercent}
                          onChange={(e) =>
                            handleItemChange(item.id, "discountPercent", e.target.value)
                          }
                          className="h-8 text-right"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          value={item.taxPercent}
                          onChange={(e) =>
                            handleItemChange(item.id, "taxPercent", e.target.value)
                          }
                          className="h-8 text-right"
                        />
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {formatCurrency(item.totalAmount)}
                      </TableCell>
                      <TableCell>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveItem(item.id)}
                          disabled={isLoading}
                          className="h-8 w-8 p-0"
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Financial Summary */}
      {items.length > 0 && (
        <Card className="border-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5" />
              Ringkasan Keuangan
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Subtotal:</span>
                <span className="font-mono">{formatCurrency(totals.subtotal)}</span>
              </div>
              {totals.totalDiscount > 0 && (
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Diskon:</span>
                  <span className="font-mono text-green-600">
                    - {formatCurrency(totals.totalDiscount)}
                  </span>
                </div>
              )}
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">PPN:</span>
                <span className="font-mono">{formatCurrency(totals.totalTax)}</span>
              </div>
              <Separator />
              <div className="flex justify-between items-center rounded-lg bg-muted/50 p-4">
                <span className="text-lg font-bold">Total:</span>
                <span className="text-2xl font-bold text-blue-600">
                  {formatCurrency(totals.total)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Form Actions */}
      <div className="flex justify-end gap-3 pt-2">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
            size="lg"
          >
            Batal
          </Button>
        )}
        <Button
          type="submit"
          disabled={isLoading || items.length === 0}
          size="lg"
          className="min-w-[150px]"
        >
          {isLoading ? (
            <>
              <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              Menyimpan...
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Simpan Perubahan
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
