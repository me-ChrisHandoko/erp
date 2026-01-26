/**
 * Purchase Order Items Manager Component
 *
 * Component for managing purchase order items (add, edit, remove)
 * Used in both create and edit purchase order forms
 *
 * Features:
 * - Keyboard shortcuts (Enter to add, Escape to clear, Ctrl+K focus)
 * - Auto-focus for faster entry
 * - Product search with code auto-complete
 * - Price comparison (Supplier Price vs HPP)
 * - Quick add panel for top supplier products
 */

"use client";

import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import {
  Plus,
  Trash2,
  AlertCircle,
  Zap,
  Package2,
  Save,
  X,
  Edit,
  FileText,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useListProductsQuery } from "@/store/services/productApi";
import { formatCurrency } from "@/types/purchase-order.types";

export interface PurchaseOrderItem {
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

interface PurchaseOrderItemsManagerProps {
  items: PurchaseOrderItem[];
  onChange: (items: PurchaseOrderItem[]) => void;
  supplierId: string;
  disabled?: boolean;
}

export function PurchaseOrderItemsManager({
  items,
  onChange,
  supplierId,
  disabled = false,
}: PurchaseOrderItemsManagerProps) {
  // Fetch products filtered by supplier
  const { data: productsData, isLoading: isLoadingProducts } = useListProductsQuery(
    { isActive: true, supplierId: supplierId, pageSize: 100 },
    { skip: !supplierId }
  );

  const products = productsData?.data || [];

  // New item state
  const [newItem, setNewItem] = useState<{
    productId: string;
    quantity: string;
    unitPrice: string;
    discountPct: string;
    notes: string;
  }>({
    productId: "",
    quantity: "1",
    unitPrice: "0",
    discountPct: "0",
    notes: "",
  });

  // Selected product info
  const [selectedProduct, setSelectedProduct] = useState<{
    id: string;
    code: string;
    name: string;
    baseUnit: string;
    baseCost: number;
    supplierPrice: number;
  } | null>(null);

  // Inline editing state
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editingData, setEditingData] = useState<{
    quantity: string;
    unitPrice: string;
    discountPct: string;
    notes: string;
  } | null>(null);

  // Refs for keyboard navigation
  const productComboboxRef = useRef<HTMLButtonElement>(null);
  const qtyInputRef = useRef<HTMLInputElement>(null);

  // Convert products to Combobox options
  const productOptions: ComboboxOption[] = useMemo(() => {
    if (!products || products.length === 0) return [];

    // Exclude already added products
    const addedProductIds = items.map((item) => item.productId);

    return products
      .filter((product) => !addedProductIds.includes(product.id))
      .map((product) => {
        // Find supplier-specific price
        const supplierInfo = product.suppliers?.find(
          (s: any) => s.supplierId === supplierId
        );
        const supplierPrice = supplierInfo?.supplierPrice
          ? parseFloat(supplierInfo.supplierPrice)
          : 0;
        const baseCost = parseFloat(product.baseCost) || 0;

        // Price indicator
        let priceIndicator = "";
        if (supplierPrice > 0 && baseCost > 0) {
          if (supplierPrice < baseCost) {
            priceIndicator = "🟢"; // Good - below HPP
          } else if (supplierPrice === baseCost) {
            priceIndicator = "🔵"; // Same as HPP
          } else {
            priceIndicator = "🟠"; // Above HPP
          }
        }

        return {
          value: product.id,
          label: `${priceIndicator} ${product.code} - ${product.name}`,
          searchLabel: `${product.code} ${product.name}`,
        };
      });
  }, [products, supplierId, items]);

  // Get top 8 products for quick add (first 8 from supplier's products)
  const quickAddProducts = useMemo(() => {
    if (!products || products.length === 0) return [];

    // Exclude already added products
    const addedProductIds = items.map((item) => item.productId);

    return products
      .filter((product) => !addedProductIds.includes(product.id))
      .slice(0, 8)
      .map((product) => {
        const supplierInfo = product.suppliers?.find(
          (s: any) => s.supplierId === supplierId
        );
        const supplierPrice = supplierInfo?.supplierPrice
          ? parseFloat(supplierInfo.supplierPrice)
          : parseFloat(product.baseCost) || 0;

        return {
          id: product.id,
          code: product.code,
          name: product.name,
          baseUnit: product.baseUnit,
          baseCost: parseFloat(product.baseCost) || 0,
          supplierPrice,
        };
      });
  }, [products, supplierId, items]);

  // Calculate subtotal
  const calculateSubtotal = useCallback(
    (qty: string, price: string, discPct: string): number => {
      const qtyNum = parseFloat(qty) || 0;
      const priceNum = parseFloat(price) || 0;
      const discNum = parseFloat(discPct) || 0;
      const discAmount = (qtyNum * priceNum * discNum) / 100;
      return qtyNum * priceNum - discAmount;
    },
    []
  );

  // Reset new item form
  const handleResetForm = useCallback(() => {
    setNewItem({
      productId: "",
      quantity: "1",
      unitPrice: "0",
      discountPct: "0",
      notes: "",
    });
    setSelectedProduct(null);
  }, []);

  // Add item to list
  const handleAddItem = useCallback(() => {
    if (!newItem.productId || !selectedProduct) return;

    const qty = parseFloat(newItem.quantity) || 0;
    const price = parseFloat(newItem.unitPrice) || 0;

    if (qty <= 0 || price <= 0) return;

    const newOrderItem: PurchaseOrderItem = {
      id: `item-${Date.now()}`,
      productId: newItem.productId,
      productName: selectedProduct.name,
      productCode: selectedProduct.code,
      productUnit: selectedProduct.baseUnit,
      quantity: newItem.quantity,
      unitPrice: newItem.unitPrice,
      discountPct: newItem.discountPct,
      subtotal: calculateSubtotal(
        newItem.quantity,
        newItem.unitPrice,
        newItem.discountPct
      ),
      notes: newItem.notes,
      baseCost: selectedProduct.baseCost,
      supplierPrice: selectedProduct.supplierPrice,
    };

    onChange([...items, newOrderItem]);
    handleResetForm();

    // Focus back to product search
    setTimeout(() => {
      productComboboxRef.current?.focus();
    }, 100);
  }, [newItem, selectedProduct, items, onChange, calculateSubtotal, handleResetForm]);

  // Quick add item (1-click)
  const handleQuickAdd = useCallback(
    (product: (typeof quickAddProducts)[0]) => {
      const newOrderItem: PurchaseOrderItem = {
        id: `item-${Date.now()}`,
        productId: product.id,
        productName: product.name,
        productCode: product.code,
        productUnit: product.baseUnit,
        quantity: "1",
        unitPrice: product.supplierPrice.toString(),
        discountPct: "0",
        subtotal: product.supplierPrice,
        notes: "",
        baseCost: product.baseCost,
        supplierPrice: product.supplierPrice,
      };

      onChange([...items, newOrderItem]);
    },
    [items, onChange]
  );

  // Remove item
  const handleRemoveItem = useCallback(
    (itemId: string) => {
      onChange(items.filter((item) => item.id !== itemId));
    },
    [items, onChange]
  );

  // Start editing item
  const handleStartEdit = useCallback((index: number) => {
    const item = items[index];
    setEditingIndex(index);
    setEditingData({
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      discountPct: item.discountPct,
      notes: item.notes,
    });
  }, [items]);

  // Save edited item
  const handleSaveEdit = useCallback(() => {
    if (editingIndex === null || !editingData) return;

    const updatedItems = [...items];
    const item = updatedItems[editingIndex];
    updatedItems[editingIndex] = {
      ...item,
      quantity: editingData.quantity,
      unitPrice: editingData.unitPrice,
      discountPct: editingData.discountPct,
      notes: editingData.notes,
      subtotal: calculateSubtotal(
        editingData.quantity,
        editingData.unitPrice,
        editingData.discountPct
      ),
    };

    onChange(updatedItems);
    setEditingIndex(null);
    setEditingData(null);
  }, [editingIndex, editingData, items, onChange, calculateSubtotal]);

  // Cancel editing
  const handleCancelEdit = useCallback(() => {
    setEditingIndex(null);
    setEditingData(null);
  }, []);

  // Update selected product when product changes
  useEffect(() => {
    if (newItem.productId && products.length > 0) {
      const product = products.find((p) => p.id === newItem.productId);
      if (product) {
        const supplierInfo = product.suppliers?.find(
          (s: any) => s.supplierId === supplierId
        );
        const supplierPrice = supplierInfo?.supplierPrice
          ? parseFloat(supplierInfo.supplierPrice)
          : parseFloat(product.baseCost) || 0;

        setSelectedProduct({
          id: product.id,
          code: product.code,
          name: product.name,
          baseUnit: product.baseUnit,
          baseCost: parseFloat(product.baseCost) || 0,
          supplierPrice,
        });

        // Auto-fill price
        setNewItem((prev) => ({
          ...prev,
          unitPrice: supplierPrice.toString(),
        }));

        // Auto-focus qty input
        setTimeout(() => {
          qtyInputRef.current?.focus();
          qtyInputRef.current?.select();
        }, 100);
      }
    } else {
      setSelectedProduct(null);
    }
  }, [newItem.productId, products, supplierId]);

  // Keyboard shortcuts handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger if disabled
      if (disabled) return;

      // Ctrl+K: Focus to product search
      if (e.ctrlKey && e.key === "k") {
        e.preventDefault();
        productComboboxRef.current?.focus();
        setTimeout(() => {
          productComboboxRef.current?.click();
        }, 100);
        return;
      }

      // Escape: Clear form or cancel edit
      if (e.key === "Escape") {
        e.preventDefault();
        if (editingIndex !== null) {
          handleCancelEdit();
        } else {
          handleResetForm();
        }
        return;
      }

      // Enter: Add item (only if not in edit mode and not typing in inputs)
      if (e.key === "Enter" && editingIndex === null) {
        const target = e.target as HTMLElement;
        if (target.tagName !== "INPUT" || target.getAttribute("type") === "number") {
          if (newItem.productId && selectedProduct) {
            e.preventDefault();
            handleAddItem();
          }
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [disabled, newItem, selectedProduct, editingIndex, handleAddItem, handleResetForm, handleCancelEdit]);

  // Reset form when supplier changes
  useEffect(() => {
    handleResetForm();
  }, [supplierId, handleResetForm]);

  // Helper function for price comparison display
  const getPriceComparisonClass = (currentPrice: number, baseCost: number) => {
    if (baseCost === 0) return "";
    if (currentPrice < baseCost) return "text-green-600 dark:text-green-400";
    if (currentPrice > baseCost) return "text-amber-600 dark:text-amber-400";
    return "text-blue-600 dark:text-blue-400";
  };

  const getPriceComparisonText = (currentPrice: number, baseCost: number) => {
    if (baseCost === 0) return null;
    const diff = currentPrice - baseCost;
    const pct = Math.abs((diff / baseCost) * 100);

    if (diff < 0) {
      return `✓ Hemat ${pct.toFixed(1)}%`;
    } else if (diff > 0) {
      return `⚠ +${pct.toFixed(1)}%`;
    }
    return "≈ HPP";
  };

  return (
    <div className="space-y-4">
      {/* Keyboard Shortcuts Hint */}
      <div className="flex items-center gap-3 text-xs text-muted-foreground bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 rounded-lg px-3 py-2">
        <Zap className="h-4 w-4 text-blue-600 dark:text-blue-400" />
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          <span>
            <kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">
              Ctrl+K
            </kbd>{" "}
            Focus Produk
          </span>
          <span>
            <kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">
              Enter
            </kbd>{" "}
            Tambah Item
          </span>
          <span>
            <kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">
              Esc
            </kbd>{" "}
            Reset
          </span>
          <span className="text-blue-600 dark:text-blue-400">
            💡 Ketik kode produk untuk quick select
          </span>
        </div>
      </div>

      {/* Quick Add Panel */}
      {quickAddProducts.length > 0 && !disabled && (
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
            <Zap className="h-4 w-4" />
            <span>Quick Add (1-Click)</span>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-2">
            {quickAddProducts.map((product) => {
              const priceClass = getPriceComparisonClass(
                product.supplierPrice,
                product.baseCost
              );

              return (
                <Button
                  key={product.id}
                  type="button"
                  variant="outline"
                  onClick={() => handleQuickAdd(product)}
                  className="h-auto py-3 px-2 flex flex-col items-start gap-1 hover:bg-accent hover:border-primary transition-all"
                  title={`Quick add: ${product.code} - ${product.name}`}
                >
                  <span className="font-bold text-xs text-primary">
                    {product.code}
                  </span>
                  <span className="text-[10px] line-clamp-2 text-left leading-tight">
                    {product.name}
                  </span>
                  <div className="flex items-center justify-between w-full mt-1">
                    <span className={`text-xs font-semibold ${priceClass}`}>
                      {formatCurrency(product.supplierPrice)}
                    </span>
                  </div>
                </Button>
              );
            })}
          </div>
        </div>
      )}

      {/* Add Item Form */}
      {!disabled && (
        <div className="grid grid-cols-1 md:grid-cols-6 gap-3 p-4 border rounded-lg bg-muted/50">
          {/* Product Selection */}
          <div className="md:col-span-2 space-y-1">
            <Label className="text-xs flex items-center gap-1">
              <Package2 className="h-3 w-3" />
              Produk
              <span className="text-[10px] text-muted-foreground ml-1">
                (Ketik kode atau nama)
              </span>
            </Label>
            <Combobox
              ref={productComboboxRef}
              value={newItem.productId}
              onValueChange={(value) =>
                setNewItem((prev) => ({ ...prev, productId: value }))
              }
              options={productOptions}
              placeholder={
                isLoadingProducts
                  ? "Memuat..."
                  : !supplierId
                  ? "Pilih supplier terlebih dahulu"
                  : productOptions.length === 0
                  ? "Semua produk sudah ditambahkan"
                  : "Ketik kode atau cari produk..."
              }
              searchPlaceholder="Ketik kode produk atau cari..."
              emptyMessage="Produk tidak ditemukan"
              disabled={isLoadingProducts || !supplierId || productOptions.length === 0}
              className="w-full"
            />
            {selectedProduct && (
              <p className="text-xs text-muted-foreground">
                Satuan: {selectedProduct.baseUnit} | HPP:{" "}
                {formatCurrency(selectedProduct.baseCost)}
              </p>
            )}
          </div>

          {/* Quantity */}
          <div className="space-y-1">
            <Label className="text-xs">Kuantitas</Label>
            <Input
              ref={qtyInputRef}
              type="number"
              step="0.001"
              min="0"
              value={newItem.quantity}
              onChange={(e) =>
                setNewItem((prev) => ({ ...prev, quantity: e.target.value }))
              }
              onKeyDown={(e) => {
                if (e.key === "Enter" && newItem.productId) {
                  e.preventDefault();
                  handleAddItem();
                }
              }}
              placeholder="0"
              disabled={!selectedProduct}
            />
          </div>

          {/* Unit Price */}
          <div className="space-y-1">
            <Label className="text-xs">Harga Satuan</Label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
                Rp
              </span>
              <Input
                type="number"
                step="0.01"
                min="0"
                value={newItem.unitPrice}
                onChange={(e) =>
                  setNewItem((prev) => ({ ...prev, unitPrice: e.target.value }))
                }
                onKeyDown={(e) => {
                  if (e.key === "Enter" && newItem.productId) {
                    e.preventDefault();
                    handleAddItem();
                  }
                }}
                className="pl-8"
                placeholder="0"
                disabled={!selectedProduct}
              />
            </div>
            {selectedProduct && (
              <p
                className={`text-xs ${getPriceComparisonClass(
                  parseFloat(newItem.unitPrice) || 0,
                  selectedProduct.baseCost
                )}`}
              >
                {getPriceComparisonText(
                  parseFloat(newItem.unitPrice) || 0,
                  selectedProduct.baseCost
                )}
              </p>
            )}
          </div>

          {/* Discount */}
          <div className="space-y-1">
            <Label className="text-xs">Diskon (%)</Label>
            <Input
              type="number"
              step="0.01"
              min="0"
              max="100"
              value={newItem.discountPct}
              onChange={(e) =>
                setNewItem((prev) => ({ ...prev, discountPct: e.target.value }))
              }
              onKeyDown={(e) => {
                if (e.key === "Enter" && newItem.productId) {
                  e.preventDefault();
                  handleAddItem();
                }
              }}
              placeholder="0"
              disabled={!selectedProduct}
            />
          </div>

          {/* Notes */}
          <div className="space-y-1">
            <Label className="text-xs flex items-center gap-1">
              <FileText className="h-3 w-3" />
              Catatan
            </Label>
            <Input
              value={newItem.notes}
              onChange={(e) =>
                setNewItem((prev) => ({ ...prev, notes: e.target.value }))
              }
              onKeyDown={(e) => {
                if (e.key === "Enter" && newItem.productId) {
                  e.preventDefault();
                  handleAddItem();
                }
              }}
              placeholder="Opsional..."
              disabled={!selectedProduct}
            />
          </div>

          {/* Add Button */}
          <div className="flex items-end">
            <Button
              type="button"
              onClick={handleAddItem}
              disabled={!newItem.productId || !selectedProduct}
              className="w-full"
            >
              <Plus className="h-4 w-4 mr-1" />
              Tambah
            </Button>
          </div>
        </div>
      )}

      {/* Empty State */}
      {items.length === 0 && (
        <div className="text-center py-8 text-muted-foreground border rounded-lg">
          <Package2 className="mx-auto h-12 w-12 mb-4 opacity-50" />
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
              <p className="text-sm">
                Gunakan Quick Add atau form di atas untuk menambahkan produk
              </p>
            </>
          )}
        </div>
      )}

      {/* Items Table */}
      {items.length > 0 && (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[50px]">#</TableHead>
                <TableHead>Produk</TableHead>
                <TableHead className="text-right w-[100px]">Qty</TableHead>
                <TableHead className="text-right w-[140px]">Harga</TableHead>
                <TableHead className="text-right w-[80px]">Diskon</TableHead>
                <TableHead className="text-right w-[140px]">Subtotal</TableHead>
                <TableHead className="w-[180px]">Catatan</TableHead>
                <TableHead className="w-[100px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item, index) => (
                <TableRow key={item.id}>
                  <TableCell className="font-medium text-muted-foreground">
                    {index + 1}
                  </TableCell>
                  <TableCell>
                    <div>
                      <p className="font-medium">
                        {item.productCode} - {item.productName}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {item.productUnit}
                        {item.baseCost > 0 && (
                          <span className="ml-2">
                            HPP: {formatCurrency(item.baseCost)}
                          </span>
                        )}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    {editingIndex === index ? (
                      <Input
                        type="number"
                        step="0.001"
                        min="0"
                        value={editingData?.quantity || ""}
                        onChange={(e) =>
                          setEditingData((prev) =>
                            prev ? { ...prev, quantity: e.target.value } : null
                          )
                        }
                        className="w-20 text-right"
                      />
                    ) : (
                      <span>{parseFloat(item.quantity).toLocaleString("id-ID")}</span>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {editingIndex === index ? (
                      <Input
                        type="number"
                        step="0.01"
                        min="0"
                        value={editingData?.unitPrice || ""}
                        onChange={(e) =>
                          setEditingData((prev) =>
                            prev ? { ...prev, unitPrice: e.target.value } : null
                          )
                        }
                        className="w-28 text-right"
                      />
                    ) : (
                      <div>
                        <span>{formatCurrency(item.unitPrice)}</span>
                        {item.baseCost > 0 && (
                          <p
                            className={`text-xs ${getPriceComparisonClass(
                              parseFloat(item.unitPrice),
                              item.baseCost
                            )}`}
                          >
                            {getPriceComparisonText(
                              parseFloat(item.unitPrice),
                              item.baseCost
                            )}
                          </p>
                        )}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {editingIndex === index ? (
                      <Input
                        type="number"
                        step="0.01"
                        min="0"
                        max="100"
                        value={editingData?.discountPct || ""}
                        onChange={(e) =>
                          setEditingData((prev) =>
                            prev ? { ...prev, discountPct: e.target.value } : null
                          )
                        }
                        className="w-16 text-right"
                      />
                    ) : (
                      <span>
                        {parseFloat(item.discountPct) > 0
                          ? `${item.discountPct}%`
                          : "-"}
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {editingIndex === index
                      ? formatCurrency(
                          calculateSubtotal(
                            editingData?.quantity || "0",
                            editingData?.unitPrice || "0",
                            editingData?.discountPct || "0"
                          )
                        )
                      : formatCurrency(item.subtotal)}
                  </TableCell>
                  <TableCell>
                    {editingIndex === index ? (
                      <Input
                        value={editingData?.notes || ""}
                        onChange={(e) =>
                          setEditingData((prev) =>
                            prev ? { ...prev, notes: e.target.value } : null
                          )
                        }
                        placeholder="Catatan..."
                        className="w-full text-xs"
                      />
                    ) : (
                      <span className="text-xs text-muted-foreground truncate block max-w-[160px]" title={item.notes || ""}>
                        {item.notes || "-"}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-1">
                      {editingIndex === index ? (
                        <>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={handleSaveEdit}
                            className="h-8 w-8 p-0 text-green-600 hover:text-green-700"
                          >
                            <Save className="h-4 w-4" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={handleCancelEdit}
                            className="h-8 w-8 p-0"
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        </>
                      ) : (
                        <>
                          {!disabled && (
                            <>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleStartEdit(index)}
                                className="h-8 w-8 p-0"
                              >
                                <Edit className="h-4 w-4" />
                              </Button>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemoveItem(item.id)}
                                className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </>
                          )}
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
