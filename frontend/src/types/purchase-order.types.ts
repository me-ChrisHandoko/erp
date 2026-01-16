/**
 * Purchase Order Types
 *
 * TypeScript type definitions for procurement order management.
 * Based on backend models/purchase.go
 */

/**
 * Purchase Order Status Enum
 * Simplified PO workflow (PHASE 0)
 */
export type PurchaseOrderStatus = "DRAFT" | "CONFIRMED" | "COMPLETED" | "CANCELLED";

/**
 * Purchase Order
 * Main purchase order entity for procurement management
 */
export interface PurchaseOrder {
  id: string;
  companyId: string;
  tenantId: string;

  // Order Information
  poNumber: string;
  poDate: string;
  supplierId: string;
  warehouseId: string;
  status: PurchaseOrderStatus;

  // Financial Information
  subtotal: string; // decimal as string
  discountAmount: string; // decimal as string
  taxAmount: string; // decimal as string
  totalAmount: string; // decimal as string

  // Additional Information
  notes?: string | null;
  expectedDeliveryAt?: string | null;

  // Workflow Information
  requestedBy?: string | null;
  approvedBy?: string | null;
  approvedAt?: string | null;
  cancelledBy?: string | null;
  cancelledAt?: string | null;
  cancellationNote?: string | null;

  // Timestamps
  createdAt: string;
  updatedAt: string;

  // Relations
  items?: PurchaseOrderItem[];
}

/**
 * Purchase Order Item
 * Line item for purchase order
 */
export interface PurchaseOrderItem {
  id: string;
  purchaseOrderId: string;
  productId: string;
  productUnitId?: string | null;

  // Quantity and Pricing
  quantity: string; // decimal as string
  unitPrice: string; // decimal as string
  discountPct: string; // decimal as string
  discountAmt: string; // decimal as string
  subtotal: string; // decimal as string
  receivedQty: string; // decimal as string - quantity received so far

  // Additional Info
  notes?: string | null;

  // Timestamps
  createdAt: string;
  updatedAt: string;

  // Relations (populated from API)
  product?: {
    id: string;
    code: string;
    name: string;
    baseUnit: string;
  };
  productUnit?: {
    id: string;
    unitName: string;
    conversionRate: string;
  } | null;
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Create Purchase Order Request
 */
export interface CreatePurchaseOrderRequest {
  supplierId: string;
  warehouseId: string;
  poDate?: string;
  expectedDeliveryAt?: string;
  notes?: string;
  items: CreatePurchaseOrderItemRequest[];
}

/**
 * Create Purchase Order Item Request
 */
export interface CreatePurchaseOrderItemRequest {
  productId: string;
  productUnitId?: string;
  quantity: string; // decimal as string
  unitPrice: string; // decimal as string
  discountPct?: string; // decimal as string
  discountAmt?: string; // decimal as string
  notes?: string;
}

/**
 * Update Purchase Order Request
 * Partial update - all fields optional
 */
export interface UpdatePurchaseOrderRequest {
  supplierId?: string;
  warehouseId?: string;
  poDate?: string;
  expectedDeliveryAt?: string;
  notes?: string;
  items?: UpdatePurchaseOrderItemRequest[];
}

/**
 * Update Purchase Order Item Request
 */
export interface UpdatePurchaseOrderItemRequest {
  id?: string; // If provided, update existing item; if not, create new
  productId: string;
  productUnitId?: string;
  quantity: string;
  unitPrice: string;
  discountPct?: string;
  discountAmt?: string;
  notes?: string;
}

/**
 * Confirm Purchase Order Request
 */
export interface ConfirmPurchaseOrderRequest {
  notes?: string;
}

/**
 * Cancel Purchase Order Request
 */
export interface CancelPurchaseOrderRequest {
  cancellationNote: string;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Purchase Order Response
 * Complete purchase order information from API
 */
export interface PurchaseOrderResponse {
  id: string;
  poNumber: string;
  poDate: string;
  supplierId: string;
  warehouseId: string;
  status: PurchaseOrderStatus;
  subtotal: string;
  discountAmount: string;
  taxAmount: string;
  totalAmount: string;
  notes?: string;
  expectedDeliveryAt?: string;
  requestedBy?: string;
  approvedBy?: string;
  approvedAt?: string;
  cancelledBy?: string;
  cancelledAt?: string;
  cancellationNote?: string;
  createdAt: string;
  updatedAt: string;

  // Populated relations
  supplier?: {
    id: string;
    code: string;
    name: string;
  };
  warehouse?: {
    id: string;
    code: string;
    name: string;
  };
  requester?: {
    id: string;
    fullName: string;
  };
  items?: PurchaseOrderItemResponse[];
}

/**
 * Purchase Order Item Response
 */
export interface PurchaseOrderItemResponse {
  id: string;
  purchaseOrderId: string;
  productId: string;
  productUnitId?: string;
  quantity: string;
  unitPrice: string;
  discountPct: string;
  discountAmt: string;
  subtotal: string;
  receivedQty: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;

  // Populated relations
  product?: {
    id: string;
    code: string;
    name: string;
    baseUnit: string;
  };
  productUnit?: {
    id: string;
    unitName: string;
    conversionRate: string;
  };
}

/**
 * Purchase Order List Response
 * Paginated list of purchase orders
 */
export interface PurchaseOrderListResponse {
  success: boolean;
  data: PurchaseOrderResponse[];
  pagination: PaginationInfo;
}

/**
 * Pagination Information
 */
export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

/**
 * Purchase Order Filters
 * Query parameters for listing purchase orders
 */
export interface PurchaseOrderFilters {
  search?: string; // search in PO number
  supplierId?: string;
  warehouseId?: string;
  status?: PurchaseOrderStatus;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  pageSize?: number;
  sortBy?: "poNumber" | "poDate" | "totalAmount" | "status" | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// Constants & Enums
// ============================================================================

/**
 * Purchase Order Status Labels (Indonesian)
 */
export const PURCHASE_ORDER_STATUS_LABELS: Record<PurchaseOrderStatus, string> = {
  DRAFT: "Draft",
  CONFIRMED: "Dikonfirmasi",
  COMPLETED: "Selesai",
  CANCELLED: "Dibatalkan",
};

/**
 * Purchase Order Status Colors for Badges
 */
export const PURCHASE_ORDER_STATUS_COLORS: Record<PurchaseOrderStatus, string> = {
  DRAFT: "bg-gray-500 text-white hover:bg-gray-600",
  CONFIRMED: "bg-blue-500 text-white hover:bg-blue-600",
  COMPLETED: "bg-green-500 text-white hover:bg-green-600",
  CANCELLED: "bg-red-500 text-white hover:bg-red-600",
};

/**
 * Purchase Order Status Options for Dropdown
 */
export const PURCHASE_ORDER_STATUS_OPTIONS = [
  { value: "DRAFT", label: "Draft" },
  { value: "CONFIRMED", label: "Dikonfirmasi" },
  { value: "COMPLETED", label: "Selesai" },
  { value: "CANCELLED", label: "Dibatalkan" },
] as const;

/**
 * Sort Options for Purchase Order List
 */
export const PURCHASE_ORDER_SORT_OPTIONS = [
  { value: "poNumber", label: "Nomor PO" },
  { value: "poDate", label: "Tanggal PO" },
  { value: "totalAmount", label: "Total" },
  { value: "status", label: "Status" },
  { value: "createdAt", label: "Tanggal Dibuat" },
] as const;

/**
 * Get status label in Indonesian
 */
export function getStatusLabel(status: PurchaseOrderStatus): string {
  return PURCHASE_ORDER_STATUS_LABELS[status] || status;
}

/**
 * Get status badge color
 */
export function getStatusBadgeColor(status: PurchaseOrderStatus): string {
  return PURCHASE_ORDER_STATUS_COLORS[status] || "bg-gray-500 text-white";
}

/**
 * Type guard to check if value is a valid purchase order status
 */
export function isPurchaseOrderStatus(value: string): value is PurchaseOrderStatus {
  return ["DRAFT", "CONFIRMED", "COMPLETED", "CANCELLED"].includes(value);
}

/**
 * Format currency for display (Indonesian Rupiah)
 */
export function formatCurrency(value: string | number): string {
  const num = typeof value === "string" ? parseFloat(value) : value;
  if (isNaN(num)) return "Rp 0";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num);
}

/**
 * Format date for display (Indonesian format)
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(date);
}
