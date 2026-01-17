/**
 * Edit Sales Order Form Component
 *
 * Form for editing existing sales orders with:
 * - Pre-filled data from existing order
 * - Customer and warehouse information
 * - Order dates
 * - Order items management with add/remove capability
 * - Financial calculations (subtotal, PPN 11%, total)
 * - Real-time validation
 */

"use client";

import { useState, useEffect } from "react";
import { Loader2, AlertCircle, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useUpdateSalesOrderMutation } from "@/store/services/salesOrderApi";
import { useListCustomersQuery } from "@/store/services/customerApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { SalesOrderItemsManager } from "./sales-order-items-manager";
import { formatCurrency } from "@/lib/utils";
import { toast } from "sonner";
import type {
  UpdateSalesOrderRequest,
  SalesOrderResponse,
  CreateSalesOrderItemRequest,
} from "@/types/sales-order.types";

interface EditSalesOrderFormProps {
  salesOrder: SalesOrderResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditSalesOrderForm({
  salesOrder,
  onSuccess,
  onCancel,
}: EditSalesOrderFormProps) {
  const [updateSalesOrder, { isLoading }] = useUpdateSalesOrderMutation();

  // Fetch customers and warehouses for dropdowns
  const { data: customersData, isLoading: isLoadingCustomers } =
    useListCustomersQuery({
      page: 1,
      pageSize: 100,
      isActive: true,
    });

  const { data: warehousesData, isLoading: isLoadingWarehouses } =
    useListWarehousesQuery({
      page: 1,
      pageSize: 100,
      isActive: true,
    });

  // Form state - initialize with sales order data
  const [formData, setFormData] = useState({
    customerId: salesOrder.customerId,
    warehouseId: salesOrder.warehouseId,
    orderDate: salesOrder.orderDate.split("T")[0], // Convert ISO to YYYY-MM-DD
    requiredDate: salesOrder.requiredDate
      ? salesOrder.requiredDate.split("T")[0]
      : "",
    notes: salesOrder.notes || "",
    discount: salesOrder.discount,
    shippingCost: salesOrder.shippingCost,
  });

  // Items state - convert existing items to CreateSalesOrderItemRequest format
  const [items, setItems] = useState<CreateSalesOrderItemRequest[]>(
    salesOrder.items?.map((item) => ({
      productId: item.productId,
      unitId: item.unitId,
      orderedQty: item.orderedQty,
      unitPrice: item.unitPrice,
      discount: item.discount,
      lineTotal: item.lineTotal,
      notes: item.notes,
    })) || []
  );

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const handleChange = (
    field: keyof typeof formData,
    value: string | number
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user types
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
  };

  // Calculate totals from items
  const calculateTotals = () => {
    const itemsSubtotal = items.reduce((sum, item) => {
      return sum + parseFloat(item.lineTotal);
    }, 0);

    const discount = parseFloat(formData.discount) || 0;
    const shipping = parseFloat(formData.shippingCost) || 0;
    const taxRate = 0.11; // 11% PPN
    const taxableAmount = itemsSubtotal - discount;
    const tax = taxableAmount * taxRate;
    const total = taxableAmount + tax + shipping;

    return {
      subtotal: itemsSubtotal.toFixed(2),
      tax: tax.toFixed(2),
      total: total.toFixed(2),
    };
  };

  const totals = calculateTotals();

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields
    if (!formData.customerId) {
      newErrors.customerId = "Pelanggan wajib dipilih";
    }
    if (!formData.warehouseId) {
      newErrors.warehouseId = "Gudang wajib dipilih";
    }
    if (!formData.orderDate) {
      newErrors.orderDate = "Tanggal pesanan wajib diisi";
    }

    // Items validation
    if (items.length === 0) {
      newErrors.items = "Minimal 1 item pesanan harus ditambahkan";
    }

    // Numeric validations
    const discount = parseFloat(formData.discount);
    const shippingCost = parseFloat(formData.shippingCost);

    if (isNaN(discount) || discount < 0) {
      newErrors.discount = "Diskon tidak valid";
    }
    if (isNaN(shippingCost) || shippingCost < 0) {
      newErrors.shippingCost = "Ongkos kirim tidak valid";
    }

    setErrors(newErrors);
    setTouched({
      customerId: true,
      warehouseId: true,
      orderDate: true,
      discount: true,
      shippingCost: true,
    });

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
      // Clean up data before sending
      const requestData: UpdateSalesOrderRequest = {
        customerId: formData.customerId,
        warehouseId: formData.warehouseId,
        orderDate: formData.orderDate,
        requiredDate: formData.requiredDate || undefined,
        notes: formData.notes || undefined,
        subtotal: totals.subtotal,
        discount: formData.discount,
        tax: totals.tax,
        shippingCost: formData.shippingCost,
        totalAmount: totals.total,
        items: items.map((item, index) => ({
          id: salesOrder.items?.[index]?.id, // Preserve existing item IDs if available
          ...item,
        })),
      };

      const result = await updateSalesOrder({
        id: salesOrder.id,
        data: requestData,
      }).unwrap();

      toast.success("Pesanan Berhasil Diperbarui", {
        description: `Pesanan ${result.orderNumber} telah diperbarui`,
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      toast.error("Gagal Memperbarui Pesanan", {
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
      {/* Order Information Card */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle>Informasi Pesanan</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Customer & Warehouse Selection - 2 Columns */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Customer Selection */}
            <div className="space-y-2 w-full">
              <Label htmlFor="customerId" className="text-sm font-medium">
                Pelanggan <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.customerId}
                onValueChange={(value) => handleChange("customerId", value)}
                disabled={isLoadingCustomers}
              >
                <SelectTrigger
                  className={
                    errors.customerId && touched.customerId
                      ? "border-destructive w-full"
                      : "w-full"
                  }
                >
                  <SelectValue placeholder={isLoadingCustomers ? "Memuat..." : "Pilih pelanggan"} />
                </SelectTrigger>
                <SelectContent>
                  {isLoadingCustomers ? (
                    <SelectItem value="loading" disabled>
                      Memuat data pelanggan...
                    </SelectItem>
                  ) : customersData?.data && customersData.data.length > 0 ? (
                    customersData.data.map((customer) => (
                      <SelectItem key={customer.id} value={customer.id}>
                        {customer.code} - {customer.name}
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value="empty" disabled>
                      Tidak ada pelanggan aktif
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
              {errors.customerId && touched.customerId && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.customerId}
                </p>
              )}
              <Alert className="bg-blue-50 border-blue-200 dark:bg-blue-900/10 dark:border-blue-900">
                <Info className="h-4 w-4 text-blue-600" />
                <AlertDescription className="text-blue-800 dark:text-blue-200 text-sm">
                  Perubahan pelanggan dapat mempengaruhi harga dan diskon yang
                  berlaku
                </AlertDescription>
              </Alert>
            </div>

            {/* Warehouse Selection */}
            <div className="space-y-2 w-full">
              <Label htmlFor="warehouseId" className="text-sm font-medium">
                Gudang <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.warehouseId}
                onValueChange={(value) => handleChange("warehouseId", value)}
                disabled={isLoadingWarehouses}
              >
                <SelectTrigger
                  className={
                    errors.warehouseId && touched.warehouseId
                      ? "border-destructive w-full"
                      : "w-full"
                  }
                >
                  <SelectValue placeholder={isLoadingWarehouses ? "Memuat..." : "Pilih gudang"} />
                </SelectTrigger>
                <SelectContent>
                  {isLoadingWarehouses ? (
                    <SelectItem value="loading" disabled>
                      Memuat data gudang...
                    </SelectItem>
                  ) : warehousesData?.data && warehousesData.data.length > 0 ? (
                    warehousesData.data.map((warehouse) => (
                      <SelectItem key={warehouse.id} value={warehouse.id}>
                        {warehouse.code} - {warehouse.name}
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value="empty" disabled>
                      Tidak ada gudang aktif
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
              {errors.warehouseId && touched.warehouseId && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.warehouseId}
                </p>
              )}
              {formData.warehouseId !== salesOrder.warehouseId && (
                <Alert className="border-amber-500/50 bg-amber-50 dark:bg-amber-950/20">
                  <Info className="h-4 w-4 text-amber-600 dark:text-amber-500" />
                  <AlertDescription className="text-amber-800 dark:text-amber-400">
                    <strong>Perhatian:</strong> Perubahan gudang dapat
                    mempengaruhi ketersediaan stok produk
                  </AlertDescription>
                </Alert>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Order Date */}
            <div className="space-y-2">
              <Label htmlFor="orderDate" className="text-sm font-medium">
                Tanggal Pesanan <span className="text-destructive">*</span>
              </Label>
              <Input
                id="orderDate"
                type="date"
                value={formData.orderDate}
                onChange={(e) => handleChange("orderDate", e.target.value)}
                onBlur={() => handleBlur("orderDate")}
                className={
                  errors.orderDate && touched.orderDate
                    ? "border-destructive"
                    : ""
                }
              />
              {errors.orderDate && touched.orderDate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.orderDate}
                </p>
              )}
            </div>

            {/* Required Date */}
            <div className="space-y-2">
              <Label htmlFor="requiredDate" className="text-sm font-medium">
                Tanggal Dibutuhkan
              </Label>
              <Input
                id="requiredDate"
                type="date"
                value={formData.requiredDate}
                onChange={(e) => handleChange("requiredDate", e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Target waktu pengiriman (opsional)
              </p>
            </div>
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes" className="text-sm font-medium">
              Catatan
            </Label>
            <Textarea
              id="notes"
              value={formData.notes}
              onChange={(e) => handleChange("notes", e.target.value)}
              placeholder="Catatan tambahan untuk pesanan ini..."
              rows={3}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground">
              Informasi tambahan tentang pesanan
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Order Items Card */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle>Item Pesanan</CardTitle>
        </CardHeader>
        <CardContent>
          <SalesOrderItemsManager
            items={items}
            onChange={setItems}
            warehouseId={formData.warehouseId}
            customerId={formData.customerId}
          />
          {errors.items && (
            <p className="flex items-center gap-1 text-sm text-destructive mt-2">
              <AlertCircle className="h-3 w-3" />
              {errors.items}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Financial Summary Card */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle>Ringkasan Keuangan</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Subtotal */}
          <div className="flex justify-between items-center">
            <span className="text-sm">Subtotal</span>
            <span className="font-medium">
              {formatCurrency(totals.subtotal)}
            </span>
          </div>

          {/* Discount */}
          <div className="flex justify-between items-center">
            <Label htmlFor="discount" className="text-sm">
              Diskon
            </Label>
            <div className="w-32">
              <Input
                id="discount"
                type="number"
                step="0.01"
                min="0"
                value={formData.discount}
                onChange={(e) => handleChange("discount", e.target.value)}
                onBlur={() => handleBlur("discount")}
                className={`text-right ${
                  errors.discount && touched.discount ? "border-destructive" : ""
                }`}
              />
            </div>
          </div>
          {errors.discount && touched.discount && (
            <p className="flex items-center gap-1 text-sm text-destructive text-right">
              <AlertCircle className="h-3 w-3" />
              {errors.discount}
            </p>
          )}

          {/* Tax (PPN 11%) */}
          <div className="flex justify-between items-center">
            <span className="text-sm">PPN (11%)</span>
            <span className="font-medium">{formatCurrency(totals.tax)}</span>
          </div>

          {/* Shipping Cost */}
          <div className="flex justify-between items-center">
            <Label htmlFor="shippingCost" className="text-sm">
              Ongkos Kirim
            </Label>
            <div className="w-32">
              <Input
                id="shippingCost"
                type="number"
                step="0.01"
                min="0"
                value={formData.shippingCost}
                onChange={(e) => handleChange("shippingCost", e.target.value)}
                onBlur={() => handleBlur("shippingCost")}
                className={`text-right ${
                  errors.shippingCost && touched.shippingCost
                    ? "border-destructive"
                    : ""
                }`}
              />
            </div>
          </div>
          {errors.shippingCost && touched.shippingCost && (
            <p className="flex items-center gap-1 text-sm text-destructive text-right">
              <AlertCircle className="h-3 w-3" />
              {errors.shippingCost}
            </p>
          )}

          <Separator />

          {/* Total */}
          <div className="flex justify-between items-center">
            <span className="text-base font-bold">Total</span>
            <span className="text-lg font-bold">
              {formatCurrency(totals.total)}
            </span>
          </div>
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
          disabled={isLoading}
          size="lg"
          className="min-w-[150px]"
        >
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Menyimpan...
            </>
          ) : (
            <>
              Simpan Perubahan
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
