/**
 * Sales Order Types
 *
 * TypeScript type definitions for sales order management.
 * Based on ERP distribution business logic and backend API contracts.
 */

/**
 * Sales Order Status Enum
 * Represents the lifecycle of a sales order
 */
export type SalesOrderStatus =
  | "DRAFT" // Being created, not submitted
  | "PENDING" // Submitted, waiting approval
  | "APPROVED" // Approved, ready for fulfillment
  | "PROCESSING" // Being prepared/picked
  | "SHIPPED" // Sent to customer
  | "DELIVERED" // Received by customer
  | "CANCELLED" // Cancelled at any stage
  | "COMPLETED"; // Fully delivered and invoiced

/**
 * Sales Order
 * Main sales order entity with multi-item support
 */
export interface SalesOrder {
  id: string;
  companyId: string;
  tenantId: string;

  // Order Information
  orderNumber: string;
  orderDate: string; // ISO 8601
  requiredDate?: string | null; // ISO 8601
  deliveryDate?: string | null; // ISO 8601

  // Customer
  customerId: string;
  customerCode?: string;
  customerName?: string;

  // Warehouse (fulfillment location)
  warehouseId: string;
  warehouseCode?: string;
  warehouseName?: string;

  // Financial (decimals as strings)
  subtotal: string;
  discount: string;
  tax: string; // PPN (typically 11% in Indonesia)
  shippingCost: string;
  totalAmount: string;

  // Status & Metadata
  status: SalesOrderStatus;
  notes?: string | null;
  paymentTerms?: string | null;

  // Timestamps
  createdAt: string;
  updatedAt: string;
  approvedAt?: string | null;
  shippedAt?: string | null;
  deliveredAt?: string | null;

  // Relations
  items?: SalesOrderItem[];
}

/**
 * Sales Order Item
 * Individual line item in a sales order
 */
export interface SalesOrderItem {
  id: string;
  salesOrderId: string;
  lineNumber: number;

  // Product
  productId: string;
  productCode?: string;
  productName?: string;

  // Unit
  unitId: string;
  unitName: string;
  conversionRate: string; // decimal as string

  // Quantity & Pricing (decimals as strings)
  orderedQty: string;
  shippedQty?: string;
  unitPrice: string;
  discount: string;
  tax: string;
  lineTotal: string;

  // Metadata
  notes?: string | null;

  // Timestamps
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Create Sales Order Request
 */
export interface CreateSalesOrderRequest {
  customerId: string;
  warehouseId: string;
  orderDate: string;
  requiredDate?: string;
  subtotal: string;
  discount?: string;
  tax?: string;
  shippingCost?: string;
  totalAmount: string;
  notes?: string;
  paymentTerms?: string;
  items: CreateSalesOrderItemRequest[];
}

/**
 * Create Sales Order Item Request
 */
export interface CreateSalesOrderItemRequest {
  productId: string;
  unitId: string;
  orderedQty: string;
  unitPrice: string;
  discount?: string;
  tax?: string;
  lineTotal: string;
  notes?: string;
}

/**
 * Update Sales Order Request
 * Partial update - all fields optional
 * Only allowed for DRAFT status
 */
export interface UpdateSalesOrderRequest {
  customerId?: string;
  warehouseId?: string;
  orderDate?: string;
  requiredDate?: string;
  subtotal?: string;
  discount?: string;
  tax?: string;
  shippingCost?: string;
  totalAmount?: string;
  notes?: string;
  paymentTerms?: string;
  items?: UpdateSalesOrderItemRequest[];
}

/**
 * Update Sales Order Item Request
 */
export interface UpdateSalesOrderItemRequest {
  id?: string; // If provided, update existing item
  productId?: string;
  unitId?: string;
  orderedQty?: string;
  unitPrice?: string;
  discount?: string;
  tax?: string;
  lineTotal?: string;
  notes?: string;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Sales Order Response
 * Complete sales order information from API
 */
export interface SalesOrderResponse {
  id: string;
  orderNumber: string;
  orderDate: string;
  requiredDate?: string;
  deliveryDate?: string;

  // Customer details
  customerId: string;
  customerCode: string;
  customerName: string;

  // Warehouse details
  warehouseId: string;
  warehouseCode: string;
  warehouseName: string;

  // Financial
  subtotal: string;
  discount: string;
  tax: string;
  shippingCost: string;
  totalAmount: string;

  // Status & metadata
  status: SalesOrderStatus;
  notes?: string;
  paymentTerms?: string;

  // Items
  items: SalesOrderItemResponse[];

  // Timestamps
  createdAt: string;
  updatedAt: string;
  approvedAt?: string;
  shippedAt?: string;
  deliveredAt?: string;
}

/**
 * Sales Order Item Response
 */
export interface SalesOrderItemResponse {
  id: string;
  salesOrderId: string;
  lineNumber: number;

  // Product
  productId: string;
  productCode: string;
  productName: string;

  // Unit
  unitId: string;
  unitName: string;
  conversionRate: string;

  // Quantity & pricing
  orderedQty: string;
  shippedQty?: string;
  unitPrice: string;
  discount: string;
  tax: string;
  lineTotal: string;

  notes?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Sales Order List Response
 * Paginated list of sales orders
 */
export interface SalesOrderListResponse {
  success: boolean;
  data: SalesOrderResponse[];
  pagination: PaginationInfo;
}

/**
 * Pagination Information
 */
export interface PaginationInfo {
  page: number;
  pageSize: number;
  limit?: number; // Alternate field name from backend
  totalItems: number;
  total?: number; // Alternate field name from backend
  totalPages: number;
  hasMore: boolean;
}

/**
 * Sales Order Filters
 * Query parameters for listing sales orders
 */
export interface SalesOrderFilters {
  search?: string; // search in order number or customer name
  customerId?: string;
  warehouseId?: string;
  status?: SalesOrderStatus;
  dateFrom?: string; // order date range start
  dateTo?: string; // order date range end
  requiredDateFrom?: string;
  requiredDateTo?: string;
  page?: number;
  pageSize?: number;
  sortBy?:
    | "orderDate"
    | "orderNumber"
    | "customerName"
    | "totalAmount"
    | "status"
    | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// Constants & Enums
// ============================================================================

/**
 * Status Labels (Indonesian)
 */
export const SALES_ORDER_STATUS_LABELS: Record<SalesOrderStatus, string> = {
  DRAFT: "Draft",
  PENDING: "Menunggu",
  APPROVED: "Disetujui",
  PROCESSING: "Diproses",
  SHIPPED: "Dikirim",
  DELIVERED: "Diterima",
  CANCELLED: "Dibatalkan",
  COMPLETED: "Selesai",
};

/**
 * Status Badge Styles
 */
export const SALES_ORDER_STATUS_STYLES: Record<SalesOrderStatus, string> = {
  DRAFT: "bg-gray-500 text-white hover:bg-gray-600",
  PENDING: "bg-yellow-500 text-white hover:bg-yellow-600",
  APPROVED: "bg-blue-500 text-white hover:bg-blue-600",
  PROCESSING: "bg-indigo-500 text-white hover:bg-indigo-600",
  SHIPPED: "bg-purple-500 text-white hover:bg-purple-600",
  DELIVERED: "bg-green-500 text-white hover:bg-green-600",
  CANCELLED: "bg-red-500 text-white hover:bg-red-600",
  COMPLETED: "bg-emerald-500 text-white hover:bg-emerald-600",
};

/**
 * Sort Options for Sales Order List
 */
export const SALES_ORDER_SORT_OPTIONS = [
  { value: "orderDate", label: "Tanggal Pesanan" },
  { value: "orderNumber", label: "Nomor Pesanan" },
  { value: "customerName", label: "Nama Pelanggan" },
  { value: "totalAmount", label: "Total Amount" },
  { value: "status", label: "Status" },
  { value: "createdAt", label: "Tanggal Dibuat" },
] as const;

/**
 * Payment Terms Options (Indonesian)
 */
export const PAYMENT_TERMS_OPTIONS = [
  { value: "COD", label: "Cash on Delivery (COD)" },
  { value: "NET30", label: "Net 30 Hari" },
  { value: "NET45", label: "Net 45 Hari" },
  { value: "NET60", label: "Net 60 Hari" },
  { value: "PREPAID", label: "Bayar Dimuka" },
] as const;

/**
 * Type guard to check if value is a valid sales order status
 */
export function isSalesOrderStatus(
  value: string
): value is SalesOrderStatus {
  return [
    "DRAFT",
    "PENDING",
    "APPROVED",
    "PROCESSING",
    "SHIPPED",
    "DELIVERED",
    "CANCELLED",
    "COMPLETED",
  ].includes(value);
}

/**
 * Check if order can be edited
 * Only DRAFT orders can be edited
 */
export function canEditOrder(status: SalesOrderStatus): boolean {
  return status === "DRAFT";
}

/**
 * Check if order can be cancelled
 * Cannot cancel DELIVERED, COMPLETED, or already CANCELLED orders
 */
export function canCancelOrder(status: SalesOrderStatus): boolean {
  return !["DELIVERED", "COMPLETED", "CANCELLED"].includes(status);
}

/**
 * Get next possible status transitions
 */
export function getNextStatusTransitions(
  status: SalesOrderStatus
): SalesOrderStatus[] {
  const transitions: Record<SalesOrderStatus, SalesOrderStatus[]> = {
    DRAFT: ["PENDING"],
    PENDING: ["APPROVED", "DRAFT"], // Approve or reject back to draft
    APPROVED: ["PROCESSING", "CANCELLED"],
    PROCESSING: ["SHIPPED", "CANCELLED"],
    SHIPPED: ["DELIVERED", "CANCELLED"],
    DELIVERED: ["COMPLETED"],
    CANCELLED: [], // Terminal state
    COMPLETED: [], // Terminal state
  };

  return transitions[status] || [];
}
