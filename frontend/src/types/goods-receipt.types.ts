/**
 * Goods Receipt Types
 *
 * TypeScript type definitions for goods receipt management (Penerimaan Barang).
 * Based on backend models/goods_receipt.go and internal/dto/goods_receipt_dto.go
 */

// ============================================================================
// STATUS TYPES
// ============================================================================

/**
 * Goods Receipt Status
 * PENDING → RECEIVED → INSPECTED → ACCEPTED/REJECTED/PARTIAL
 */
export type GoodsReceiptStatus =
  | "PENDING"
  | "RECEIVED"
  | "INSPECTED"
  | "ACCEPTED"
  | "REJECTED"
  | "PARTIAL";

export const GOODS_RECEIPT_STATUS_OPTIONS = [
  { value: "PENDING", label: "Menunggu", color: "bg-yellow-100 text-yellow-800" },
  { value: "RECEIVED", label: "Diterima", color: "bg-blue-100 text-blue-800" },
  { value: "INSPECTED", label: "Diperiksa", color: "bg-purple-100 text-purple-800" },
  { value: "ACCEPTED", label: "Disetujui", color: "bg-green-100 text-green-800" },
  { value: "REJECTED", label: "Ditolak", color: "bg-red-100 text-red-800" },
  { value: "PARTIAL", label: "Sebagian", color: "bg-orange-100 text-orange-800" },
] as const;

// ============================================================================
// BASIC RESPONSE TYPES
// ============================================================================

export interface GoodsReceiptProductResponse {
  id: string;
  code: string;
  name: string;
  baseUnit: string;
}

export interface GoodsReceiptProductUnitResponse {
  id: string;
  unitName: string;
  conversionRate: string;
}

export interface GoodsReceiptPurchaseOrderResponse {
  id: string;
  poNumber: string;
  poDate: string;
}

export interface WarehouseBasicResponse {
  id: string;
  code: string;
  name: string;
}

export interface SupplierBasicResponse {
  id: string;
  code: string;
  name: string;
}

export interface UserBasicResponse {
  id: string;
  fullName: string;
  email: string;
}

// ============================================================================
// GOODS RECEIPT ITEM
// ============================================================================

export interface GoodsReceiptItemResponse {
  id: string;
  goodsReceiptId: string;
  purchaseOrderItemId: string;
  productId: string;
  product?: GoodsReceiptProductResponse;
  productUnitId?: string;
  productUnit?: GoodsReceiptProductUnitResponse;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  orderedQty: string;
  receivedQty: string;
  acceptedQty: string;
  rejectedQty: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// GOODS RECEIPT
// ============================================================================

export interface GoodsReceiptResponse {
  id: string;
  grnNumber: string;
  grnDate: string;
  purchaseOrderId: string;
  purchaseOrder?: GoodsReceiptPurchaseOrderResponse;
  warehouseId: string;
  warehouse?: WarehouseBasicResponse;
  supplierId: string;
  supplier?: SupplierBasicResponse;
  status: GoodsReceiptStatus;
  receivedBy?: string;
  receiver?: UserBasicResponse;
  receivedAt?: string;
  inspectedBy?: string;
  inspector?: UserBasicResponse;
  inspectedAt?: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;
  items?: GoodsReceiptItemResponse[];
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// API REQUEST TYPES
// ============================================================================

export interface CreateGoodsReceiptItemRequest {
  purchaseOrderItemId: string;
  productId: string;
  productUnitId?: string;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  receivedQty: string;
  acceptedQty?: string;
  rejectedQty?: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
}

export interface CreateGoodsReceiptRequest {
  purchaseOrderId: string;
  grnDate: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;
  items: CreateGoodsReceiptItemRequest[];
}

export interface UpdateGoodsReceiptItemRequest {
  id?: string;
  receivedQty: string;
  acceptedQty?: string;
  rejectedQty?: string;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
}

export interface UpdateGoodsReceiptRequest {
  grnDate?: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;
  items?: UpdateGoodsReceiptItemRequest[];
}

export interface ReceiveGoodsRequest {
  notes?: string;
}

export interface InspectGoodsRequest {
  notes?: string;
  items?: UpdateGoodsReceiptItemRequest[];
}

export interface AcceptGoodsRequest {
  notes?: string;
}

export interface RejectGoodsRequest {
  rejectionReason: string;
}

// ============================================================================
// API RESPONSE TYPES
// ============================================================================

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface GoodsReceiptListResponse {
  success: boolean;
  data: GoodsReceiptResponse[];
  pagination: PaginationInfo;
}

// ============================================================================
// FILTER TYPES
// ============================================================================

export interface GoodsReceiptFilters {
  search?: string;
  status?: GoodsReceiptStatus;
  purchaseOrderId?: string;
  supplierId?: string;
  warehouseId?: string;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  pageSize?: number;
  sortBy?: "grnNumber" | "grnDate" | "status" | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

export function getGoodsReceiptStatusLabel(status: GoodsReceiptStatus): string {
  const option = GOODS_RECEIPT_STATUS_OPTIONS.find((o) => o.value === status);
  return option?.label || status;
}

export function getGoodsReceiptStatusColor(status: GoodsReceiptStatus): string {
  const option = GOODS_RECEIPT_STATUS_OPTIONS.find((o) => o.value === status);
  return option?.color || "bg-gray-100 text-gray-800";
}
