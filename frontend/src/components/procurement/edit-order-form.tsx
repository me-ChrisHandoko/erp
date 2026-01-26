/**
 * Edit Purchase Order Form Component
 *
 * Professional form for editing DRAFT purchase orders with:
 * - Pre-filled order data
 * - Supplier and warehouse selection
 * - Dynamic line items
 * - Real-time calculation
 * - Validation
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import {
  Save,
  AlertCircle,
  Building,
  Warehouse,
  Calendar,
  ShoppingCart,
  Wallet,
  Package,
  Receipt,
  Tag,
  Percent,
  DollarSign,
} from "lucide-react";
import { selectActiveCompany } from "@/store/slices/companySlice";
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
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";
import { useUpdatePurchaseOrderMutation } from "@/store/services/purchaseOrderApi";
import { useListSuppliersQuery } from "@/store/services/supplierApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import {
  PurchaseOrderItemsManager,
  type PurchaseOrderItem,
} from "@/components/procurement/purchase-order-items-manager";
import type {
  PurchaseOrderResponse,
  UpdatePurchaseOrderRequest,
  CreatePurchaseOrderItemRequest,
} from "@/types/purchase-order.types";
import { formatCurrency } from "@/types/purchase-order.types";

interface EditOrderFormProps {
  order: PurchaseOrderResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditOrderForm({ order, onSuccess, onCancel }: EditOrderFormProps) {
  const [updateOrder, { isLoading }] = useUpdatePurchaseOrderMutation();
  const activeCompany = useSelector(selectActiveCompany);

  // Track if form has been initialized from order data
  const [isInitialized, setIsInitialized] = useState(false);

  // Form state - initialize directly from order prop
  const [supplierId, setSupplierId] = useState(order.supplierId || "");
  const [warehouseId, setWarehouseId] = useState(order.warehouseId || "");
  const [poDate, setPoDate] = useState(order.poDate ? order.poDate.split("T")[0] : "");
  const [expectedDeliveryAt, setExpectedDeliveryAt] = useState(
    order.expectedDeliveryAt ? order.expectedDeliveryAt.split("T")[0] : ""
  );
  const [notes, setNotes] = useState(order.notes || "");
  const [discountAmount, setDiscountAmount] = useState(order.discountAmount || "0");
  const [lineItems, setLineItems] = useState<PurchaseOrderItem[]>(() => {
    // Initialize line items from order
    if (order.items && order.items.length > 0) {
      return order.items.map((item, index) => ({
        id: `item-${index}`,
        productId: item.productId || "",
        productName: item.product?.name || "",
        productCode: item.product?.code || "",
        productUnit: item.product?.baseUnit || "",
        quantity: item.quantity || "0",
        unitPrice: item.unitPrice || "0",
        discountPct: item.discountPct || "0",
        subtotal: parseFloat(item.subtotal || "0"),
        notes: item.notes || "",
        baseCost: parseFloat((item.product as any)?.baseCost || "0"),
        supplierPrice: parseFloat(item.unitPrice || "0"),
      }));
    }
    return [];
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Discount mode: 'nominal' or 'percentage'
  const [discountMode, setDiscountMode] = useState<'nominal' | 'percentage'>('nominal');
  const [discountPercentage, setDiscountPercentage] = useState<string>("0");

  // Fetch suppliers and warehouses
  const { data: suppliersData, isLoading: isLoadingSuppliers } = useListSuppliersQuery({ isActive: true, pageSize: 100 });
  const { data: warehousesData, isLoading: isLoadingWarehouses } = useListWarehousesQuery({ isActive: true, pageSize: 100 });

  // Initialize form only once when order changes (handles case where order prop updates)
  useEffect(() => {
    if (order && !isInitialized) {
      setSupplierId(order.supplierId || "");
      setWarehouseId(order.warehouseId || "");
      setPoDate(order.poDate ? order.poDate.split("T")[0] : "");
      setExpectedDeliveryAt(
        order.expectedDeliveryAt ? order.expectedDeliveryAt.split("T")[0] : ""
      );
      setNotes(order.notes || "");
      setDiscountAmount(order.discountAmount || "0");

      // Convert order items to PurchaseOrderItem format
      if (order.items && order.items.length > 0) {
        const items: PurchaseOrderItem[] = order.items.map((item, index) => ({
          id: `item-${index}`,
          productId: item.productId || "",
          productName: item.product?.name || "",
          productCode: item.product?.code || "",
          productUnit: item.product?.baseUnit || "",
          quantity: item.quantity || "0",
          unitPrice: item.unitPrice || "0",
          discountPct: item.discountPct || "0",
          subtotal: parseFloat(item.subtotal || "0"),
          notes: item.notes || "",
          baseCost: parseFloat((item.product as any)?.baseCost || "0"),
          supplierPrice: parseFloat(item.unitPrice || "0"),
        }));
        setLineItems(items);
      }
      setIsInitialized(true);
    }
  }, [order, isInitialized]);

  // Clear line items when supplier changes (only if supplier actually changes)
  const handleSupplierChange = (newSupplierId: string) => {
    // Only clear items if supplier is actually changing to a different value
    if (newSupplierId !== supplierId) {
      setSupplierId(newSupplierId);
      if (lineItems.length > 0) {
        setLineItems([]);
      }
    }
  };

  // Handle discount preset selection
  const handleDiscountPreset = (percentage: number) => {
    setDiscountMode('percentage');
    setDiscountPercentage(percentage.toString());
    const itemsSubtotal = lineItems.reduce((sum, item) => sum + item.subtotal, 0);
    const discountValue = (itemsSubtotal * percentage) / 100;
    setDiscountAmount(discountValue.toFixed(2));
  };

  // Toggle discount mode
  const handleToggleDiscountMode = () => {
    if (discountMode === 'nominal') {
      // Switch to percentage: calculate percentage from current nominal
      const itemsSubtotal = lineItems.reduce((sum, item) => sum + item.subtotal, 0);
      const currentDiscount = parseFloat(discountAmount) || 0;
      const percentage = itemsSubtotal > 0 ? (currentDiscount / itemsSubtotal) * 100 : 0;
      setDiscountPercentage(percentage.toFixed(2));
      setDiscountMode('percentage');
    } else {
      // Switch to nominal: keep current discount value
      setDiscountMode('nominal');
    }
  };

  // Handle percentage input change
  const handleDiscountPercentageChange = (value: string) => {
    setDiscountPercentage(value);
    const itemsSubtotal = lineItems.reduce((sum, item) => sum + item.subtotal, 0);
    const percentage = parseFloat(value) || 0;
    const discountValue = (itemsSubtotal * percentage) / 100;
    setDiscountAmount(discountValue.toFixed(2));
  };

  // Calculate totals with discount and tax
  const subtotal = lineItems.reduce((sum, item) => sum + item.subtotal, 0);
  const discount = parseFloat(discountAmount) || 0;
  const afterDiscount = subtotal - discount;

  // Tax calculation based on company PKP status (PPN calculated from DPP = subtotal - discount)
  let taxRate = 0;
  if (activeCompany?.isPKP) {
    taxRate = activeCompany.ppnRate ? activeCompany.ppnRate / 100 : 0.11;
  }
  const taxAmount = afterDiscount * taxRate;
  const totalAmount = afterDiscount + taxAmount;

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!supplierId) {
      newErrors.supplierId = "Supplier wajib dipilih";
    }

    if (!warehouseId) {
      newErrors.warehouseId = "Gudang tujuan wajib dipilih";
    }

    if (!poDate) {
      newErrors.poDate = "Tanggal PO wajib diisi";
    }

    if (lineItems.length === 0) {
      newErrors.items = "Minimal 1 item produk diperlukan";
    }

    // Validate each line item
    lineItems.forEach((item, index) => {
      if (!item.productId) {
        newErrors[`item_${index}_product`] = "Produk wajib dipilih";
      }
      if (parseFloat(item.quantity) <= 0) {
        newErrors[`item_${index}_qty`] = "Kuantitas harus lebih dari 0";
      }
      if (parseFloat(item.unitPrice) <= 0) {
        newErrors[`item_${index}_price`] = "Harga harus lebih dari 0";
      }
    });

    setErrors(newErrors);
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
      const items: CreatePurchaseOrderItemRequest[] = lineItems.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        discountPct: item.discountPct || undefined,
        notes: item.notes || undefined,
      }));

      const data: UpdatePurchaseOrderRequest = {
        supplierId,
        warehouseId,
        poDate,
        expectedDeliveryAt: expectedDeliveryAt || undefined,
        discountAmount: discountAmount || undefined,
        notes: notes || undefined,
        items,
      };

      await updateOrder({ id: order.id, data }).unwrap();

      toast.success("PO Berhasil Diperbarui", {
        description: `${order.poNumber} telah diperbarui`,
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      toast.error("Gagal Memperbarui PO", {
        description:
          error?.data?.error?.message ||
          error?.data?.message ||
          error?.message ||
          "Terjadi kesalahan pada server",
      });
    }
  };

  const suppliers = suppliersData?.data || [];
  const warehouses = warehousesData?.data || [];

  // Ensure current supplier is in the list (for display purposes)
  const supplierOptions = [...suppliers];
  if (supplierId && order.supplier && !suppliers.find(s => s.id === supplierId)) {
    supplierOptions.unshift({
      id: order.supplier.id,
      code: order.supplier.code,
      name: order.supplier.name,
    } as typeof suppliers[0]);
  }

  // Ensure current warehouse is in the list (for display purposes)
  const warehouseOptions = [...warehouses];
  if (warehouseId && order.warehouse && !warehouses.find(w => w.id === warehouseId)) {
    warehouseOptions.unshift({
      id: order.warehouse.id,
      code: order.warehouse.code,
      name: order.warehouse.name,
    } as typeof warehouses[0]);
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Header Information */}
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <ShoppingCart className="h-5 w-5" />
            Informasi Pesanan Pembelian
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2">
            {/* Supplier */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Building className="h-4 w-4" />
                Supplier <span className="text-destructive">*</span>
              </Label>
              <Select value={supplierId} onValueChange={handleSupplierChange} disabled={isLoadingSuppliers}>
                <SelectTrigger
                  className={`w-full ${errors.supplierId ? "border-destructive" : ""}`}
                >
                  <SelectValue placeholder={isLoadingSuppliers ? "Memuat..." : "Pilih supplier..."} />
                </SelectTrigger>
                <SelectContent>
                  {supplierOptions.map((supplier) => (
                    <SelectItem key={supplier.id} value={supplier.id}>
                      {supplier.code} - {supplier.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.supplierId && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.supplierId}
                </p>
              )}
            </div>

            {/* Warehouse */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Warehouse className="h-4 w-4" />
                Gudang Tujuan <span className="text-destructive">*</span>
              </Label>
              <Select value={warehouseId} onValueChange={setWarehouseId} disabled={isLoadingWarehouses}>
                <SelectTrigger
                  className={`w-full ${errors.warehouseId ? "border-destructive" : ""}`}
                >
                  <SelectValue placeholder={isLoadingWarehouses ? "Memuat..." : "Pilih gudang..."} />
                </SelectTrigger>
                <SelectContent>
                  {warehouseOptions.map((warehouse) => (
                    <SelectItem key={warehouse.id} value={warehouse.id}>
                      {warehouse.code} - {warehouse.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.warehouseId && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.warehouseId}
                </p>
              )}
            </div>

            {/* PO Date */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                Tanggal PO <span className="text-destructive">*</span>
              </Label>
              <Input
                type="date"
                value={poDate}
                onChange={(e) => setPoDate(e.target.value)}
                className={errors.poDate ? "border-destructive" : ""}
              />
              {errors.poDate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.poDate}
                </p>
              )}
            </div>

            {/* Expected Delivery */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                Estimasi Pengiriman
              </Label>
              <Input
                type="date"
                value={expectedDeliveryAt}
                onChange={(e) => setExpectedDeliveryAt(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Opsional</p>
            </div>

            {/* Notes */}
            <div className="space-y-2 sm:col-span-2">
              <Label>Catatan</Label>
              <Textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Catatan tambahan untuk PO ini..."
                rows={2}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Line Items */}
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Item Produk</CardTitle>
        </CardHeader>
        <CardContent>
          {errors.items && (
            <p className="flex items-center gap-1 text-sm text-destructive mb-4">
              <AlertCircle className="h-3 w-3" />
              {errors.items}
            </p>
          )}

          <PurchaseOrderItemsManager
            items={lineItems}
            onChange={setLineItems}
            supplierId={supplierId}
            disabled={isLoading}
          />
        </CardContent>
      </Card>

      {/* Financial Summary Card */}
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Wallet className="h-5 w-5" />
            Ringkasan Keuangan
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {lineItems.length === 0 ? (
            /* Empty State */
            <div className="text-center py-8 text-muted-foreground">
              <Wallet className="h-12 w-12 mx-auto mb-3 opacity-20" />
              <p className="text-sm font-medium">Belum ada item ditambahkan</p>
              <p className="text-xs mt-1">
                Tambahkan produk terlebih dahulu untuk melihat ringkasan
              </p>
            </div>
          ) : (
            <>
              {/* Subtotal with Item Count */}
              <div className="flex justify-between items-center text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Package className="h-4 w-4" />
                  <span className="text-sm">
                    Subtotal ({lineItems.length} item)
                  </span>
                </div>
                <span className="font-medium text-foreground">
                  {formatCurrency(subtotal)}
                </span>
              </div>

              {/* Discount */}
              <div className="space-y-2">
                <div className="flex justify-between items-center text-destructive">
                  <div className="flex items-center gap-2">
                    <Tag className="h-4 w-4" />
                    <span className="text-sm">Diskon</span>
                  </div>
                  <div className="flex flex-col items-end gap-1">
                    <div className="flex items-center gap-2">
                      {/* Toggle Mode Button */}
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={handleToggleDiscountMode}
                        className="h-7 px-2"
                      >
                        {discountMode === 'nominal' ? (
                          <Percent className="h-3 w-3" />
                        ) : (
                          <DollarSign className="h-3 w-3" />
                        )}
                      </Button>
                      {/* Input Field */}
                      <div className="relative w-32">
                        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
                          {discountMode === 'nominal' ? 'Rp' : '%'}
                        </span>
                        <Input
                          type="number"
                          step="0.01"
                          min="0"
                          max={discountMode === 'percentage' ? 100 : undefined}
                          value={discountMode === 'nominal' ? discountAmount : discountPercentage}
                          onChange={(e) => {
                            if (discountMode === 'nominal') {
                              setDiscountAmount(e.target.value);
                            } else {
                              handleDiscountPercentageChange(e.target.value);
                            }
                          }}
                          className="pl-8 text-right h-9 transition-all duration-200"
                          placeholder="0"
                        />
                      </div>
                    </div>
                    {/* Discount Breakdown */}
                    {parseFloat(discountAmount) > 0 && (
                      <div className="text-xs text-muted-foreground transition-all duration-300">
                        {discountMode === 'nominal' ? (
                          // Show percentage when in nominal mode
                          (() => {
                            const percentage = subtotal > 0 ? (parseFloat(discountAmount) / subtotal) * 100 : 0;
                            return `(${percentage.toFixed(2)}% dari subtotal)`;
                          })()
                        ) : (
                          // Show nominal when in percentage mode
                          `= ${formatCurrency(discountAmount)}`
                        )}
                      </div>
                    )}
                  </div>
                </div>
                {/* Preset Buttons - Show only in percentage mode */}
                {discountMode === 'percentage' && (
                  <div className="flex gap-1 justify-end">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleDiscountPreset(5)}
                      className="h-6 px-2 text-xs"
                    >
                      5%
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleDiscountPreset(10)}
                      className="h-6 px-2 text-xs"
                    >
                      10%
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleDiscountPreset(15)}
                      className="h-6 px-2 text-xs"
                    >
                      15%
                    </Button>
                  </div>
                )}
              </div>

              {/* Tax (PPN) with breakdown */}
              <div className="flex justify-between items-center text-amber-600 dark:text-amber-500">
                <div className="flex items-center gap-2">
                  <Receipt className="h-4 w-4" />
                  <span className="text-sm">
                    PPN ({activeCompany?.isPKP ? (activeCompany?.ppnRate ?? 11) : 0}%)
                  </span>
                </div>
                <div className="text-right">
                  <div className="font-medium text-foreground transition-all duration-300">
                    {formatCurrency(taxAmount)}
                  </div>
                  <div className="text-xs text-muted-foreground transition-all duration-300">
                    dari {formatCurrency(afterDiscount)}
                  </div>
                  {!activeCompany?.isPKP && (
                    <p className="text-xs text-muted-foreground">
                      Perusahaan non-PKP
                    </p>
                  )}
                </div>
              </div>

              <Separator />

              {/* Total with Highlight */}
              <div className="bg-primary/10 -mx-6 -mb-6 px-6 py-4 rounded-b-lg">
                <div className="flex justify-between items-center">
                  <span className="text-base font-bold">Total Pesanan</span>
                  <span className="text-2xl font-bold text-primary">
                    {formatCurrency(totalAmount)}
                  </span>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

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
          disabled={isLoading || lineItems.length === 0}
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
