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
import {
  Plus,
  Trash2,
  Save,
  AlertCircle,
  Building,
  Warehouse,
  Calendar,
  ShoppingCart,
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
import { toast } from "sonner";
import { useUpdatePurchaseOrderMutation } from "@/store/services/purchaseOrderApi";
import { useListSuppliersQuery } from "@/store/services/supplierApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { useListProductsQuery } from "@/store/services/productApi";
import type {
  PurchaseOrderResponse,
  UpdatePurchaseOrderRequest,
  CreatePurchaseOrderItemRequest,
} from "@/types/purchase-order.types";
import { formatCurrency } from "@/types/purchase-order.types";

interface LineItem {
  id: string;
  productId: string;
  productName: string;
  productCode: string;
  productUnit: string;
  quantity: string;
  unitPrice: string;
  discountPct: string;
  subtotal: number;
  notes: string;
  baseCost: number;
  supplierPrice: number;
}

interface EditOrderFormProps {
  order: PurchaseOrderResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditOrderForm({ order, onSuccess, onCancel }: EditOrderFormProps) {
  const [updateOrder, { isLoading }] = useUpdatePurchaseOrderMutation();

  // Form state
  const [supplierId, setSupplierId] = useState("");
  const [warehouseId, setWarehouseId] = useState("");
  const [poDate, setPoDate] = useState("");
  const [expectedDeliveryAt, setExpectedDeliveryAt] = useState("");
  const [notes, setNotes] = useState("");
  const [lineItems, setLineItems] = useState<LineItem[]>([]);

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Fetch suppliers and warehouses
  const { data: suppliersData } = useListSuppliersQuery({ isActive: true, pageSize: 100 });
  const { data: warehousesData } = useListWarehousesQuery({ isActive: true, pageSize: 100 });

  // Fetch products filtered by selected supplier
  const { data: productsData, isFetching: isLoadingProducts } = useListProductsQuery(
    { isActive: true, supplierId: supplierId, pageSize: 100 },
    { skip: !supplierId }
  );

  // Pre-fill form with existing order data
  useEffect(() => {
    if (order) {
      setSupplierId(order.supplierId || "");
      setWarehouseId(order.warehouseId || "");
      setPoDate(order.poDate ? order.poDate.split("T")[0] : "");
      setExpectedDeliveryAt(
        order.expectedDeliveryAt ? order.expectedDeliveryAt.split("T")[0] : ""
      );
      setNotes(order.notes || "");

      // Convert order items to line items
      if (order.items && order.items.length > 0) {
        const items: LineItem[] = order.items.map((item, index) => ({
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
          baseCost: 0, // Not available in response, will be filled from products data
          supplierPrice: 0, // Will be filled from products data
        }));
        setLineItems(items);
      }
    }
  }, [order]);

  // Update supplier prices when products data loads
  useEffect(() => {
    if (productsData?.data && lineItems.length > 0 && supplierId) {
      setLineItems((prev) =>
        prev.map((item) => {
          const product = productsData.data.find((p) => p.id === item.productId);
          if (product) {
            const supplierInfo = product.suppliers?.find(
              (s) => s.supplierId === supplierId
            );
            const supplierPriceValue = supplierInfo?.supplierPrice
              ? parseFloat(supplierInfo.supplierPrice)
              : 0;
            return {
              ...item,
              supplierPrice: supplierPriceValue,
            };
          }
          return item;
        })
      );
    }
  }, [productsData, supplierId]);

  // Clear line items when supplier changes
  const handleSupplierChange = (newSupplierId: string) => {
    setSupplierId(newSupplierId);
    if (lineItems.length > 0) {
      setLineItems([]);
    }
  };

  // Calculate totals
  const subtotal = lineItems.reduce((sum, item) => sum + item.subtotal, 0);
  const totalAmount = subtotal;

  const addLineItem = () => {
    const newItem: LineItem = {
      id: `item-${Date.now()}`,
      productId: "",
      productName: "",
      productCode: "",
      productUnit: "",
      quantity: "1",
      unitPrice: "0",
      discountPct: "0",
      subtotal: 0,
      notes: "",
      baseCost: 0,
      supplierPrice: 0,
    };
    setLineItems([...lineItems, newItem]);
  };

  const removeLineItem = (id: string) => {
    setLineItems(lineItems.filter((item) => item.id !== id));
  };

  const updateLineItem = (id: string, field: keyof LineItem, value: string) => {
    setLineItems((prev) =>
      prev.map((item) => {
        if (item.id !== id) return item;

        const updated = { ...item, [field]: value };

        // If product changed, update product info
        if (field === "productId" && productsData?.data) {
          const product = productsData.data.find((p) => p.id === value);
          if (product) {
            updated.productName = product.name;
            updated.productCode = product.code;
            updated.productUnit = product.baseUnit;
            updated.baseCost = parseFloat(product.baseCost) || 0;

            const supplierInfo = product.suppliers?.find(
              (s) => s.supplierId === supplierId
            );
            const supplierPriceValue = supplierInfo?.supplierPrice
              ? parseFloat(supplierInfo.supplierPrice)
              : 0;
            updated.supplierPrice = supplierPriceValue;
            updated.unitPrice = supplierInfo?.supplierPrice || product.baseCost;
          }
        }

        // Recalculate subtotal
        const qty = parseFloat(updated.quantity) || 0;
        const price = parseFloat(updated.unitPrice) || 0;
        const discPct = parseFloat(updated.discountPct) || 0;
        const discAmount = (qty * price * discPct) / 100;
        updated.subtotal = qty * price - discAmount;

        return updated;
      })
    );
  };

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
  const products = productsData?.data || [];

  // Helper function to calculate and format price difference
  const getPriceDifference = (currentPrice: number, baseCost: number) => {
    if (baseCost === 0) return null;
    const diff = currentPrice - baseCost;
    const pct = (diff / baseCost) * 100;
    return {
      diff,
      pct,
      isHigher: diff > 0,
      isLower: diff < 0,
      isSame: diff === 0,
    };
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Header Information */}
      <Card className="border-2">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <ShoppingCart className="h-5 w-5" />
            Informasi Purchase Order
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
              <Select value={supplierId} onValueChange={handleSupplierChange}>
                <SelectTrigger
                  className={`w-full ${errors.supplierId ? "border-destructive" : ""}`}
                >
                  <SelectValue placeholder="Pilih supplier..." />
                </SelectTrigger>
                <SelectContent>
                  {suppliers.map((supplier) => (
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
              <Select value={warehouseId} onValueChange={setWarehouseId}>
                <SelectTrigger
                  className={`w-full ${errors.warehouseId ? "border-destructive" : ""}`}
                >
                  <SelectValue placeholder="Pilih gudang..." />
                </SelectTrigger>
                <SelectContent>
                  {warehouses.map((warehouse) => (
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
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Item Produk</CardTitle>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addLineItem}
              disabled={!supplierId || isLoadingProducts}
            >
              <Plus className="mr-2 h-4 w-4" />
              Tambah Item
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {errors.items && (
            <p className="flex items-center gap-1 text-sm text-destructive mb-4">
              <AlertCircle className="h-3 w-3" />
              {errors.items}
            </p>
          )}

          {lineItems.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <ShoppingCart className="mx-auto h-12 w-12 mb-4 opacity-50" />
              {!supplierId ? (
                <>
                  <p>Pilih supplier terlebih dahulu</p>
                  <p className="text-sm">
                    Produk akan ditampilkan berdasarkan supplier yang dipilih
                  </p>
                </>
              ) : isLoadingProducts ? (
                <>
                  <p>Memuat produk...</p>
                  <p className="text-sm">Sedang mengambil daftar produk dari supplier</p>
                </>
              ) : products.length === 0 ? (
                <>
                  <p>Tidak ada produk untuk supplier ini</p>
                  <p className="text-sm">
                    Hubungkan produk dengan supplier terlebih dahulu di Master Produk
                  </p>
                </>
              ) : (
                <>
                  <p>Belum ada item produk</p>
                  <p className="text-sm">Klik "Tambah Item" untuk menambahkan produk</p>
                </>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              {lineItems.map((item, index) => (
                <div
                  key={item.id}
                  className="grid gap-4 p-4 border rounded-lg bg-muted/30"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-muted-foreground">
                      Item #{index + 1}
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeLineItem(item.id)}
                      className="text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>

                  <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4">
                    {/* Product */}
                    <div className="space-y-2 sm:col-span-2">
                      <Label>
                        Produk <span className="text-destructive">*</span>
                      </Label>
                      <Select
                        value={item.productId}
                        onValueChange={(value) =>
                          updateLineItem(item.id, "productId", value)
                        }
                        disabled={!supplierId || isLoadingProducts}
                      >
                        <SelectTrigger
                          className={`w-full ${
                            errors[`item_${index}_product`] ? "border-destructive" : ""
                          }`}
                        >
                          <SelectValue
                            placeholder={
                              !supplierId
                                ? "Pilih supplier terlebih dahulu"
                                : isLoadingProducts
                                  ? "Memuat produk..."
                                  : products.length === 0
                                    ? "Tidak ada produk untuk supplier ini"
                                    : "Pilih produk..."
                            }
                          />
                        </SelectTrigger>
                        <SelectContent>
                          {products.map((product) => (
                            <SelectItem key={product.id} value={product.id}>
                              {product.code} - {product.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {item.productUnit && (
                        <p className="text-xs text-muted-foreground">
                          Satuan: {item.productUnit}
                        </p>
                      )}
                    </div>

                    {/* Quantity */}
                    <div className="space-y-2">
                      <Label>
                        Kuantitas <span className="text-destructive">*</span>
                      </Label>
                      <Input
                        type="number"
                        step="0.001"
                        min="0"
                        value={item.quantity}
                        onChange={(e) =>
                          updateLineItem(item.id, "quantity", e.target.value)
                        }
                        className={
                          errors[`item_${index}_qty`] ? "border-destructive" : ""
                        }
                      />
                    </div>

                    {/* Unit Price */}
                    <div className="space-y-2">
                      <Label>
                        Harga Satuan <span className="text-destructive">*</span>
                      </Label>
                      <div className="relative">
                        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                          Rp
                        </span>
                        <Input
                          type="number"
                          step="0.01"
                          min="0"
                          value={item.unitPrice}
                          onChange={(e) =>
                            updateLineItem(item.id, "unitPrice", e.target.value)
                          }
                          className={`pl-10 ${
                            errors[`item_${index}_price`] ? "border-destructive" : ""
                          }`}
                        />
                      </div>
                      {/* Price Comparison Info */}
                      {item.productId && item.baseCost > 0 && (
                        <div className="text-xs space-y-0.5 pt-1">
                          {item.supplierPrice > 0 && (
                            <p className="text-muted-foreground">
                              Harga Supplier: {formatCurrency(item.supplierPrice)}
                            </p>
                          )}
                          <p className="text-muted-foreground">
                            HPP: {formatCurrency(item.baseCost)}
                          </p>
                          {(() => {
                            const currentPrice = parseFloat(item.unitPrice) || 0;
                            const priceInfo = getPriceDifference(
                              currentPrice,
                              item.baseCost
                            );
                            if (!priceInfo) return null;

                            if (priceInfo.isSame) {
                              return (
                                <p className="text-blue-600 dark:text-blue-400 font-medium">
                                  ≈ Sama dengan HPP
                                </p>
                              );
                            } else if (priceInfo.isLower) {
                              return (
                                <p className="text-green-600 dark:text-green-400 font-medium">
                                  ✓ Hemat {Math.abs(priceInfo.pct).toFixed(1)}% dari HPP
                                </p>
                              );
                            } else {
                              return (
                                <p className="text-amber-600 dark:text-amber-400 font-medium">
                                  ⚠ Lebih mahal {priceInfo.pct.toFixed(1)}% dari HPP
                                </p>
                              );
                            }
                          })()}
                        </div>
                      )}
                    </div>

                    {/* Discount */}
                    <div className="space-y-2">
                      <Label>Diskon (%)</Label>
                      <Input
                        type="number"
                        step="0.01"
                        min="0"
                        max="100"
                        value={item.discountPct}
                        onChange={(e) =>
                          updateLineItem(item.id, "discountPct", e.target.value)
                        }
                      />
                    </div>

                    {/* Subtotal */}
                    <div className="space-y-2">
                      <Label>Subtotal</Label>
                      <div className="h-10 flex items-center px-3 bg-muted rounded-md font-medium">
                        {formatCurrency(item.subtotal)}
                      </div>
                    </div>

                    {/* Notes */}
                    <div className="space-y-2 sm:col-span-2">
                      <Label>Catatan Item</Label>
                      <Input
                        value={item.notes}
                        onChange={(e) =>
                          updateLineItem(item.id, "notes", e.target.value)
                        }
                        placeholder="Catatan untuk item ini..."
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Summary */}
      {lineItems.length > 0 && (
        <Card className="border-2 bg-muted/30">
          <CardContent className="pt-6">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div>
                <p className="text-sm text-muted-foreground">
                  {lineItems.length} item produk
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-muted-foreground">Total</p>
                <p className="text-2xl font-bold">{formatCurrency(totalAmount)}</p>
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
