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

import { useState } from "react";
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
  Receipt,
  ReceiptText,
  Truck,
  HandCoins,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  id: string; // Local ID for UI tracking
  originalId?: string; // Original item ID from backend (for reference in audit)
  purchaseOrderItemId?: string; // Reference to PO item
  goodsReceiptItemId?: string; // Reference to GRN item (for tracking invoiced qty)
  productId: string;
  productCode: string;
  productName: string;
  quantity: string;
  unitId: string;
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
    notes: invoice.notes || "",
    taxPercentage: invoice.taxRate || "11",
    // Header-level discount (diskon faktur)
    headerDiscountAmount: invoice.discountAmount || "0",
    // Biaya non-barang (header level)
    shippingCost: invoice.shippingCost || "0",
    handlingCost: invoice.handlingCost || "0",
    otherCost: invoice.otherCost || "0",
    otherCostDescription: invoice.otherCostDescription || "",
  });

  // Initialize items from invoice
  const [items, setItems] = useState<InvoiceItem[]>(
    invoice.items?.map((item, index) => ({
      id: `item-${index}`, // Local ID for UI tracking
      originalId: item.id, // Original item ID from backend
      purchaseOrderItemId: item.purchaseOrderItemId, // PO item reference
      goodsReceiptItemId: item.goodsReceiptItemId, // GRN item reference (for invoiced qty tracking)
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

  // Track if items have been modified to decide whether to show original or recalculated values
  const [itemsModified, setItemsModified] = useState(false);

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
    // Check if number has decimal part
    const hasDecimal = num % 1 !== 0;
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: hasDecimal ? 2 : 0,
    }).format(num);
  };

  const handleAddItem = () => {
    const newItem: InvoiceItem = {
      id: `item-${Date.now()}`,
      productId: "",
      productCode: "",
      productName: "",
      quantity: "1",
      unitId: "",
      unitName: "PCS",
      unitPrice: "0",
      discountPercent: "0",
      discountAmount: "0",
      taxPercent: formData.taxPercentage,
      taxAmount: "0",
      totalAmount: "0",
    };
    setItems([...items, newItem]);
    setItemsModified(true);
  };

  const handleRemoveItem = (itemId: string) => {
    setItems(items.filter((item) => item.id !== itemId));
    setItemsModified(true);
  };

  const handleItemChange = (itemId: string, field: keyof InvoiceItem, value: string) => {
    setItemsModified(true);
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
    // Calculate subtotal (qty × price)
    const subtotal = items.reduce((sum, item) => {
      const qty = parseFloat(item.quantity || "0");
      const price = parseFloat(item.unitPrice || "0");
      return sum + qty * price;
    }, 0);

    // Line-level discount total
    const lineDiscount = items.reduce((sum, item) => {
      return sum + parseFloat(item.discountAmount || "0");
    }, 0);

    // Header-level discount (additional discount on entire invoice)
    const headerDiscount = parseFloat(formData.headerDiscountAmount || "0");

    // Total discount = line discount + header discount
    const totalDiscount = lineDiscount + headerDiscount;

    // Line-level tax total
    const totalTax = items.reduce((sum, item) => {
      return sum + parseFloat(item.taxAmount || "0");
    }, 0);

    // Non-goods costs (header level)
    const shippingCost = parseFloat(formData.shippingCost || "0");
    const handlingCost = parseFloat(formData.handlingCost || "0");
    const otherCost = parseFloat(formData.otherCost || "0");
    const totalNonGoodsCost = shippingCost + handlingCost + otherCost;

    // Total = subtotal - total discount + tax + non-goods costs
    const total = subtotal - totalDiscount + totalTax + totalNonGoodsCost;

    return {
      subtotal,
      lineDiscount,
      headerDiscount,
      totalDiscount,
      totalTax,
      shippingCost,
      handlingCost,
      otherCost,
      totalNonGoodsCost,
      total,
    };
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
      // Use header discount (not total discount which includes line-level)
      const headerDiscount = parseFloat(formData.headerDiscountAmount || "0");

      const updateData: UpdatePurchaseInvoiceRequest = {
        invoiceDate: formData.invoiceDate,
        dueDate: formData.dueDate,
        supplierId: formData.supplierId,
        discountAmount: headerDiscount > 0 ? headerDiscount.toString() : undefined,
        taxRate: formData.taxPercentage,
        notes: formData.notes || undefined,
        // Non-Goods Costs (Biaya Tambahan)
        shippingCost: parseFloat(formData.shippingCost || "0") > 0 ? formData.shippingCost : undefined,
        handlingCost: parseFloat(formData.handlingCost || "0") > 0 ? formData.handlingCost : undefined,
        otherCost: parseFloat(formData.otherCost || "0") > 0 ? formData.otherCost : undefined,
        otherCostDescription: formData.otherCostDescription || undefined,
        // Include items for update
        items: items.map(item => ({
          id: item.originalId, // Reference to original item for audit
          purchaseOrderItemId: item.purchaseOrderItemId, // Preserve PO item linkage
          goodsReceiptItemId: item.goodsReceiptItemId, // Preserve GRN item linkage for invoiced qty tracking
          productId: item.productId,
          unitId: item.unitId,
          quantity: item.quantity,
          unitPrice: item.unitPrice,
          discountAmount: parseFloat(item.discountAmount || "0") > 0 ? item.discountAmount : undefined,
          discountPct: parseFloat(item.discountPercent || "0") > 0 ? item.discountPercent : undefined,
          taxAmount: parseFloat(item.taxAmount || "0") > 0 ? item.taxAmount : undefined,
          notes: undefined, // Add notes if available
        })),
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

            {/* Purchase Order Reference - Read-only */}
            <div className="space-y-2">
              <Label className="text-sm font-medium">Referensi PO</Label>
              <div className="flex items-center gap-2 rounded-md border bg-muted px-3 py-2">
                <Receipt className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">
                  {invoice.poNumber || "-"}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                Referensi PO tidak dapat diubah setelah faktur dibuat
              </p>
            </div>

            {/* Goods Receipt Reference - Read-only */}
            <div className="space-y-2">
              <Label className="text-sm font-medium">Referensi Penerimaan Barang (GRN)</Label>
              <div className="flex items-center gap-2 rounded-md border bg-muted px-3 py-2">
                <Package className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">
                  {invoice.grNumber || "-"}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                Referensi GRN tidak dapat diubah setelah faktur dibuat
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

            {/* Header Discount */}
            <div className="space-y-2">
              <Label htmlFor="headerDiscount" className="text-sm font-medium">
                Diskon Faktur (Rp)
              </Label>
              <div className="flex items-center gap-2">
                <DollarSign className="h-4 w-4 text-muted-foreground" />
                <Input
                  id="headerDiscount"
                  type="number"
                  min="0"
                  step="0.01"
                  value={formData.headerDiscountAmount}
                  onChange={(e) => handleChange("headerDiscountAmount", e.target.value)}
                  placeholder="0"
                  className="font-mono"
                />
              </div>
              <p className="text-xs text-muted-foreground">
                Diskon tambahan pada level faktur (selain diskon per item)
              </p>
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

      {/* Additional Costs (Non-Goods Costs) */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ReceiptText className="h-5 w-5" />
            Biaya Tambahan
          </CardTitle>
          <p className="text-sm text-muted-foreground">
            Biaya non-barang seperti ongkir, handling, dan biaya lainnya
          </p>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-3">
            {/* Shipping Cost */}
            <div className="space-y-2">
              <Label htmlFor="shippingCost" className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <Truck className="h-4 w-4" />
                  Biaya Pengiriman
                </div>
              </Label>
              <Input
                id="shippingCost"
                type="number"
                min="0"
                step="1000"
                value={formData.shippingCost}
                onChange={(e) => handleChange("shippingCost", e.target.value)}
                placeholder="0"
                className="text-right"
              />
              <p className="text-xs text-muted-foreground">Ongkos kirim dari supplier</p>
            </div>

            {/* Handling Cost */}
            <div className="space-y-2">
              <Label htmlFor="handlingCost" className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <HandCoins className="h-4 w-4" />
                  Biaya Handling
                </div>
              </Label>
              <Input
                id="handlingCost"
                type="number"
                min="0"
                step="1000"
                value={formData.handlingCost}
                onChange={(e) => handleChange("handlingCost", e.target.value)}
                placeholder="0"
                className="text-right"
              />
              <p className="text-xs text-muted-foreground">Biaya bongkar muat</p>
            </div>

            {/* Other Cost */}
            <div className="space-y-2">
              <Label htmlFor="otherCost" className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <DollarSign className="h-4 w-4" />
                  Biaya Lain-lain
                </div>
              </Label>
              <Input
                id="otherCost"
                type="number"
                min="0"
                step="1000"
                value={formData.otherCost}
                onChange={(e) => handleChange("otherCost", e.target.value)}
                placeholder="0"
                className="text-right"
              />
              <p className="text-xs text-muted-foreground">Biaya tambahan lainnya</p>
            </div>

            {/* Other Cost Description - Only show if other cost > 0 */}
            {parseFloat(formData.otherCost || "0") > 0 && (
              <div className="space-y-2 sm:col-span-3">
                <Label htmlFor="otherCostDescription" className="text-sm font-medium">
                  Keterangan Biaya Lain-lain
                </Label>
                <Input
                  id="otherCostDescription"
                  value={formData.otherCostDescription}
                  onChange={(e) => handleChange("otherCostDescription", e.target.value)}
                  placeholder="Jelaskan biaya lain-lain..."
                />
              </div>
            )}
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
              <p className="text-sm">Klik &quot;Tambah Item&quot; untuk menambahkan produk</p>
            </div>
          ) : (
            <div className="rounded-md border overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[250px]">Produk</TableHead>
                    <TableHead className="w-[80px]">Kuantitas</TableHead>
                    <TableHead className="w-[120px] text-right">Harga Satuan</TableHead>
                    <TableHead className="w-[120px] text-right">Diskon</TableHead>
                    <TableHead className="w-[120px] text-right">PPN</TableHead>
                    <TableHead className="w-[120px] text-right">Total</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item, index) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div className="space-y-1">
                          <Input
                            value={item.productName}
                            onChange={(e) =>
                              handleItemChange(item.id, "productName", e.target.value)
                            }
                            placeholder="Nama produk"
                            className="h-8"
                          />
                          <p className="text-xs text-muted-foreground">{item.productCode}</p>
                        </div>
                        {errors[`item-${index}-product`] && (
                          <p className="text-xs text-destructive mt-1">
                            {errors[`item-${index}-product`]}
                          </p>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Input
                            type="number"
                            value={item.quantity}
                            onChange={(e) =>
                              handleItemChange(item.id, "quantity", e.target.value)
                            }
                            className="h-8 w-16"
                          />
                          <span className="text-sm text-muted-foreground whitespace-nowrap">
                            {item.unitName}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <Input
                          type="number"
                          value={item.unitPrice}
                          onChange={(e) =>
                            handleItemChange(item.id, "unitPrice", e.target.value)
                          }
                          className="h-8 text-right"
                        />
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="space-y-1">
                          <div className="font-mono text-sm">
                            {formatCurrency(item.discountAmount)}
                          </div>
                          <div className="flex items-center justify-end gap-1">
                            <Input
                              type="number"
                              value={item.discountPercent}
                              onChange={(e) =>
                                handleItemChange(item.id, "discountPercent", e.target.value)
                              }
                              className="h-6 w-14 text-right text-xs"
                            />
                            <span className="text-xs text-muted-foreground">%</span>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="space-y-1">
                          <div className="font-mono text-sm">
                            {formatCurrency(item.taxAmount)}
                          </div>
                          <div className="flex items-center justify-end gap-1">
                            <Input
                              type="number"
                              value={item.taxPercent}
                              onChange={(e) =>
                                handleItemChange(item.id, "taxPercent", e.target.value)
                              }
                              className="h-6 w-14 text-right text-xs"
                            />
                            <span className="text-xs text-muted-foreground">%</span>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm font-medium text-primary">
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
          <CardContent className="space-y-4">
            {/* Use original invoice values when items haven't been modified */}
            {/* Subtotal */}
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-muted-foreground">Subtotal</p>
              <p className="font-mono font-semibold">
                {formatCurrency(itemsModified ? totals.subtotal : invoice.subtotalAmount)}
              </p>
            </div>

            <Separator />

            {/* Diskon Item (dari masing-masing item) */}
            {(itemsModified ? totals.lineDiscount > 0 : false) && (
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-muted-foreground">Diskon Item</p>
                <p className="font-mono font-semibold text-green-600">
                  - {formatCurrency(totals.lineDiscount)}
                </p>
              </div>
            )}

            {/* Diskon Faktur (header-level discount) */}
            {(itemsModified ? totals.headerDiscount > 0 : parseFloat(invoice.discountAmount || "0") > 0) && (
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-muted-foreground">Diskon Faktur</p>
                <p className="font-mono font-semibold text-green-600">
                  - {formatCurrency(itemsModified ? totals.headerDiscount : invoice.discountAmount)}
                </p>
              </div>
            )}

            {/* Separator setelah diskon jika ada */}
            {(itemsModified ? totals.totalDiscount > 0 : parseFloat(invoice.discountAmount || "0") > 0) && (
              <Separator />
            )}

            {/* PPN */}
            {(itemsModified ? totals.totalTax > 0 : parseFloat(invoice.taxAmount || "0") > 0) && (
              <>
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium text-muted-foreground">
                    PPN ({formData.taxPercentage || 0}%)
                  </p>
                  <p className="font-mono font-semibold">
                    {formatCurrency(itemsModified ? totals.totalTax : invoice.taxAmount)}
                  </p>
                </div>
                <Separator />
              </>
            )}

            {/* Non-Goods Costs Section */}
            {(itemsModified ? totals.totalNonGoodsCost > 0 : parseFloat(invoice.totalNonGoodsCost || "0") > 0) && (
              <>
                <div className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Biaya Tambahan
                </div>
                {(itemsModified ? totals.shippingCost > 0 : parseFloat(invoice.shippingCost || "0") > 0) && (
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-muted-foreground flex items-center gap-1">
                      <Truck className="h-3 w-3" />
                      Biaya Pengiriman
                    </p>
                    <p className="font-mono font-semibold">
                      {formatCurrency(itemsModified ? totals.shippingCost : invoice.shippingCost)}
                    </p>
                  </div>
                )}
                {(itemsModified ? totals.handlingCost > 0 : parseFloat(invoice.handlingCost || "0") > 0) && (
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-muted-foreground flex items-center gap-1">
                      <HandCoins className="h-3 w-3" />
                      Biaya Handling
                    </p>
                    <p className="font-mono font-semibold">
                      {formatCurrency(itemsModified ? totals.handlingCost : invoice.handlingCost)}
                    </p>
                  </div>
                )}
                {(itemsModified ? totals.otherCost > 0 : parseFloat(invoice.otherCost || "0") > 0) && (
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-muted-foreground flex items-center gap-1">
                      <DollarSign className="h-3 w-3" />
                      Biaya Lain-lain
                    </p>
                    <p className="font-mono font-semibold">
                      {formatCurrency(itemsModified ? totals.otherCost : invoice.otherCost)}
                    </p>
                  </div>
                )}
                <Separator />
              </>
            )}

            {/* Total */}
            <div className="flex items-center justify-between rounded-lg bg-muted/50 p-4">
              <p className="text-base font-bold">Total</p>
              <p className="text-2xl font-bold text-blue-600">
                {formatCurrency(itemsModified ? totals.total : invoice.totalAmount)}
              </p>
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
