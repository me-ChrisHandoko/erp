/**
 * Create Sales Order Form Component
 *
 * Form for creating new sales orders with:
 * - Customer and warehouse selection
 * - Order date and required date
 * - Order items management with product selection
 * - Financial calculations (subtotal, PPN 11%, total)
 * - Form validation and submission
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import { selectActiveCompany } from "@/store/slices/companySlice";
import { Loader2, Plus, AlertTriangle, Package, Tag, Receipt, Truck, Wallet, Percent, DollarSign, CreditCard, TrendingUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Separator } from "@/components/ui/separator";
import { useForm } from "react-hook-form";
import { useCreateSalesOrderMutation } from "@/store/services/salesOrderApi";
import { useListCustomersQuery, useGetCustomerCreditInfoQuery } from "@/store/services/customerApi";
import { useListWarehousesQuery } from "@/store/services/warehouseApi";
import { SalesOrderItemsManager } from "./sales-order-items-manager";
import { formatCurrency } from "@/lib/utils";
import type { CreateSalesOrderRequest, CreateSalesOrderItemRequest } from "@/types/sales-order.types";

interface CreateSalesOrderFormProps {
  onSuccess: (orderId: string) => void;
  onCancel: () => void;
}

export function CreateSalesOrderForm({
  onSuccess,
  onCancel,
}: CreateSalesOrderFormProps) {
  const [createSalesOrder, { isLoading, error }] =
    useCreateSalesOrderMutation();

  // Get active company for tax rate
  const activeCompany = useSelector(selectActiveCompany);

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

  // Form state
  const [formData, setFormData] = useState({
    customerId: "",
    warehouseId: "",
    orderDate: new Date().toISOString().split("T")[0],
    requiredDate: "",
    notes: "",
    discount: "0",
    shippingCost: "0",
  });

  // Fetch customer credit info for credit limit validation
  const { data: creditInfo } = useGetCustomerCreditInfoQuery(
    formData.customerId,
    {
      skip: !formData.customerId, // Only fetch when customer is selected
    }
  );

  const [items, setItems] = useState<CreateSalesOrderItemRequest[]>([]);

  // Warehouse change confirmation state
  const [showWarehouseChangeDialog, setShowWarehouseChangeDialog] = useState(false);
  const [pendingWarehouseId, setPendingWarehouseId] = useState<string>("");

  // Discount mode: 'nominal' or 'percentage'
  const [discountMode, setDiscountMode] = useState<'nominal' | 'percentage'>('nominal');
  const [discountPercentage, setDiscountPercentage] = useState<string>("0");

  // Calculate totals from items
  const calculateTotals = () => {
    const itemsSubtotal = items.reduce((sum, item) => {
      return sum + parseFloat(item.lineTotal);
    }, 0);

    const discount = parseFloat(formData.discount) || 0;
    const shipping = parseFloat(formData.shippingCost) || 0;

    // Get tax rate from active company
    // IMPORTANT: If company is not PKP, tax rate should be 0% regardless of ppnRate value
    let taxRate = 0;
    if (activeCompany?.isPKP) {
      taxRate = activeCompany.ppnRate ? activeCompany.ppnRate / 100 : 0.11;
    }

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

  // Handle discount preset selection
  const handleDiscountPreset = (percentage: number) => {
    setDiscountMode('percentage');
    setDiscountPercentage(percentage.toString());
    const itemsSubtotal = items.reduce((sum, item) => sum + parseFloat(item.lineTotal), 0);
    const discountAmount = (itemsSubtotal * percentage) / 100;
    setFormData((prev) => ({ ...prev, discount: discountAmount.toFixed(2) }));
  };

  // Toggle discount mode
  const handleToggleDiscountMode = () => {
    if (discountMode === 'nominal') {
      // Switch to percentage: calculate percentage from current nominal
      const itemsSubtotal = items.reduce((sum, item) => sum + parseFloat(item.lineTotal), 0);
      const currentDiscount = parseFloat(formData.discount) || 0;
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
    const itemsSubtotal = items.reduce((sum, item) => sum + parseFloat(item.lineTotal), 0);
    const percentage = parseFloat(value) || 0;
    const discountAmount = (itemsSubtotal * percentage) / 100;
    setFormData((prev) => ({ ...prev, discount: discountAmount.toFixed(2) }));
  };

  // Handle warehouse change with confirmation if items exist
  const handleWarehouseChange = (newWarehouseId: string) => {
    // If warehouse is being changed and there are items, show confirmation
    if (formData.warehouseId && items.length > 0 && newWarehouseId !== formData.warehouseId) {
      setPendingWarehouseId(newWarehouseId);
      setShowWarehouseChangeDialog(true);
    } else {
      // No items or same warehouse, just update
      setFormData((prev) => ({ ...prev, warehouseId: newWarehouseId }));
    }
  };

  // Confirm warehouse change and clear items
  const handleConfirmWarehouseChange = () => {
    setFormData((prev) => ({
      ...prev,
      warehouseId: pendingWarehouseId,
      discount: "0",
      shippingCost: "0",
    }));
    setItems([]); // Clear all items
    setShowWarehouseChangeDialog(false);
    setPendingWarehouseId("");
  };

  // Cancel warehouse change
  const handleCancelWarehouseChange = () => {
    setShowWarehouseChangeDialog(false);
    setPendingWarehouseId("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Validation
      if (!formData.customerId || !formData.warehouseId) {
        alert("Mohon pilih pelanggan dan gudang");
        return;
      }

      if (items.length === 0) {
        alert("Mohon tambahkan minimal 1 item pesanan");
        return;
      }

      const requestData: CreateSalesOrderRequest = {
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
        items: items,
      };

      const result = await createSalesOrder(requestData).unwrap();
      onSuccess(result.id);
    } catch (err) {
      console.error("Failed to create sales order:", err);
      // Error is handled by RTK Query
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="space-y-6">
        {/* Order Information Card */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Pesanan</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Customer & Warehouse Selection - 2 Columns */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Customer Selection */}
              <div className="space-y-2 w-full">
                <label className="text-sm font-medium">
                  Pelanggan <span className="text-destructive">*</span>
                </label>
                <Select
                  value={formData.customerId}
                  onValueChange={(value) =>
                    setFormData((prev) => ({ ...prev, customerId: value }))
                  }
                  disabled={isLoadingCustomers}
                >
                  <SelectTrigger className="w-full">
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
              </div>

              {/* Warehouse Selection */}
              <div className="space-y-2 w-full">
                <label className="text-sm font-medium">
                  Gudang <span className="text-destructive">*</span>
                </label>
                <Select
                  value={formData.warehouseId}
                  onValueChange={handleWarehouseChange}
                  disabled={isLoadingWarehouses}
                >
                  <SelectTrigger className="w-full">
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
              </div>
            </div>

            {/* PHASE 3: Customer Credit Info Display */}
            {creditInfo && formData.customerId && (
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <CreditCard className="h-4 w-4" />
                  <span>Informasi Kredit Customer</span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 bg-muted/50 p-4 rounded-lg">
                  {/* Credit Limit */}
                  <div className="space-y-1">
                    <div className="text-xs text-muted-foreground">Limit Kredit</div>
                    <div className="text-lg font-semibold">
                      {formatCurrency(creditInfo.creditLimit)}
                    </div>
                  </div>

                  {/* Outstanding Amount */}
                  <div className="space-y-1">
                    <div className="text-xs text-muted-foreground">Saldo Outstanding</div>
                    <div className="text-lg font-semibold text-orange-600 dark:text-orange-400">
                      {formatCurrency(creditInfo.outstandingAmount)}
                    </div>
                    {parseFloat(creditInfo.overdueAmount) > 0 && (
                      <div className="text-xs text-red-600 dark:text-red-400">
                        Overdue: {formatCurrency(creditInfo.overdueAmount)}
                      </div>
                    )}
                  </div>

                  {/* Available Credit */}
                  <div className="space-y-1">
                    <div className="text-xs text-muted-foreground">Kredit Tersedia</div>
                    <div className={`text-lg font-semibold ${
                      parseFloat(creditInfo.availableCredit) < 0
                        ? 'text-red-600 dark:text-red-400'
                        : parseFloat(creditInfo.availableCredit) < parseFloat(creditInfo.creditLimit) * 0.2
                        ? 'text-orange-600 dark:text-orange-400'
                        : 'text-green-600 dark:text-green-400'
                    }`}>
                      {formatCurrency(creditInfo.availableCredit)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Utilisasi: {creditInfo.utilizationPercent}%
                    </div>
                  </div>
                </div>

                {/* PHASE 3: Credit Limit Validation Warning */}
                {(() => {
                  const currentOrderTotal = parseFloat(totals.total);
                  const availableCredit = parseFloat(creditInfo.availableCredit);
                  const newBalance = availableCredit - currentOrderTotal;

                  if (creditInfo.isExceedingLimit) {
                    return (
                      <Alert variant="destructive">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>
                          <strong>Customer sudah melebihi limit kredit!</strong>
                          <div className="mt-1">
                            Outstanding: {formatCurrency(creditInfo.outstandingAmount)} sudah melebihi limit {formatCurrency(creditInfo.creditLimit)}
                          </div>
                        </AlertDescription>
                      </Alert>
                    );
                  }

                  if (currentOrderTotal > availableCredit) {
                    return (
                      <Alert variant="destructive">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>
                          <strong>Order ini akan melebihi limit kredit customer!</strong>
                          <div className="mt-1">
                            Total order: {formatCurrency(totals.total)} melebihi kredit tersedia: {formatCurrency(creditInfo.availableCredit)}
                          </div>
                          <div className="text-xs mt-1">
                            Kekurangan: {formatCurrency(Math.abs(newBalance).toFixed(2))}
                          </div>
                        </AlertDescription>
                      </Alert>
                    );
                  }

                  if (newBalance < parseFloat(creditInfo.creditLimit) * 0.1) {
                    return (
                      <Alert className="border-orange-300 dark:border-orange-800">
                        <TrendingUp className="h-4 w-4 text-orange-600 dark:text-orange-400" />
                        <AlertDescription className="text-orange-900 dark:text-orange-200">
                          <strong>Peringatan: Kredit hampir habis!</strong>
                          <div className="mt-1 text-sm">
                            Setelah order ini, kredit tersisa: {formatCurrency(newBalance.toFixed(2))}
                            ({((newBalance / parseFloat(creditInfo.creditLimit)) * 100).toFixed(1)}% dari limit)
                          </div>
                        </AlertDescription>
                      </Alert>
                    );
                  }

                  return null;
                })()}
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Order Date */}
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Tanggal Pesanan <span className="text-destructive">*</span>
                </label>
                <Input
                  type="date"
                  value={formData.orderDate}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      orderDate: e.target.value,
                    }))
                  }
                  required
                />
              </div>

              {/* Required Date */}
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Tanggal Dibutuhkan
                </label>
                <Input
                  type="date"
                  value={formData.requiredDate}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      requiredDate: e.target.value,
                    }))
                  }
                />
              </div>
            </div>

            {/* Notes */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Catatan</label>
              <Textarea
                placeholder="Catatan tambahan untuk pesanan ini..."
                value={formData.notes}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, notes: e.target.value }))
                }
                rows={3}
              />
            </div>
          </CardContent>
        </Card>

        {/* Order Items Card */}
        <Card>
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
          </CardContent>
        </Card>

        {/* Financial Summary Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wallet className="h-5 w-5" />
              Ringkasan Keuangan
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {items.length === 0 ? (
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
                      Subtotal ({items.length} item{items.length > 1 ? "" : ""})
                    </span>
                  </div>
                  <span className="font-medium text-foreground transition-all duration-300">
                    {formatCurrency(totals.subtotal)}
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
                            value={discountMode === 'nominal' ? formData.discount : discountPercentage}
                            onChange={(e) => {
                              if (discountMode === 'nominal') {
                                setFormData((prev) => ({
                                  ...prev,
                                  discount: e.target.value,
                                }));
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
                      {parseFloat(formData.discount) > 0 && (
                        <div className="text-xs text-muted-foreground transition-all duration-300">
                          {discountMode === 'nominal' ? (
                            // Show percentage when in nominal mode
                            (() => {
                              const itemsSubtotal = items.reduce((sum, item) => sum + parseFloat(item.lineTotal), 0);
                              const percentage = itemsSubtotal > 0 ? (parseFloat(formData.discount) / itemsSubtotal) * 100 : 0;
                              return `(${percentage.toFixed(2)}% dari subtotal)`;
                            })()
                          ) : (
                            // Show nominal when in percentage mode
                            `= ${formatCurrency(formData.discount)}`
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
                      {formatCurrency(totals.tax)}
                    </div>
                    <div className="text-xs text-muted-foreground transition-all duration-300">
                      dari {formatCurrency((parseFloat(totals.subtotal) - parseFloat(formData.discount)).toFixed(2))}
                    </div>
                  </div>
                </div>

                {/* Shipping Cost */}
                <div className="flex justify-between items-center text-blue-600 dark:text-blue-500">
                  <div className="flex items-center gap-2">
                    <Truck className="h-4 w-4" />
                    <span className="text-sm">Ongkos Kirim</span>
                  </div>
                  <div className="relative w-36">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
                      Rp
                    </span>
                    <Input
                      type="number"
                      step="0.01"
                      min="0"
                      value={formData.shippingCost}
                      onChange={(e) =>
                        setFormData((prev) => ({
                          ...prev,
                          shippingCost: e.target.value,
                        }))
                      }
                      className="pl-8 text-right h-9 transition-all duration-200"
                      placeholder="0"
                    />
                  </div>
                </div>

                <Separator />

                {/* Total with Highlight */}
                <div className="bg-primary/10 -mx-6 -mb-6 px-6 py-4 rounded-b-lg transition-all duration-300">
                  <div className="flex justify-between items-center">
                    <span className="text-base font-bold">Total Pembayaran</span>
                    <span className="text-2xl font-bold text-primary transition-all duration-300 hover:scale-105">
                      {formatCurrency(totals.total)}
                    </span>
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Error Display */}
        {error && (
          <Card className="border-destructive">
            <CardContent className="pt-6">
              <p className="text-sm text-destructive">
                {"data" in error && typeof error.data === "object" && error.data !== null && "message" in error.data
                  ? String((error.data as any).message)
                  : "Gagal membuat pesanan. Silakan coba lagi."}
              </p>
            </CardContent>
          </Card>
        )}

        {/* Warehouse Change Confirmation Dialog */}
        <AlertDialog open={showWarehouseChangeDialog} onOpenChange={setShowWarehouseChangeDialog}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle className="flex items-center gap-2">
                <AlertTriangle className="h-5 w-5 text-destructive" />
                Konfirmasi Ganti Gudang
              </AlertDialogTitle>
              <AlertDialogDescription>
                Anda sudah menambahkan {items.length} item pesanan. Jika mengganti gudang,
                semua item yang sudah ditambahkan akan dihapus.
                <br /><br />
                Apakah Anda yakin ingin mengganti gudang?
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={handleCancelWarehouseChange}>
                Batal
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirmWarehouseChange}
                className="bg-destructive hover:bg-destructive/90"
              >
                Ya, Ganti Gudang
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        {/* Action Buttons */}
        <div className="flex gap-2 justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
          >
            Batal
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {isLoading ? "Menyimpan..." : "Simpan Pesanan"}
          </Button>
        </div>
      </div>
    </form>
  );
}
