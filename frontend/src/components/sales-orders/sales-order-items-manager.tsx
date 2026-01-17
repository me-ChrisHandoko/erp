/**
 * Sales Order Items Manager Component
 *
 * Component for managing sales order items (add, edit, remove)
 * Used in both create and edit sales order forms
 *
 * PHASE 1 IMPROVEMENTS:
 * - Keyboard shortcuts (Enter to add, Escape to clear, Ctrl+K focus)
 * - Auto-focus for faster entry
 * - Product code auto-complete
 * - Stock indicator with color coding
 *
 * PHASE 2 IMPROVEMENTS:
 * - Inline editing pada items table
 * - Duplicate item button functionality
 * - Quick add panel for frequent items (HYBRID APPROACH - OPSI 3)
 *   * If customer has purchase history: show top 8 most frequently purchased products
 *   * If new customer: fallback to first 8 available products
 *   * Smart sorting: frequency > total quantity > recent purchase date
 * - Visual feedback for inline edits
 *
 * PHASE 3 IMPROVEMENTS (Validation & Business Rules):
 * - Real-time stock validation:
 *   * Block adding if qty <= 0
 *   * Block adding if stock is 0
 *   * Block adding if requested > available
 *   * Warning if requesting > 80% of stock
 * - Duplicate item detection & merge:
 *   * Detect duplicate product+unit combinations
 *   * Dialog asking user to merge or add as new line
 *   * Quick add auto-merges duplicates
 * - Price validation:
 *   * Warning if selling price < cost price
 *   * Show calculated loss amount
 * - Visual feedback:
 *   * Alert components for warnings and errors
 *   * Color-coded alerts (red for errors, orange for warnings)
 *   * Clear validation messages
 */

"use client";

import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import { Plus, Trash2, AlertCircle, Zap, Package2, Copy, Save, X, Edit, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Combobox } from "@/components/ui/combobox";
import { Alert, AlertDescription } from "@/components/ui/alert";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useListProductsQuery } from "@/store/services/productApi";
import { useGetFrequentProductsQuery } from "@/store/services/customerApi";
import { formatCurrency } from "@/lib/utils";
import type { CreateSalesOrderItemRequest } from "@/types/sales-order.types";

interface SalesOrderItemsManagerProps {
  items: CreateSalesOrderItemRequest[];
  onChange: (items: CreateSalesOrderItemRequest[]) => void;
  warehouseId?: string;
  customerId?: string; // NEW: For fetching customer purchase history
}

interface Product {
  id: string;
  code: string;
  name: string;
  units: Array<{
    id: string;
    unitName: string;
    isBaseUnit: boolean;
    conversionRate: string;
    sellPrice: string;
  }>;
}

export function SalesOrderItemsManager({
  items,
  onChange,
  warehouseId,
  customerId,
}: SalesOrderItemsManagerProps) {
  const { data: productsData, isLoading: isLoadingProducts } =
    useListProductsQuery({
      page: 1,
      pageSize: 100,
      isActive: true,
    });

  // Fetch frequent products from backend (Optimized Opsi 3 - Backend Implementation)
  // Only fetch if both customerId and warehouseId are provided
  const { data: frequentProductsData } = useGetFrequentProductsQuery(
    customerId && warehouseId
      ? {
          customerId,
          warehouseId,
          limit: 8,
        }
      : { customerId: customerId || "", warehouseId: warehouseId },
    {
      skip: !customerId || !warehouseId, // Skip if either is missing
    }
  );

  // Filter produk berdasarkan warehouse yang dipilih
  const filteredProducts = useMemo(() => {
    const filtered = productsData?.data?.filter((product) => {
      // Jika warehouse belum dipilih, tampilkan semua produk
      if (!warehouseId) return true;

      // Jika produk tidak punya data warehouse stock, tidak tampilkan
      if (!product.currentStock?.warehouses) {
        return false;
      }

      // Cek apakah produk ada stok di warehouse yang dipilih
      const hasStockInWarehouse = product.currentStock.warehouses.some(
        (wh: any) => wh.warehouseId === warehouseId && parseFloat(wh.quantity || "0") > 0
      );

      return hasStockInWarehouse;
    });

    return filtered;
  }, [productsData?.data, warehouseId]);

  // Helper function to get stock info for a product
  const getProductStock = useCallback((productId: string): { quantity: number; unit: string } => {
    if (!warehouseId) return { quantity: 0, unit: '' };

    const product = filteredProducts?.find(p => p.id === productId);
    if (!product?.currentStock?.warehouses) return { quantity: 0, unit: '' };

    const warehouseStock = product.currentStock.warehouses.find(
      (wh: any) => wh.warehouseId === warehouseId
    );

    const baseUnit = product.units?.find((u: any) => u.isBaseUnit);
    const quantity = parseFloat(warehouseStock?.quantity || "0");
    const unit = baseUnit?.unitName || '';

    return { quantity, unit };
  }, [filteredProducts, warehouseId]);

  // Helper function to get stock color indicator
  const getStockColor = useCallback((quantity: number): string => {
    if (quantity === 0) return 'text-red-600 dark:text-red-400';
    if (quantity < 10) return 'text-orange-600 dark:text-orange-400';
    if (quantity < 50) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-green-600 dark:text-green-400';
  }, []);

  // Convert filtered products to Combobox options with stock info
  const productOptions = useMemo(() => {
    if (!filteredProducts || filteredProducts.length === 0) return [];

    return filteredProducts.map((product) => {
      const stock = getProductStock(product.id);
      const stockColor = getStockColor(stock.quantity);
      const stockIndicator = stock.quantity === 0
        ? '🔴'
        : stock.quantity < 10
        ? '🟠'
        : stock.quantity < 50
        ? '🟡'
        : '🟢';

      return {
        value: product.id,
        label: `${stockIndicator} ${product.code} - ${product.name} • Stock: ${stock.quantity.toLocaleString('id-ID')} ${stock.unit}`,
        searchLabel: `${product.code} ${product.name}`,
        code: product.code, // For quick code matching
        stockColor,
      };
    });
  }, [filteredProducts, warehouseId, getProductStock, getStockColor]);

  const [newItem, setNewItem] = useState<{
    productId: string;
    unitId: string;
    orderedQty: string;
    unitPrice: string;
    discount: string;
  }>({
    productId: "",
    unitId: "",
    orderedQty: "1",
    unitPrice: "0",
    discount: "0",
  });

  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);

  // Inline editing state
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editingData, setEditingData] = useState<{
    orderedQty: string;
    unitPrice: string;
    discount: string;
  } | null>(null);

  // PHASE 3: Validation state
  const [validationWarnings, setValidationWarnings] = useState<{
    stock?: string;
    duplicate?: string;
    price?: string;
    quantity?: string;
  }>({});
  const [showDuplicateDialog, setShowDuplicateDialog] = useState(false);
  const [duplicateAction, setDuplicateAction] = useState<'merge' | 'add' | null>(null);

  // Refs for keyboard navigation and auto-focus
  const productComboboxRef = useRef<HTMLButtonElement>(null);
  const qtyInputRef = useRef<HTMLInputElement>(null);

  // Convert selected product units to Combobox options
  const unitOptions = useMemo(() => {
    if (!selectedProduct?.units || selectedProduct.units.length === 0) {
      return [];
    }

    return selectedProduct.units.map((unit) => ({
      value: unit.id,
      label: `${unit.unitName}${unit.isBaseUnit ? " (Dasar)" : ""}`,
      searchLabel: unit.unitName,
    }));
  }, [selectedProduct]);

  // Get frequent/top products for quick add panel (top 8)
  // OPTIMIZED BACKEND APPROACH (Opsi 3):
  // - Backend calculates frequent products based on customer purchase history
  // - Falls back to first 8 products if customer has no history
  // - Pre-filtered by warehouse availability
  const frequentProducts = useMemo(() => {
    if (!filteredProducts || filteredProducts.length === 0) return [];

    // Use backend-calculated frequent products if available
    if (frequentProductsData && frequentProductsData.frequentProducts.length > 0) {
      // Map backend frequent product IDs to actual product objects
      const frequentProductIds = frequentProductsData.frequentProducts.map(
        (fp) => fp.productId
      );

      // Filter and sort products to match backend order
      return filteredProducts
        .filter((p) => frequentProductIds.includes(p.id))
        .sort((a, b) => {
          // Maintain the frequency order from backend
          return frequentProductIds.indexOf(a.id) - frequentProductIds.indexOf(b.id);
        });
    }

    // FALLBACK: Show first 8 available products if backend data not loaded yet
    // or customer has no purchase history
    return filteredProducts.slice(0, 8);
  }, [filteredProducts, frequentProductsData]);

  // ============================================================================
  // HELPER FUNCTIONS (must be defined before handlers that use them)
  // ============================================================================

  const calculateLineTotal = useCallback((
    qty: string,
    price: string,
    discount: string
  ): string => {
    const qtyNum = parseFloat(qty) || 0;
    const priceNum = parseFloat(price) || 0;
    const discountNum = parseFloat(discount) || 0;
    const subtotal = qtyNum * priceNum;
    const total = subtotal - discountNum;
    return total.toFixed(2);
  }, []);

  // PHASE 3: Validation helpers

  // Check if product already exists in items
  const checkDuplicateProduct = useCallback((productId: string, unitId: string): number => {
    return items.findIndex(
      (item) => item.productId === productId && item.unitId === unitId
    );
  }, [items]);

  // Validate stock availability
  const validateStock = useCallback((productId: string, requestedQty: string): {
    isValid: boolean;
    warning?: string;
    availableQty: number;
  } => {
    const stock = getProductStock(productId);
    const requested = parseFloat(requestedQty) || 0;

    if (requested <= 0) {
      return {
        isValid: false,
        warning: "Quantity harus lebih dari 0",
        availableQty: stock.quantity
      };
    }

    if (stock.quantity === 0) {
      return {
        isValid: false,
        warning: "Stok habis! Tidak dapat menambahkan item.",
        availableQty: 0
      };
    }

    if (requested > stock.quantity) {
      return {
        isValid: false,
        warning: `Stok tidak cukup! Tersedia: ${stock.quantity} ${stock.unit}`,
        availableQty: stock.quantity
      };
    }

    // Warning if requesting more than 80% of stock
    if (requested > stock.quantity * 0.8) {
      return {
        isValid: true,
        warning: `⚠️ Mengambil ${((requested / stock.quantity) * 100).toFixed(0)}% dari total stok`,
        availableQty: stock.quantity
      };
    }

    return { isValid: true, availableQty: stock.quantity };
  }, [getProductStock]);

  // Validate price against cost
  const validatePrice = useCallback((price: string, productId: string): {
    isValid: boolean;
    warning?: string;
  } => {
    const product = filteredProducts?.find(p => p.id === productId);
    if (!product) return { isValid: true };

    const priceNum = parseFloat(price) || 0;
    const costNum = parseFloat(product.baseCost || "0");

    if (priceNum < costNum) {
      const loss = costNum - priceNum;
      return {
        isValid: true, // Allow but warn
        warning: `⚠️ Harga di bawah cost! Rugi: ${formatCurrency(loss.toString())}`
      };
    }

    return { isValid: true };
  }, [filteredProducts]);

  // ============================================================================
  // HANDLER FUNCTIONS (must be defined before useEffect that use them)
  // ============================================================================

  const handleAddItem = useCallback(() => {
    if (!newItem.productId || !newItem.unitId) {
      return;
    }

    // Clear previous warnings
    setValidationWarnings({});

    // PHASE 3 VALIDATION: Stock validation
    const stockValidation = validateStock(newItem.productId, newItem.orderedQty);
    if (!stockValidation.isValid) {
      setValidationWarnings({ stock: stockValidation.warning });
      return; // Block adding if stock invalid
    }

    // PHASE 3 VALIDATION: Check for duplicates
    const duplicateIndex = checkDuplicateProduct(newItem.productId, newItem.unitId);
    if (duplicateIndex !== -1) {
      // Show duplicate dialog
      setShowDuplicateDialog(true);
      return; // Wait for user action
    }

    // PHASE 3 VALIDATION: Price validation (warning only, not blocking)
    const priceValidation = validatePrice(newItem.unitPrice, newItem.productId);
    if (priceValidation.warning) {
      setValidationWarnings({ price: priceValidation.warning });
      // Continue with adding, just show warning
    }

    // Stock warning (high usage warning, not blocking)
    if (stockValidation.warning) {
      setValidationWarnings((prev) => ({ ...prev, stock: stockValidation.warning }));
    }

    const lineTotal = calculateLineTotal(
      newItem.orderedQty,
      newItem.unitPrice,
      newItem.discount
    );

    const item: CreateSalesOrderItemRequest = {
      productId: newItem.productId,
      unitId: newItem.unitId,
      orderedQty: newItem.orderedQty,
      unitPrice: newItem.unitPrice,
      discount: newItem.discount,
      lineTotal: lineTotal,
    };

    onChange([...items, item]);

    // Reset form
    setNewItem({
      productId: "",
      unitId: "",
      orderedQty: "1",
      unitPrice: "0",
      discount: "0",
    });
    setSelectedProduct(null);

    // Auto-focus back to product combobox for rapid entry
    setTimeout(() => {
      productComboboxRef.current?.focus();
    }, 100);
  }, [newItem, items, onChange, calculateLineTotal, validateStock, checkDuplicateProduct, validatePrice]);

  // PHASE 3: Handle duplicate merge action
  const handleDuplicateMerge = useCallback(() => {
    const duplicateIndex = checkDuplicateProduct(newItem.productId, newItem.unitId);
    if (duplicateIndex === -1) return;

    const existingItem = items[duplicateIndex];
    const newQty = (parseFloat(existingItem.orderedQty) + parseFloat(newItem.orderedQty)).toString();

    // Recalculate line total with merged quantity
    const lineTotal = calculateLineTotal(
      newQty,
      existingItem.unitPrice,
      existingItem.discount || "0"
    );

    const updatedItems = [...items];
    updatedItems[duplicateIndex] = {
      ...existingItem,
      orderedQty: newQty,
      lineTotal,
    };

    onChange(updatedItems);

    // Reset form and close dialog
    setNewItem({
      productId: "",
      unitId: "",
      orderedQty: "1",
      unitPrice: "0",
      discount: "0",
    });
    setSelectedProduct(null);
    setShowDuplicateDialog(false);
    setValidationWarnings({});

    setTimeout(() => {
      productComboboxRef.current?.focus();
    }, 100);
  }, [newItem, items, onChange, checkDuplicateProduct, calculateLineTotal]);

  // PHASE 3: Handle duplicate add as new action
  const handleDuplicateAddNew = useCallback(() => {
    const lineTotal = calculateLineTotal(
      newItem.orderedQty,
      newItem.unitPrice,
      newItem.discount
    );

    const item: CreateSalesOrderItemRequest = {
      productId: newItem.productId,
      unitId: newItem.unitId,
      orderedQty: newItem.orderedQty,
      unitPrice: newItem.unitPrice,
      discount: newItem.discount,
      lineTotal: lineTotal,
    };

    onChange([...items, item]);

    // Reset form and close dialog
    setNewItem({
      productId: "",
      unitId: "",
      orderedQty: "1",
      unitPrice: "0",
      discount: "0",
    });
    setSelectedProduct(null);
    setShowDuplicateDialog(false);
    setValidationWarnings({});

    setTimeout(() => {
      productComboboxRef.current?.focus();
    }, 100);
  }, [newItem, items, onChange, calculateLineTotal]);

  // Reset form handler
  const handleResetForm = useCallback(() => {
    setNewItem({
      productId: "",
      unitId: "",
      orderedQty: "1",
      unitPrice: "0",
      discount: "0",
    });
    setSelectedProduct(null);
    setValidationWarnings({}); // Clear warnings
    productComboboxRef.current?.focus();
  }, []);

  // Product code auto-complete: Find product by code and auto-select
  const handleProductCodeSearch = useCallback((searchValue: string) => {
    if (!searchValue || !productOptions.length) return;

    // Check if searchValue matches any product code exactly
    const matchedOption = productOptions.find(
      opt => opt.code?.toUpperCase() === searchValue.toUpperCase()
    );

    if (matchedOption) {
      // Auto-select the matched product
      setNewItem(prev => ({ ...prev, productId: matchedOption.value }));
      // Focus to qty input for rapid entry
      setTimeout(() => {
        qtyInputRef.current?.focus();
        qtyInputRef.current?.select();
      }, 100);
    }
  }, [productOptions]);

  const handleRemoveItem = useCallback((index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    onChange(newItems);
  }, [items, onChange]);

  // Start inline editing
  const handleStartEdit = useCallback((index: number) => {
    const item = items[index];
    setEditingIndex(index);
    setEditingData({
      orderedQty: item.orderedQty,
      unitPrice: item.unitPrice,
      discount: item.discount || "0",
    });
  }, [items]);

  // Save inline edit
  const handleSaveEdit = useCallback(() => {
    if (editingIndex === null || !editingData) return;

    const updatedItems = [...items];
    const item = updatedItems[editingIndex];

    // Recalculate line total
    const lineTotal = calculateLineTotal(
      editingData.orderedQty,
      editingData.unitPrice,
      editingData.discount
    );

    updatedItems[editingIndex] = {
      ...item,
      orderedQty: editingData.orderedQty,
      unitPrice: editingData.unitPrice,
      discount: editingData.discount,
      lineTotal,
    };

    onChange(updatedItems);
    setEditingIndex(null);
    setEditingData(null);
  }, [editingIndex, editingData, items, onChange, calculateLineTotal]);

  // Cancel inline edit
  const handleCancelEdit = useCallback(() => {
    setEditingIndex(null);
    setEditingData(null);
  }, []);

  // Duplicate item
  const handleDuplicateItem = useCallback((index: number) => {
    const itemToDuplicate = items[index];
    const duplicatedItem: CreateSalesOrderItemRequest = {
      ...itemToDuplicate,
    };
    onChange([...items, duplicatedItem]);
  }, [items, onChange]);

  // Quick add from frequent items (with validation)
  const handleQuickAdd = useCallback((product: Product) => {
    const baseUnit = product.units?.find((u: any) => u.isBaseUnit);
    if (!baseUnit) return;

    // Clear previous warnings
    setValidationWarnings({});

    // PHASE 3 VALIDATION: Stock validation for quick add
    const stockValidation = validateStock(product.id, "1");
    if (!stockValidation.isValid) {
      setValidationWarnings({ stock: stockValidation.warning });
      return; // Block adding if stock invalid
    }

    // PHASE 3 VALIDATION: Check for duplicates
    const duplicateIndex = checkDuplicateProduct(product.id, baseUnit.id);
    if (duplicateIndex !== -1) {
      // For quick add, auto-merge quantity instead of showing dialog
      const existingItem = items[duplicateIndex];
      const newQty = (parseFloat(existingItem.orderedQty) + 1).toString();

      const lineTotal = calculateLineTotal(
        newQty,
        existingItem.unitPrice,
        existingItem.discount || "0"
      );

      const updatedItems = [...items];
      updatedItems[duplicateIndex] = {
        ...existingItem,
        orderedQty: newQty,
        lineTotal,
      };

      onChange(updatedItems);
      return;
    }

    // Stock warning (high usage warning, not blocking)
    if (stockValidation.warning) {
      setValidationWarnings({ stock: stockValidation.warning });
    }

    const lineTotal = calculateLineTotal(
      "1",
      baseUnit.sellPrice || "0",
      "0"
    );

    const item: CreateSalesOrderItemRequest = {
      productId: product.id,
      unitId: baseUnit.id,
      orderedQty: "1",
      unitPrice: baseUnit.sellPrice || "0",
      discount: "0",
      lineTotal,
    };

    onChange([...items, item]);
  }, [items, onChange, calculateLineTotal, validateStock, checkDuplicateProduct]);

  // ============================================================================
  // EFFECTS (now handlers are defined above)
  // ============================================================================

  // Auto-focus on mount for immediate keyboard entry
  useEffect(() => {
    const timer = setTimeout(() => {
      productComboboxRef.current?.focus();
    }, 300);
    return () => clearTimeout(timer);
  }, []);

  // Keyboard shortcuts handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl+K: Focus to product search and auto-open dropdown
      if (e.ctrlKey && e.key === 'k') {
        e.preventDefault();
        productComboboxRef.current?.focus();
        // Auto-click to open dropdown for immediate typing
        setTimeout(() => {
          productComboboxRef.current?.click();
        }, 100);
        return;
      }

      // Escape: Clear form
      if (e.key === 'Escape') {
        e.preventDefault();
        handleResetForm();
        return;
      }

      // Enter: Add item (only if not typing in input fields)
      if (e.key === 'Enter') {
        const target = e.target as HTMLElement;
        // Don't trigger if user is typing in combobox search or other inputs
        if (target.tagName !== 'INPUT' || target.getAttribute('type') === 'number') {
          const canAdd = newItem.productId && newItem.unitId;
          if (canAdd) {
            e.preventDefault();
            handleAddItem();
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [newItem, handleAddItem, handleResetForm]);

  // Reset form when warehouse changes
  useEffect(() => {
    setNewItem({
      productId: "",
      unitId: "",
      orderedQty: "1",
      unitPrice: "0",
      discount: "0",
    });
    setSelectedProduct(null);
  }, [warehouseId]);

  // Reset unit when product changes
  useEffect(() => {
    if (newItem.productId && filteredProducts) {
      const product = filteredProducts.find((p) => p.id === newItem.productId);
      if (product) {
        setSelectedProduct(product as any);
        // Auto-select base unit and its price
        const baseUnit = product.units?.find((u: any) => u.isBaseUnit);
        if (baseUnit) {
          setNewItem((prev) => ({
            ...prev,
            unitId: baseUnit.id,
            unitPrice: baseUnit.sellPrice || "0",
          }));
          // If product was selected, focus to qty for rapid entry
          setTimeout(() => {
            qtyInputRef.current?.focus();
            qtyInputRef.current?.select();
          }, 150);
        }
      }
    } else {
      setSelectedProduct(null);
      setNewItem((prev) => ({
        ...prev,
        unitId: "",
        unitPrice: "0",
      }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [newItem.productId]);

  // Update price when unit changes
  useEffect(() => {
    if (newItem.unitId && selectedProduct) {
      const unit = selectedProduct.units.find((u) => u.id === newItem.unitId);
      if (unit) {
        const newPrice = unit.sellPrice || "0";
        // Only update if price is different
        if (newItem.unitPrice !== newPrice) {
          setNewItem((prev) => ({
            ...prev,
            unitPrice: newPrice,
          }));
        }
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [newItem.unitId, selectedProduct]);

  // ============================================================================
  // DISPLAY HELPERS
  // ============================================================================

  const getProductName = (productId: string): string => {
    // Cari di semua produk (tidak hanya filtered) untuk display item yang sudah ditambahkan
    const product = productsData?.data.find((p) => p.id === productId);
    return product ? `${product.code} - ${product.name}` : productId;
  };

  const getUnitName = (productId: string, unitId: string): string => {
    // Cari di semua produk (tidak hanya filtered) untuk display item yang sudah ditambahkan
    const product = productsData?.data.find((p) => p.id === productId);
    if (!product) return unitId;
    const unit = product.units?.find((u: any) => u.id === unitId);
    return unit ? unit.unitName : unitId;
  };

  return (
    <div className="space-y-4">
      {/* Keyboard Shortcuts Hint */}
      <div className="flex items-center gap-3 text-xs text-muted-foreground bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 rounded-lg px-3 py-2">
        <Zap className="h-4 w-4 text-blue-600 dark:text-blue-400" />
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          <span><kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">Ctrl+K</kbd> Focus Produk</span>
          <span><kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">Enter</kbd> Tambah Item</span>
          <span><kbd className="px-1.5 py-0.5 bg-white dark:bg-gray-800 border rounded text-[10px]">Esc</kbd> Reset</span>
          <span className="text-blue-600 dark:text-blue-400">💡 Type kode produk untuk quick select</span>
        </div>
      </div>

      {/* Quick Add Panel - Frequent Items */}
      {frequentProducts.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <Zap className="h-4 w-4" />
              <span>Quick Add (1-Click)</span>
            </div>
            {frequentProductsData && frequentProductsData.totalOrders > 0 && (
              <span className="text-xs text-muted-foreground italic">
                Produk yang sering dibeli customer ini ({frequentProductsData.totalOrders} transaksi)
              </span>
            )}
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-2">
            {frequentProducts.map((product) => {
              const stock = getProductStock(product.id);
              const baseUnit = product.units?.find((u: any) => u.isBaseUnit);
              const stockColor = getStockColor(stock.quantity);

              return (
                <Button
                  key={product.id}
                  type="button"
                  variant="outline"
                  onClick={() => handleQuickAdd(product as any)}
                  disabled={stock.quantity === 0}
                  className="h-auto py-3 px-2 flex flex-col items-start gap-1 hover:bg-accent hover:border-primary transition-all"
                  title={`Quick add: ${product.code} - ${product.name}`}
                >
                  <span className="font-bold text-xs text-primary">{product.code}</span>
                  <span className="text-[10px] line-clamp-2 text-left leading-tight">{product.name}</span>
                  <div className="flex items-center justify-between w-full mt-1">
                    <span className="text-xs font-semibold">
                      {formatCurrency(baseUnit?.sellPrice || "0")}
                    </span>
                    <span className={`text-[10px] font-medium ${stockColor}`}>
                      {stock.quantity > 0 ? `${stock.quantity.toFixed(0)}` : 'Habis'}
                    </span>
                  </div>
                </Button>
              );
            })}
          </div>
        </div>
      )}

      {/* Add Item Form */}
      <div className="grid grid-cols-1 md:grid-cols-6 gap-3 p-4 border rounded-lg bg-muted/50">
        {/* Product Selection */}
        <div className="md:col-span-2 space-y-1">
          <Label className="text-xs flex items-center gap-1">
            <Package2 className="h-3 w-3" />
            Produk
            <span className="text-[10px] text-muted-foreground ml-1">(Type kode atau nama)</span>
          </Label>
          <Combobox
            ref={productComboboxRef}
            value={newItem.productId}
            onValueChange={(value) =>
              setNewItem((prev) => ({ ...prev, productId: value }))
            }
            onSearchChange={handleProductCodeSearch}
            options={productOptions}
            placeholder={
              isLoadingProducts
                ? "Memuat..."
                : !warehouseId
                ? "Pilih gudang terlebih dahulu"
                : productOptions.length === 0
                ? "Tidak ada produk tersedia"
                : "Type kode atau cari produk..."
            }
            searchPlaceholder="Type kode produk atau cari..."
            emptyMessage="Produk tidak ditemukan"
            disabled={isLoadingProducts || !warehouseId || productOptions.length === 0}
            className="w-full"
          />
        </div>

        {/* Unit Selection */}
        <div className="space-y-1">
          <Label className="text-xs">Satuan</Label>
          <Combobox
            value={newItem.unitId}
            onValueChange={(value) =>
              setNewItem((prev) => ({ ...prev, unitId: value }))
            }
            options={unitOptions}
            placeholder={
              !selectedProduct
                ? "Pilih produk dulu"
                : unitOptions.length === 0
                ? "Tidak ada satuan"
                : "Pilih satuan"
            }
            searchPlaceholder="Cari satuan..."
            emptyMessage="Satuan tidak ditemukan"
            disabled={!selectedProduct || unitOptions.length === 0}
            className="w-full"
          />
        </div>

        {/* Quantity */}
        <div className="space-y-1">
          <Label className="text-xs">Qty</Label>
          <Input
            ref={qtyInputRef}
            type="number"
            step="0.01"
            min="0.01"
            value={newItem.orderedQty}
            onChange={(e) =>
              setNewItem((prev) => ({ ...prev, orderedQty: e.target.value }))
            }
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newItem.productId && newItem.unitId) {
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
          <Label className="text-xs">Harga</Label>
          <Input
            type="number"
            step="0.01"
            min="0"
            value={newItem.unitPrice}
            onChange={(e) =>
              setNewItem((prev) => ({ ...prev, unitPrice: e.target.value }))
            }
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newItem.productId && newItem.unitId) {
                e.preventDefault();
                handleAddItem();
              }
            }}
            placeholder="0"
            disabled={!selectedProduct}
          />
        </div>

        {/* Discount */}
        <div className="space-y-1">
          <Label className="text-xs">Diskon</Label>
          <Input
            type="number"
            step="0.01"
            min="0"
            value={newItem.discount}
            onChange={(e) =>
              setNewItem((prev) => ({ ...prev, discount: e.target.value }))
            }
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newItem.productId && newItem.unitId) {
                e.preventDefault();
                handleAddItem();
              }
            }}
            placeholder="0"
            disabled={!selectedProduct}
          />
        </div>

        {/* Add Button */}
        <div className="flex items-end">
          <Button
            type="button"
            onClick={handleAddItem}
            disabled={!newItem.productId || !newItem.unitId}
            className="w-full"
          >
            <Plus className="h-4 w-4 mr-1" />
            Tambah
          </Button>
        </div>
      </div>

      {/* PHASE 3: Validation Warnings Display */}
      {(validationWarnings.stock || validationWarnings.price || validationWarnings.quantity) && (
        <div className="space-y-2">
          {validationWarnings.stock && (
            <Alert variant={validationWarnings.stock.includes('⚠️') ? 'default' : 'destructive'}>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>{validationWarnings.stock}</AlertDescription>
            </Alert>
          )}
          {validationWarnings.price && (
            <Alert variant="default" className="border-orange-300 dark:border-orange-800">
              <AlertTriangle className="h-4 w-4 text-orange-600 dark:text-orange-400" />
              <AlertDescription className="text-orange-900 dark:text-orange-200">
                {validationWarnings.price}
              </AlertDescription>
            </Alert>
          )}
          {validationWarnings.quantity && (
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>{validationWarnings.quantity}</AlertDescription>
            </Alert>
          )}
        </div>
      )}

      {/* PHASE 3: Duplicate Item Detection Dialog */}
      <AlertDialog open={showDuplicateDialog} onOpenChange={setShowDuplicateDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Produk Sudah Ada</AlertDialogTitle>
            <AlertDialogDescription>
              Produk dengan satuan yang sama sudah ada dalam daftar item.
              Apakah Anda ingin menggabungkan quantity atau menambahkan sebagai item baru?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setShowDuplicateDialog(false)}>
              Batal
            </AlertDialogCancel>
            <Button
              variant="outline"
              onClick={handleDuplicateAddNew}
              className="mr-2"
            >
              Tambah Baris Baru
            </Button>
            <AlertDialogAction onClick={handleDuplicateMerge}>
              Gabung Quantity
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Items Table */}
      {items.length > 0 ? (
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[40%]">Produk</TableHead>
                <TableHead>Satuan</TableHead>
                <TableHead className="text-right">Qty</TableHead>
                <TableHead className="text-right">Harga</TableHead>
                <TableHead className="text-right">Diskon</TableHead>
                <TableHead className="text-right">Total</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item, index) => {
                const isEditing = editingIndex === index;

                return (
                  <TableRow key={index} className={isEditing ? 'bg-accent' : ''}>
                    <TableCell className="font-medium">
                      {getProductName(item.productId)}
                    </TableCell>
                    <TableCell>
                      {getUnitName(item.productId, item.unitId)}
                    </TableCell>

                    {/* Editable Qty */}
                    <TableCell className="text-right">
                      {isEditing ? (
                        <Input
                          type="number"
                          step="0.01"
                          value={editingData?.orderedQty || ''}
                          onChange={(e) => setEditingData(prev => prev ? {...prev, orderedQty: e.target.value} : null)}
                          className="w-20 text-right"
                          autoFocus
                        />
                      ) : (
                        parseFloat(item.orderedQty).toLocaleString("id-ID")
                      )}
                    </TableCell>

                    {/* Editable Price */}
                    <TableCell className="text-right">
                      {isEditing ? (
                        <Input
                          type="number"
                          step="0.01"
                          value={editingData?.unitPrice || ''}
                          onChange={(e) => setEditingData(prev => prev ? {...prev, unitPrice: e.target.value} : null)}
                          className="w-24 text-right"
                        />
                      ) : (
                        formatCurrency(item.unitPrice)
                      )}
                    </TableCell>

                    {/* Editable Discount */}
                    <TableCell className="text-right">
                      {isEditing ? (
                        <Input
                          type="number"
                          step="0.01"
                          value={editingData?.discount || ''}
                          onChange={(e) => setEditingData(prev => prev ? {...prev, discount: e.target.value} : null)}
                          className="w-24 text-right"
                        />
                      ) : (
                        formatCurrency(item.discount || "0")
                      )}
                    </TableCell>

                    {/* Total (calculated) */}
                    <TableCell className="text-right font-medium">
                      {isEditing && editingData
                        ? formatCurrency(calculateLineTotal(editingData.orderedQty, editingData.unitPrice, editingData.discount))
                        : formatCurrency(item.lineTotal)
                      }
                    </TableCell>

                    {/* Action buttons */}
                    <TableCell>
                      <div className="flex gap-1">
                        {isEditing ? (
                          <>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={handleSaveEdit}
                              className="h-8 w-8 text-green-600 hover:text-green-700 hover:bg-green-50 dark:hover:bg-green-950"
                              title="Save"
                            >
                              <Save className="h-4 w-4" />
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={handleCancelEdit}
                              className="h-8 w-8 text-gray-600 hover:text-gray-700 hover:bg-gray-50 dark:hover:bg-gray-900"
                              title="Cancel"
                            >
                              <X className="h-4 w-4" />
                            </Button>
                          </>
                        ) : (
                          <>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDuplicateItem(index)}
                              className="h-8 w-8 text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950"
                              title="Duplicate"
                            >
                              <Copy className="h-4 w-4" />
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => handleStartEdit(index)}
                              className="h-8 w-8 text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:hover:bg-orange-950"
                              title="Edit"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => handleRemoveItem(index)}
                              className="h-8 w-8 text-destructive hover:text-destructive hover:bg-red-50 dark:hover:bg-red-950"
                              title="Delete"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-8 text-muted-foreground border-2 border-dashed rounded-lg">
          <AlertCircle className="h-8 w-8 mb-2" />
          <p className="text-sm">Belum ada item ditambahkan</p>
          <p className="text-xs mt-1">
            Gunakan form di atas untuk menambahkan produk
          </p>
        </div>
      )}
    </div>
  );
}
