/**
 * Stock and Inventory Types
 *
 * TypeScript type definitions for warehouse stock management.
 * Based on backend models/warehouse.go (WarehouseStock) and internal/dto/warehouse_dto.go
 */

/**
 * Warehouse Stock
 * Stock tracking per warehouse per product
 */
export interface WarehouseStock {
  id: string;
  warehouseID: string;
  productID: string;
  quantity: string; // decimal as string
  minimumStock: string; // decimal as string
  maximumStock: string; // decimal as string
  location?: string | null; // Rack or zone location (e.g., "RAK-A-01")
  lastCountDate?: string | null;
  lastCountQty?: string | null; // decimal as string
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Warehouse Stock Response
 * Complete warehouse stock information with product details
 */
export interface WarehouseStockResponse {
  id: string;
  warehouseID: string;
  warehouseName?: string; // From join
  warehouseCode?: string; // From join
  productID: string;
  productCode: string; // From join
  productName: string; // From join
  productCategory?: string; // From join
  productUnit?: string; // From join
  quantity: string;
  minimumStock: string;
  maximumStock: string;
  location?: string;
  lastCountDate?: string;
  lastCountQty?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Warehouse Stock List Response
 * Paginated list of warehouse stocks
 */
export interface WarehouseStockListResponse {
  success: boolean;
  data: WarehouseStockResponse[];
  pagination: StockPaginationInfo;
}

/**
 * Pagination Information for Stock List
 */
export interface StockPaginationInfo {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasMore?: boolean;
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Stock Filters
 * Query parameters for listing warehouse stocks
 */
export interface StockFilters {
  search?: string; // Search by product code or name
  warehouseID?: string;
  productID?: string;
  lowStock?: boolean; // Filter products below minimum stock
  zeroStock?: boolean; // Filter products with zero stock
  page?: number;
  pageSize?: number;
  sortBy?: "productCode" | "productName" | "quantity" | "createdAt";
  sortOrder?: "asc" | "desc";
}

/**
 * Update Warehouse Stock Request
 * Update stock settings (not actual quantity - that's done via inventory movements)
 */
export interface UpdateWarehouseStockRequest {
  minimumStock?: string; // decimal as string
  maximumStock?: string; // decimal as string
  location?: string; // Rack or zone location
}

// ============================================================================
// Utility Types
// ============================================================================

/**
 * Stock Status Enum
 * Calculated stock status based on quantity and thresholds
 */
export type StockStatus = "NORMAL" | "LOW" | "CRITICAL" | "OUT_OF_STOCK" | "OVERSTOCK";

/**
 * Stock Alert
 * Alert information for stock levels
 */
export interface StockAlert {
  stockId: string;
  productCode: string;
  productName: string;
  warehouseName: string;
  status: StockStatus;
  currentQuantity: string;
  minimumStock: string;
  message: string;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Calculate stock status based on quantity and thresholds
 */
export function getStockStatus(
  quantity: string,
  minimumStock: string,
  maximumStock: string
): StockStatus {
  const qty = Number(quantity);
  const min = Number(minimumStock);
  const max = Number(maximumStock);

  if (qty === 0) return "OUT_OF_STOCK";
  if (qty < min * 0.5) return "CRITICAL"; // Below 50% of minimum
  if (qty < min) return "LOW";
  if (max > 0 && qty > max) return "OVERSTOCK";
  return "NORMAL";
}

/**
 * Get stock status badge color
 */
export function getStockStatusColor(status: StockStatus): string {
  const colors: Record<StockStatus, string> = {
    NORMAL: "bg-green-500 text-white hover:bg-green-600",
    LOW: "bg-yellow-500 text-white hover:bg-yellow-600",
    CRITICAL: "bg-orange-500 text-white hover:bg-orange-600",
    OUT_OF_STOCK: "bg-red-500 text-white hover:bg-red-600",
    OVERSTOCK: "bg-purple-500 text-white hover:bg-purple-600",
  };
  return colors[status] || "bg-gray-500 text-white";
}

/**
 * Get stock status label in Indonesian
 */
export function getStockStatusLabel(status: StockStatus): string {
  const labels: Record<StockStatus, string> = {
    NORMAL: "Normal",
    LOW: "Stok Rendah",
    CRITICAL: "Kritis",
    OUT_OF_STOCK: "Habis",
    OVERSTOCK: "Stok Berlebih",
  };
  return labels[status] || status;
}

/**
 * Format stock quantity with unit
 */
export function formatStockQuantity(quantity: string, unit?: string): string {
  const num = Number(quantity);
  if (isNaN(num)) return "-";

  const formatted = num.toLocaleString("id-ID", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 3,
  });

  return unit ? `${formatted} ${unit}` : formatted;
}

/**
 * Check if stock is below minimum threshold
 */
export function isLowStock(quantity: string, minimumStock: string): boolean {
  return Number(quantity) < Number(minimumStock);
}

/**
 * Check if stock is critically low (below 50% of minimum)
 */
export function isCriticalStock(quantity: string, minimumStock: string): boolean {
  return Number(quantity) < Number(minimumStock) * 0.5;
}

/**
 * Calculate stock percentage of minimum
 */
export function getStockPercentage(quantity: string, minimumStock: string): number {
  const qty = Number(quantity);
  const min = Number(minimumStock);

  if (min === 0) return 100; // No minimum set
  return (qty / min) * 100;
}

// ============================================================================
// Constants
// ============================================================================

/**
 * Sort Options for Stock List
 */
export const STOCK_SORT_OPTIONS = [
  { value: "productCode", label: "Kode Produk" },
  { value: "productName", label: "Nama Produk" },
  { value: "quantity", label: "Jumlah Stok" },
  { value: "createdAt", label: "Tanggal Dibuat" },
] as const;

/**
 * Stock Filter Options
 */
export const STOCK_FILTER_OPTIONS = [
  { value: "all", label: "Semua Stok" },
  { value: "lowStock", label: "Stok Rendah" },
  { value: "zeroStock", label: "Stok Habis" },
] as const;
